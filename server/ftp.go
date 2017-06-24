package server

import (
	"io"
	"bytes"
	"fmt"
	"os"
	"strings"
	"path/filepath"
	"github.com/dutchcoders/goftp"
)

type ftp struct {
	conn     *goftp.FTP
	basepath string
}

var createdDirs []string

// Connect to the FTP server.
func ConnectFtp(cfg Params) (*ftp, error) {
	conn, err := goftp.Connect(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("Couldn't connect to FTP server. System response: %s\n", err.Error())
	}

	if err = conn.Login(cfg.Username, cfg.Password); err != nil {
		return nil, fmt.Errorf("Couldn't login to FTP server. Double check the username and password.\n")
	}

	if err = conn.Cwd(cfg.Path); err != nil {
		return nil, err
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

	if err = f.conn.Stor(filepath.Base(destination), file); err != nil {
		return fmt.Errorf("%s couldn't be uploaded.\n", path)
	}

	return nil
}

// Read a file's contents.
func (f *ftp) Read(path string) (string, error) {
	f.conn.Cwd(f.basepath)

	var contents string
	_, err := f.conn.Retr(path, func(r io.Reader) error {
		var buff bytes.Buffer
		buff.ReadFrom(r)
		contents = buff.String()

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("File %s couldn't be read from server.", path)
	}

	return contents, nil
}

// Delete a file.
func (f *ftp) Delete(path string) error {
	if err := f.conn.Dele(path); err != nil {
		return err
	}

	return nil
}

// Close connection.
func (f *ftp) Close() {
	f.conn.Quit()
}

// Create directories for a given path.
func (f *ftp) createDirs(dir string) error {
	components := strings.Split(dir, string(os.PathSeparator))
	f.conn.Cwd(f.basepath)
	currentDir := strings.TrimRight(f.basepath, "/")

	for _, c := range components {
		if c == "." || c == ".." {
			continue
		}

		currentDir += "/" + c

		// Don't create a directory if it's already been created or
		// it exists.
		if !f.directoryAlreadyCreated(currentDir) && !f.directoryExists(currentDir) {
			if err := f.conn.Mkd(c); err != nil {
				return err
			}
		}

		// Save directories that have been already created or checked
		// if they exists. It lowers the expense of the operation.
		f.addToCreatedDirs(currentDir)

		if err := f.conn.Cwd(c); err != nil {
			return err
		}
	}

	return nil
}

// Check if a directory exists in the server.
func (f *ftp) directoryExists(fullpath string) bool {
	idx := len(fullpath) - 1
	for idx >= 0 && fullpath[idx] != '/' {
		idx--
	}

	// Split the fullpath to a path + basedir. This works
	// very similar to filepath.Dir(), but instead it doesn't
	// replace slashes with os.PathSeparator.
	path := fullpath[:idx]
	dirname := fullpath[idx+1:]

	// Ignore if there's no path or basedir. It means it's a
	// single path that can't be checked.
	if path == "" || dirname == "" {
		return false
	}

	list, err := f.conn.List(path)

	if err != nil {
		return false
	}

	for _, k := range list {
		// FTP returns a list of files and directories in the format:
		// type=dir;...; dirname
		// type=file: ...; file.ext
		parts := strings.Split(k, ";")
		if len(parts) > 2 {
			typepart := strings.Split(parts[0], "=")
			filename := strings.TrimSpace(parts[len(parts)-1])

			if len(typepart) == 2 {
				filetype := strings.TrimSpace(typepart[1])
				if filetype == "dir" && filename == dirname {
					return true
				}
			}
		}
	}

	return false
}

// Check if a directory was already created.
func (f *ftp) directoryAlreadyCreated(dir string) bool {
	for _, k := range createdDirs {
		if dir == k {
			return true
		}
	}

	return false
}

// Add a directory to the list of created directories if
// it wasn't added before.
func (f *ftp) addToCreatedDirs(dir string) {
	for _, k := range createdDirs {
		if dir == k {
			return
		}
	}

	createdDirs = append(createdDirs, dir)
}
