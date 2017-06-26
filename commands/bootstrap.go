package commands

import (
	"fmt"
	"os"
	"time"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/fadion/steer/config"
	"github.com/fadion/steer/server"
)

// Make the initial setup.
func bootstrap(all bool, servers []string) []config.SectionConfig {
	if !checkIfGitRepo() {
		color.Red("Not a git repository. You sure you're inside the correct directory?")
		os.Exit(1)
	}

	srvs, err := readConfig(all, servers)
	if err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}

	return srvs
}

// Connect to each server.
func eachServer(servers []config.SectionConfig, fn func(config.SectionConfig, *server.Connection)) {
	for i, srv := range servers {
		if i > 0 {
			fmt.Println()
		}

		showBadge(srv)

		conn, err := connectToServer(srv)
		if err != nil {
			beep()
			color.Red(err.Error())
			continue
		}

		fn(srv, conn)

		conn.Close()
	}
}

// Parse the local config and transfer sections to servers.
func readConfig(all bool, servers []string) ([]config.SectionConfig, error) {
	localcfg := config.NewLocal()
	cfg, err := localcfg.Read()
	if err != nil {
		return nil, err
	}

	if all {
		return cfg.Sections, nil
	}

	// When no server argument is set, treat the first
	// server in the config as the default.
	if len(servers) == 0 {
		return cfg.Sections[0:1], nil
	}

	var srvs []config.SectionConfig

	for _, c := range cfg.Sections {
		for _, s := range servers {
			if s == c.Section {
				srvs = append(srvs, c)
			}
		}
	}

	return srvs, nil
}

// Connect to server.
func connectToServer(cfg config.SectionConfig) (*server.Connection, error) {
	// Ask interactively for username.
	if cfg.Username == "" {
		cfg.Username = askForUsername(fmt.Sprintf("Enter user for %s: ", cfg.Host))
		fmt.Println()
	}

	// Ask interactively for password.
	if cfg.Password == "" && cfg.Privatekey == "" {
		cfg.Password = askForPassword(fmt.Sprintf("Enter password for %s with user '%s': ", cfg.Host, cfg.Username))
		fmt.Println()
	}

	spin := spinner.New(spinner.CharSets[21], 100*time.Millisecond)
	spin.Prefix = fmt.Sprintf("Connecting to %s ", cfg.Host)
	spin.Start()
	defer spin.Stop()

	serverparams := server.Params{
		Host:       cfg.Host,
		Port:       cfg.Port,
		Username:   cfg.Username,
		Password:   cfg.Password,
		Privatekey: cfg.Privatekey,
		Path:       cfg.Path,
		Maxclients: cfg.Maxclients,
	}

	var driver server.Driver
	var err error
	switch cfg.Scheme {
	case "ftp":
		driver, err = server.ConnectFtp(serverparams)
	case "sftp", "ssh":
		driver, err = server.ConnectSsh(serverparams)
	}

	if err != nil {
		return nil, err
	}

	conn := server.Manage(driver)

	return conn, nil
}
