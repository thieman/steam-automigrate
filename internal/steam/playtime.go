package steam

import (
	"os"
	"path"
	"strconv"
	"time"

	"github.com/andygrunwald/vdf"
)

// Return a map from Steam App ID to the most recent time
// each game was played. If multiple Steam users are defined
// on the system, the most recent time across all users is returned.
func GetPlaytimes(config *Config) (map[string]time.Time, error) {
	ids, err := steamIds(config)
	if err != nil {
		return nil, err
	}

	result := make(map[string]time.Time)

	for _, id := range ids {
		path := path.Join(config.SteamMainDir, "userdata", id, "config", "localconfig.vdf")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}

		parser := vdf.NewParser(f)
		data, err := parser.Parse()
		if err != nil {
			return nil, err
		}

		appMap := data["UserLocalConfigStore"].(map[string]interface{})["Software"].(map[string]interface{})["Valve"].(map[string]interface{})["Steam"].(map[string]interface{})["apps"].(map[string]interface{})

		for appId, values := range appMap {
			i, err := strconv.ParseInt(values.(map[string]interface{})["LastPlayed"].(string), 10, 64)
			if err != nil {
				return nil, err
			}

			lastPlayedTime := time.Unix(i, 0)
			current, found := result[appId]

			if !found || current.Before(lastPlayedTime) {
				result[appId] = lastPlayedTime
			}
		}
	}

	return result, nil
}
