package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Ref struct {
	pathName string
}

func (r *Ref) UpdateHead(oid string) error {
	lockfile := NewLockFile(r.headPath())

	err := lockfile.HoldForUpdate()
	if err != nil {
		return errors.New(fmt.Sprintf("Could not acquire lock on file: %s", r.headPath()))
	}

	err = lockfile.Write(oid + "\n")
	if err != nil {
		fmt.Println("error writing lockfile")
		fmt.Println(err)
		os.Exit(1)
	}

	err = lockfile.Commit()
	if err != nil {
		fmt.Println("error committing lock file")
		fmt.Println(err)
		os.Exit(1)
	}

	return nil
}

func (r Ref) ReadHead() string {
	if _, err := os.Stat(r.headPath()); errors.Is(err, os.ErrNotExist) {
		return ""
	}

	contents, err := os.ReadFile(r.headPath())
	if err != nil {
		fmt.Println("error reading head file")
		fmt.Println(err)
		os.Exit(1)
	}

	return strings.TrimSpace(string(contents))
}

func (r *Ref) headPath() string {
	return filepath.Join(r.pathName, "HEAD")
}

func NewRef(pathName string) Ref {
	return Ref{
		pathName: pathName,
	}
}
