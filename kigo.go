package main

import (
	"github.com/AngelVI13/kigo/utils"
	"github.com/alexflint/go-arg"
	"log"
)

import _ "embed"

//go:embed defaults.json
var DEFAULT_CONFIGS []byte

// todo 1. add tests
// todo 2. update readme
func main() {
	arg.MustParse(&utils.Args)
	switch {
	case utils.Args.Run != nil:
		config, err := utils.LoadConfig(utils.Args.Run.Config)
		if err != nil {
			log.Fatal(err)
		}

		utils.Run(&config)
	case utils.Args.Create != nil:
		if err := utils.CreateDefaultConfig(DEFAULT_CONFIGS, utils.Args.Create.Tag, utils.Args.Create.OutFile); err != nil {
			log.Fatal(err)
		}
		log.Println(utils.CommandStyle("Config created successfully!"))
	}
}
