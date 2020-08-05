package steam

import (
	"io/ioutil"
	"path/filepath"
)

// Returns all Steam IDs found in the local Steam install's userdata folder
func steamIds(config *Config) ([]string, error) {
	dir := filepath.Join(config.SteamMainDir, "userdata")

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
