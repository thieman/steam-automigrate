package summary

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/minio/minio/pkg/disk"
	"github.com/thieman/steam-automigrate/internal/steam"
)

type summarized struct {
	Library         *steam.Library
	VolumeName      string
	FreeSpaceBytes  uint64
	TotalSpaceBytes uint64
	PercentFree     float64
}

func compileSummaryStats(config *steam.Config) (map[string][]summarized, error) {
	stats := make(map[string][]summarized)

	stats["ssd"] = make([]summarized, 0)
	stats["hdd"] = make([]summarized, 0)

	merged := append(config.SSDs, config.HDDs...)
	for i := range merged {
		library := merged[i]

		info, err := disk.GetInfo(filepath.VolumeName(library.Path))
		if err != nil {
			return nil, err
		}

		this := summarized{
			Library:         library,
			VolumeName:      filepath.VolumeName(library.Path),
			FreeSpaceBytes:  info.Free,
			TotalSpaceBytes: info.Total,
			PercentFree:     100 * float64(info.Free) / float64(info.Total),
		}

		if library.Type == "ssd" {
			stats["ssd"] = append(stats["ssd"], this)
		} else {
			stats["hdd"] = append(stats["hdd"], this)
		}
	}

	return stats, nil
}

func printTypeSummary(libraryType string, summaries []summarized) {
	var total, free uint64

	for _, summary := range summaries {
		total += summary.TotalSpaceBytes
		free += summary.FreeSpaceBytes
	}

	gbFree := float64(free) / (1024 * 1024 * 1024)
	percentFree := 100 * float64(free) / float64(total)

	bold := color.New(color.Bold)
	bold.Printf(strings.ToUpper(libraryType)+"s: %2.0f%% Free (%2.0f GB)\n\n", percentFree, gbFree)
}

func printLibrarySummary(summarized summarized) {
	gbFree := float64(summarized.FreeSpaceBytes) / (1024 * 1024 * 1024)
	fmt.Printf("    "+summarized.Library.Path+": %2.0f%% Free (%2.0f GB)\n", summarized.PercentFree, gbFree)
}

func printDetails(library *steam.Library, installed *[]steam.InstalledApp, detailed bool) {
	var filtered []steam.InstalledApp
	for _, app := range *installed {
		if app.Library == library {
			filtered = append(filtered, app)
		}
	}

	sort.Slice(filtered, func(i, j int) bool { return filtered[i].SizeOnDiskBytes > filtered[j].SizeOnDiskBytes })

	var shown uint

	for _, app := range filtered {
		if !detailed && shown >= 5 {
			break
		}

		gb := float64(app.SizeOnDiskBytes) / (1024 * 1024 * 1024)
		fmt.Printf("        "+app.AppName+" (%2.1f GB)\n", gb)

		shown++
	}

	fmt.Println()
}

func DoSummary(detailed bool) error {
	config, err := steam.GetConfig()
	if err != nil {
		return err
	}

	installed, err := steam.GetInstalledApps(config)
	if err != nil {
		return err
	}

	stats, err := compileSummaryStats(config)
	if err != nil {
		return err
	}

	for _, libraryType := range []string{"ssd", "hdd"} {
		if len(stats[libraryType]) == 0 {
			continue
		}

		printTypeSummary(libraryType, stats[libraryType])

		for _, summarized := range stats[libraryType] {
			printLibrarySummary(summarized)
			printDetails(summarized.Library, &installed, detailed)
		}
	}

	return nil
}
