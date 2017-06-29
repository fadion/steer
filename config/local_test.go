package config

import (
	"testing"
	"os"
	"io/ioutil"
	"reflect"
)

func TestLocalConfigDoesntExist(t *testing.T) {
	cfg := NewLocal()
	if cfg.Exists() {
		t.Fatalf("Config file shouldn't exist.")
	}
}

func TestLocalConfigCreate(t *testing.T) {
	cfg := NewLocal()
	cfg.file = "./.steer"

	err := cfg.Create()
	defer os.Remove(cfg.file)
	if err != nil {
		t.Fatalf("Config file couldn't be created.")
	}

	actual, err := ioutil.ReadFile(cfg.file)
	if err != nil {
		t.Fatalf("Config file was created, but couldn't be read.")
	}

	expected := `[production]
scheme = ftp
host = ftp.example.com
port = 21
username = user
password = pass
path = /
branch = master`

	if string(actual) != expected {
		t.Fatalf("Config file created but not with the correct contents.")
	}
}

func TestLocalConfigRead(t *testing.T) {
	cfg := NewLocal()
	cfg.file = "./.steer"

	err := cfg.Create()
	defer os.Remove(cfg.file)
	if err != nil {
		t.Fatalf("Config file couldn't be created.")
	}

	contents, err := cfg.Read()
	if err != nil {
		t.Fatalf("Config file couldn't be read.")
	}

	actual := contents.Sections[0]

	expected := SectionConfig{
		Scheme:     "ftp",
		Section:    "production",
		Host:       "ftp.example.com",
		Port:       21,
		Username:   "user",
		Password:   "pass",
		Privatekey: "",
		Path:       "/",
		Branch:     "master",
		Atomic:     false,
		Reldir:     "releases",
		Currdir:    "current",
		Include:    []string{},
		Exclude:    []string{},
		Logger:     false,
		Maxclients: 3,
		Predeploy:  []string{},
		Postdeploy: []string{},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Config file contents not as expected.")
	}
}
