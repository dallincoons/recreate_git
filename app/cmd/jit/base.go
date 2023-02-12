package jit

import "io"

var pager *Pager

type Pager struct {
	stdout io.Writer
	stderr io.Writer
}

func setUpPager(stdout io.Writer, stderr io.Writer) {
	if pager != nil {
		return
	}

	pager = &Pager{stdout: stdout, stderr: stderr}
}
