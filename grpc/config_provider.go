package grpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	grpcJsonFileName = "grpc.json" //for client/server sample. will be moved out from here
)

type ConfigProvider interface {
	GetConfig(resourceName string, target interface{}) error
}

// ConfigProviderFromFile must be without the trailing slash
// example: "dir/subdir"
type ConfigProviderFromFile struct {
	DirectoryPath string
}

func (cfd *ConfigProviderFromFile) GetConfig(resourceName string, target interface{}) error {
	path, _ := filepath.Abs(fmt.Sprint(cfd.DirectoryPath, "/", resourceName))
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
