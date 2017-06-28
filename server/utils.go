package server

import (
	"io"
	"os"
	"golang.org/x/crypto/ssh"
)

var createdDirs []string

// Check if a directory was already created.
func directoryAlreadyCreated(dir string) bool {
	for _, k := range createdDirs {
		if dir == k {
			return true
		}
	}

	return false
}

// Add a directory to the list of created directories if
// it wasn't added before.
func addToCreatedDirs(dir string) {
	for _, k := range createdDirs {
		if dir == k {
			return
		}
	}

	createdDirs = append(createdDirs, dir)
}

// Parse private key.
func parsePrivateKey(file string) (ssh.AuthMethod, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	buffer := make([]byte, 5*1024)
	read := 0
	for {
		read, err = f.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
	}

	key, err := ssh.ParsePrivateKey(buffer[0:read])
	if err != nil {
		return nil, err
	}

	return ssh.PublicKeys(key), nil
}
