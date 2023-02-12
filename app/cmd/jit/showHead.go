package jit

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"jit/app"
	"jit/app/repository"
	"os"
	"path"
)

var showHeadCmd = &cobra.Command{
	Use:   "show_head",
	Short: "Show all objects in commit tree.",
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		ShowHeadRun()
	},
}

func ShowHeadRun() {
	rootPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	repo := repository.NewRepository(rootPath)

	headOid := repo.GetRefs().ReadHead()
	obj := repo.GetDatabase().Load(headOid)
	bytes, err := json.Marshal(*obj)
	if err != nil {
		fmt.Println(err)
		return
	}
	commit := &app.Commit{}
	err = json.Unmarshal(bytes, commit)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%+v\n", *obj)

	showTree(repo, commit.GetTreeOid(), "")
}

func showTree(repo repository.Repository, oid string, prefix string) {
	obj := repo.GetDatabase().Load(oid)

	bytes, err := json.Marshal(*obj)
	if err != nil {
		fmt.Println(err)
	}
	tree := &app.Tree{}
	err = json.Unmarshal(bytes, tree)
	if err != nil {
		fmt.Println(err)
		return
	}

	for name, e := range tree.GetEntryMap() {
		pth := path.Join(prefix, name)
		if e.IsATree() {
			showTree(repo, e.GetOid(), pth)
		} else {
			mode := e.GetMode()
			fmt.Println(fmt.Sprintf("%d %s %s", mode, e.GetOid(), pth))
		}
	}
}

func init() {
	rootCmd.AddCommand(showHeadCmd)
}
