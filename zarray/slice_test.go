//go:build go1.18
// +build go1.18

package zarray_test

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zarray"
	"github.com/sohaha/zlsgo/ztype"
)

var l = []int{0, 1, 2, 3, 4, 5}
var l2 = []int{0, 1, 2, 3, 4, 5, 2, 34, 5, 6, 7, 98, 6, 67, 54, 543, 345, 435, 43543, 435, 3, 2, 42, 3423, 54, 6, 5}

func TestShuffle(t *testing.T) {
	t.Log(zarray.Shuffle(l))
	t.Log(zarray.Shuffle(l2))
}

func TestRand(t *testing.T) {
	t.Log(zarray.Rand(l))
}

func TestFilter(t *testing.T) {
	tt := zlsgo.NewTest(t)
	nl := zarray.Filter(l, func(index int, item int) bool {
		t.Log(index, item)
		return item%2 == 0
	})
	tt.Equal([]int{0, 2, 4}, nl)
}

func TestMap(t *testing.T) {
	tt := zlsgo.NewTest(t)
	s := []int{1, 2, 3}
	nl := zarray.Map(s, func(i int, v int) string {
		return ztype.ToString(v) + "//"
	})
	tt.Equal([]string{"1//", "2//", "3//"}, nl)
}

func TestDiff(t *testing.T) {
	tt := zlsgo.NewTest(t)

	n1, n2 := zarray.Diff(l2, l)

	tt.Equal([]int{34, 6, 7, 98, 6, 67, 54, 543, 345, 435, 43543, 435, 42, 3423, 54, 6}, n1)
	tt.Equal([]int{}, n2)
}

func TestContains(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tt.EqualTrue(!zarray.Contains(l, 54))
	tt.EqualTrue(!zarray.Contains(l, 6))
	tt.EqualTrue(zarray.Contains(l2, 5))
	tt.EqualTrue(zarray.Contains(l2, 6))
	tt.EqualTrue(zarray.Contains(l2, 54))
}

func TestUnique(t *testing.T) {
	tt := zlsgo.NewTest(t)
	a := append(l, l2...)
	unia := zarray.Unique(a)
	tt.Equal(18, len(unia))
	tt.EqualTrue(len(a) != len(unia))
	t.Log(unia)
}

func TestFind(t *testing.T) {
	tt := zlsgo.NewTest(t)
	a := []map[string]string{
		{"name": "a"},
		{"name": "b"},
		{"name": "c"},
	}

	v, ok := zarray.Find(a, func(_ int, v map[string]string) bool {
		return v["name"] == "b"
	})
	tt.EqualTrue(ok)
	tt.Equal("b", v["name"])

	v, ok = zarray.Find(a, func(_ int, v map[string]string) bool {
		return v["name"] == "z"
	})
	tt.EqualTrue(!ok)
	tt.Equal("", v["name"])
}
