package steam

import (
	"github.com/mitchellh/go-ps"
)

func IsRunning() (bool, error) {
	procs, err := ps.Processes()
	if err != nil {
		return false, err
	}

	for _, proc := range procs {
		if proc.Executable() == "steam.exe" {
			return true, nil
		}
	}

	return false, nil
}
