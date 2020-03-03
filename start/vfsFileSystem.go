package start

import (
	"net/http"
	"os"
	"github.com/c2fo/vfs/v5"
	"time"
	"strings"
)

type VFSFileInfo struct {
	f vfs.File
}

func (f VFSFileInfo) Name() string {
	return f.f.Name()
}
func (f VFSFileInfo) IsDir() bool {
	return false
}
func (f VFSFileInfo) ModTime() time.Time {
	m, err := f.f.LastModified()
	if err != nil {
		return time.Time{}
	} else {
		return *m
	}
	return time.Time{}
}
func (f VFSFileInfo) Mode() os.FileMode {
	return 0
}
func (f VFSFileInfo) Size() int64 {
	s, err := f.f.Size()
	if err != nil {
		return 0
	} else {
		return int64(s)
	}
}
func (f VFSFileInfo) Sys() interface{} {
	return nil
}

type VFSFile struct {
	vfs.File
}
func (f VFSFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}
func (f VFSFile) Stat() (os.FileInfo, error) {
	return VFSFileInfo{f}, nil
}

type VFSFileSystem struct {}
func (VFSFileSystem) Open(name string) (http.File, error) {
	f, e := DB.ContentStorage.NewFile(strings.ReplaceAll(name, "/", ""))
	if e != nil {
		panic(e)
	}
	return VFSFile{f}, e
}