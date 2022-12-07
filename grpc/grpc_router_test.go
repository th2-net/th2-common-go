package grpc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	path              = "../resources"
	validGrpcFileName = "grpc.json"
)

func assertNilWithMsg(t *testing.T, err error) {
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func getGrpcConfig(fileName string) (GrpcConfig, error) {
	gr := CommonGrpcRouter{Config: GrpcConfig{}}
	cp := ConfigProviderFromFile{DirectoryPath: path}
	err := cp.GetConfig(fileName, &gr.Config)

	return gr.Config, err
}

func TestGetConfig(t *testing.T) {
	_, err := getGrpcConfig(validGrpcFileName)
	assertNilWithMsg(t, err)
}

func TestValidatePinsSingleEndpoint(t *testing.T) {
	gc, _ := getGrpcConfig(validGrpcFileName)
	err := gc.validatePins()
	assertNilWithMsg(t, err)
}

func TestValidatePinsMultipleEndpoints(t *testing.T) {
	gc, _ := getGrpcConfig("grpc_invalid1.json")
	err := gc.validatePins()
	assert.NotNil(t, err)
}

func TestFindEndpointAddrViaAttributesExactMatch(t *testing.T) {
	gc, _ := getGrpcConfig(validGrpcFileName)
	addr, err := gc.findEndpointAddrViaAttributes([]string{"actAttr", "otherAttr1", "otherAttr2"})
	assertNilWithMsg(t, err)
	assert.Equal(t, ":8080", addr.asColonSeparatedString(), "Expected and actual addresses not equal")
}

func TestFindEndpointAddrViaAttributesPartlyMatch(t *testing.T) {
	gc, _ := getGrpcConfig(validGrpcFileName)
	addr, err := gc.findEndpointAddrViaAttributes([]string{"actAttr"})
	assertNilWithMsg(t, err)
	assert.Equal(t, ":8080", addr.asColonSeparatedString(), "Expected and actual addresses not equal")
}

func TestFindEndpointAddrViaAttributesNonExistingTarget(t *testing.T) {
	gc, _ := getGrpcConfig(validGrpcFileName)
	_, err := gc.findEndpointAddrViaAttributes([]string{"wrongAttr"})
	assert.NotNil(t, err)
}

func TestFindEndpointAddrViaAttributesNotEnoughInclusive(t *testing.T) {
	gc, _ := getGrpcConfig(validGrpcFileName)
	_, err := gc.findEndpointAddrViaAttributes([]string{"wrongAttr", "actAttr"})
	assert.NotNil(t, err)
}
