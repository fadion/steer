package server

import (
	"fmt"
	"strings"
	"path"
	"os"
	"path/filepath"
	"bufio"
	"io"
	srv "github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type sftp struct {
	conn     *srv.Client
	basepath string
}

// Connect to the SFTP server.
func ConnectSsh(cfg Params) (*sftp, error) {
	var method []ssh.AuthMethod

	if cfg.Privatekey != "" {
		if mth, err := parsePrivateKey(cfg.Privatekey); err == nil {
			method = append(method, mth)
		}
	} else {
		method = append(method, ssh.Password(cfg.Password))
	}

	conncfg := &ssh.ClientConfig{
		User: cfg.Username,
		Auth: method,
		Config: ssh.Config{
			Ciphers: []string{"aes128-cbc", "hmac-sha1"},
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), conncfg)
	if err != nil {
		return nil, fmt.Errorf("Couldn't connect to SFTP server. System response: %s\n", err.Error())
	}

	client, err := srv.NewClient(conn)
	if err != nil {
		return nil, fmt.Errorf("Couldn't connect to SFTP server. System response: %s\n", err.Error())
	}

	return &sftp{conn: client, basepath: cfg.Path}, nil
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

// Create directory.
func (s *sftp) MkDir(path string) error {
	if err := s.createDirs(s.makePath(path)); err != nil {
		return err
	}

	return nil
}

// Upload a file.
func (s *sftp) Upload(path, destination string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("%s couldn't be opened. Make sure it exists.\n", path)
	}

	defer file.Close()

	if err = s.MkDir(filepath.Dir(destination)); err != nil {
		return err
	}

	f, err := s.conn.Create(s.makePath(destination))
	if err != nil {
		return fmt.Errorf("%s couldn't be uploaded.\n", path)
	}

	defer f.Close()

	_, err = f.ReadFrom(bufio.NewReader(file))
	if err != nil {
		s.Delete(s.makePath(destination))
		return fmt.Errorf("%s couldn't be uploaded.\n", path)
	}

	return nil
}

// Read a file's contents.
func (s *sftp) Read(path string) (string, error) {
	file, err := s.conn.Open(s.makePath(path))
	if err != nil {
		return "", fmt.Errorf("File %s couldn't be read from server.", path)
	}

	defer file.Close()

	buffer := make([]byte, 512)
	read := 0
	for {
		read, err = file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return "", fmt.Errorf("File %s couldn't be read from server.", path)
			}
		}
	}

	return string(buffer[:read]), nil
}

// Delete a file.
func (s *sftp) Delete(path string) error {
	if err := s.conn.Remove(path); err != nil {
		return err
	}

	return nil
}

// Close connection.
func (s *sftp) Close() {
	s.conn.Close()
}

// Append the basepath to path.
func (s *sftp) makePath(path string) string {
	return strings.TrimRight(s.basepath, "/") + "/" + strings.TrimLeft(path, "/")
}

// Create directories for a given path.
// Taken from the example at:
// http://godoc.org/github.com/pkg/sftp#Client.Mkdir
func (s *sftp) createDirs(dir string) error {
	var parents string
	var err error

	ssh_fx_failure := uint32(4)

	for _, name := range strings.Split(dir, string(os.PathSeparator)) {
		parents = path.Join(parents, name)
		err = s.conn.Mkdir(parents)
		if status, ok := err.(*srv.StatusError); ok {
			if status.Code == ssh_fx_failure {
				var fi os.FileInfo
				fi, err = s.conn.Stat(parents)
				if err == nil {
					if !fi.IsDir() {
						return nil
					}
				}
			}
		}
		if err != nil {
			break
		}
	}

	return err
}
