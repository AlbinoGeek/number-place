// +build ignore

// This program generates git.go. It can be invoked by running `go generate`
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"
)

var tmpl = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
// This file was generated at: {{.Timestamp}}
// This file is in sync with: {{.CommitSHA}}
package main

// PROGRAM is the human readable product name
const PROGRAM = "{{.Program}}"

// VERSION is the human readable product version
const VERSION = "{{.Version}}"
`))

func getCommitSHA() string {
	str, err := ioutil.ReadFile(".git/FETCH_HEAD")
	if err != nil {
		return "unknown-commit"
	}

	return strings.Split(string(str), "\t")[0]
}

func getPackageName() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Base(filepath.Dir(b))
}

func getVersion() string {
	// If the `git` tool is available, we can get our version number reliably
	cmd := exec.Command("git", "describe", "--match", "v*", "--always", "--tags")

	// Should return a string like "v0.3.65-g3a407d2"
	// "v" tag "." commits_since_tag "-g" commit_sha
	out, err := cmd.CombinedOutput()

	// Otherwise, return the commit SHA of HEAD
	if err != nil {
		return fmt.Sprintf("git-%s", getCommitSHA())
	}

	return strings.Trim(string(out), " \t\r\n")
}

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "PROGRAM" {
			fmt.Println(getPackageName())
		}
		if os.Args[1] == "VERSION" {
			fmt.Println(getVersion())
		}

		return
	}

	f, err := os.Create("version.go")
	if err != nil {
		panic(err)
	}

	tmpl.Execute(f, struct {
		CommitSHA string
		Program   string
		Timestamp string
		Version   string
	}{
		getCommitSHA(),
		getPackageName(),
		time.Now().String(),
		getVersion(),
	})

	if err := f.Close(); err != nil {
		panic(err)
	}
}
