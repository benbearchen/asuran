package net

import "testing"

import (
	"runtime"
)

func TestShiftRightV4(t *testing.T) {
	check := func(a, b []string) {
		_, _, line, _ := runtime.Caller(1)

		if len(a) != len(b) {
			t.Errorf("TestShiftRightV4() L%d: input length unmatch: %d vs %d", line, len(a), len(b))
			return
		}

		c := ShiftRightV4(a)
		if len(c) != len(b) {
			t.Errorf("TestShiftRightV4() L%d: ShiftRightV4(%v) return len(%v) -> %d != %d <- len(%v)", line, a, c, len(c), len(b), b)
		} else {
			for i := range c {
				if c[i] != b[i] {
					t.Errorf("TestShiftRightV4() L%d: ShiftRightV4(%v) return [%d] %s != %s", line, a, i, c[i], b[i])
				}
			}
		}
	}

	check(
		[]string{},
		[]string{},
	)

	check(
		[]string{
			"",
		},
		[]string{
			"",
		},
	)

	check(
		[]string{
			"::1",
		},
		[]string{
			"::1",
		},
	)

	check(
		[]string{
			"192",
		},
		[]string{
			"192",
		},
	)

	check(
		[]string{
			"::1",
			"192",
		},
		[]string{
			"::1",
			"192",
		},
	)

	check(
		[]string{
			"192",
			"::1",
		},
		[]string{
			"::1",
			"192",
		},
	)

	check(
		[]string{
			"193",
			"::1",
			"192",
		},
		[]string{
			"::1",
			"193",
			"192",
		},
	)
}
