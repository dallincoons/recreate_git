package repository

import (
	"fmt"
	"jit/app"
	"jit/app/index"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
)

type ChangeType string

const Deleted = "deleted"
const Modified = "modified"
const Added = "added"

var LONG_STATUS = map[ChangeType]string{
	Added:    "new file:",
	Deleted:  "deleted:",
	Modified: "modified:",
}

type Status struct {
	repo             Repository
	changed          map[string]bool
	untracked        *[]string
	stats            map[string]os.FileInfo
	indexChanges     map[string]ChangeType
	workspaceChanges map[string]ChangeType
	headTree         map[string]*app.Entry
}

func NewStatus(repo Repository) Status {
	changed := make(map[string]bool)
	untracked := make([]string, 0)

	status := Status{
		repo:             repo,
		changed:          changed,
		untracked:        &untracked,
		stats:            make(map[string]os.FileInfo),
		indexChanges:     make(map[string]ChangeType),
		workspaceChanges: make(map[string]ChangeType),
		headTree:         make(map[string]*app.Entry),
	}

	status.scanWorkspace(repo, "")
	status.loadHeadTree(repo)
	status.checkIndexEntries(repo)

	return status
}

func (s *Status) GetChanged() map[string]bool {
	return s.changed
}

func (s *Status) GetUntracked() *[]string {
	sort.Strings(*s.untracked)
	return s.untracked
}

func (s *Status) GetHeadTree() map[string]*app.Entry {
	return s.headTree
}

func (s *Status) GetIndexChanges() map[string]ChangeType {
	return s.indexChanges
}

func (s *Status) GetWorkspaceChanges() map[string]ChangeType {
	return s.workspaceChanges
}

func (s *Status) GetStats() map[string]os.FileInfo {
	return s.stats
}

func (s *Status) scanWorkspace(repo Repository, prefix string) {
	workspace := repo.GetWorkspace()
	index := repo.GetIndex()
	statMap := workspace.ListDirectories(prefix)

	for path, stat := range statMap {
		prefixedPath := filepath.Join(prefix, path)
		if index.IsTracked(prefixedPath) {
			s.stats[prefixedPath] = stat
			if stat.IsDir() {
				s.scanWorkspace(repo, prefixedPath)
			}
		} else if isTrackableFile(prefixedPath, stat, index, workspace) {
			if stat.IsDir() {
				prefixedPath += string(filepath.Separator)
			}
			*s.untracked = append(*s.untracked, prefixedPath)
		}
	}
}

func (s *Status) loadHeadTree(repo Repository) {
	headOid := repo.GetRefs().ReadHead()

	if headOid == "" {
		return
	}

	obj := repo.GetDatabase().Load(headOid)
	commit := app.ObjToCommit(obj)
	s.readTree(repo, commit.GetTreeOid(), "")
}

func (s *Status) readTree(repo Repository, treeOid, pathName string) {
	obj := repo.GetDatabase().Load(treeOid)

	tree := app.ObjToTree(obj)

	for name, entry := range tree.EntryMap.EMap {
		addedName := path.Join(pathName, name)
		if entry.Entry.Mode == 40000 {
			s.readTree(repo, entry.Entry.Oid, addedName)
		} else {
			s.headTree[addedName] = entry.Entry
		}
	}
}

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

func (s *Status) checkIndexEntries(repo Repository) {
	idx := repo.GetIndex()
	for _, entry := range idx.SortedEntries() {
		s.checkIndexEntryAgainstWorkspace(repo, entry)
		s.checkIndexEntryAgainstHeadTree(entry)
		s.collectDeletedHeadFiles(repo)
	}
}

func (s *Status) collectDeletedHeadFiles(repo Repository) {
	for pathName := range s.headTree {
		if !repo.GetIndex().IsTracked(pathName) {
			s.recordChange(pathName, s.indexChanges, Deleted)
		}
	}
}

func (s *Status) checkIndexEntryAgainstWorkspace(repo Repository, entry index.Entry) {
	stat, exists := s.stats[entry.Path]

	if !exists {
		s.recordChange(entry.Path, s.workspaceChanges, Deleted)
		return
	}

	if !entry.StatMatches(stat) {
		s.recordChange(entry.Path, s.workspaceChanges, Modified)
		return
	}

	if !entry.TimeMatches(index.ToFileInfo(stat)) {
		return
	}

	workspace := repo.GetWorkspace()
	database := repo.GetDatabase()
	idx := repo.GetIndex()
	data, err := workspace.ReadFile(entry.Path)
	if err != nil {
		fmt.Println(err)
		return
	}
	blob := app.NewBlob(data)
	oid := database.HashObject(blob)

	if oid == entry.Oid {
		idx.UpdateEntryStat(entry, index.ToFileInfo(stat))
	} else {
		s.recordChange(entry.Path, s.workspaceChanges, Modified)
	}
}

func (s *Status) recordChange(path string, set map[string]ChangeType, changeType ChangeType) {
	s.changed[path] = true

	set[path] = changeType
}

func (s *Status) checkIndexEntryAgainstHeadTree(entry index.Entry) {
	item, exists := s.headTree[entry.Path]

	if exists {
		eMode, _ := strconv.Atoi(fmt.Sprintf("%o", entry.Mode))
		if int32(eMode) != item.Mode || entry.Oid != item.Oid {
			s.recordChange(entry.Path, s.indexChanges, Modified)
		}
	} else {
		s.recordChange(entry.Path, s.indexChanges, Added)
	}
}
