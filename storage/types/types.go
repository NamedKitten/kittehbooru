package storageTypes

import "net/http"
import "io"

type ReadableFile interface {
	io.ReadCloser
}

type WriteableFile interface {
	io.WriteCloser
}

type Storage interface {
	Exists(string) bool
	ReadFile(string) (ReadableFile, error)
	WriteFile(string) (WriteableFile, error)
	Open(string) (http.File, error)
}
