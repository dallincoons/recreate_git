package dirB

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"jit/app/cmd/jit"
	jit_testing "jit/testing"
	"os"
	"testing"
)

func xTestListUntrackedDirectoriesNotTheirContents(t *testing.T) {
	err := os.RemoveAll(".git")
	if err != nil {
		fmt.Println("error removing git dir")
		fmt.Println(err)
	}
	jit.RunInitCmd([]string{jit_testing.RepoPath()})

	buff := &bytes.Buffer{}

	_, err = jit.StatusCmdRun(jit_testing.RepoPath(), buff)
	if err != nil {
		fmt.Println(err)
	}

	assert.Equal(t, "?? dir/\n?? file.txt\n?? status_test.go\n", buff.String())
}

func TestListFilesAsUntrackedIfNotInIndex(t *testing.T) {
	err := os.RemoveAll(".git")
	if err != nil {
		fmt.Println("error removing git dir")
		fmt.Println(err)
	}
	jit.RunInitCmd([]string{jit_testing.RepoPath()})

	_, err = jit.AddCmdRun(jit_testing.RepoPath(), []string{"dir/subdir/another2.txt"})

	if err != nil {
		fmt.Println("error adding committed.txt to index")
		fmt.Println(err)
	}

	buff2 := &bytes.Buffer{}

	_, err = jit.StatusCmdRun(jit_testing.RepoPath(), buff2)
	if err != nil {
		fmt.Println(err)
	}

	assert.Equal(t, "?? dir/another.txt\n?? file.txt\n?? status_test.go\n", buff2.String())
}
