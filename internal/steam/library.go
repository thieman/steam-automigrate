package steam

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

type Library struct {
	Type            string
	Path            string `toml:"path"`
	VolumeName      string
	FreeSpaceBytes  uint64
	TotalSpaceBytes uint64
	PercentFree     float64
}

func (l *Library) GetManifestPaths() ([]string, error) {
	dir := filepath.Join(l.Path, "steamapps")

	fileinfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var paths []string

	for _, fileinfo := range fileinfos {
		if strings.HasPrefix(fileinfo.Name(), "appmanifest") && strings.HasSuffix(fileinfo.Name(), "acf") {
			paths = append(paths, filepath.Join(dir, fileinfo.Name()))
		}
	}

	return paths, nil
}
