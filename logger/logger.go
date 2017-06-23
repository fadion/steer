package logger

import (
	"time"
	"fmt"
	"os"
	"strings"
	"github.com/fadion/steer/server"
)

// Logger.
type Log struct {
	file string
	conn *server.Connection
}

// Initialise a logger.
func New(conn *server.Connection) *Log {
	return &Log{
		file: ".steer-log",
		conn: conn,
	}
}

func (l *Log) Read() (string, error) {
	contents, err := l.conn.Read(l.file)

	if err != nil {
		return "", err
	}

	return contents, nil
}

// Write the log file in the server.
func (l *Log) Write(nrfiles int, branch string, commit string, message string) error {
	now := time.Now().Format("2006-01-02 03:04:05 -0700")
	contents := now + " | Commit: " + commit + " | Branch: " + branch +
		" | Changed Files: " + fmt.Sprintf("%d", nrfiles)
	if message != "" {
		contents += " | Message: " + message
	}

	remote, err := l.Read()

	// If the file isn't empty, the log line is appended
	// to the existing contents.
	if err != nil {
		remote = contents
	} else {
		remote += "\n" + contents
	}

	f, err := os.Create(l.file)
	if err != nil {
		return err
	}

	_, err = f.WriteString(remote)

	if err != nil {
		return err
	}

	f.Sync()

	defer f.Close()
	defer os.Remove(l.file)

	err = l.conn.Upload(l.file, l.file)
	if err != nil {
		return err
	}

	return nil
}

// Parse a line from the log.
func (l *Log) ParseLine(line string) string {
	parts := strings.Split(line, " | ")
	output := ""

	if len(parts) >= 4 {
		date := strings.TrimSpace(parts[0])
		commit := strings.TrimSpace(strings.Split(parts[1], ":")[1])
		branch := strings.TrimSpace(strings.Split(parts[2], ":")[1])
		files := strings.TrimSpace(strings.Split(parts[3], ":")[1])

		output = fmt.Sprintf("Date: %s\nCommit %s on branch [%s] with %s files changed\n", date, commit, branch, files)

		if len(parts) > 4 {
			output += fmt.Sprintf("Message: %s\n", strings.TrimSpace(strings.Split(parts[4], ":")[1]))
		}
	}

	return output
}

// Write the log file in the server.
func (l *Log) Clear() bool {
	if err := l.conn.Delete(l.file); err != nil {
		return false
	}

	return true
}
