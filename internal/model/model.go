package model

import (
    "errors"
    "strconv"
)

var ErrURLExists = errors.New("URL already exists")

type Task struct {
    UserUUID string
    Item string
}

type ShortURLResult struct {
    ShortURL string
    IsDeleted bool
}

type LongURLResult struct {
    OriginURL string
    IsDeleted bool
}

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type BatchRequest struct {
    CorrelationID string `json:"correlation_id"`
    OriginalURL string `json:"original_url"`
}

type BatchResponse struct {
    CorrelationID string `json:"correlation_id"`
    ShortURL string `json:"short_url"`
}

type Data struct {
    UUID        string `json:"uuid"`
    ShortURL    string `json:"short_url"`
    OriginalURL string `json:"original_url"`
    UserUUID    string `json:"user_uuid"`
    IsDeleted   bool   `json:"is_deleted"`
}

type UserURL struct {
    ShortURL    string `json:"short_url"`
    OriginalURL string `json:"original_url"`
}

type URLRecord struct {
    LongURL  string
    UserUUID string
    IsDeleted bool
}

func DataSliceToMap(dataSlice []Data) map[string]URLRecord {
    dataMap := make(map[string]URLRecord, len(dataSlice))
    for _, d := range dataSlice {
        dataMap[d.ShortURL] = URLRecord{
            LongURL:   d.OriginalURL,
            UserUUID: d.UserUUID, 
            IsDeleted: d.IsDeleted,
        }
    }
    return dataMap
}

func MapToDataSlice(dataMap map[string]URLRecord) []Data {
    dataSlice := make([]Data, 0, len(dataMap))
    i := 1
    for short, rec := range dataMap {
        dataSlice = append(dataSlice, Data{
            UUID:        strconv.Itoa(i),
            ShortURL:    short,
            OriginalURL: rec.LongURL,
            UserUUID:    rec.UserUUID,
            IsDeleted:   rec.IsDeleted,
        })
        i++
    }
    return dataSlice
}