package commands

import (
	"os"
	"github.com/urfave/cli"
	"github.com/fatih/color"
	"github.com/fadion/steer/config"
)

// Create a template configuration file.
func Init(ctx *cli.Context) error {
	localcfg := config.NewLocal()
	force := ctx.Bool("force")

	if localcfg.Exists() && !force {
		color.Red(".steer file already exists in the project.")

		// Ask to override the config file.
		if !askForConfirmation("Want to override it with the template?") {
			os.Exit(0)
		}
	}

	err := localcfg.Create()
	if err != nil {
		color.Red(err.Error())
	} else {
		color.Green(".steer file created successfully. Edit it with your server details before deploying.")
	}

	return nil
}
