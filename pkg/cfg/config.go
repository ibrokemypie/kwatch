package cfg

import (
	"os"
	"path/filepath"

	"github.com/ibrokemypie/kwatch/pkg/source"
	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Bookmarks       []source.Bookmark
	DefaultBookmark int
	FileViewer      string
}

func (config *Config) WriteConfig(confFilePath string) error {
	bytes, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(confFilePath), os.ModeAppend)
	if err != nil {
		return err
	}

	file, err := os.Create(confFilePath)
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

func (config *Config) ReadConfig(confFilePath string) error {
	bytes, err := os.ReadFile(confFilePath)
	if err != nil {
		return err
	}

	err = toml.Unmarshal(bytes, config)
	if err != nil {
		return err
	}

	return nil
}
