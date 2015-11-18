package pack

import "testing"

import (
	"os"
)

func TestPack(t *testing.T) {
	path1 := "p.test"
	defer os.Remove(path1)

	p := newPack("a", "b", "cmd\n")
	err := p.WriteTo(path1)
	if err != nil {
		t.Errorf("write pack failed: %v", err)
	}

	r, err := loadPack(path1)
	if err != nil {
		t.Errorf("read pack failed: %v", err)
	}

	if r.Name() != p.Name() || r.Author() != p.Author() || !r.CreateTime().Equal(p.CreateTime()) || r.Command() != p.Command() {
		t.Errorf("write&load failed: %v != %v", p, r)
	}
}
