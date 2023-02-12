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

func TestListFilesAsUntrackedIfNotInIndex(t *testing.T) {
	err := os.RemoveAll(".git")
	if err != nil {
		fmt.Println("error removing git dir")
		fmt.Println(err)
	}
	jit.RunInitCmd([]string{jit_testing.RepoPath()})

	_, err = jit.AddCmdRun(jit_testing.RepoPath(), []string{"committed.txt"})

	if err != nil {
		fmt.Println("error adding committed.txt to index")
		fmt.Println(err)
	}

	buff := &bytes.Buffer{}

	jit.CommitCmdRun(jit_testing.RepoPath(), buff)

	buff2 := &bytes.Buffer{}

	_, err = jit.StatusCmdRun(jit_testing.RepoPath(), buff2)
	if err != nil {
		fmt.Println(err)
	}

	assert.Equal(t, `?? status_test.go
?? uncommitted.txt
`, buff2.String())
}
