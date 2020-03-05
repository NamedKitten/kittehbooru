package storageTypes

import "net/http"
import "io"

type ReadableFile interface {
	http.File
}

type WriteableFile interface {
	io.Writer
	http.File
}

type Storage interface {
	Exists(string) bool
	ReadFile(string) (ReadableFile, error)
	WriteFile(string) (WriteableFile, error)
	Open(string) (http.File, error) 
}