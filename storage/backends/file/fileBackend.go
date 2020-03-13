package fileBackend

import (
	"context"
	"net/http"
	"os"
	"runtime/trace"

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
	defer trace.StartRegion(ctx, "FileStorage/ReadFile").End()
	return os.OpenFile(fb.path+s, os.O_RDONLY, 0666)
}

func (fb FileBackend) WriteFile(ctx context.Context, s string) (types.WriteableFile, error) {
	defer trace.StartRegion(ctx, "FileStorage/WriteFile").End()
	return os.OpenFile(fb.path+s, os.O_WRONLY|os.O_CREATE, 0666)
}

func New(s string) FileBackend {
	return FileBackend{s}
}
