package file

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	"github.com/ibrokemypie/kwatch/pkg/cfg"
)

func OpenFile(cfg *cfg.Config, path []string, filePath string) error {
	addressCopy := cfg.Address

	addressCopy.User = url.UserPassword(cfg.Username, cfg.Password)
	addressCopy.Path = "/" + strings.Join(path, "/") + "/" + filePath
	runCMD := exec.Command(cfg.FileViewer, addressCopy.String())

	err := runCMD.Run()
	if err != nil {
		return fmt.Errorf("%s: %s", runCMD.String(), err.Error())
	}

	return nil
}
