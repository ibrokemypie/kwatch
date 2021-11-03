package cfg

import (
	"net/url"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Username   string
	Password   string
	Address    url.URL
	FileViewer string
}

func WriteConfig(cfg *Config, confDir, confFile *string) error {
	bytes, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	err = os.MkdirAll(*confDir, os.ModeAppend)
	if err != nil {
		return err
	}

	file, err := os.Create(*confFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func ReadConfig(cfg *Config, confFile *string) error {
	var readCfg Config
	bytes, err := os.ReadFile(*confFile)
	if err != nil {
		return err
	}
	err = toml.Unmarshal(bytes, &readCfg)
	if err != nil {
		return err
	}

	if len(cfg.Username) <= 0 {
		cfg.Username = readCfg.Username
	}
	if len(cfg.Password) <= 0 {
		cfg.Password = readCfg.Password
	}
	if len(cfg.Address.Host) <= 0 {
		cfg.Address = readCfg.Address
	}
	if len(cfg.FileViewer) <= 0 {
		cfg.FileViewer = readCfg.FileViewer
	}

	return nil
}
