package app

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Tree struct {
	Oid      string
	Entries  []Entry
	EntryMap EntryMap
}

type ByName []Entry

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

func (t *Tree) ToString() string {
	packStrings := []string{}
	sort.Strings(t.EntryMap.keys)
	for _, k := range t.EntryMap.keys {
		entryOrTree := t.EntryMap.EMap[k]
		if entryOrTree.IsATree() {
			packStrings = append(packStrings, entryToPackString(k, entryOrTree.Tree))
		} else {
			packStrings = append(packStrings, entryToPackString(k, entryOrTree.Entry))
		}
	}

	return strings.Join(packStrings, "")
}

func (t *Tree) GetEntryMap() map[string]EntryOrTree {
	return t.EntryMap.EMap
}

func entryToPackString(name string, entry StorableEntry) string {
	binaryOid, err := hex.DecodeString(entry.GetOid())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	modeString := entry.GetMode()
	return fmt.Sprintf("%o %s\000%s", modeString, name, string(binaryOid))
}

func (*Tree) Type() string {
	return "tree"
}

func (t *Tree) GetMode() int {
	return DIRECTORY_MODE
}

func (t *Tree) SetOid(oid string) {
	t.Oid = oid
}

func (c Tree) parse(contents []byte) Object {
	state := "mode"

	var mode strings.Builder
	var name strings.Builder
	var oid []byte
	oidCounter := 1

	var entries []Entry

	for _, ch := range contents {
		if state == "mode" {
			if ch == '\x00' {
				continue
			}
			if ch == ' ' {
				state = "name"
				continue
			}
			mode.WriteString(string(ch))
		} else if state == "name" {
			if ch == '\x00' {
				state = "oid"
				continue
			}

			name.WriteString(string(ch))
		} else if state == "oid" {
			oid = append(oid, ch)

			if oidCounter == 20 {
				state = "mode"

				modeInt, err := strconv.Atoi(mode.String())

				if err != nil {
					fmt.Println(err)
					return nil
				}

				entries = append(entries, NewEntry(name.String(), int32(modeInt), hex.EncodeToString(oid)))

				mode.Reset()
				name.Reset()
				oid = []byte{}
				oidCounter = 1
				continue
			}

			oidCounter++
		}
	}

	return BuildTree(entries)
}

func (t *Tree) GetOid() string {
	return t.Oid
}

type EntryMap struct {
	EMap map[string]EntryOrTree
	keys []string
}

type EntryOrTree struct {
	Entry *Entry
	Tree  *Tree
}

func (eot EntryOrTree) GetMode() int {
	if eot.IsATree() {
		return eot.Tree.GetMode()
	}

	return eot.Entry.GetMode()
}

func (eot EntryOrTree) GetOid() string {
	if eot.IsATree() {
		return eot.Tree.Oid
	}

	return eot.Entry.Oid
}

func (eot EntryOrTree) IsATree() bool {
	return eot.Tree != nil
}

func (eot EntryOrTree) isAnEntry() bool {
	return eot.Entry != nil
}

func newOrEntry(entry *Entry) EntryOrTree {
	return EntryOrTree{
		Entry: entry,
	}
}

func newOrTree(tree *Tree) EntryOrTree {
	return EntryOrTree{
		Tree: tree,
	}
}

func NewEntryMap() EntryMap {
	return EntryMap{
		EMap: make(map[string]EntryOrTree),
	}
}

func (em *EntryMap) SetEntry(key string, e *Entry) {
	_, exists := em.EMap[key]
	if !exists {
		em.EMap[key] = newOrEntry(e)
		em.keys = append(em.keys, key)
	}
}

func (em *EntryMap) SetTree(key string, t *Tree) {
	_, exists := em.EMap[key]
	if !exists {
		em.EMap[key] = newOrTree(t)
		em.keys = append(em.keys, key)
	}
}

func (t *Tree) addEntry(parentDirs []string, entry Entry) {
	if len(parentDirs) == 0 {
		baseName := filepath.Base(entry.Name)
		t.EntryMap.SetEntry(baseName, &entry)
		return
	}

	parent := parentDirs[0]
	parentBase := filepath.Base(parent)
	var tree *Tree
	if existingEntry, exists := t.EntryMap.EMap[parentBase]; exists {
		tree = existingEntry.Tree
	} else {
		tree = NewTree()
	}

	t.EntryMap.SetTree(parentBase, tree)
	tree.addEntry(parentDirs[1:], entry)
}

func (t *Tree) Traverse(action func(t *Tree)) {
	eMap := t.EntryMap.EMap
	for _, k := range t.EntryMap.keys {
		entry := eMap[k]
		if entry.IsATree() {
			entry.Tree.Traverse(action)
		}
	}

	action(t)
}

func (t *Tree) GetEntries() []Entry {
	return t.Entries
}

func BuildTree(entries []Entry) *Tree {
	root := NewTree()

	for _, entry := range entries {
		root.addEntry(entry.getParentDirectories(), entry)
	}

	return root
}

func NewTree() *Tree {
	return &Tree{
		EntryMap: NewEntryMap(),
	}
}
