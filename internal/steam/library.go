package steam

import (
	"io/ioutil"
	"path"
	"strings"
)

type Library struct {
	Path string `toml:"path"`
	Type string
}

func (l *Library) GetManifestPaths() ([]string, error) {
	dir := path.Join(l.Path, "steamapps")

	fileinfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var paths []string

	for _, fileinfo := range fileinfos {
		if strings.HasPrefix(fileinfo.Name(), "appmanifest") && strings.HasSuffix(fileinfo.Name(), "acf") {
			paths = append(paths, path.Join(dir, fileinfo.Name()))
		}
	}

	return paths, nil
}
