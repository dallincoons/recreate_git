package index

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const REGULAR_MODE = 0100644
const EXECUTABLE_MODE = 0100755
const MAX_PATH_SIZE = 0xfff
const ENTRY_BLOCK_SIZE = 8
const ENTRY_MIN_SIZE = 64

type Entry struct {
	ctime      int64
	ctime_nsec int64
	mtime      int64
	mtime_nsec int64
	dev        int32
	ino        uint64
	Mode       int32
	uid        uint32
	gid        uint32
	size       int64
	Name       string
	Oid        string
	flags      uint16
	Path       string
}

func (e Entry) octalMode() interface{} {
	return fmt.Sprintf("%o", e.Mode)
}

func (e *Entry) getParentDirectories() []string {
	parentDirs := []string{}

	dir := filepath.Dir(e.Path)
	parts := strings.Split(dir, string(os.PathSeparator))
	fullString := ""

	for _, part := range parts {
		if part == "." {
			continue
		}
		fullString = filepath.Join(fullString, part)
		parentDirs = append(parentDirs, fullString)
	}

	return parentDirs
}

func (e *Entry) TimeMatches(stat *Stat) bool {
	return e.ctime == stat.CtimeSec && e.ctime_nsec == stat.CtimeNsec &&
		e.mtime == stat.MtimeSec && e.mtime_nsec == stat.MtimeSec
}

func (e *Entry) UpdateStat(stat *Stat) {
	//mode, _ := strconv.Atoi(fmt.Sprintf("%o", GetStatMode(stat.Mode)))

	e.ctime = stat.CtimeSec
	e.ctime_nsec = stat.CtimeNsec
	e.mtime = stat.MtimeSec
	e.mtime_nsec = stat.MtimeSec
	e.dev = stat.Dev
	e.ino = stat.Ino
	e.Mode = int32(stat.Mode)
	e.uid = stat.Uid
	e.gid = stat.Gid
	e.size = stat.Size
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (e *Entry) StatMatches(stat os.FileInfo) bool {
	fileInfo := ToFileInfo(stat)

	return e.Mode == int32(fileInfo.Mode) && (e.size == 0 || e.size == stat.Size())
}

func (e *Entry) ToString() string {
	oid, _ := hex.DecodeString(e.Oid)
	flag := min(len(e.Path), MAX_PATH_SIZE)

	str := fmt.Sprintf("%08x%08x%08x%08x%08x%08x%08x%08x%08x%08x%08x%04x%x", e.ctime, e.ctime_nsec, e.mtime, e.mtime_nsec, e.dev, e.ino, e.Mode, e.uid, e.gid, e.size, oid, flag, e.Path)
	return str
}

func parse(data []byte) Entry {
	//fmt.Println("about to parse full")
	//fmt.Println(data)
	ctime := int64(binary.BigEndian.Uint32(data[0:4]))
	//fmt.Println("read ctime")
	//fmt.Println(ctime)
	ctime_nsec := int64(binary.BigEndian.Uint32(data[4:8]))
	//fmt.Println("read ctime_nsec")
	//fmt.Println(ctime_nsec)
	mtime := int64(binary.BigEndian.Uint32(data[8:12]))
	//fmt.Println("read mtime")
	//fmt.Println(mtime)
	mtime_nsec := int64(binary.BigEndian.Uint32(data[12:16]))
	//fmt.Println("read mtime_nsec")
	//fmt.Println(mtime_nsec)
	dev := int32(binary.BigEndian.Uint32(data[16:20]))
	//fmt.Println("read dev")
	//fmt.Println(dev)
	ino := uint64(binary.BigEndian.Uint32(data[20:24]))
	//fmt.Println("read ino")
	//fmt.Println(ino)

	mode := int32(binary.BigEndian.Uint32(data[24:28]))
	//fmt.Println("read Mode")
	//fmt.Println(mode)

	uid := binary.BigEndian.Uint32(data[28:32])
	//fmt.Println("read uid")
	//fmt.Println(uid)
	gid := binary.BigEndian.Uint32(data[32:36])
	//fmt.Println("read gid")
	//fmt.Println(gid)
	size := int64(binary.BigEndian.Uint32(data[36:40]))
	//fmt.Println("read size")
	//fmt.Println(size)
	oid := hex.EncodeToString(data[40:60])
	//fmt.Println("read Oid")
	//fmt.Println(oid)
	flags, _ := strconv.Atoi(string(data[60:62]))
	//fmt.Println("read flags")
	//fmt.Println(flags)
	pathCandidate := data[62:]
	path := make([]byte, 0)
	for _, pathByte := range pathCandidate {
		if pathByte != '\000' {
			path = append(path, pathByte)
		}
	}

	return Entry{
		ctime:      ctime,
		ctime_nsec: ctime_nsec,
		mtime:      mtime,
		mtime_nsec: mtime_nsec,
		dev:        dev,
		ino:        ino,
		Mode:       int32(mode),
		uid:        uid,
		gid:        gid,
		size:       size,
		Oid:        oid,
		flags:      uint16(flags),
		Path:       string(path),
	}
}

type FileInfo interface {
	Stat() *Stat
}

type Stat struct {
	fileinfo  os.FileInfo
	CtimeSec  int64
	CtimeNsec int64
	MtimeSec  int64
	MtimNsec  int64
	Uid       uint32
	Dev       int32
	Ino       uint64
	Gid       uint32
	Size      int64
	Mode      uint16
	Flags     uint32
	Path      string
}

func (s *Stat) Stat() *Stat {
	return s
}

func NewEntry(path, oid string, fi FileInfo) Entry {
	stat := fi.Stat()

	//mode, _ := strconv.Atoi(fmt.Sprintf("%o", GetStatMode(stat.Mode)))

	return Entry{
		ctime:      stat.CtimeSec,
		ctime_nsec: stat.CtimeNsec,
		mtime:      stat.MtimeSec,
		mtime_nsec: stat.MtimNsec,
		dev:        stat.Dev,
		ino:        stat.Ino,
		Mode:       int32(stat.Mode),
		uid:        stat.Uid,
		gid:        stat.Gid,
		size:       stat.Size,
		Oid:        oid,
		flags:      uint16(stat.Flags),
		Path:       path,
	}
}

func GetStatMode(statMode uint16) int32 {
	var mode int32
	if statMode&0111 != 0 {
		mode = EXECUTABLE_MODE
	} else {
		mode = REGULAR_MODE
	}

	return mode
}
