package steam

import (
	"os"
	"strconv"

	"github.com/andygrunwald/vdf"
)

type InstalledApp struct {
	AppID           string
	AppName         string
	InstallDir      string
	SizeOnDiskBytes int64
	Library         *Library
}

func GetInstalledApps(config *Config) ([]InstalledApp, error) {
	var installed []InstalledApp

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

			p := vdf.NewParser(f)
			data, err := p.Parse()
			if err != nil {
				return nil, err
			}

			state := data["AppState"].(map[string]interface{})
			size, err := strconv.ParseInt(state["SizeOnDisk"].(string), 10, 64)
			if err != nil {
				return nil, err
			}

			app := InstalledApp{
				AppID:           state["appid"].(string),
				AppName:         state["name"].(string),
				InstallDir:      state["installdir"].(string),
				SizeOnDiskBytes: size,
				Library:         library,
			}

			installed = append(installed, app)
		}
	}

	return installed, nil
}