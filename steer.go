package main

import (
	"os"
	"fmt"
	"github.com/urfave/cli"
	"github.com/fadion/steer/commands"
)

func main() {
	app := cli.NewApp()
	app.Name = "steer"
	app.Usage = "deploy with git via ftp and ssh"
	app.Authors = []cli.Author{{
		Name:  "Fadion Dashi",
		Email: "jonidashi@gmail.com",
	}}
	app.Version = "0.3.1"

	app.Commands = []cli.Command{
		{
			Name:  "init",
			Usage: "Create a template .steer file",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force",
					Usage: "Override existing .steer file if it exists",
				},
			},
			Action: commands.Init,
		},
		{
			Name:  "preview",
			Usage: "Preview changed files",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "all",
					Usage: "Preview all servers",
				},
				cli.StringFlag{
					Name:  "commit, c",
					Usage: "Changes from `COMMIT`",
				},
			},
			Action: commands.Preview,
		},
		{
			Name:  "deploy",
			Usage: "Deploy to the server",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "fresh",
					Usage: "Upload every file as it is a fresh deploy",
				},
				cli.BoolFlag{
					Name:  "all",
					Usage: "Deploy to all servers",
				},
				cli.BoolFlag{
					Name:  "versions",
					Usage: "Enable versions",
				},
				cli.BoolFlag{
					Name:  "no-versions",
					Usage: "Disable versions",
				},
				cli.StringFlag{
					Name:  "commit, c",
					Usage: "Changes from `COMMIT`",
				},
				cli.StringFlag{
					Name:  "message, m",
					Usage: "`MESSAGE` for the log",
				},
			},
			Action: commands.Deploy,
		},
		{
			Name:  "sync",
			Usage: "Sync remote revision to current head",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "commit, c",
					Usage: "Sync to `COMMIT` hash",
				},
				cli.BoolFlag{
					Name:  "all",
					Usage: "Sync all servers",
				},
			},
			Action: commands.Sync,
		},
		{
			Name:  "log",
			Usage: "Get information from the remote log",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "latest",
					Usage: "Get the latest `NUMBER` of lines",
				},
				cli.BoolFlag{
					Name:  "all",
					Usage: "Get log info from all servers",
				},
				cli.BoolFlag{
					Name:  "clear",
					Usage: "Clear the log",
				},
			},
			Action: commands.Log,
		},
		{
			Name:  "update",
			Usage: "Update steer to the latest version",
			Action: commands.Update,
		},
	}

	app.CommandNotFound = func(ctx *cli.Context, command string) {
		fmt.Fprintf(ctx.App.Writer, "Command %q doesn't exist.\n", command)
	}

	app.Run(os.Args)
}
