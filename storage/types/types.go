package storageTypes

import (
	"context"
	"io"
	"net/http"
)

type ReadableFile interface {
	io.ReadCloser
}

type WriteableFile interface {
	io.WriteCloser
}

type Storage interface {
	Exists(context.Context, string) bool
	ReadFile(context.Context, string) (ReadableFile, error)
	WriteFile(context.Context, string) (WriteableFile, error)
	Open(string) (http.File, error)
}
