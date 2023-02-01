package metrics

import (
	"encoding/json"
	"fmt"
	"os"
)

type PrometheusConfiguration struct {
	Host    string `json:"host"`
	Port    string `json:"port"`
	Enabled bool   `json:"enabled"`
}

func (promConfig *PrometheusConfiguration) Init(path string) error {
	content, err := os.ReadFile(path) // Read json file
	if err != nil {
		fmt.Println("Error happened when reading")
		return err
	}
	if err := json.Unmarshal(content, promConfig); err != nil {
		fmt.Println("Error when unmarshaling")
		return err
	}
	return nil
}
