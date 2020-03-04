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
	return fb.File(s)
}

func (fb FileBackend) File(s string) (storageTypes.File, error) {
	f, e := os.OpenFile(fb.path + s, os.O_RDWR|os.O_CREATE, 0666)
	return f, e
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

