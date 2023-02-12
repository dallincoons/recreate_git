package jit

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"jit/app"
	"jit/app/index"
	"os"
	"path"
	"path/filepath"
	"time"
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Save a snapshot of your project.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		code, err := CommitCmdRun(rootPath, os.Stdout)
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(code)
	},
}

func CommitCmdRun(rootPath string, out io.Writer) (int, error) {
	gitPath := path.Join(rootPath, ".git")
	dbPath := path.Join(gitPath, "objects")

	database := app.NewDatabase(dbPath)
	index := index.NewIndex(path.Join(gitPath, "index"))
	refs := app.NewRef(gitPath)

	index.Load()

	sortedEntries := index.SortedEntries()
	entries := make([]app.Entry, 0, len(sortedEntries))
	for _, e := range index.SortedEntries() {
		newEntry := app.NewEntry(e.Path, e.Mode, e.Oid)
		entries = append(entries, newEntry)
	}

	tree := app.BuildTree(entries)
	tree.Traverse(func(t *app.Tree) {
		database.Store(t)
	})

	parentID := refs.ReadHead()
	authorName := os.Getenv("GIT_AUTHOR_NAME")
	authorEmail := os.Getenv("GIT_AUTHOR_EMAIL")
	author := app.NewAuthor(authorName, authorEmail, time.Now())

	reader := bufio.NewReader(os.Stdin)
	message, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(out, "error reading stdin\n")
		fmt.Fprintf(out, err.Error()+"\n")
		return 1, err
	}

	commit := app.NewCommit(parentID, tree.GetOid(), author, message)
	database.Store(commit)
	err = refs.UpdateHead(commit.GetOid())
	if err != nil {
		fmt.Fprintf(out, "error writing to head\n")
		return 1, err
	}

	file, err := os.OpenFile(filepath.Join(gitPath, "HEAD"), os.O_WRONLY|os.O_CREATE, 0777)
	defer file.Close()
	if err != nil {
		fmt.Fprintf(out, "error opening HEAD file\n")
		return 1, err
	}

	_, err = file.WriteString(commit.GetOid() + "\n")
	if err != nil {
		fmt.Fprintf(out, "error writing commit file\n")
		return 1, err
	}

	isRoot := ""
	if parentID == "" {
		isRoot = "(root-commit)"
	}
	fmt.Println(fmt.Sprintf("%s %s %s", isRoot, commit.GetOid(), message))
	return 0, nil
}

func init() {
	rootCmd.AddCommand(commitCmd)
}
