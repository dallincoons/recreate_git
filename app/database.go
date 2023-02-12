package app

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

const TEMP_CHARS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Database struct {
	path    string
	objects map[string]*Object
}

type DBEntry struct {
	Oid  string
	Mode int32
}

func (db Database) Load(oid string) *Object {
	if _, exists := db.objects[oid]; !exists {
		object, err := db.readObject(oid)
		if err != nil {
			fmt.Println("error reading object", err)
			return nil
		}
		db.objects[oid] = object
	}

	return db.objects[oid]
}

func (db *Database) ShortOid(oid string) string {
	return oid[0:6]
}

func (db *Database) readObject(oid string) (*Object, error) {
	data, err := os.ReadFile(db.objectPath(oid))
	if err != nil {
		return nil, err
	}
	b := bytes.NewReader(data)
	z, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	contents, err := io.ReadAll(z)
	if err != nil {
		return nil, err
	}

	objType := string(contents[:strings.IndexRune(string(contents), ' ')])

	emptyObject := newObjectFromType(objType)
	object := emptyObject.parse(contents[strings.IndexRune(string(contents), '\x00'):])
	object.SetOid(oid)

	return &object, nil
}

type Parseable interface {
	parse(contents []byte) Object
	SetOid(oid string)
}

func newObjectFromType(objType string) Parseable {
	switch objType {
	case "blob":
		return &Blob{}
	case "tree":
		return &Tree{}
	case "commit":
		return &Commit{}
	}

	return nil
}

func (db *Database) Store(obj Object) {
	content := db.SerializeObject(obj)
	oid := db.HashContent(content)
	obj.SetOid(oid)

	db.writeObject(oid, content)
}

func (db *Database) SerializeObject(obj Object) string {
	str := obj.ToString()
	content := fmt.Sprintf("%s %d\000%s", obj.Type(), len(str), str)
	return content
}

func (db *Database) HashObject(obj Object) string {
	return db.HashContent(db.SerializeObject(obj))
}

func (db *Database) HashContent(content string) string {
	hasher := sha1.New()
	hasher.Write([]byte(content))
	oid := hex.EncodeToString(hasher.Sum(nil))
	return oid
}

func (db *Database) writeObject(oid string, content string) {
	objectPath := db.objectPath(oid)
	if _, err := os.Stat(objectPath); !errors.Is(err, os.ErrNotExist) {
		fmt.Println("object already exists", objectPath)
		return
	}

	dirName := filepath.Dir(objectPath)

	tmpName := db.generateTempName()
	tmpPath := filepath.Join(dirName, tmpName)

	file, err := os.OpenFile(tmpPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0777)
	if err != nil {
		err := os.MkdirAll(dirName, 0777)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		file, err = os.OpenFile(tmpPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0777)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	var in bytes.Buffer
	b := []byte(content)
	w := zlib.NewWriter(&in)
	w.Write(b)
	w.Close()

	_, err = file.Write(in.Bytes())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = os.Rename(file.Name(), objectPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (db *Database) objectPath(oid string) string {
	return filepath.Join(db.path, oid[0:2], oid[2:])
}

func (db *Database) generateTempName() string {
	name := ""
	for i := 1; i <= 6; i++ {
		randomIndex := rand.Intn(len(TEMP_CHARS))
		name += string(TEMP_CHARS[randomIndex])
	}

	return name
}

func NewDatabase(path string) Database {
	return Database{
		path:    path,
		objects: make(map[string]*Object),
	}
}
