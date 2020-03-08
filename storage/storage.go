package storage

import (
	"strings"

	fileBackend "github.com/NamedKitten/kittehimageboard/storage/backends/file"
	"github.com/NamedKitten/kittehimageboard/types"
)

func GetStorage(s string) types.Storage {
	if strings.HasPrefix(s, "file://") {
		return fileBackend.New(strings.TrimPrefix(s, "file://"))
	}
	return nil
}
