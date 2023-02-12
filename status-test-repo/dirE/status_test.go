package dirD

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"jit/app/cmd/jit"
	jit_testing "jit/testing"
	"os"
	"testing"
)

func TestDontListAnythingIfNothingHasChanged(t *testing.T) {
	err := os.RemoveAll(".git")
	if err != nil {
		fmt.Println("error removing git dir")
		fmt.Println(err)
	}
	jit.RunInitCmd([]string{jit_testing.RepoPath()})

	err = os.WriteFile("file1.txt", []byte("file1"), 0777)
	if err != nil {
		fmt.Println("error writing file1.txt")
	}

	_, err = jit.AddCmdRun(jit_testing.RepoPath(), []string{"file1.txt"})
	_, err = jit.AddCmdRun(jit_testing.RepoPath(), []string{"status_test.go"})

	buff := &bytes.Buffer{}

	jit.CommitCmdRun(jit_testing.RepoPath(), buff)

	buff2 := &bytes.Buffer{}

	_, err = jit.StatusCmdRun(jit_testing.RepoPath(), buff2)
	if err != nil {
		fmt.Println(err)
	}

	assert.Equal(t, "", buff2.String())
}

func TestListChangedFiles(t *testing.T) {
	err := os.RemoveAll(".git")
	if err != nil {
		fmt.Println("error removing git dir")
		fmt.Println(err)
	}
	jit.RunInitCmd([]string{jit_testing.RepoPath()})

	err = os.WriteFile("file1.txt", []byte("file1"), 0777)
	if err != nil {
		fmt.Println("error writing file1.txt")
	}

	_, err = jit.AddCmdRun(jit_testing.RepoPath(), []string{"file1.txt"})
	_, err = jit.AddCmdRun(jit_testing.RepoPath(), []string{"status_test.go"})

	buff := &bytes.Buffer{}

	jit.CommitCmdRun(jit_testing.RepoPath(), buff)

	err = os.WriteFile("file1.txt", []byte("file1 changed"), 0777)
	if err != nil {
		fmt.Println("error writing file1.txt")
	}

	buff2 := &bytes.Buffer{}

	_, err = jit.StatusCmdRun(jit_testing.RepoPath(), buff2)
	if err != nil {
		fmt.Println(err)
	}

	assert.Equal(t, " M file1.txt\n", buff2.String())
}
