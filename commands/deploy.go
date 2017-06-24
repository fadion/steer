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
	versions := ctx.Bool("versions")
	noversions := ctx.Bool("no-versions")
	message := ctx.String("message")
	spin := spinner.New(spinner.CharSets[21], 100*time.Millisecond)

	if fresh && !askForConfirmation("A fresh deploy will discard all the files. Are you sure you want this?") {
		os.Exit(1)
	}

	eachServer(bootstrap(all, servers), func(cfg config.SectionConfig, conn *server.Connection) {
		var rev string
		var err error
		var versionpath string
		var versionfolder string

		isversioned := versions || cfg.Versions
		if noversions {
			isversioned = false
		}

		// Read the remote revision if it's not a fresh deploy or
		// a versioned one.
		if !fresh && !isversioned {
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

		// Create the versions folder either when a config or cli
		// flag are set.
		if isversioned {
			versionfolder = "/version-" + fmt.Sprintf("%d", time.Now().Unix()) + "/"
			versionpath = strings.Trim(cfg.Vfolder, "/") + versionfolder
			color.Yellow("Starting a versioned deployment on: %s", strings.TrimRight(versionfolder, "/"))
			fmt.Println()
		}

		vcs, err := git.New(cfg.Branch)
		if err != nil {
			color.Red(err.Error())
			return
		}

		if commit == "" {
			commit = vcs.RefHead()
		}

		files := vcs.Changes(rev, commit)
		files = addIncludes(files, cfg.Include)
		files = removeExcludes(files, cfg.Exclude)

		for _, file := range files {
			switch file.Operation {
			case git.ADDED, git.COPIED, git.MODIFIED, git.TYPE:
				spin.Prefix = fmt.Sprintf("Uploading %s ", file.Name)
				spin.Start()

				err := conn.Upload(file.Name, versionpath+file.Name)
				spin.Stop()
				if err != nil {
					color.Red("× %s couldn't be uploaded\n", file.Name)
				} else {
					color.Green("✓ %s was uploaded\n", file.Name)
				}
			case git.DELETED:
				spin.Prefix = fmt.Sprintf("Deleting %s ", file.Name)
				spin.Start()

				err := conn.Delete(versionpath + file.Name)
				spin.Stop()
				if err != nil {
					color.Red("× %s couldn't be deleted\n", file.Name)
				} else {
					color.Green("✓ %s was deleted\n", file.Name)
				}
			}
		}

		if len(files) == 0 {
			color.Yellow("Nothing changed since the last deploy.")
		} else {
			// Write to the log if it's active in the config.
			if cfg.Logger {
				spin.Prefix = "Writing log "
				spin.Start()

				log := logger.New(conn)
				err = log.Write(len(files), cfg.Branch, commit, message)
				spin.Stop()

				if err != nil {
					color.Red("Couldn't write to log file.")
				}
			}

			if isversioned {
				color.Yellow("Project deployed successfully on: %s", strings.TrimRight(versionfolder, "/"))
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
