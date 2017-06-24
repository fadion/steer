package git

import (
	"os/exec"
	"strings"
	"bytes"
	"bufio"
	"fmt"
)

// Version control.
type Version struct {
	Branch string
}

// Represents a local file.
type File struct {
	Name      string
	Operation string
}

const (
	ADDED    = "added"
	COPIED   = "copied"
	DELETED  = "deleted"
	MODIFIED = "modified"
	RENAMED  = "renamed"
	TYPE     = "type"
	UNKNOWN  = "unknown"
)

var operation = map[rune]string{
	'A': ADDED,
	'C': COPIED,
	'D': DELETED,
	'M': MODIFIED,
	'R': RENAMED,
	'T': TYPE,
	'U': UNKNOWN,
}

// Initialise a Version struct.
func New(branch string) (*Version, error) {
	v := Version{Branch: branch}
	if err := v.Checkout(branch); err != nil {
		return nil, err
	}

	return &v, nil
}

// List files that have changed.
func (v *Version) Changes(remote, local string) []File {
	if remote == "" {
		return v.lsfiles()
	} else {
		return v.diff(remote, local)
	}
}

// Get the HEAD commit hash.
func (v *Version) RefHead() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	out, _ := cmd.Output()

	return strings.Trim(string(out), "\n ")
}

// Checkout to a branch.
func (v *Version) Checkout(branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	out, _ := cmd.Output()
	lines := strings.Split(strings.Trim(string(out), "\n "), "\n")

	if len(lines) == 0 {
		return fmt.Errorf("No output from 'git checkout'. This shouldn't have happened.")
	}

	// Non existing branches or those that have local modifications,
	// produce an error message as in: "error:..."
	if strings.Contains(lines[0], "error:") {
		return fmt.Errorf("Branch '%s' doesn't exist or modifications need to be stashed.", branch)
	}

	return nil
}

// List files without diffing.
func (v *Version) lsfiles() []File {
	cmd := exec.Command("git", "-c", "core.quotepath=false", "ls-files")
	cmdOut := &bytes.Buffer{}
	cmd.Stdout = cmdOut
	cmd.Run()

	scanner := bufio.NewScanner(cmdOut)
	var list []File

	// Each line represents a different file.
	for scanner.Scan() {
		line := scanner.Text()

		list = append(list, File{
			Name:      strings.TrimSpace(line),
			Operation: ADDED,
		})
	}

	return list
}

// List files by running diff.
func (v *Version) diff(remote, local string) []File {
	// Set the local commit to HEAD just for consistence,
	// as it doesn't make much of a difference.
	if local == "" {
		local = "HEAD"
	}

	cmd := exec.Command("git", "-c", "core.quotepath=false", "diff", "--name-status", "--no-renames", remote, local)
	cmdOut := &bytes.Buffer{}
	cmd.Stdout = cmdOut
	cmd.Run()

	scanner := bufio.NewScanner(cmdOut)
	var list []File

	// Each line represents a different file with the type
	// of change. Ie:
	// M file.ext
	// A file2.ext
	for scanner.Scan() {
		line := scanner.Text()
		op, file := line[0], strings.TrimSpace(line[1:])

		operat := operation[rune(op)]
		if operat == "" {
			operat = UNKNOWN
		}

		list = append(list, File{
			Name:      file,
			Operation: operat,
		})
	}

	return list
}
