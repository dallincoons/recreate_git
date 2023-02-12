package index

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"jit/app"
	"os"
	"sort"
	"strconv"
	"syscall"
)

const HEADER_SIZE = 12
const SIGNATURE = "DIRC"
const VERSION = 2

type Index struct {
	path     string
	entries  map[string]Entry
	changed  bool
	lockfile *app.LockFile
	hash     hash.Hash
	paths    []string
	parents  map[string]map[string]bool
}

func (i *Index) GetPath() string {
	return i.path
}

func (i *Index) EntryForPath(path string) Entry {
	return i.entries[path]
}

func (i *Index) UpdateEntryStat(entry Entry, stat *Stat) {
	entry.UpdateStat(stat)
	i.changed = true
}

func (i *Index) Add(path, oid string, fileinfo FileInfo) {
	entry := NewEntry(path, oid, fileinfo)
	i.discardConflicts(entry)
	i.storeEntry(entry)
	i.changed = true
}

func (i *Index) discardConflicts(entry Entry) {
	for _, dir := range entry.getParentDirectories() {
		for idx, p := range i.paths {
			if p == dir {
				i.paths = append(i.paths[:idx], i.paths[idx+1:]...)
			}
		}
		delete(i.entries, dir)
	}
	i.removeChildren(entry.Path)
}

func (i *Index) removeChildren(path string) {
	if _, exists := i.parents[path]; !exists {
		return
	}

	for child := range i.parents[path] {
		i.removeEntry(child)
	}
}

func (i *Index) removeEntry(pathname string) {
	entry, exists := i.entries[pathname]
	if !exists {
		return
	}

	for idx, p := range i.paths {
		if p == entry.Path {
			i.paths = append(i.paths[:idx], i.paths[idx+1:]...)
		}
	}
	delete(i.entries, entry.Path)

	for _, dirname := range entry.getParentDirectories() {
		delete(i.parents[dirname], entry.Path)
		if len(i.parents[dirname]) == 0 {
			delete(i.parents, dirname)
		}
	}
}

func (i *Index) storeEntry(entry Entry) {
	if _, exists := i.entries[entry.Path]; exists == false {
		i.paths = append(i.paths, entry.Path)
	}
	i.entries[entry.Path] = entry

	for _, dir := range entry.getParentDirectories() {
		if i.parents[dir] == nil {
			i.parents[dir] = make(map[string]bool)
		}
		i.parents[dir][entry.Path] = true
	}
}

func (i *Index) SortedEntries() []Entry {
	sort.Strings(i.paths)
	sortedEntries := make([]Entry, 0)
	for _, key := range i.paths {
		sortedEntries = append(sortedEntries, i.entries[key])
	}

	return sortedEntries
}

func (i *Index) MarkChanged() {
	i.changed = true
}

func (i *Index) WriteUpdates() bool {
	if !i.changed {
		err := i.lockfile.Rollback()
		if err != nil {
			fmt.Println("failed to roll back")
			fmt.Println(err)
		}
		return false
	}

	writer := NewCheckSum(i.lockfile.GetFile())

	i.beginWrite()
	header, _ := hex.DecodeString("4449524300000002" + fmt.Sprintf("%08x", len(i.entries)))
	writer.write(header)
	for _, entry := range i.SortedEntries() {
		b, _ := hex.DecodeString(entry.ToString())
		b = append(b, '\000')
		for len(b)%ENTRY_BLOCK_SIZE != 0 {
			b = append(b, '\000')
		}
		writer.write(b)
	}

	writer.writeChecksum(writer.hash.Sum(nil))

	err := i.lockfile.Commit()
	if err != nil {
		fmt.Println("failed to commit index contents in lock file")
		fmt.Println(err)
		os.Exit(1)
	}

	i.changed = false
	return true
}

func (i *Index) beginWrite() {
	i.hash = sha1.New()
}

func (i *Index) write(data []byte) {
	err := i.lockfile.Write(string(data))
	if err != nil {
		fmt.Println("failed to write to lock file")
		fmt.Println(err)
		os.Exit(1)
	}
	io.WriteString(i.hash, string(data))
}

func (i *Index) finishWrite() {
	err := i.lockfile.Write(string(i.hash.Sum(nil)))
	if err != nil {
		fmt.Println("failed to write index contents hash to lock file")
		fmt.Println(err)
		os.Exit(1)
	}
	err = i.lockfile.Commit()
	if err != nil {
		fmt.Println("failed to commit index contents in lock file")
		fmt.Println(err)
		os.Exit(1)
	}
}

func (i *Index) ReleaseLock() {
	i.lockfile.Rollback()
}

func (i *Index) LoadForUpdate() error {
	if err := i.lockfile.HoldForUpdate(); err != nil {
		return err
	}
	i.Load()
	return nil
}

func (i *Index) Load() bool {
	i.clear()
	file := i.openIndexFile()
	defer file.Close()

	if file != nil {
		reader := NewCheckSum(file)
		count, err := i.readHeader(reader)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		i.readEntries(reader, count)
		err = reader.verifyChecksum()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	return true
}

func (i *Index) readHeader(checksum *Checksum) (int, error) {
	data, err := checksum.read(HEADER_SIZE)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	signature := string(data[0:4])
	version, err := strconv.Atoi(hex.EncodeToString(data[5:8]))
	if err != nil {
		return 0, err
	}

	entriesCount := int(binary.BigEndian.Uint32(data[8:12]))
	if err != nil {
		return 0, err
	}

	if string(signature) != SIGNATURE {
		return 0, errors.New(fmt.Sprintf("Signature expected %s but got %s", SIGNATURE, signature))
	}
	if version != VERSION {
		return 0, errors.New(fmt.Sprintf("Version expected %v but got %v", VERSION, version))
	}

	return entriesCount, nil
}

func (idx *Index) readEntries(checksum *Checksum, count int) {
	for i := 0; i < count; i++ {
		entryBytes, err := checksum.read(ENTRY_MIN_SIZE)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for entryBytes[len(entryBytes)-1] != '\000' {
			newBytes, err := checksum.read(ENTRY_BLOCK_SIZE)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			entryBytes = append(entryBytes, newBytes...)
		}

		idx.storeEntry(parse(entryBytes))
	}
}

func (i *Index) IsTracked(path string) bool {
	_, exists := i.entries[path]

	_, parentExists := i.parents[path]

	return exists || parentExists
}

func (i *Index) IsTrackedFile(path string) bool {
	_, exists := i.entries[path]

	return exists
}

func (i *Index) openIndexFile() *os.File {
	file, err := os.Open(i.path)
	if err != nil {
		return nil
	}
	return file
}

func NewIndex(path string) *Index {
	index := &Index{
		path:     path,
		lockfile: app.NewLockFile(path),
	}
	index.clear()
	return index
}

func (i *Index) clear() {
	i.entries = make(map[string]Entry)
	i.paths = []string{}
	i.changed = false
	i.parents = make(map[string]map[string]bool)
}

func ToFileInfo(osf os.FileInfo) *Stat {
	stat := osf.Sys().(*syscall.Stat_t)

	return &Stat{
		CtimeSec:  stat.Ctimespec.Sec,
		CtimeNsec: stat.Ctimespec.Nsec,
		MtimeSec:  stat.Mtimespec.Sec,
		MtimNsec:  stat.Mtimespec.Nsec,
		Dev:       stat.Dev,
		Ino:       stat.Ino,
		Uid:       stat.Uid,
		Gid:       stat.Gid,
		Size:      stat.Size,
		Mode:      stat.Mode,
		Flags:     stat.Flags,
	}
}
