package jit

import (
	"fmt"
	"github.com/spf13/cobra"
	"jit/app"
	"jit/app/diff"
	"jit/app/index"
	"jit/app/repository"
	"jit/format"
	"os"
	"path"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Calculate diff.",
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		exitCode, err := DiffRun(rootPath, args)
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(exitCode)
	},
}

func DiffRun(rootPath string, args []string) (int, error) {
	cachedFlag, _ := rootCmd.Flags().GetBool("cached")
	stagingFlag, _ := rootCmd.Flags().GetBool("staging")
	if cachedFlag || stagingFlag {
		return diffHeadIndex(rootPath)
	}

	return diffIndexWorkspace(rootPath)
}

func diffHeadIndex(rootPath string) (int, error) {
	repo := repository.NewRepository(rootPath)
	repo.GetIndex().Load()

	status := repository.NewStatus(repo)

	for p, state := range status.GetIndexChanges() {
		switch state {
		case repository.Added:
			printDiff(NewTargetFromNothing(repo, p), NewTargetFromIndex(repo, p))
		case repository.Modified:
			printDiff(NewTargetFromHead(repo, p), NewTargetFromFile(repo, p))
		case repository.Deleted:
			printDiff(NewTargetFromHead(repo, p), NewTargetFromNothing(repo, p))
		}
	}

	return 0, nil
}

func diffIndexWorkspace(rootPath string) (int, error) {
	repo := repository.NewRepository(rootPath)
	repo.GetIndex().Load()

	status := repository.NewStatus(repo)

	for p, state := range status.GetWorkspaceChanges() {
		switch state {
		case repository.Modified:
			printDiff(NewTargetFromIndex(repo, p), NewTargetFromFile(repo, p))
		case repository.Deleted:
			printDiff(NewTargetFromIndex(repo, p), NewTargetFromNothing(repo, p))
		}
	}

	return 0, nil
}

const NULL_OID = "0000000"
const NULL_PATH = "/dev/null"

type Target struct {
	Path string
	Oid  string
	Mode string
	Data []string
}

func NewTarget(path, oid, mode string, data []string) *Target {
	return &Target{
		Path: path,
		Oid:  oid,
		Mode: mode,
		Data: data,
	}
}

func NewTargetFromIndex(repo repository.Repository, path string) *Target {
	entry := repo.GetIndex().EntryForPath(path)

	obj := repo.GetDatabase().Load(entry.Oid)
	blob := app.ObjToBlob(obj)

	return &Target{
		Path: path,
		Oid:  entry.Oid,
		Mode: fmt.Sprintf("%o", entry.Mode),
		Data: blob.GetLines(),
	}
}

func NewTargetFromHead(repo repository.Repository, path string) *Target {
	status := repository.NewStatus(repo)
	entry := status.GetHeadTree()[path]
	obj := repo.GetDatabase().Load(entry.Oid)
	blob := app.ObjToBlob(obj)

	return NewTarget(entry.Name, entry.Oid, fmt.Sprintf("%o", entry.Mode), blob.GetLines())
}

func NewTargetFromFile(repo repository.Repository, path string) *Target {
	database := repo.GetDatabase()
	workspace := repo.GetWorkspace()
	status := repository.NewStatus(repo)

	fileBytes, err := workspace.ReadFile(path)
	if err != nil {
		fmt.Println(err)
	}

	blob := app.NewBlob(fileBytes)

	bOid := database.HashObject(blob)
	bMode := fmt.Sprintf("%o", index.ToFileInfo(status.GetStats()[path]).Mode)

	return NewTarget(path, bOid, bMode, blob.GetLines())
}

func NewTargetFromNothing(repo repository.Repository, path string) *Target {
	return &Target{
		Path: path,
		Oid:  NULL_OID,
		Mode: "",
		Data: []string{""},
	}
}

func (t *Target) diffPath() string {
	if t.Mode != "" {
		return t.Path
	}

	return NULL_PATH
}

func printDiff(a, b *Target) {
	if a.Oid == b.Oid && a.Mode == b.Mode {
		return
	}

	a.Path = path.Join("a", a.Path)
	b.Path = path.Join("b", b.Path)

	fmt.Println(fmt.Sprintf("diff --git %s %s", a.Path, b.Path))
	printDiffMode(a, b)
	printDiffContents(a, b)
}

func printDiffMode(a, b *Target) {
	if a == nil || b == nil {
		fmt.Println("a or b target is nil")
		return
	}

	if a.Mode == "" {
		fmt.Println(fmt.Sprintf("new file mode %s", b.Mode))
	} else if b.Mode == "" {
		fmt.Println(fmt.Sprintf("deleted file mode %s", a.Mode))
	} else if a.Mode != b.Mode {
		fmt.Println("old mode", a.Mode)
		fmt.Println("new mode", b.Mode)
	}
}

func printDiffContents(a, b *Target) {
	if a.Oid == b.Oid {
		return
	}

	oidRange := fmt.Sprintf("index %s..%s", shortOid(a.Oid), shortOid(b.Oid))
	if a.Mode == b.Mode {
		oidRange = fmt.Sprintf("%s %s", oidRange, a.Mode)
	}

	fmt.Println(oidRange)
	fmt.Println(fmt.Sprintf("--- %s", a.diffPath()))
	fmt.Println(fmt.Sprintf("+++ %s", b.diffPath()))

	hunks := diff.DiffHunks(diff.Lines(a.Data), diff.Lines(b.Data))
	for _, hunk := range hunks {
		printDiffHunk(hunk)
	}
}

func printDiffHunk(hunk *diff.Hunk) {
	fmt.Println(format.Color("cyan", hunk.Header()))
	for _, edit := range hunk.Edits {
		fmt.Println(edit.ToString())
	}
}

func shortOid(oid string) string {
	return oid[0:6]
}

func init() {
	rootCmd.AddCommand(diffCmd)

	rootCmd.PersistentFlags().Bool("cached", false, "Look at difference between staging area and index.")
}
