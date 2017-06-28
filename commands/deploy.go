package commands

import (
	"fmt"
	"strings"
	"time"
	"os"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/urfave/cli"
	"github.com/fadion/steer/logger"
	"github.com/fadion/steer/config"
	"github.com/fadion/steer/server"
	"github.com/fadion/steer/git"
)

// Deploy file changes to the server.
func Deploy(ctx *cli.Context) error {
	servers := ctx.Args()
	fresh := ctx.Bool("fresh")
	all := ctx.Bool("all")
	commit := ctx.String("commit")
	message := ctx.String("message")
	spin := spinner.New(spinner.CharSets[21], 100*time.Millisecond)

	if fresh && !askForConfirmation("A fresh deploy will discard all the files. Are you sure you want this?") {
		os.Exit(1)
	}

	eachServer(bootstrap(all, servers), func(cfg config.SectionConfig, conn *server.Connection) {
		var rev string
		var err error
		var atomicpath string
		var releasefolder string

		isatomic := cfg.Atomic

		// Read the remote revision if it's not a fresh deploy or
		// an atomic one.
		if !fresh && !isatomic {
			spin.Prefix = "Reading remote revision file "
			spin.Start()
			remotecfg := config.NewRemote(conn)
			rev, err = remotecfg.Read()
			spin.Stop()

			if err != nil || rev == "" {
				color.Red("Remote revision file not found. Considering it a first-time deployment.")
				fmt.Println()
			}
		}

		// Create the atomic folder when the config option is set.
		if isatomic {
			releasefolder = "/" + fmt.Sprintf("%d", time.Now().Unix()) + "/"
			atomicpath = strings.Trim(cfg.Releasedir, "/") + releasefolder
			color.Yellow("Starting an atomic deployment on: %s", strings.TrimRight(releasefolder, "/"))
			fmt.Println()
		}

		vcs, err := git.New(cfg.Branch)
		if err != nil {
			color.Red(err.Error())
			return
		}

		files := vcs.Changes(rev, commit)
		files = addIncludes(files, cfg.Include)
		files = removeExcludes(files, cfg.Exclude)

		// Write a temp file to indicate deployment progress.
		go func() { createProgressIndicator(conn) }()
		defer deleteProgressIndicator(conn)

		// Predeploy commands.
		if len(cfg.Predeploy) > 0 {
			color.Yellow("Executing pre deployment commands:")
			executeCommands(cfg.Predeploy, conn)
			if len(files) > 0 {
				fmt.Println()
			}
		}

		// A channel with a buffer the size of the maximum
		// number of clients read from the config.
		sem := make(chan bool, cfg.Maxclients)

		spin.Prefix = "Starting deploy "
		spin.Start()

		for _, file := range files {
			sem <- true

			go func(file git.File) {
				switch file.Operation {
				case git.ADDED, git.COPIED, git.MODIFIED, git.TYPE:
					err := conn.Upload(file.Name, atomicpath+file.Name)
					spin.Stop()
					if err != nil {
						color.Red("× %s couldn't be uploaded", file.Name)
					} else {
						color.Green("✓ %s was uploaded", file.Name)
					}
				case git.DELETED:
					err := conn.Delete(atomicpath + file.Name)
					spin.Stop()
					if err != nil {
						color.Red("× %s couldn't be deleted", file.Name)
					} else {
						color.Green("✓ %s was deleted", file.Name)
					}
				}

				<-sem
			}(file)
		}

		// Wait for the last goroutines (buffer size) to finish.
		for i := 0; i < cap(sem); i++ {
			sem <- true
		}

		spin.Stop()

		// Postdeploy commands.
		if len(cfg.Postdeploy) > 0 {
			fmt.Println()
			color.Yellow("Executing post deployment commands:")
			executeCommands(cfg.Postdeploy, conn)
		}

		if len(files) == 0 {
			color.Yellow("\nNothing changed since the last deploy.")
		} else {
			// Write to the log if it's active in the config.
			if cfg.Logger {
				spin.Prefix = "Writing log "
				spin.Start()

				log := logger.New(conn)
				_, err = log.Write(len(files), cfg.Branch, commit, message)
				spin.Stop()

				if err != nil {
					color.Red("Couldn't write to log file.")
				}
			}

			if isatomic {
				color.Yellow("\nProject deployed successfully on: %s", strings.TrimRight(releasefolder, "/"))
			} else {
				spin.Prefix = "Writing remote revision file "
				spin.Start()

				remoteCfg := config.NewRemote(conn)
				err := remoteCfg.Write(commit)
				spin.Stop()

				if err != nil {
					color.Red("\nProject deployed, but remote revision couldn't be written. Try running 'steer sync'.")
				} else {
					color.Yellow("\nProject deployed successfully.")
				}
			}
		}
	})

	return nil
}
