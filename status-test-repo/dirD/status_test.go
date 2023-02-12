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

func TestDontListEmptyDirAndListRootDirWhenSubdirHasTrackableFile(t *testing.T) {
	err := os.RemoveAll(".git")
	if err != nil {
		fmt.Println("error removing git dir")
		fmt.Println(err)
	}
	jit.RunInitCmd([]string{jit_testing.RepoPath()})

	if err != nil {
		fmt.Println("error adding committed.txt to index")
		fmt.Println(err)
	}

	buff := &bytes.Buffer{}

	_, err = jit.StatusCmdRun(jit_testing.RepoPath(), buff)
	if err != nil {
		fmt.Println(err)
	}

	assert.Equal(t, `?? dir/
?? status_test.go
`, buff.String())
}
