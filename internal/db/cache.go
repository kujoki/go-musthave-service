package cache

import (
	"os"
	"encoding/json"
    "io"
    "github.com/kujoki/go-musthave-service/internal/model"
)

func Load(fname string) ([]model.Data, error) {
	file, err := os.OpenFile(fname, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dataBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if len(dataBytes) == 0 {
		return []model.Data{}, nil
	}

	var items []model.Data
	if err := json.Unmarshal(dataBytes, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func Save(fname string, items []model.Data) error {
	file, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(items)
}