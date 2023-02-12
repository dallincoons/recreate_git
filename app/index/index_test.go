package index

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"testing"
)

type FakeFileInfo struct {
}

func (ffi *FakeFileInfo) Stat() *Stat {
	return &Stat{}
}

func TestAddEntryToIndex(t *testing.T) {
	index := NewIndex("blah blah")
	fileInfo := &FakeFileInfo{}

	index.Add("blah.txt", "1234", fileInfo)

	entries := index.SortedEntries()

	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "blah.txt", entries[0].Path)
}

func TestAddReplaceFileWithDirectory(t *testing.T) {
	index := NewIndex("blah blah")
	fileInfo := &FakeFileInfo{}

	index.Add("alice.txt", randomHex(20), fileInfo)
	index.Add("nested/bob.txt", randomHex(20), fileInfo)
	index.Add("alice.txt/nested.txt", randomHex(20), fileInfo)

	entries := index.SortedEntries()

	assert.Equal(t, 2, len(entries))
	assert.Equal(t, "alice.txt/nested.txt", entries[0].Path)
	assert.Equal(t, "nested/bob.txt", entries[1].Path)
}

func TestAddReplaceDirectoryWithFile(t *testing.T) {
	index := NewIndex("blah blah")
	fileInfo := &FakeFileInfo{}

	index.Add("alice.txt", randomHex(20), fileInfo)
	index.Add("nested/bob.txt", randomHex(20), fileInfo)

	index.Add("nested", randomHex(20), fileInfo)

	entries := index.SortedEntries()

	assert.Equal(t, 2, len(entries))
	assert.Equal(t, "alice.txt", entries[0].Path)
	assert.Equal(t, "nested", entries[1].Path)
}

func TestAddRecursivelyReplaceDirectoryWithFile(t *testing.T) {
	index := NewIndex("blah blah")
	fileInfo := &FakeFileInfo{}

	index.Add("alice.txt", randomHex(20), fileInfo)
	index.Add("nested/bob.txt", randomHex(20), fileInfo)
	index.Add("nested/inner/claire.txt", randomHex(20), fileInfo)

	index.Add("nested", randomHex(20), fileInfo)

	entries := index.SortedEntries()

	assert.Equal(t, 2, len(entries))
	assert.Equal(t, "alice.txt", entries[0].Path)
	assert.Equal(t, "nested", entries[1].Path)
}

func randomHex(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}
