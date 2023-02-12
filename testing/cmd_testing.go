package jit_testing

import (
	"bytes"
	"fmt"
	"jit/app/cmd/jit"
	"jit/app/repository"
	"os"
	"path/filepath"
)

func RepoPath(args ...string) string {
	path := "."
	if len(args) == 1 {
		path = args[0]
	}
	absPath, _ := filepath.Abs(path)

	return absPath
}

func Repo() repository.Repository {
	return repository.NewRepository(RepoPath())
}

func WriteFile(name string, contents string) string {
	err := os.MkdirAll(RepoPath(), os.ModePerm)
	if err != nil {
		fmt.Println("failed making repo dir")
		fmt.Println(err)
	}

	path := filepath.Join(RepoPath(), name)
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	_, err = file.WriteString(contents)
	if err != nil {
		fmt.Println("error writing to fake repo")
		fmt.Println(err)
	}

	return path
}

func Commit(message string) {
	os.Setenv("GIT_AUTHOR_NAME", "Dallin")
	os.Setenv("GIT_AUTHOR_EMAIL", "dallincoons@gmail.com")
	rootPath, _ := filepath.Abs(".")
	buff := &bytes.Buffer{}

	jit.CommitCmdRun(rootPath, buff)
}
