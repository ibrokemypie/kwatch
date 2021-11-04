package file

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	"github.com/ibrokemypie/kwatch/pkg/cfg"
)

func OpenFile(config *cfg.Config, openBookmark int, path []string, filePath string) error {
	bookmark := config.Bookmarks[openBookmark]

	address, err := url.Parse(bookmark.Address)
	if err != nil {
		return err
	}

	address.User = url.UserPassword(bookmark.Username, bookmark.Password)
	address.Path = "/" + strings.Join(path, "/") + "/" + filePath
	runCMD := exec.Command(config.FileViewer, address.String())

	err = runCMD.Run()
	if err != nil {
		return fmt.Errorf("%s: %s", runCMD.String(), err.Error())
	}

	return nil
}
