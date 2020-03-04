package storageTypes

import "net/http"
import "io"

type File interface {
	io.Writer
	http.File
}


type Storage interface {
	Exists(string) bool
	File(string) (File, error)
	Open(string) (http.File, error) 
}