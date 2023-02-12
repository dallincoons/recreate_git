package jit

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"jit/app"
	"jit/app/index"
	"jit/app/repository"
	"jit/format"
	"os"
	"path/filepath"
	"sort"
)

var stats = make(map[string]os.FileInfo)

const LABEL_WIDTH = 16

type StatusCmd struct {
	rootPath string
	out      io.Writer
	status   repository.Status
}

var porcelainFlag bool

var statusCobraCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of the index.",
	Run: func(cmd *cobra.Command, args []string) {
		rootPath, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		statusCmd := &StatusCmd{
			rootPath: rootPath,
			out:      os.Stdout,
		}

		status, err := statusCmd.Run(porcelainFlag)
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(status)
	},
}

func (s *StatusCmd) Run(porcelainFlag bool) (int, error) {
	repo := repository.NewRepository(s.rootPath)

	idx := repo.GetIndex()
	err := idx.LoadForUpdate()

	if err != nil {
		fmt.Println("error loading index for update")
		fmt.Println(err)
		return 1, err
	}

	status := repository.NewStatus(repo)

	s.status = status

	//s.scanWorkspace(repo, "")
	//s.loadHeadTree(repo)
	//s.checkIndexEntries(repo)

	idx.WriteUpdates()

	s.printResults(porcelainFlag)

	return 0, nil
}

func (s *StatusCmd) printResults(porcelainFlag bool) {
	if porcelainFlag {
		s.printPorcelainFormat()
	} else {
		s.printLongFormat()
	}
}

const (
	PRINT_GREEN = "green"
	PRINT_RED   = "red"
)

func (s *StatusCmd) printLongFormat() {
	s.printChanges("Changes to be committed", s.status.GetIndexChanges(), PRINT_GREEN)
	s.printChanges("Changes not staged for commit", s.status.GetWorkspaceChanges(), PRINT_RED)
	s.printUntrackedFiles(s.status.GetUntracked())

	s.printCommitStatus()
}

func (s *StatusCmd) printUntrackedFiles(files *[]string) {
	if files == nil || len(*files) == 0 {
		return
	}
	fmt.Println()
	fmt.Println("Untracked files:")
	for _, file := range *files {
		fmt.Println(colorFmt(PRINT_RED, fmt.Sprintf("%*s", LABEL_WIDTH, file)))
	}
}

func (s *StatusCmd) printChanges(message string, changeSet map[string]repository.ChangeType, color string) {
	if len(changeSet) == 0 {
		return
	}

	msgStr := fmt.Sprintf("%s:", message)
	fmt.Println(msgStr)

	for _, pathStr := range sortChangeSet(changeSet) {
		t := changeSet[pathStr]

		longStatus, exists := repository.LONG_STATUS[t]
		if exists {
			fmt.Println(colorFmt(color, fmt.Sprintf("%*s %s", LABEL_WIDTH, longStatus, pathStr)))
		}
	}
}

// I didn't care enough to figure out how to determine if stdout is a tty or not
// but if not then we wouldn't want to colorize the text
func colorFmt(color, string string) string {
	return format.Color(color, string)
}

func sortChangeSet(changeSet map[string]repository.ChangeType) []string {
	changes := []string{}
	for pathChanged := range changeSet {
		changes = append(changes, pathChanged)
	}
	sort.Strings(changes)

	return changes
}

func (s *StatusCmd) sortedIndexChanges() []string {
	changes := []string{}
	for pathChanged := range s.status.GetIndexChanges() {
		changes = append(changes, pathChanged)
	}
	sort.Strings(changes)

	return changes
}

func (s *StatusCmd) printCommitStatus() {
	if len(s.status.GetIndexChanges()) > 0 {
		return
	}

	if len(s.status.GetWorkspaceChanges()) > 0 {
		fmt.Println("no changes added to commit")
	} else if s.status.GetUntracked() != nil && len(*s.status.GetUntracked()) > 0 {
		fmt.Println("nothing added to commit but untracked files present")
	} else {
		fmt.Println("nothing to commit, working tree clean")
	}
}

func (s *StatusCmd) printPorcelainFormat() {
	sort.Strings(*s.status.GetUntracked())

	for _, pathChanged := range s.sortedChanges() {
		status := s.statusFor(pathChanged)
		fmt.Fprintf(s.out, fmt.Sprintf("%s %s\n", status, pathChanged))
	}
	for _, pathChanged := range *s.status.GetUntracked() {
		fmt.Fprintf(s.out, fmt.Sprintf("?? %s\n", pathChanged))
	}
}

