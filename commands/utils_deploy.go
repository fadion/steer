package commands

import (
	"fmt"
	"time"
	"github.com/fatih/color"
	"github.com/briandowns/spinner"
	"github.com/fadion/steer/server"
	"os"
)

var progressindicator = ".steer-process"

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

func createProgressIndicator(conn *server.Connection) {
	f, err := os.Create(progressindicator)
	if err != nil {
		return
	}

	defer f.Close()
	defer os.Remove(progressindicator)

	conn.Upload(progressindicator, progressindicator)
}

func deleteProgressIndicator(conn *server.Connection) {
	conn.Delete(progressindicator)
}

func deployInProgress(conn *server.Connection) bool {
	_, err := conn.Read(progressindicator)
	if err != nil {
		return false
	}

	return true
}