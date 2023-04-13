package modules

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/th2-net/th2-common-go/pkg/modules/queue"
	"testing"
	"testing/fstest"
)

const (
	containerName = "rabbitmq-container-test"
	mqPort        = "5672"
)

func TestCanRegisterMqModule(t *testing.T) {
	if testing.Short() {
		t.Skip()
		return
	}
	host, port, err := startContainer(t)
	connectionCfg := connectionConfiguration(host, port)

	mqCfg := `{
	  "queues": {
		"pin":  {
		  "attributes": [
			"group",
			"publish"
		  ],
		  "exchange": "",
		  "filters": [],
		  "name": "",
		  "queue": ""
		}
	  }
	}`

	factory := CreateTestFactory(fstest.MapFS{
		"rabbitMQ": &fstest.MapFile{
			Data: []byte(connectionCfg),
		},
		"mq": &fstest.MapFile{
			Data: []byte(mqCfg),
		},
	})

	err = factory.Register(queue.NewRabbitMqModule)
	if err != nil {
		t.Fatal(err)
	}
	var mod queue.Module
	mod, err = queue.ModuleID.GetModule(factory)
	if err != nil {
		t.Fatal(err)
	}
	if mod == nil {
		t.Fatal("module is nil")
	}
}

func connectionConfiguration(host string, port nat.Port) string {
	connectionCfg := fmt.Sprintf(`{
	  "host":"%s",
	  "vHost": "",
	  "port": "%d",
	  "username": "guest",
	  "password": "guest",
	  "exchangeName": "amq.direct"
	}`, host, port.Int())
	return connectionCfg
}

func startContainer(t *testing.T) (string, nat.Port, error) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Name:         containerName,
		Image:        "rabbitmq:3.10",
		ExposedPorts: []string{mqPort},
		WaitingFor:   wait.ForLog("Server startup complete"),
	}
	rabbit, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            true,
	})
	if err != nil {
		t.Fatal("cannot create container", err)
	}
	t.Cleanup(func() {
		err := rabbit.Terminate(ctx)
		if err != nil {
			t.Logf("cannot stop rabbitmq container: %v", err)
		}
	})
	host, err := rabbit.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := rabbit.MappedPort(ctx, mqPort)
	if err != nil {
		t.Fatal(err)
	}
	return host, port, err
}
