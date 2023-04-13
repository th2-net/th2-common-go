package modules

import (
	"github.com/th2-net/th2-common-go/pkg/modules/prometheus"
	"testing"
	"testing/fstest"
)

func TestCanRegisterPrometheus(t *testing.T) {
	cfg := `
	{
		"enabled": false,
		"host": "0.0.0.0",
		"port": "9752"
	}
	`
	factory := CreateTestFactory(fstest.MapFS{
		"prometheus": &fstest.MapFile{
			Data: []byte(cfg),
		},
	})

	if err := factory.Register(prometheus.NewModule); err != nil {
		t.Fatal(err)
	}

	mod, err := prometheus.ModuleID.GetModule(factory)
	if err != nil {
		t.Fatal(err)
	}
	if mod == nil {
		t.Fatal("module is nil")
	}
}
