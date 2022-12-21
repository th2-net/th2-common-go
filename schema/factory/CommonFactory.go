package factory

import (
	"errors"
	"fmt"
	"github.com/th2-net/th2-common-go/schema/common"
	"reflect"
)

const (
	sourceDirectory = "../resources"
)

type CommonFactory struct {
	modules     map[common.ModuleKey]common.Module
	cfgProvider ConfigProvider
}

func NewProvider(args []string) ConfigProvider {
	return &ConfigProviderFromFile{directoryPath: sourceDirectory, files: args}
}

func NewFactory(args []string) *CommonFactory {
	provider := NewProvider(args[1:])
	return &CommonFactory{
		modules:     make(map[common.ModuleKey]common.Module),
		cfgProvider: provider,
	}
}

func (cf *CommonFactory) Register(factories ...func(ConfigProvider) common.Module) error {
	for _, factory := range factories {
		module := factory(cf.cfgProvider)
		if oldModule, exist := cf.modules[module.GetKey()]; exist {
			return fmt.Errorf("module %s with key %s already registered", reflect.TypeOf(oldModule), module.GetKey())
		}
		cf.modules[module.GetKey()] = module
	}
	return nil
}

func (cf *CommonFactory) Get(key common.ModuleKey) (common.Module, error) {
	if module, exist := cf.modules[key]; !exist {
		return nil, errors.New("module " + string(key) + " does not exist")
	} else {
		return module, nil
	}
}
