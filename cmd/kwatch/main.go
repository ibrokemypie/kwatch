package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/ibrokemypie/kwatch/pkg/cfg"
	"github.com/ibrokemypie/kwatch/pkg/ui"
)

func main() {
	confDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("Unable to find user config dir: %s", err)
		confDir, err = os.Getwd()
		if err != nil {
			log.Fatalf("Unable to find current working directory dir: %s", err)
		}
	}

	confFile := flag.String("c", confDir+"/kwatch.toml", "Configuration file [optional]")
	flag.Parse()

	confFilePath := filepath.Clean(*confFile)

	config := new(cfg.Config)
	err = config.ReadConfig(confFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
	}

	program := ui.NewProgram(config, confFilePath)

	if err := program.Start(); err != nil {
		log.Fatal(err)
	}
}
