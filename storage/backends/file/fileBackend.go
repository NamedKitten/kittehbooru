package fileBackend

import (
	"context"
	"net/http"
	"os"

	"github.com/NamedKitten/kittehimageboard/types"
)

type FileBackend struct {
	path string
}

func (fb FileBackend) Open(s string) (http.File, error) {
	return os.OpenFile(fb.path+s, os.O_RDONLY, 0666)
}

func (fb FileBackend) Delete(s string) error {
	return os.Remove(fb.path + s)
}

func (fb FileBackend) ReadFile(ctx context.Context, s string) (types.ReadableFile, error) {
	return os.OpenFile(fb.path+s, os.O_RDONLY, 0666)
}

func (fb FileBackend) WriteFile(ctx context.Context, s string) (types.WriteableFile, error) {
	return os.OpenFile(fb.path+s, os.O_WRONLY|os.O_CREATE, 0666)
}

func (fb FileBackend) Exists(ctx context.Context, s string) bool {
	f, err := os.Stat(fb.path + s)
	if os.IsNotExist(err) {
		return false
	}
	return !f.IsDir()
}

func New(s string) FileBackend {
	return FileBackend{s}
}
