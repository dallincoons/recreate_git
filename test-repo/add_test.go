package test_repo

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"jit/app/cmd/jit"
	"jit/app/index"
	jit_testing "jit/testing"
	"os"
	"path/filepath"
	"testing"
)

func TestAddFileToIndex(t *testing.T) {
	os.RemoveAll(".git")
	jit.RunInitCmd([]string{jit_testing.RepoPath()})

	jit_testing.WriteFile("hello.txt", "hello")

	repo := jit_testing.Repo()
	index := repo.GetIndex()

	rootPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_, err = jit.AddCmdRun(rootPath, []string{"hello.txt"})
	if err != nil {
		fmt.Println(err)
	}

	index.Load()

	assert.Equal(t, 1, len(index.SortedEntries()))
	assert.Equal(t, true, indexEntryExists(index, "hello.txt"))
}

func TestAddMultipleFilesToIndex(t *testing.T) {
	os.RemoveAll(".git")
	jit.RunInitCmd([]string{jit_testing.RepoPath()})

	jit_testing.WriteFile("hello.txt", "hello")
	jit_testing.WriteFile("hello2.txt", "hello2")

	repo := jit_testing.Repo()
	index := repo.GetIndex()

	rootPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_, err = jit.AddCmdRun(rootPath, []string{"hello.txt", "hello2.txt"})
	if err != nil {
		fmt.Println(err)
	}

	index.Load()

	assert.Equal(t, 2, len(index.SortedEntries()))
	assert.Equal(t, true, indexEntryExists(index, "hello.txt"))
	assert.Equal(t, true, indexEntryExists(index, "hello2.txt"))
}

func TestIncrementallyAddsFilesToIndex(t *testing.T) {
	os.RemoveAll(".git")
	jit.RunInitCmd([]string{jit_testing.RepoPath()})

	jit_testing.WriteFile("hello.txt", "hello")
	jit_testing.WriteFile("hello2.txt", "hello2")

	repo := jit_testing.Repo()
	index := repo.GetIndex()

	rootPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_, err = jit.AddCmdRun(rootPath, []string{"hello.txt"})
	if err != nil {
		fmt.Println(err)
	}

	index.Load()

	assert.Equal(t, 1, len(index.SortedEntries()))
	assert.Equal(t, true, indexEntryExists(index, "hello.txt"))

	_, err = jit.AddCmdRun(rootPath, []string{"hello2.txt"})
	if err != nil {
		fmt.Println(err)
	}

	index.Load()

	assert.Equal(t, 2, len(index.SortedEntries()))
	assert.Equal(t, true, indexEntryExists(index, "hello.txt"))
	assert.Equal(t, true, indexEntryExists(index, "hello2.txt"))
}

func TestAddsDirectoryToIndex(t *testing.T) {
	os.RemoveAll(".git")
	jit.RunInitCmd([]string{jit_testing.RepoPath()})

	err := os.MkdirAll(filepath.Join(jit_testing.RepoPath(), "a-dir"), os.ModePerm)
	if err != nil {
		fmt.Println("failed making repo dir")
		fmt.Println(err)
	}
	jit_testing.WriteFile("a-dir/hello.txt", "hello")

	repo := jit_testing.Repo()
	index := repo.GetIndex()

	rootPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_, err = jit.AddCmdRun(rootPath, []string{"a-dir/hello.txt"})
	if err != nil {
		fmt.Println(err)
	}

	index.Load()

	assert.Equal(t, 1, len(index.SortedEntries()))
	assert.Equal(t, true, indexEntryExists(index, "a-dir/hello.txt"))
}

func TestAddsRepositoryRootToIndex(t *testing.T) {
	os.RemoveAll(".git")
	jit.RunInitCmd([]string{jit_testing.RepoPath()})

	err := os.MkdirAll(filepath.Join(jit_testing.RepoPath(), "a/b/c"), os.ModePerm)
	if err != nil {
		fmt.Println("failed making repo dir")
		fmt.Println(err)
	}
	jit_testing.WriteFile("a/b/c/x.txt", "ex")

	repo := jit_testing.Repo()
	index := repo.GetIndex()

	rootPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_, err = jit.AddCmdRun(rootPath, []string{"."})
	if err != nil {
		fmt.Println(err)
	}

	index.Load()

	assert.Equal(t, true, indexEntryExists(index, "a/b/c/x.txt"))
}

func TestFailsToAddNonExistingIndex(t *testing.T) {
	os.RemoveAll(".git")
	jit.RunInitCmd([]string{jit_testing.RepoPath()})

	repo := jit_testing.Repo()
	index := repo.GetIndex()

	rootPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_, err = jit.AddCmdRun(rootPath, []string{"file-dont-exist"})
	if err != nil {
		fmt.Println(err)
	}

	index.Load()

	assert.Equal(t, 0, len(index.SortedEntries()))
}

func TestFailsIfIndexIsLocked(t *testing.T) {
	os.RemoveAll(".git")
	jit_testing.WriteFile("hello.txt", "hello")

	jit.RunInitCmd([]string{jit_testing.RepoPath()})

	jit_testing.WriteFile(".git/index.lock", "")

	repo := jit_testing.Repo()
	index := repo.GetIndex()

	rootPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_, err = jit.AddCmdRun(rootPath, []string{"hello.txt"})
	if err != nil {
		fmt.Println(err)
	}

	index.Load()

	assert.Equal(t, 0, len(index.SortedEntries()))
}

func indexEntryExists(index index.Index, fileName string) bool {
	fileFound := false
	for _, entry := range index.SortedEntries() {
		if entry.Path == fileName {
			fileFound = true
		}
	}

	return fileFound
}
