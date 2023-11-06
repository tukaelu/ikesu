package file

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/tukaelu/ikesu/internal/config/loader"
)

var (
	ErrNoCheckRules     = fmt.Errorf("No check rules defined.")
	ErrNoSuchConfigFile = fmt.Errorf("No such config file.")
	ErrEmptyConfigFile  = fmt.Errorf("The specified config file is empty.")
)

func init() {
	loader.Register("file", &Loader{})
}

type Loader struct{}

func (d *Loader) LoadWithContext(ctx context.Context, u *url.URL) ([]byte, error) {
	f, err := os.Open(u.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return nil, ErrNoSuchConfigFile
	}
	if fi.Size() == 0 {
		return nil, ErrEmptyConfigFile
	}
	return os.ReadFile(u.Path)
}
