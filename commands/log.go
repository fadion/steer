package commands

import (
	"fmt"
	"time"
	"strings"
	"github.com/urfave/cli"
	"github.com/fatih/color"
	"github.com/briandowns/spinner"
	"github.com/fadion/steer/config"
	"github.com/fadion/steer/server"
	"github.com/fadion/steer/logger"
)

// Sync the remote revision.
func Log(ctx *cli.Context) error {
	servers := ctx.Args()
	all := ctx.Bool("all")
	latest := ctx.Int("latest")
	clear := ctx.Bool("clear")
	spin := spinner.New(spinner.CharSets[21], 100*time.Millisecond)

	eachServer(bootstrap(all, servers), func(cfg config.SectionConfig, conn *server.Connection) {
		spin.Prefix = "Reading log "
		spin.Start()

		log := logger.New(conn)
		contents, err := log.Read()
		spin.Stop()

		if err != nil {
			color.Red("No log file found on the server.")
			return
		}

		if clear {
			if log.Clear() {
				color.Yellow("Log file was deleted.")
			} else {
				color.Red("Couldn't delete the log file.")
			}

			return
		}

		if contents == "" {
			color.Red("Log file found, but it's empty.")
			return
		}

		lines := strings.Split(contents, "\n")

		// If the latest flag isn't set, take only one line.
		if latest == 0 {
			latest = 1
		}

		for i, line := range lines {
			if i >= len(lines)-latest {
				color.White(log.ParseLine(line))
				fmt.Println()
			}
		}
	})

	return nil
}
