package commands

import (
	"net/http"
	"encoding/json"
	"strings"
	"fmt"
	"io/ioutil"
	"io"
	"bytes"
	"archive/zip"
	"os"
	"compress/gzip"
	"archive/tar"
)

// A GitHub release.
type Release struct {
	Version     string `json:"tag_name"`
	Assets      []Asset `json:"assets"`
	Description string `json:"body"`
}

// A release asset.
type Asset struct {
	Id   int `json:"id"`
	Name string `json:"name"`
}

// Supported systems and their friendly names.
var systems = map[string]string{
	"darwin":  "macos",
	"linux":   "linux",
	"windows": "windows",
}

// Supported architectures.
var architectures = map[string]string{
	"amd64": "64bit",
	"386":   "32bit",
}

const (
	LATEST_RELEASE = "https://api.github.com/repos/fadion/steer/releases/latest"
	ASSET_URL      = "https://api.github.com/repos/fadion/steer/releases/assets/%d"
	FILE_FLAGS     = os.O_RDWR | os.O_CREATE | os.O_TRUNC
)

// Download a release.
func downloadRelease(id int) (*bytes.Reader, error) {
	url := fmt.Sprintf(ASSET_URL, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/octet-stream")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b), nil
}

// Get the latest release.
func getLatestRelease() (*Release, error) {
	client, err := http.Get(LATEST_RELEASE)
	if err != nil {
		return nil, err
	}

	defer client.Body.Close()

	var release Release
	if err = json.NewDecoder(client.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// Parse a semver string.
func parseSemver(ver string) string {
	if len(ver) > 1 && (ver[0] == 'v' || ver[0] == 'V') {
		ver = ver[1:]
	}

	parts := strings.Split(ver, ".")
	if len(parts) != 3 {
		return ""
	}

	return strings.Join(parts, "")
}

// Generate the complete filename for the archive.
func releaseFilename(os, arch string) string {
	filename := fmt.Sprintf("steer-%s-%s", os, arch)
	if os == "linux" {
		filename += ".tar.gz"
	} else {
		filename += ".zip"
	}

	return filename
}

// Extract a tar.gz archive.
func extractTar(source *bytes.Reader, dest string) error {
	gr, err := gzip.NewReader(source)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		destBackup := dest + ".bak"
		if _, err := os.Stat(dest); err == nil {
			os.Rename(dest, destBackup)
		}

		fileCopy, err := os.OpenFile(dest, FILE_FLAGS, hdr.FileInfo().Mode())
		if err != nil {
			os.Rename(destBackup, dest)
			return err
		}

		if _, err = io.Copy(fileCopy, tr); err != nil {
			os.Rename(destBackup, dest)
			return err
		} else {
			os.Remove(destBackup)
		}

		fileCopy.Close()
	}

	return nil
}

// Extract zip archive.
func extractZip(source *bytes.Reader, dest string) error {
	zr, err := zip.NewReader(source, int64(source.Len()))
	if err != nil {
		return err
	}

	for _, f := range zr.File {
		fileCopy, err := os.OpenFile(dest, FILE_FLAGS, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(fileCopy, rc)
		if err != nil {
			return err
		}

		fileCopy.Close()
		rc.Close()
	}

	return nil
}
