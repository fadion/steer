package commands

import (
	"fmt"
	"time"
	"github.com/fatih/color"
	"github.com/briandowns/spinner"
	"github.com/fadion/steer/server"
)

func executeCommands(commands []string, conn *server.Connection) {
	spin := spinner.New(spinner.CharSets[21], 100*time.Millisecond)

	for _, cmd := range commands {
		spin.Prefix = fmt.Sprintf("Executing %s ", cmd)
		spin.Start()
		_, err := conn.Exec(cmd)
		spin.Stop()

		if err != nil {
			color.Red("Command '%s' failed with error: %s", cmd, err.Error())
			continue
		}

		color.Green("'%s' executed successfully.", cmd)
	}
}