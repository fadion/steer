package logger

import (
	"testing"
	"github.com/fadion/steer/server"
	"time"
	"fmt"
)

var connection = &server.Connection{Driver: &MockServerDriver{}}
var logcontents = "2017-06-27 05:00:00 +0200 | Commit: abc | Branch: master | Changed Files: 10 | Message: Testing logger"

type MockServerDriver struct{}

func (d *MockServerDriver) MkDir(path string) error               { return nil }
func (d *MockServerDriver) Upload(path, destination string) error { return nil }
func (d *MockServerDriver) Read(path string) (string, error)      { return logcontents, nil }
func (d *MockServerDriver) Delete(path string) error              { return nil }
func (d *MockServerDriver) Close()                                {}

func TestLogRead(t *testing.T) {
	log := New(connection)
	actual, _ := log.Read()

	if actual != logcontents {
		t.Fatalf("Expected %s but got %s", logcontents, actual)
	}
}

func TestLogWrite(t *testing.T) {
	log := New(connection)
	actual, err := log.Write(2, "master", "abc", "Testing")
	expected := fmt.Sprintf("%s | Commit: %s | Branch: %s | Changed Files: %d | Message: %s", time.Now().Format("2006-01-02 03:04:05 -0700"), "abc", "master", 2, "Testing")

	if err != nil {
		t.Fatalf("Log file couldn't be written.")
	}

	if actual != expected {
		t.Fatalf("Expected %s but got %s", expected, actual)
	}
}

func TestLogParseLine(t *testing.T) {
	log := New(connection)
	actual := log.ParseLine(logcontents)
	expected := fmt.Sprintf("Date: %s\nCommit %s on branch [%s] with %s files changed\nMessage: %s\n", "2017-06-27 05:00:00 +0200", "abc", "master", "10", "Testing logger")

	if actual != expected {
		t.Fatalf("Expected %s but got %s", expected, actual)
	}
}

func TestLogClear(t *testing.T) {
	log := New(connection)

	if !log.Clear() {
		t.Fatalf("Log clear didn't work.")
	}
}