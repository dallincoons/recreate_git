package app

import (
	"errors"
	"fmt"
	"os"
)

type LockFile struct {
	filePath string
	lockPath string
	lock     *os.File
}

func (lf *LockFile) GetFile() *os.File {
	return lf.lock
}

func (lf *LockFile) HoldForUpdate() error {
	if lf.lock != nil {
		return errors.New("lock file already exists")
	}

	file, err := os.OpenFile(lf.lockPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0777)
	if err != nil {
		return err
	}
	lf.lock = file
	return nil
}

func (lf *LockFile) Rollback() error {
	if lf.lock == nil {
		return errors.New(fmt.Sprintf("Not holding lock on file: %s", lf.lockPath))
	}

	lf.lock.Close()
	os.Remove(lf.lockPath)
	lf.lock = nil

	return nil
}

func (lf *LockFile) Read(b []byte) (int, error) {
	return lf.lock.Read(b)
}

func (lf *LockFile) Write(contents string) error {
	if lf.lock == nil {
		return errors.New(fmt.Sprintf("Not holding lock on file: %s", lf.lockPath))
	}
	_, err := lf.lock.WriteString(contents)
	if err != nil {
		fmt.Println("error writing to lock file")
		fmt.Println(err)
		os.Exit(1)
	}

	return nil
}

func (lf *LockFile) Commit() error {
	if lf.lock == nil {
		return errors.New(fmt.Sprintf("Not holding lock on file: %s", lf.lockPath))
	}
	lf.lock.Close()
	err := os.Rename(lf.lockPath, lf.filePath)
	if err != nil {
		fmt.Println("error renaming file")
		fmt.Println(err)
		os.Exit(1)
	}
	lf.lock = nil

	return nil
}

func NewLockFile(filePath string) *LockFile {
	return &LockFile{
		filePath: filePath,
		lockPath: fmt.Sprintf("%s%s", filePath, ".lock"),
	}
}
