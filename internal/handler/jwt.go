package handler

import (
	"net/http"
	"github.com/golang-jwt/jwt/v4"
	"errors"
	"strings"
	"context"
	"crypto/rand"
    "encoding/hex"
	"go.uber.org/zap"
)

type contextKey string

const userUUIDKey contextKey = "userUUID"

type Claims struct {
	UserUUID string `json:"user_uuid"`
	jwt.RegisteredClaims
}

type JWTBuilder struct {
    secretKey 	   string
	issuer 		   string
	UserCookieName string
	suLog 		   zap.SugaredLogger
}

func NewJWTBuild(secret string, issuer string, userCookieName string, suLog zap.SugaredLogger) *JWTBuilder {
	return &JWTBuilder{
		secretKey: secret,
		issuer: issuer,
		UserCookieName: userCookieName,
		suLog: suLog,
	}
}

var ErrNoUserUUID = errors.New("there is no user UUID")

func (b *JWTBuilder) GenerateUserUUID() string {
	userUUID := make([]byte, 16)
	if _, err := rand.Read(userUUID); err != nil {
		b.suLog.Infow("while uuid generating", "error", err)
	}
	return hex.EncodeToString(userUUID)
}

func (b *JWTBuilder) Sign (token *jwt.Token) (string, error) {
	tokenString, err := token.SignedString([]byte(b.secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (b *JWTBuilder) SignUserUUID(userToken string) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		Claims {
			UserUUID: userToken,
		},
	)
	tokenString, err := b.Sign(token)
	if err != nil {
		b.suLog.Infow("there is an error with sign", err)
		return "", err
	}
	return tokenString, nil
}

func (b *JWTBuilder) ParseUserUUID(tokenString string) (string, error) {
    claims := &Claims{}
    token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
    return []byte(b.secretKey), nil
	})
	if err != nil {
		b.suLog.Infow("error during parse", err)
        return "", err
    }

    if !token.Valid {
        b.suLog.Infow("token is invalid")
        return "", errors.New("there is invalid token")
    }

    if claims.UserUUID == "" {
        b.suLog.Infow("token does not contain user UUID")
        return "", ErrNoUserUUID
    }

    b.suLog.Infow("Token is valid")
    return claims.UserUUID, nil
}

func getOrCreateUserUUID(b *JWTBuilder, w http.ResponseWriter, req *http.Request) (string, error) {
	var userUUID string
	var err error
	authHeader := req.Header.Get("Authorization")

	if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			userUUID, err = b.ParseUserUUID(tokenString)
			b.suLog.Infow("failed to parse user UUID from token", "error", err)
		}
	if userUUID != "" {
		return userUUID, nil
	}

    cookie, err := req.Cookie(b.UserCookieName)
    if err != nil {
        b.suLog.Infoln("no cookie/ auth header, generating new one")
        userUUID := b.GenerateUserUUID()
        token, err := b.SignUserUUID(userUUID)
        if err != nil {
			b.suLog.Infow("during sign uuid", "user", userUUID[:6], "error", err)
            return "", err
        }
        http.SetCookie(w, &http. Cookie{
            Name:     b.UserCookieName,
            Value:    token,
            Path:     "/",
            HttpOnly: true,
            Secure:   true, 
        })
		b.suLog.Infoln("set cookie")
		w.Header().Set("Authorization", "Bearer "+token)
		b.suLog.Infoln("set authorization header")
        return userUUID, nil
    }
    userUUID, err = b.ParseUserUUID(cookie.Value)
    if err != nil {
        b.suLog.Infoln("cookie parsing failed, issuing new one")
        newUUID := b.GenerateUserUUID()
        token, err := b.SignUserUUID(newUUID)
        if err != nil {
			b.suLog.Infow("during sign uuid", "user", userUUID[:6], "error", err)
            return "", err
        }
        http.SetCookie(w, &http.Cookie{
            Name:     b.UserCookieName,
            Value:    token,
            Path:     "/",
            HttpOnly: true,
            Secure:   true,
        })
        return "", ErrNoUserUUID
    }

    b.suLog.Infoln("valid cookie with user UUID")
    return userUUID, nil
}

func AuthMiddleware(b *JWTBuilder) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
            userUUID, err := getOrCreateUserUUID(b, w, req)
			b.suLog.Infoln("userUUID was got")
			if err != nil {
				b.suLog.Infoln("there is an authorization error")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("there is an authorization error"))
				return
			}

            ctx := context.WithValue(req.Context(), userUUIDKey, userUUID)

            next.ServeHTTP(w, req.WithContext(ctx))
        })
    }
}
