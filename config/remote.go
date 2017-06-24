package config

import (
	"strings"
	"os"
	"github.com/fadion/steer/server"
)

// Remote configuration.
type RemoteConfig struct {
	conn *server.Connection
	file string
}

// Initialise a new remote config.
func NewRemote(conn *server.Connection) *RemoteConfig {
	return &RemoteConfig{
		conn: conn,
		file: ".steer-revision",
	}
}

// Read the remote config.
func (c *RemoteConfig) Read() (string, error) {
	rev, err := c.conn.Read(c.file)
	rev = strings.Trim(rev, "\n ")

	return rev, err
}

// Write to the remote config.
func (c *RemoteConfig) Write(rev string) error {
	// Create a local copy of the revision file, so
	// it can be copied to the server.
	f, err := os.Create(c.file)
	if err != nil {
		return err
	}

	defer f.Close()
	defer os.Remove(c.file)

	_, err = f.WriteString(rev)
	if err != nil {
		return err
	}

	f.Sync()

	if err = c.conn.Upload(c.file, c.file); err != nil {
		return err
	}

	return nil
}
