package factory

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type ConfigProvider interface {
	GetConfig(resourceName string, target interface{}) error
}

// ConfigProviderFromFile must be without the trailing slash
// example: "dir/subdir"
type ConfigProviderFromFile struct {
	configurationPath string
	files             []string
}

func (cfd *ConfigProviderFromFile) getPath(resourceName string) string {
	if len(cfd.files) == 0 {
		path, err := filepath.Abs(fmt.Sprint(cfd.configurationPath, "/", resourceName, ".json"))
		if err != nil {
			log.Fatalln(err)
		}
		return path
	} else {
		for _, filePath := range cfd.files {
			if strings.Contains(filePath, "/") {
				fileSlice := strings.Split(filePath, "/")
				if fileSlice[len(fileSlice)-1] == resourceName {
					path, err := filepath.Abs(fmt.Sprint(filePath, ".json"))
					if err != nil {
						log.Fatalln(err)

					}
					return path
				}
			} else {
				if filePath == resourceName {
					path, err := filepath.Abs(fmt.Sprint(cfd.configurationPath, "/", resourceName, ".json"))
					if err != nil {
						log.Fatalln(err)
					}
					return path
				}
			}
		}
	}
	return ""
}

func (cfd *ConfigProviderFromFile) GetConfig(resourceName string, target interface{}) error {

	path := cfd.getPath(resourceName)
	fileContentBytes, fileReadErr := os.ReadFile(path)
	if fileReadErr != nil {
		return errors.New(fmt.Sprintf("file with path %s couldn't be read", path))
	}

	err := json.Unmarshal(fileContentBytes, target)

	if err != nil {
		return errors.New(fmt.Sprintf("deserialization error: %s ", err.Error()))
	}

	return nil
}
