package storage

import (
	"strings"

	fileBackend "github.com/NamedKitten/kittehimageboard/storage/backends/file"
	storageTypes "github.com/NamedKitten/kittehimageboard/storage/types"
)

func GetStorage(s string) storageTypes.Storage {
	if strings.HasPrefix(s, "file://") {
		return fileBackend.New(strings.TrimPrefix(s, "file://"))
	}
	return nil
}
