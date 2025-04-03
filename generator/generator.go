// go:build ignore
//go:build ignore
// +build ignore

// ^^ This build tag ensures this file is not compiled with your main application

package main

import (
	"fmt"

	"gorm.io/gen"

	"lqkhoi-go-http-api/internal/models"
)

const outPath = "../internal/query"

func main() {

	g := gen.NewGenerator(gen.Config{
		OutPath:      outPath,
		ModelPkgPath: "models",

		Mode: gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	modelsToGenerate := []any{
		models.Project{},
		models.Sprint{},
		models.Task{},
		models.User{},
	}

	g.ApplyBasic(modelsToGenerate...)

	g.Execute()

	fmt.Printf("Successfully generated code in %s\n", outPath)
}
