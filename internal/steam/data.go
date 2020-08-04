package steam

import (
	"io/ioutil"
	"path"

	"github.com/thieman/steam-automigrate/internal/config"
)

// Returns all Steam IDs found in the local Steam install's userdata folder
func steamIds(config *config.Config) ([]string, error) {
	dir := path.Join(config.SteamMainDir, "userdata")

	fileinfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var ids []string
	for _, fileinfo := range fileinfos {
		ids = append(ids, fileinfo.Name())
	}

	return ids, nil
}
