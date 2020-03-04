package storage

import (
	"github.com/NamedKitten/kittehimageboard/storage/types"
	"github.com/NamedKitten/kittehimageboard/storage/backends/file"
	"strings"
)

func GetStorage(s string) storageTypes.Storage {
	if strings.HasPrefix(s, "file://") {
		return fileBackend.New(strings.TrimLeft(s, "file://"))
	}
	return nil
}