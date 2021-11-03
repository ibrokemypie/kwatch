package file

import (
	"net/url"
	"os/exec"
	"strings"

	"github.com/ibrokemypie/kwatch/pkg/cfg"
)

func OpenFile(cfg *cfg.Config, path []string, filePath string) error {
	addressCopy := cfg.Address

	addressCopy.User = url.UserPassword(cfg.Username, cfg.Password)
	addressCopy.Path = strings.Join(path, "/") + filePath
	runCMD := exec.Command(cfg.FileViewer, addressCopy.String())

	return runCMD.Run()
}
