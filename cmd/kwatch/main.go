package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"

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

	username := flag.String("u", "", "HTTP Username [optional]")
	password := flag.String("p", "", "HTTP Password [optional]")
	addressString := flag.String("a", "", "Root caddy fileserver address [required]")
	fileViewer := flag.String("o", "", "Program to open files with [optional] [defaults to mpv]")
	confFile := flag.String("c", confDir+"/kwatch.toml", "Configuration file [optional]")
	writeToConfig := flag.Bool("w", false, "Write current arguments to config file [optional]")
	flag.Parse()

	address := &url.URL{}
	if len(*addressString) > 0 {
		address, err = url.Parse(*addressString)
		if err != nil {
			log.Fatal(err)
		}
		if len(address.Scheme) <= 0 {
			fmt.Println("Address requires scheme (http/https)")
			os.Exit(1)
		}
	}

	config := &cfg.Config{
		Address:    *address,
		Username:   *username,
		Password:   *password,
		FileViewer: *fileViewer,
	}

	if *writeToConfig {
		fmt.Println("Attempting to write current options to config file " + *confFile)
		err = cfg.WriteConfig(config, &confDir, confFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = cfg.ReadConfig(config, confFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
	}

	if len(config.Address.Host) <= 0 {
		flag.Usage()
		os.Exit(1)
	}

	config.Address.Path = ""

	if len(config.FileViewer) <= 0 {
		config.FileViewer = "mpv"
	}

	_, err = exec.LookPath(config.FileViewer)
	if err != nil {
		log.Fatal(err)
	}

	program := ui.NewProgram(config)

	if err := program.Start(); err != nil {
		log.Fatal(err)
	}
}
