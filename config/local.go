package config

import (
	"os"
	"fmt"
	"github.com/go-ini/ini"
)

// Local configuration.
type LocalConfig struct {
	file     string
	defaults localDefaults
}

// Maps the local ini file to a struct.
type ServerConfig struct {
	Sections []SectionConfig
}

// Maps the server sections of the local init file.
type SectionConfig struct {
	Scheme     string
	Section    string
	Host       string
	Port       int
	Username   string
	Password   string
	Privatekey string
	Path       string
	Branch     string
	Atomic     bool
	Releasedir string
	Include    []string
	Exclude    []string
	Logger     bool
	Maxclients int
}

// Default configuration.
type localDefaults struct {
	scheme     string
	port       int
	path       string
	branch     string
	atomic     bool
	releasedir string
	logger     bool
	maxclients int
}

// Initialise a new local config.
func NewLocal() *LocalConfig {
	return &LocalConfig{
		file: ".steer",
		defaults: localDefaults{
			scheme:     "ftp",
			port:       21,
			path:       "/",
			branch:     "master",
			atomic:     false,
			releasedir: "releases",
			logger:     false,
			maxclients: 3,
		},
	}
}

// Check if the config file exists.
func (c *LocalConfig) Exists() bool {
	if _, err := os.Stat(c.file); err == nil {
		return true
	} else {
		return false
	}
}

// Create a config file from a template.
func (c *LocalConfig) Create() error {
	f, err := os.Create(c.file)
	if err != nil {
		return fmt.Errorf(".steer file couldn't be created. Please give it another try.")
	}

	defer f.Close()

	_, err = f.WriteString(`[production]
scheme = ftp
host = ftp.example.com
port = 21
username = user
password = pass
path = /
branch = master`)

	if err != nil {
		return fmt.Errorf(".steer file created, but failed to write the default template.")
	}

	f.Sync()

	return nil
}

// Read and parse the config file into a struct.
func (c *LocalConfig) Read() (*ServerConfig, error) {
	if !c.Exists() {
		return nil, fmt.Errorf(".steer file doesn't exist. Create one by running: steer init.")
	}

	// Do a case insensitive load for sections and keys.
	// In contrast, Load() is case sensitive.
	cfg, err := ini.Load(".steer")
	if err != nil {
		return nil, fmt.Errorf(".steer file doesn't exist or it's not correctly formatted.")
	}

	// ini adds a "DEFAULT" section to the section
	// array, so the actual sections are from the second
	// one and on
	sections := cfg.SectionStrings()[1:]
	if len(sections) == 0 {
		return nil, fmt.Errorf("No server found in .steer file. Check if it's correctly formatted.")
	}

	var out []SectionConfig
	for _, section := range sections {
		sec, _ := cfg.GetSection(section)
		out = append(out, SectionConfig{
			Section:    section,
			Scheme:     sec.Key("scheme").In(c.defaults.scheme, []string{"ftp", "sftp", "ssh"}),
			Host:       sec.Key("host").MustString(""),
			Port:       sec.Key("port").MustInt(c.defaults.port),
			Username:   sec.Key("username").MustString(""),
			Password:   sec.Key("password").MustString(""),
			Privatekey: sec.Key("privatekey").MustString(""),
			Path:       sec.Key("path").MustString(c.defaults.path),
			Branch:     sec.Key("branch").MustString(c.defaults.branch),
			Atomic:     sec.Key("atomic").MustBool(c.defaults.atomic),
			Releasedir: sec.Key("releasedir").MustString(c.defaults.releasedir),
			Include:    sec.Key("include").Strings(","),
			Exclude:    sec.Key("exclude").Strings(","),
			Logger:     sec.Key("logger").MustBool(c.defaults.logger),
			Maxclients: sec.Key("maxclients").MustInt(c.defaults.maxclients),
		})
	}

	return &ServerConfig{Sections: out}, nil
}
