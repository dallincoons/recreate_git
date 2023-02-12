package diff

import (
	"fmt"
	"jit/format"
)

type Myers struct {
	a []Line
	b []Line
}

func NewMyers(a, b []Line) *Myers {
	return &Myers{
		a: a,
		b: b,
	}
}

func (e *Edit) ToString() string {
	var edit string
	if e.aLine != nil {
		edit = e.aLine.Text
	} else if e.bLine != nil {
		edit = e.bLine.Text
	}

	switch e.kind {
	case EditTypeEqual:
		edit = fmt.Sprintf("%s%s", e.getDisplaySymbol(), edit)
	case EditTypeInsert:
		edit = fmt.Sprintf("%s%s", e.getDisplaySymbol(), format.Color("green", edit))
	case EditTypeDelete:
		edit = fmt.Sprintf("%s%s", e.getDisplaySymbol(), format.Color("red", edit))
	}

	return edit
}

func (e *Edit) getDisplaySymbol() string {
	if e.kind == EditTypeInsert {
		return "+"
	}

	if e.kind == EditTypeDelete {
		return "-"
	}

	return " "
}

func (m *Myers) diff() []Edit {
	diff := make([]Edit, 0)

	m.backtrack(func(prev_x, prev_y, x, y int) {
		var bLine *Line
		var aLine *Line

		if len(m.b) > 0 {
			bLine = &m.b[getPositiveIndex(len(m.b), prev_y)]
		}

		if len(m.a) > 0 {
			aLine = &m.a[getPositiveIndex(len(m.a), prev_x)]
		}

		if x == prev_x {
			diff = append(diff, NewEdit(EditTypeInsert, nil, bLine))
		} else if y == prev_y {
			diff = append(diff, NewEdit(EditTypeDelete, aLine, nil))
		} else {
			diff = append(diff, NewEdit(EditTypeEqual, aLine, bLine))
		}
	})

	return reverse(diff)
}

func reverse(arr []Edit) []Edit {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}

	return arr
}

func (m *Myers) backtrack(yield func(prev_x, prev_y, x, y int)) {
	x, y := len(m.a), len(m.b)

	var prev_k, prev_x, prev_y int

	steps := m.shortestEdit()
	for d := len(steps) - 1; d > 0; d-- {
		v := steps[d]

		k := x - y

		if k == -d || (k != d && v[getPositiveIndex(len(v), k-1)] < v[getPositiveIndex(len(v), k+1)]) {
			prev_k = k + 1
		} else {
			prev_k = k - 1
		}

		prev_x = v[getPositiveIndex(len(v), prev_k)]
		prev_y = prev_x - prev_k

		for x > prev_x && y > prev_y {
			//make a diagonal move backward
			yield(x-1, y-1, x, y)
			x, y = x-1, y-1
		}

		if d > 0 {
			yield(prev_x, prev_y, x, y)
		}
		x, y = prev_x, prev_y
	}
}

func (meyers *Myers) shortestEdit() [][]int {
	n, m := len(meyers.a), len(meyers.b)
	max := n + m

	v := make([]int, (2*max)+1)
	trace := make([][]int, 0)

	var x int
	var y int
	for d := 0; d < max; d++ {
		vCopy := make([]int, len(v))
		copy(vCopy, v)
		trace = append(trace, vCopy)
		//if k equals negative d then we must have moved down to get there
		//moving down means keeping x the same as before (y will be recalculated as x - k)
		for k := -d; k <= d; k += 2 {
			if k == -d || (k != d && v[getPositiveIndex(len(v), k-1)] < v[getPositiveIndex(len(v), k+1)]) {
				x = v[getPositiveIndex(len(v), k+1)]
			} else {
				x = v[getPositiveIndex(len(v), k-1)] + 1
			}

			y = x - k

			for x < n && y < m && meyers.a[x].Text == meyers.b[y].Text {
				x, y = x+1, y+1
			}

			v[getPositiveIndex(len(v), k)] = x

			if x >= n && y >= m {
				return trace
			}
		}
	}

	return trace
}

func getPositiveIndex(v int, i int) int {
	var ki int
	if i < 0 {
		ki = v + i
	} else {
		ki = i
	}

	if ki >= v {
		return 0
	}

	return ki
}
