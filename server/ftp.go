package server

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"path/filepath"
	"github.com/secsy/goftp"
)

type ftp struct {
	conn     *goftp.Client
	basepath string
}

// Connect to the FTP server.
func ConnectFtp(cfg Params) (*ftp, error) {
	conn, err := goftp.DialConfig(goftp.Config{
		User: cfg.Username,
		Password: cfg.Password,
		ConnectionsPerHost: cfg.Maxclients,
	}, fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("Couldn't connect to FTP server. System response: %s\n", err.Error())
	}

	return &ftp{conn: conn, basepath: cfg.Path}, nil
}

// Create directory.
func (f *ftp) MkDir(path string) error {
	// Try creating all the directories in the path.
	if err := f.createDirs(path); err != nil {
		return err
	}

	return nil
}

// Upload a file.
func (f *ftp) Upload(path, destination string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("%s couldn't be opened. Make sure it exists.\n", path)
	}

	defer file.Close()

	if err = f.MkDir(filepath.Dir(destination)); err != nil {
		return err
	}

	if err = f.conn.Store(f.makePath(destination), file); err != nil {
		return fmt.Errorf("%s couldn't be uploaded.\n", path)
	}

	return nil
}

// Read a file's contents.
func (f *ftp) Read(path string) (string, error) {
	contents := &bytes.Buffer{}
	err := f.conn.Retrieve(f.makePath(path), contents)

	if err != nil {
		return "", fmt.Errorf("File %s couldn't be read from server.", path)
	}

	return contents.String(), nil
}

// Delete a file.
func (f *ftp) Delete(path string) error {
	if err := f.conn.Delete(f.makePath(path)); err != nil {
		return err
	}

	return nil
}

// Close connection.
func (f *ftp) Close() {
	f.conn.Close()
}

// Append the basepath to path.
func (f *ftp) makePath(path string) string {
	return strings.TrimRight(f.basepath, "/") + "/" + strings.TrimLeft(path, "/")
}

// Create directories for a given path.
func (f *ftp) createDirs(dir string) error {
	components := strings.Split(dir, string(os.PathSeparator))
	currentDir := strings.TrimRight(f.basepath, "/")

	for _, c := range components {
		if c == "." || c == ".." {
			continue
		}

		currentDir += "/" + c

		_, err := f.conn.Stat(currentDir)
		if err != nil {
			if _, err := f.conn.Mkdir(currentDir); err != nil {
				return err
			}
		}
	}

	return nil
}