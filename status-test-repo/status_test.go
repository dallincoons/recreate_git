package status_test_repo

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"jit/app/cmd/jit"
	jit_testing "jit/testing"
	"os"
	"path/filepath"
	"testing"
)

func TestShowUntrackedFiles(t *testing.T) {
	os.RemoveAll("dirA/.git")
	jit.RunInitCmd([]string{jit_testing.RepoPath("./dirA")})

	rootPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	buff := &bytes.Buffer{}
	_, err = jit.StatusCmdRun(filepath.Join(rootPath, "./dirA"), buff)
	if err != nil {
		fmt.Println(err)
	}

	assert.Equal(t, buff.String(), `?? fileA.txt
?? fileB.txt
`)
}

func TestListFilesAsUntrackedIfNotInIndex(t *testing.T) {
	err := os.RemoveAll("dirB/.git")
	if err != nil {
		fmt.Println("error removing git dir")
		fmt.Println(err)
	}
	jit.RunInitCmd([]string{jit_testing.RepoPath("./dirB")})

	//rootPath, err := os.Getwd()
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}

	_, err = jit.AddCmdRun(jit_testing.RepoPath("./dirB"), []string{"committed.txt"})

	if err != nil {
		fmt.Println("error adding committed.txt to index")
		fmt.Println(err)
	}

	//buff := &bytes.Buffer{}

	//jit.CommitCmdRun(filepath.Join(rootPath, "dirB"), buff)

	//buff2 := &bytes.Buffer{}

	//_, err = jit.StatusCmdRun(filepath.Join(rootPath, "dirB"), buff2)
	//if err != nil {
	//	fmt.Println(err)
	//}

	//assert.Equal(t, `?? /Users/dallincoons/Projects/jit/status-test-repo/dirB/uncommitted.txt`, buff2.String())
}
