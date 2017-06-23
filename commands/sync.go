package commands

import (
	"time"
	"github.com/urfave/cli"
	"github.com/fatih/color"
	"github.com/briandowns/spinner"
	"github.com/fadion/steer/config"
	"github.com/fadion/steer/server"
	"github.com/fadion/steer/git"
)

// Sync the remote revision.
func Sync(ctx *cli.Context) error {
	servers := ctx.Args()
	all := ctx.Bool("all")
	commit := ctx.String("commit")
	spin := spinner.New(spinner.CharSets[21], 100*time.Millisecond)

	eachServer(bootstrap(all, servers), func(cfg config.SectionConfig, conn *server.Connection) {
		// Read the head commit if a specific commit isn't set.
		if commit == "" {
			vcs, err := git.New(cfg.Branch)
			if err != nil {
				color.Red(err.Error())
				return
			}

			head := vcs.RefHead()
			if head == "" {
				color.Red("Couldn't get the latest commit hash from the git repo.")
				return
			}
			commit = head
		}

		spin.Prefix = "Writing remote revision file "
		spin.Start()

		remoteCfg := config.NewRemote(conn)
		err := remoteCfg.Write(commit)
		spin.Stop()

		if err != nil {
			color.Red("Remote revision file couldn't be written.")
			return
		}

		color.Green("Remote revision file updated.")
	})

	return nil
}
