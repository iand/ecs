//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/iand/ecs/ecsgen"
)

func main() {
	gen := new(ecsgen.Generator)

	gen.Package("example")
	gen.ManagerName("EM")
	gen.Filename("ecsem.go")
	gen.AddComponent(1, "Dynamics")
	gen.AddComponent(2, "Behavior")
	gen.AddArchetype("Creature", "Behavior", "Dynamics")

	if err := gen.Generate(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
