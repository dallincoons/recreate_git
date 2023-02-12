package app

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type Workspace struct {
	rootPath string
}

var IGNORES = []string{".git", ".idea"}

func (ws *Workspace) GetFiles(rootPath string) ([]string, error) {

	fileList := []string{}

	fileInfo, err := os.Stat(rootPath)
	if err != nil {
		return nil, err
	}
	if !fileInfo.IsDir() {
		return []string{rootPath}, nil
	}

	err = filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		for _, ignore := range IGNORES {
			if info.IsDir() && info.Name() == ignore {
				return filepath.SkipDir
			}
		}

		if !info.IsDir() {
			fileList = append(fileList, path)
		}

		return nil
	})
	if err != nil {
		fmt.Println("error recursively walking filepath")
		fmt.Println(err)
		os.Exit(1)
	}

	return fileList, nil
}

func (ws *Workspace) ListDirectories(dirname string) map[string]os.FileInfo {
	path := filepath.Join(ws.rootPath, dirname)
	stats := make(map[string]os.FileInfo)

	dirEntries, err := os.ReadDir(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for _, dirEntry := range dirEntries {
		if isIgnoredDir(dirEntry) {
			continue
		}
		info, err := dirEntry.Info()
		if err != nil {
			fmt.Println(err)
		}
		stats[dirEntry.Name()] = info
	}

	return stats
}

func isIgnoredDir(dirEntry os.DirEntry) bool {
	for _, ignore := range IGNORES {
		if ignore == dirEntry.Name() {
			return true
		}
	}

	return false
}

func (ws *Workspace) ReadFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

func (ws *Workspace) StatFile(filePath string) (os.FileInfo, error) {
	return os.Stat(filePath)
}

func NewWorkspace(rootPath string) Workspace {
	return Workspace{
		rootPath: rootPath,
	}
}