func (s *StatusCmd) sortedChanges() []string {
	changes := []string{}
	for pathChanged := range s.status.GetChanged() {
		changes = append(changes, pathChanged)
	}
	sort.Strings(changes)

	return changes
}

func (s *StatusCmd) statusFor(path string) string {
	indexChange := s.status.GetIndexChanges()[path]
	workspaceChange := s.status.GetWorkspaceChanges()[path]

	left := " "
	right := " "

	if indexChange == repository.Added {
		left = "A"
	}

	if indexChange == repository.Modified {
		left = "M"
	}

	if indexChange == repository.Deleted {
		left = "D"
	}

	if workspaceChange == repository.Added {
		right = "A"
	}

	if workspaceChange == repository.Modified {
		right = "M"
	}

	if workspaceChange == repository.Deleted {
		right = "D"
	}

	return left + right
}

//func (s *StatusCmd) scanWorkspace(repo repository.Repository, prefix string) {
//	workspace := repo.GetWorkspace()
//	index := repo.GetIndex()
//	statMap := workspace.ListDirectories(prefix)
//
//	for path, stat := range statMap {
//		prefixedPath := filepath.Join(prefix, path)
//		if index.IsTracked(prefixedPath) {
//			stats[prefixedPath] = stat
//			if stat.IsDir() {
//				s.scanWorkspace(repo, prefixedPath)
//			}
//		} else if isTrackableFile(prefixedPath, stat, index, workspace) {
//			if stat.IsDir() {
//				prefixedPath += string(filepath.Separator)
//			}
//			*s.untracked = append(*s.untracked, prefixedPath)
//		}
//	}
//}

//func (s *StatusCmd) loadHeadTree(repo repository.Repository) {
//	headOid := repo.GetRefs().ReadHead()
//
//	if headOid == "" {
//		return
//	}
//
//	obj := repo.GetDatabase().Load(headOid)
//	commit := app.ObjToCommit(obj)
//	s.readTree(repo, commit.GetTreeOid(), "")
//}

//func (s *StatusCmd) readTree(repo repository.Repository, treeOid, pathName string) {
//	obj := repo.GetDatabase().Load(treeOid)
//
//	tree := app.ObjToTree(obj)
//
//	for name, entry := range tree.EntryMap.EMap {
//		addedName := path.Join(pathName, name)
//		if entry.IsATree() {
//			s.readTree(repo, entry.GetOid(), path.Join(pathName, addedName))
//		} else {
//			s.headTree[addedName] = entry.Entry
//		}
//	}
//}

//func (s *StatusCmd) checkIndexEntries(repo repository.Repository) {
//	idx := repo.GetIndex()
//	for _, entry := range idx.SortedEntries() {
//		s.checkIndexEntryAgainstWorkspace(repo, entry)
//		s.checkIndexEntryAgainstHeadTree(entry)
//		s.collectDeletedHeadFiles(repo)
//	}
//}

//func (s *StatusCmd) checkIndexEntryAgainstWorkspace(repo repository.Repository, entry index.Entry) {
//	stat, exists := stats[entry.Path]
//
//	if !exists {
//		s.recordChange(entry.Path, s.workspaceChanges, Deleted)
//		return
//	}
//
//	if !entry.StatMatches(stat) {
//		s.recordChange(entry.Path, s.workspaceChanges, Modified)
//		return
//	}
//
//	if !entry.TimeMatches(index.ToFileInfo(stat)) {
//		return
//	}
//
//	workspace := repo.GetWorkspace()
//	database := repo.GetDatabase()
//	idx := repo.GetIndex()
//	data, err := workspace.ReadFile(entry.Path)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	blob := app.NewBlob(data)
//	oid := database.HashObject(blob)
//
//	if oid == entry.Oid {
//		idx.UpdateEntryStat(entry, index.ToFileInfo(stat))
//	} else {
//		s.recordChange(entry.Path, s.workspaceChanges, Modified)
//	}
//}

func isTrackableFile(path string, stat os.FileInfo, index *index.Index, workspace app.Workspace) bool {
	if !stat.IsDir() {
		return !index.IsTracked(path)
	}

	untracked := workspace.ListDirectories(path)

	for file, stat := range untracked {
		if !stat.IsDir() {
			if isTrackableFile(filepath.Join(path, file), stat, index, workspace) {
				return true
			}
		}
	}

	for file, stat := range untracked {
		if stat.IsDir() {
			if isTrackableFile(filepath.Join(path, file), stat, index, workspace) {
				return true
			}
		}
	}

	return false
}

func init() {
	rootCmd.AddCommand(statusCobraCmd)

	statusCobraCmd.Flags().BoolVarP(&porcelainFlag, "porcelain", "p", false, "")
}
