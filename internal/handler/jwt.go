package handler

import (
	"github.com/golang-jwt/jwt/v4"
	"errors"
	"log"
	"crypto/rand"
    "encoding/hex"
)

type Claims struct {
	UserUUID string `json:"user_uuid"`
	jwt.RegisteredClaims
}

type JWTBuilder struct {
    secretKey 	   string
	issuer 		   string
	UserCookieName string
}

func NewJWTBuild(secret string, issuer string, userCookieName string) *JWTBuilder {
	return &JWTBuilder{
		secretKey: secret,
		issuer: issuer,
		UserCookieName: userCookieName,
	}
}

var ErrNoUserUUID = errors.New("there is no user UUID")
var ErrInvalidToken = errors.New("there is unvalid token")
var ErrUnexpectedAlg = errors.New("unexpected signing method")

func (b *JWTBuilder) GenerateUserUUID() string {
	userUUID := make([]byte, 16)
	if _, err := rand.Read(userUUID); err != nil {
		log.Printf("error while uuid generating: %v\n", err)
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
		log.Println("there is an error with sign", err)
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
		log.Println("error during parse ", err)
        return "", err
    }

    if !token.Valid {
        log.Println("token is invalid")
        return "", ErrInvalidToken
    }

    if claims.UserUUID == "" {
        log.Println("token does not contain user UUID")
        return "", ErrNoUserUUID
    }

    log.Println("Token is valid")
    return claims.UserUUID, nil
}