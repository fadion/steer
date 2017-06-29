package config

import (
	"testing"
	"github.com/fadion/steer/server"
)

var connection = &server.Connection{Driver: &MockServerDriver{}}
var revisioncontents = "mock-revision-contents"

type MockServerDriver struct{}

func (d *MockServerDriver) MkDir(path string) error               { return nil }
func (d *MockServerDriver) Upload(path, destination string) error { return nil }
func (d *MockServerDriver) Read(path string) (string, error)      { return revisioncontents, nil }
func (d *MockServerDriver) Delete(path string) error              { return nil }
func (d *MockServerDriver) Exec(command string) (string, error)   { return "", nil }
func (d *MockServerDriver) Close()                                {}

func TestRemoteConfigRead(t *testing.T) {
	rmt := NewRemote(connection)
	actual, _ := rmt.Read()

	if actual != revisioncontents {
		t.Fatalf("Expected %s but got %s", revisioncontents, actual)
	}
}

func TestRemoteConfigWrite(t *testing.T) {
	rmt := NewRemote(connection)
	err := rmt.Write(revisioncontents)

	if err != nil {
		t.Fatalf("Remote config file couldn't be written.")
	}
}
