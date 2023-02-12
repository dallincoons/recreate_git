package app

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Object interface {
	ToString() string
	Type() string
	SetOid(oid string)
}

func ObjToCommit(obj *Object) *Commit {
	bytes, err := json.Marshal(*obj)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	commit := &Commit{}
	err = json.Unmarshal(bytes, commit)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return commit
}

func ObjToTree(obj *Object) *Tree {
	bytes, err := json.Marshal(*obj)
	if err != nil {
		fmt.Println(err)
	}
	tree := &Tree{}
	err = json.Unmarshal(bytes, tree)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return tree
}

func ObjToBlob(obj *Object) *Blob {
	bytes, err := json.Marshal(*obj)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	blob := &Blob{}
	err = json.Unmarshal(bytes, blob)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return blob
}

type Blob struct {
	Data []byte
	Oid  string
}

func (b Blob) parse(contents []byte) Object {
	return NewBlob(contents)
}

func (b Blob) GetData() []byte {
	return b.Data
}

func (b Blob) GetLines() []string {
	return strings.Split(string(b.Data), "\n")
}

func (b *Blob) ToString() string {
	return string(b.Data[:])
}

func (*Blob) Type() string {
	return "blob"
}

func (b *Blob) SetOid(oid string) {
	b.Oid = oid
}

func NewBlob(data []byte) *Blob {
	return &Blob{
		Data: data,
	}
}
