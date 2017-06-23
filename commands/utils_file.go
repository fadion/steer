package commands

import (
	"fmt"
	"os"
	"io/ioutil"
	"strings"
	"github.com/fadion/steer/git"
)

// Add includes to the list of files.
func addIncludes(current []git.File, files []string) []git.File {
	if len(files) == 0 {
		return current
	}

	for _, file := range expandFiles(files) {
		current = append(current, git.File{
			Name:      file,
			Operation: git.ADDED,
		})
	}

	return current
}

// Remove excludes from the list of files.
func removeExcludes(current []git.File, files []string) []git.File {
	// Don't deploy steer configuration file.
	files = append(files, ".steer")
	output := []git.File{}

	for _, c := range current {
		remove := false
		for _, f := range expandFiles(files) {
			if strings.Trim(f, "/") == strings.Trim(c.Name, "/") {
				remove = true
				break
			}
		}

		if !remove {
			output = append(output, c)
		}
	}

	return output
}

// Read files and directories.
func expandFiles(files []string) []string {
	var output []string

	for _, k := range files {
		file, err := os.Stat(k)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		if file.IsDir() {
			dirfiles, err := ioutil.ReadDir(k)
			if err != nil {
				continue
			}

			for _, dirfile := range dirfiles {
				if !dirfile.IsDir() {
					output = append(output, fmt.Sprintf("%s%s%s", k, string(os.PathSeparator), dirfile.Name()))
				}
			}
		} else {
			output = append(output, k)
		}
	}

	return output
}
