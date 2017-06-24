package commands

import (
	"fmt"
	"time"
	"os"
	"runtime"
	"bytes"
	"github.com/urfave/cli"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

// Update steer executable to the latest version.
// Implementation based on https://github.com/exercism/cli
func Update(ctx *cli.Context) error {
	spin := spinner.New(spinner.CharSets[21], 100*time.Millisecond)
	spin.Prefix = "Checking the latest version "
	spin.Start()

	release, err := getLatestRelease()
	spin.Stop()
	if err != nil {
		color.Red("Couldn't retrieve version information.")
	}

	if parseSemver(ctx.App.Version) == parseSemver(release.Version) {
		color.Green("You already have the latest version.")
		os.Exit(1)
	}

	// Try and get the absolute path to the excutable.
	dest, err := os.Executable()
	if err != nil {
		color.Red("Couldn't find the path to the steer executable.")
		os.Exit(1)
	}

	curros := systems[runtime.GOOS]
	currarch := architectures[runtime.GOARCH]

	if curros == "" || currarch == "" {
		color.Red("Your OS and Architecture aren't supported.")
		os.Exit(1)
	}

	filename := releaseFilename(curros, currarch)

	// Loop each release file and check if there's one with
	// the correct OS and Architecture. If found, download it.
	var download *bytes.Reader
	for _, ast := range release.Assets {
		if ast.Name == filename {
			spin.Prefix = fmt.Sprintf("Downloading %s ", release.Version)
			spin.Start()
			download, err = downloadRelease(ast.Id)
			spin.Stop()

			if err != nil {
				color.Red("Couldn't download executable.")
				os.Exit(1)
			}

			break
		}
	}

	if download == nil {
		color.Red("Couldn't download executable.")
		os.Exit(1)
	}

	// Linux archives are .tar.gz. Zip for MacOS and Windows.
	if curros == "linux" {
		err = extractTar(download, dest)
	} else {
		err = extractZip(download, dest)
	}

	if err != nil {
		color.Red("Failed while extracting archive.")
		os.Exit(1)
	}

	color.Green("Steer updated successfully to %s", release.Version)
	if release.Description != "" {
		color.Yellow("\nChanges on this release:")
		color.Yellow(release.Description)
	}

	return nil
}
