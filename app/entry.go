package app

import (
	"os"
	"path/filepath"
	"strings"
)

const REGULAR_MODE = 0100644
const EXECUTABLE_MODE = 0100755
const DIRECTORY_MODE = 040000

type StorableEntry interface {
	GetOid() string
	GetMode() int
}

type Entry struct {
	Name string
	Oid  string
	Mode int32
}

func (e *Entry) GetMode() int {
	if e.isExecAny(e.Mode) {
		return EXECUTABLE_MODE
	}

	return REGULAR_MODE
}

func (e *Entry) GetOid() string {
	return e.Oid
}

func (e *Entry) getParentDirectories() []string {
	parentDirs := []string{}

	dir := filepath.Dir(e.Name)
	parts := strings.Split(dir, string(os.PathSeparator))
	fullString := ""

	for _, part := range parts {
		if part == "." {
			continue
		}
		fullString = filepath.Join(fullString, part)
		parentDirs = append(parentDirs, fullString)
	}

	return parentDirs
}

func (e *Entry) isExecAny(mode int32) bool {
	return mode&0111 != 0
}

func NewEntry(name string, mode int32, oid string) Entry {
	return Entry{
		Name: name,
		Mode: mode,
		Oid:  oid,
	}
}
