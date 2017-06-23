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
	"github.com/fadion/steer/git"
)

// Preview file changes.
func Preview(ctx *cli.Context) error {
	servers := ctx.Args()
	all := ctx.Bool("all")
	commit := ctx.String("commit")
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

		if commit == "" {
			commit = vcs.RefHead()
		}

		files := vcs.Changes(rev, commit)
		files = addIncludes(files, cfg.Include)
		files = removeExcludes(files, cfg.Exclude)

		for _, file := range files {
			switch file.Operation {
			case git.ADDED, git.COPIED:
				color.Set(color.FgGreen)
			case git.MODIFIED, git.RENAMED, git.TYPE:
				color.Set(color.FgBlue)
			case git.DELETED:
				color.Set(color.FgRed)
			case git.UNKNOWN:
				color.Set(color.FgBlack)
			default:
				color.Set(color.FgWhite)
			}

			fmt.Printf("[%s] %s\n", strings.ToUpper(file.Operation[0:3]), file.Name)

			color.Unset()
		}

		if len(files) > 0 {
			fmt.Println()
		}

		color.White("Remote commit: %s", rev)
		color.White("Local HEAD: %s", vcs.RefHead())
		color.Green("%d file(s) changed", len(files))

		if len(files) == 0 {
			color.Yellow("\nNothing changed since the last deploy.")
		}
	})

	return nil
}
