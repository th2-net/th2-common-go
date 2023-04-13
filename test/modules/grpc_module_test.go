package modules

import (
	"github.com/th2-net/th2-common-go/pkg/modules/grpc"
	"testing"
	"testing/fstest"
)

func TestCanRegisterGrpc(t *testing.T) {
	cfg := `
    {
		"server": {
		},
		"services": {
		}
	}
    `
	factory := CreateTestFactory(fstest.MapFS{
		"grpc": &fstest.MapFile{
			Data: []byte(cfg),
		},
	})

	err := factory.Register(grpc.NewModule)
	if err != nil {
		t.Fatal(err)
	}
	var mod grpc.Module
	mod, err = grpc.ModuleID.GetModule(factory)
	if err != nil {
		t.Fatal(err)
	}
	if mod == nil {
		t.Fatal("module is nil")
	}
}
