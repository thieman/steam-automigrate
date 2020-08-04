package steam

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/thieman/steam-automigrate/internal/config"
)

type Library struct {
	Path          string
	Type          string
	ManifestPaths []string
}

func LibraryFromConfig(config *config.LibraryConfig) (*Library, error) {
	manifestPaths, err := getManifestPaths(config.Path)
	if err != nil {
		return nil, err
	}

	return &Library{
		Path:          config.Path,
		Type:          config.Type,
		ManifestPaths: manifestPaths,
	}, nil
}

func getManifestPaths(libraryPath string) ([]string, error) {
	dir := path.Join(libraryPath, "steamapps")

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
