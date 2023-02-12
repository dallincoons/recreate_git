package app

import (
	"fmt"
	"strconv"
	"time"
)

type Author struct {
	name  string
	email string
	time  time.Time
}

func (a *Author) ToString() string {
	now := time.Now()
	tz := now.Format("-0700")
	return fmt.Sprintf("%s <%s> %s %s", a.name, a.email, strconv.FormatInt(now.Unix(), 10), tz)
}

func NewAuthor(name, email string, time time.Time) *Author {
	return &Author{
		name:  name,
		email: email,
		time:  time,
	}
}
