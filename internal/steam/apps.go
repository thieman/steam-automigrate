package steam

import (
	"os"
	"sort"
	"strconv"

	"github.com/andygrunwald/vdf"
)

type InstalledApp struct {
	AppID           string
	AppName         string
	InstallDirBase  string
	SizeOnDiskBytes uint64
	ManifestPath    string
	Library         *Library
}

func AppsBySizeForLibrary(library *Library, apps []*InstalledApp) []*InstalledApp {
	var filtered []*InstalledApp
	for _, app := range apps {
		if app.Library == library {
			filtered = append(filtered, app)
		}
	}

	sort.Slice(filtered, func(i, j int) bool { return filtered[i].SizeOnDiskBytes > filtered[j].SizeOnDiskBytes })

	return filtered
}

func TotalSizeOfLibraryBytes(library *Library, apps []*InstalledApp) uint64 {
	var size uint64
	for _, app := range apps {
		if app.Library == library {
			size += app.SizeOnDiskBytes
		}
	}
	return size
}

func GetInstalledApps(config *Config) ([]*InstalledApp, error) {
	var installed []*InstalledApp

	merged := append(config.HDDs, config.SSDs...)
	for i := range merged {
		library := merged[i]
		paths, err := library.GetManifestPaths()
		if err != nil {
			return nil, err
		}

		for _, manifestPath := range paths {
			f, err := os.Open(manifestPath)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, err
			}
			defer f.Close()

			p := vdf.NewParser(f)
			data, err := p.Parse()
			if err != nil {
				return nil, err
			}

			state := data["AppState"].(map[string]interface{})
			size, err := strconv.ParseUint(state["SizeOnDisk"].(string), 10, 64)
			if err != nil {
				return nil, err
			}

			app := InstalledApp{
				AppID:           state["appid"].(string),
				AppName:         state["name"].(string),
				InstallDirBase:  state["installdir"].(string),
				SizeOnDiskBytes: size,
				ManifestPath:    manifestPath,
				Library:         library,
			}

			installed = append(installed, &app)
		}
	}

	return installed, nil
}
