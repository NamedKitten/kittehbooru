package fileBackend

import (
	"os"
	"github.com/NamedKitten/kittehimageboard/storage/types"
	"net/http"
)

type FileBackend struct {
	path string
}

func (fb FileBackend) Open(s string) (http.File, error) {
	return fb.ReadFile(s)
}

func (fb FileBackend) ReadFile(s string) (storageTypes.ReadableFile, error) {
	return os.OpenFile(fb.path + s, os.O_RDONLY|os.O_CREATE, 0666)
}


func (fb FileBackend) WriteFile(s string) (storageTypes.WriteableFile, error) {
	return os.OpenFile(fb.path + s, os.O_WRONLY|os.O_CREATE, 0666)
}

func (fb FileBackend) Exists(s string) bool {
	f, err := os.Stat(fb.path + s)
	if os.IsNotExist(err) {
		return false 
	}
	return !f.IsDir()
}


func New(s string) FileBackend {
	return FileBackend{s}
}

