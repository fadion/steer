package commands

import (
	"fmt"
	"time"
	"github.com/urfave/cli"
	"github.com/fatih/color"
	"github.com/briandowns/spinner"
	"github.com/fadion/steer/config"
	"github.com/fadion/steer/server"
	"github.com/fadion/steer/git"
)

// Preview file changes.
func Status(ctx *cli.Context) error {
	servers := ctx.Args()
	all := ctx.Bool("all")
	spin := spinner.New(spinner.CharSets[21], 100*time.Millisecond)

	eachServer(bootstrap(all, servers), func(cfg config.SectionConfig, conn *server.Connection) {
		var rev string
		var err error

		spin.Prefix = "Reading remote revision file "
		spin.Start()
		remotecfg := config.NewRemote(conn)
		rev, err = remotecfg.Read()
		spin.Stop()

		if err != nil || rev == "" {
			color.Red("Remote revision file not found.")
			fmt.Println()
		}

		vcs, err := git.New(cfg.Branch)
		if err != nil {
			color.Red(err.Error())
			return
		}

		files := vcs.Changes(rev, "")
		files = addIncludes(files, cfg.Include)
		files = removeExcludes(files, cfg.Exclude)

		color.White("Remote commit: %s", rev)
		color.White("Local HEAD: %s", vcs.RefHead())
		color.Green("%d file(s) changed since last commit", len(files))

		if deployInProgress(conn) {
			color.Red("A deployment is already in progress.")
		}
	})

	return nil
}
