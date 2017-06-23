package commands

import (
	"fmt"
	"os"
	"bufio"
	"strings"
	"syscall"
	"github.com/fatih/color"
	"golang.org/x/crypto/ssh/terminal"
	"github.com/fadion/steer/config"
)

// Show a nice badge with some useful info.
func showBadge(cfg config.SectionConfig) {
	color.Yellow("+ ---------------------------------+")
	color.Yellow("+ Server %s [%s]", cfg.Host, cfg.Section)
	color.Yellow("+ Branch [%s]", cfg.Branch)
	color.Yellow("+ --------------------------------- +\n\n")
}

// Ask for y/n confirmation.
func askForConfirmation(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	color.New(color.FgWhite).Print(message + " (y/N): ")
	response, _ := reader.ReadString('\n')

	if strings.Trim(strings.ToLower(response), " \n") == "y" {
		return true
	}

	return false
}

// Ask for username interactively.
func askForUsername(message string) string {
	reader := bufio.NewReader(os.Stdin)
	color.New(color.FgWhite).Print(message)
	response, _ := reader.ReadString('\n')

	return strings.Trim(response, " \n")
}

// Ask for password interactively.
func askForPassword(message string) string {
	color.New(color.FgWhite).Printf(message)
	password, _ := terminal.ReadPassword(int(syscall.Stdin))

	return strings.TrimSpace(string(password))
}

// Check if it's inside a git repo.
func checkIfGitRepo() bool {
	if _, err := os.Stat(".git"); err != nil {
		return false
	}

	return true
}

func beep() {
	fmt.Print("\a")
}