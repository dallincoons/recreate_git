package jit

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes jit project",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RunInitCmd(args)
	},
}

func RunInitCmd(args []string) {
	var rootPath string
	var err error

	if len(args) > 0 {
		rootPath, err = filepath.Abs(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
			return
		}
	}

	if rootPath == "" {
		rootPath, err = os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
			return
		}
	}

	jitpath := filepath.Join(rootPath, ".git")

	for _, dirName := range []string{"objects", "refs"} {
		err = os.MkdirAll(filepath.Join(jitpath, dirName), os.ModePerm)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
			return
		}
	}

	fmt.Println(fmt.Sprintf("Initialized empty Jit repository in %s", jitpath))
}

func init() {
	rootCmd.AddCommand(initCmd)
}
