package fileBackend

import (
	"context"
	"net/http"
	"os"

	storageTypes "github.com/NamedKitten/kittehimageboard/storage/types"
)

type FileBackend struct {
	path string
}

func (fb FileBackend) Open(s string) (http.File, error) {
	return os.OpenFile(fb.path+s, os.O_RDONLY, 0666)
}

func (fb FileBackend) ReadFile(ctx context.Context, s string) (storageTypes.ReadableFile, error) {
	return os.OpenFile(fb.path+s, os.O_RDONLY, 0666)
}

func (fb FileBackend) WriteFile(ctx context.Context, s string) (storageTypes.WriteableFile, error) {
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
