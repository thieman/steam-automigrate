package main

import (
	"errors"
	"log"
	"os"

	"github.com/thieman/steam-automigrate/internal/migrate"
	"github.com/thieman/steam-automigrate/internal/steam"
	"github.com/thieman/steam-automigrate/internal/summary"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "steam-automigrate",
		Usage: "Automatically move Steam games between SSDs and HDDs based on last play time",
		Commands: []*cli.Command{
			{
				Name:  "summary",
				Usage: "Summarize currently installed games",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "detailed"},
				},
				Action: func(c *cli.Context) error {
					return summary.DoSummary(c.Bool("detailed"))
				},
			},
			{
				Name:  "migrate",
				Usage: "Automatically migrate games between SSDs and HDDs",
				Action: func(c *cli.Context) error {
					steamRunning, err := steam.IsRunning()
					if err != nil {
						return err
					}

					if steamRunning {
						return errors.New("You must quit Steam before migrating")
					}

					return migrate.DoMigrate()
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
