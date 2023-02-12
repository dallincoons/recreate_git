package app

import (
	"bytes"
	"fmt"
	"strings"
)

type Commit struct {
	Oid      string
	ParentID string
	TreeOid  string
	author   *Author
	message  string
}

func (c *Commit) Type() string {
	return "commit"
}

func (c *Commit) ToString() string {
	lines := []string{
		fmt.Sprintf("tree %s", c.GetTreeOid()),
		fmt.Sprintf("author %s", c.author.ToString()),
		fmt.Sprintf("committer %s", c.author.ToString()),
		"",
		c.message,
	}

	if c.ParentID != "" {
		lines[1] = fmt.Sprintf("parent %s", c.ParentID)
	}

	return strings.Join(lines, "\n")
}

func (c Commit) GetTreeOid() string {
	return c.TreeOid
}

func (c Commit) parse(contents []byte) Object {
	parsed := make(map[string]string)
	for _, line := range strings.Split(string(contents), "\n") {
		if line == "" {
			break
		}
		splitLine := strings.SplitN(line, " ", 2)
		if len(splitLine) == 2 {
			key := bytes.Trim([]byte(splitLine[0]), "\000")
			parsed[string(key)] = splitLine[1]
		}
	}

	parentID := ""
	if _, parentExists := parsed["parent"]; parentExists {
		parentID = parsed["parent"]
	}

	return NewCommit(parentID, parsed["tree"], nil, "")
}

func (c *Commit) GetOid() string {
	return c.Oid
}

func (c *Commit) SetOid(oid string) {
	c.Oid = oid
}

func NewCommit(parentID, treeOid string, author *Author, message string) *Commit {
	return &Commit{
		ParentID: parentID,
		TreeOid:  treeOid,
		author:   author,
		message:  message,
	}
}
