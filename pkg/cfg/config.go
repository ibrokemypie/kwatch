package cfg

import (
	"os"
	"path/filepath"

	"github.com/ibrokemypie/kwatch/pkg/source/bookmark"
	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Bookmarks       []bookmark.Bookmark
	DefaultBookmark int
}

func (cfg Config) GetBookmarks() []bookmark.Bookmark {
	return cfg.Bookmarks
}

func (cfg Config) GetDefaultBookmark() int {
	if len(cfg.Bookmarks) > 0 {
		return cfg.DefaultBookmark
	} else {
		return -1
	}
}

func (cfg *Config) SetDefaultBookmark(newDefaultBookmark int) {
	cfg.DefaultBookmark = newDefaultBookmark
}

func (cfg Config) GetBookmark(index int) bookmark.Bookmark {
	return cfg.Bookmarks[index]
}

func (cfg *Config) AddBookmark(b bookmark.Bookmark) {
	cfg.Bookmarks = append(cfg.Bookmarks, b)
}

func (cfg *Config) UpdateBookmark(index int, b bookmark.Bookmark) {
	cfg.Bookmarks[index] = b
}

func (cfg Config) WriteConfig(confFilePath string) error {
	bytes, err := toml.Marshal(cfg)
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

func (cfg *Config) ReadConfig(confFilePath string) error {
	bytes, err := os.ReadFile(confFilePath)
	if err != nil {
		return err
	}

	err = toml.Unmarshal(bytes, cfg)
	if err != nil {
		return err
	}

	return nil
}
