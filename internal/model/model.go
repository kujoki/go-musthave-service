package model

import "strconv"

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
    UUID string    `json:"uuid"`
    ShortURL string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func MapToDataSlice(dataMap map[string]string) []Data {
    dataSlice := make([]Data, 0, len(dataMap))
    i := 1
    for short, orig := range dataMap {
        dataSlice = append(dataSlice, Data{
            UUID:        strconv.Itoa(i),
            ShortURL:    short,
            OriginalURL: orig,
        })
        i++
    }
    return dataSlice
}

func DataSliceToMap(dataSlice []Data) map[string]string {
    dataMap := make(map[string]string)
    for _, data := range dataSlice {
        dataMap[data.ShortURL] = data.OriginalURL
    }
    return dataMap
}