package steam

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/minio/minio/pkg/disk"
)

const DEFAULT_CONFIG = `# Be sure to use single quotes for the paths
steam_main_dir = "C:\Program Files (x86)\Steam"
migrate_threshold_seconds = 2592000  # 1 month
`

type Config struct {
	SteamMainDir            string     `toml:"steam_main_dir"`
	SSDs                    []*Library `toml:"SSD"`
	HDDs                    []*Library `toml:"HDD"`
	MigrateThresholdSeconds int        `toml:"migrate_threshold_seconds"`
}

// Return the path of the config file. If it does not already exist,
// will write the default config file and error out. Prompt the user
// to go fix it before trying to run the program again.
func configFilePath() (string, error) {
	dir := filepath.Join(os.Getenv("APPDATA"), "SteamAutomigrate")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return "", err
		}
	}

	path := filepath.Join(dir, "config.toml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = ioutil.WriteFile(path, []byte(DEFAULT_CONFIG), 0644)
		if err != nil {
			return "", err
		}

		return "", errors.New("Wrote default config to " + path + ". Please configure it before running again.")
	}

	return path, nil
}

func validateConfig(config Config) error {
	if len(config.SSDs) == 0 {
		return errors.New("No SSDs defined")
	}

	for _, ssd := range config.SSDs {
		if _, err := os.Stat(ssd.Path); os.IsNotExist(err) {
			return err
		}
	}

	if len(config.HDDs) == 0 {
		return errors.New("No HDDs defined")
	}

	for _, hdd := range config.HDDs {
		if _, err := os.Stat(hdd.Path); os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

func GetConfig() (*Config, error) {
	path, err := configFilePath()
	if err != nil {
		return nil, err
	}

	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}

	err = validateConfig(config)
	if err != nil {
		return nil, err
	}

	merged := append(config.SSDs, config.HDDs...)
	for i := range merged {
		library := merged[i]
		info, err := disk.GetInfo(filepath.VolumeName(library.Path))
		if err != nil {
			return nil, err
		}

		library.VolumeName = filepath.VolumeName(library.Path)
		library.FreeSpaceBytes = info.Free
		library.TotalSpaceBytes = info.Total
		library.PercentFree = 100 * float64(info.Free) / float64(info.Total)
	}

	for i := range config.SSDs {
		config.SSDs[i].Type = "ssd"
	}

	for i := range config.HDDs {
		config.HDDs[i].Type = "hdd"
	}

	return &config, nil
}
