package settings

import (
	"encoding/json"
	"os"
)

type Settings struct {
	Volume              float64 `json:"volume"`
	Language            string  `json:"language"`
	Difficulty          string  `json:"difficulty"`
	BackgroundAnimation bool    `json:"background_animation"`
}

var Current Settings

func Load() error {
	data, err := os.ReadFile("settings.json")
	if err != nil {
		Current = Settings{
			Volume:              0.7,
			Language:            "ru",
			Difficulty:          "normal",
			BackgroundAnimation: true,
		}
		return Save()
	}
	return json.Unmarshal(data, &Current)
}

func Save() error {
	data, err := json.MarshalIndent(Current, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("settings.json", data, 0644)
}