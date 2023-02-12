package index

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
)

const CHECKSUM_SIZE = 20

type Checksum struct {
	file *os.File
	hash hash.Hash
}

func NewCheckSum(file *os.File) *Checksum {
	return &Checksum{
		file: file,
		hash: sha1.New(),
	}
}

func (cs *Checksum) write(contents []byte) {
	_, err := cs.file.Write(contents)
	if err != nil {
		fmt.Println("failed to write to lock file")
		fmt.Println(err)
		os.Exit(1)
	}
	_, err = io.WriteString(cs.hash, string(contents))
	if err != nil {
		fmt.Println("failed to update hash")
		fmt.Println(err)
	}
}

func (cs *Checksum) writeChecksum(contents []byte) {
	_, err := cs.file.Write(contents)
	if err != nil {
		fmt.Println("failed to write index contents hash to lock file")
		fmt.Println(err)
		os.Exit(1)
	}
}

func (cs *Checksum) read(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := cs.file.Read(b)
	if err != nil {
		fmt.Println("error reading file")
		fmt.Println(err)
		os.Exit(1)
	}

	if len(b) != size {
		return nil, errors.New("Unexpected end of file while reading index.")
	}

	cs.hash.Sum(b)
	return b, nil
}

func (cs *Checksum) verifyChecksum() error {
	b := make([]byte, CHECKSUM_SIZE)
	_, err := cs.file.Read(b)
	if err != nil {
		return err
	}

	if len(b) != cs.hash.Size() {
		return errors.New("Checksum does not equal value stored on disk.")
	}

	return nil
}
