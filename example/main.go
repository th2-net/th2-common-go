package main

import (
	"fmt"
	"os"
	"reflect"
)

import (
	fc "github.com/th2-net/th2-common-go/schema/factory"
	md "github.com/th2-net/th2-common-go/schema/modules"
)

func main() {
	fmt.Println("test")
	factory := fc.NewFactory(os.Args)
	if err := factory.Register(md.NewQueueModule); err != nil {
		panic(err)
	}

	if module, err := md.ModuleID.GetModule(factory); err != nil {
		panic("no module")
	} else {
		fmt.Println("module found", reflect.TypeOf(module))
	}

}
