package factory_test

import (
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/th2-net/th2-common-go/pkg/factory"
	"os"
	"testing"
	"testing/fstest"
)

var logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

type testStr struct {
	Key string `json:"key"`
}

func TestFindsResource(t *testing.T) {
	provider := factory.NewFileProviderForFS(
		fstest.MapFS{
			"test.json": &fstest.MapFile{
				Data: []byte(`{ "key": "value" }`),
			},
		},
		".json",
		logger,
	)
	var str testStr
	if err := provider.GetConfig("test", &str); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, str.Key, "value", "unexpected value deserialized")
}

func TestNoResource(t *testing.T) {
	provider := factory.NewFileProviderForFS(
		fstest.MapFS{},
		".json",
		logger,
	)
	var str testStr
	err := provider.GetConfig("test", &str)
	assert.ErrorIs(t, err, factory.ResourceNotFound, "unexpected error")
}

func TestReturnsDeserializationError(t *testing.T) {
	provider := factory.NewFileProviderForFS(
		fstest.MapFS{
			"test.json": &fstest.MapFile{
				Data: []byte(`{ "key": { "sun": 42 } }`),
			},
		},
		".json",
		logger,
	)
	var str testStr
	err := provider.GetConfig("test", &str)
	assert.Error(t, err, "should get deserialization error")
}

func TestProcessesEnvVariables(t *testing.T) {
	provider := factory.NewFileProviderForFS(
		fstest.MapFS{
			"test.json": &fstest.MapFile{
				Data: []byte(`{ "key": "${TEST_ENV_VAR}" }`),
			},
		},
		".json",
		logger,
	)
	if err := os.Setenv("TEST_ENV_VAR", "value"); err != nil {
		t.Error("cannot set var", err)
	}
	var str testStr
	if err := provider.GetConfig("test", &str); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, str.Key, "value", "unexpected value deserialized")
}
