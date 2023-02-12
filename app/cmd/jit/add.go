package jit

import (
	"fmt"
	"github.com/spf13/cobra"
	"jit/app"
	"jit/app/index"
	"jit/app/repository"
	"os"
	"path/filepath"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add files to the staging environment.",
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		exitCode, err := AddCmdRun(rootPath, args)
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(exitCode)
	},
}

func AddCmdRun(rootPath string, args []string) (int, error) {
	repo := repository.NewRepository(rootPath)

	idx := repo.GetIndex()
	database := repo.GetDatabase()
	workspace := repo.GetWorkspace()

	err := idx.LoadForUpdate()
	if err != nil {
		fmt.Println(`
				Another git process seems to be running in this repository, e.g.
				an editor opened by 'git commit'. Please make sure all processes
				are terminated then try again. If it still fails, a git process
				may have crashed in this repository earlier:
				remove the file manually to continue.
			`)
		return 1, err
	}

	paths := []string{}
	for _, argPath := range args {
		absPath, err := filepath.Abs(argPath)
		if err != nil {
			return 1, err
		}

		filePaths, err := workspace.GetFiles(absPath)
		if err != nil {
			idx.ReleaseLock()
			return 1, err
		}
		paths = append(paths, filePaths...)
	}

	for _, path := range paths {
		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			fmt.Println("Could not get relative filepath.")
			return 1, err
		}
		data, err := workspace.ReadFile(relPath)
		if err != nil {
			fmt.Println(fmt.Sprintf("Could not read file: %s", relPath))
			return 1, err
		}
		fileStat, err := workspace.StatFile(relPath)
		if err != nil {
			return 1, err
		}

		blob := app.NewBlob(data)
		database.Store(blob)
		idx.Add(relPath, blob.Oid, index.ToFileInfo(fileStat))
	}

	idx.WriteUpdates()

	return 0, nil
}

func init() {
	rootCmd.AddCommand(addCmd)
}
