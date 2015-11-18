package pack

import "testing"

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func TestPackMap(t *testing.T) {
	name := "a"
	p1 := newPack(name, "1", "cmd\n")
	p2 := newPack(name, "2", "cmd2j\n")
	p3 := newPack(name, "3", "cmd3\n")
	p4 := newPack(name, "4", "cmd4\n")
	p5 := newPack(name, "5", "cmd5\n")

	p2.create = p2.create.Add(time.Second * 2)
	p3.create = p3.create.Add(-time.Second * 2)
	p4.create = p4.create.Add(time.Second * 1)
	p5.create = p1.create

	m := make(packMap)
	m.Add(p1)
	if m[name][0] != p1 {
		t.Errorf("TestPackMap Add(p1) wrong pos, %v", m)
	}

	m.Add(p2)
	if m[name][1] != p2 {
		t.Errorf("TestPackMap Add(p2) wrong pos, %v", m)
	}

	m.Add(p3)
	if m[name][0] != p3 {
		t.Errorf("TestPackMap Add(p3) wrong pos, %v", m)
	}

	m.Add(p4)
	if m[name][2] != p4 {
		t.Errorf("TestPackMap Add(p4) wrong pos, %v", m)
	}

	m.Add(p5)
	if m[name][2] != p5 {
		t.Errorf("TestPackMap Add(p5) wrong pos, %v", m)
	}

	s := m[name]
	if s[0] != p3 || s[1] != p1 || s[2] != p5 || s[3] != p4 || s[4] != p2 {
		t.Errorf("TestPackMap final 5 wrong")
	}
}

func TestDir(t *testing.T) {
	dirname := "crazy-tmp-dir-name-aha-and-very-loong"
	packname := "a"

	d := new(Dir)
	d.dir = dirname

	defer os.RemoveAll(d.dir)

	err := os.RemoveAll(d.dir)
	if err != nil {
		t.Errorf("TestDir clean path %s failed: %v", d.dir, err)
	}

	err = d.mkdir()
	if err != nil {
		t.Errorf("TestDir d.mkdir() failed: %v", err)
	}

	ps := packList{}
	for i := 0; i < 5; i++ {
		p := newPack(packname, strconv.Itoa(i+1), "cmd\n")
		ps = append(ps, p)
		timeoffset := (i*97 + 5) % (7 + len(ps))
		p.create = p.create.Add(time.Second * time.Duration(timeoffset))
		err = p.WriteTo(filepath.Join(d.dir, p.Author()))
		if err != nil {
			t.Errorf("TestDir write pack(%v) failed: %v", p, err)
		}
	}

	err = d.load()
	if err != nil {
		t.Errorf("TestDir d.load() failed: %v", err)
	}

	s := d.packs[packname]
	if len(s) != len(ps) {
		t.Errorf("TestDir d.load() count wrong: %d != %d", len(s), len(ps))
	}

	for i := 1; i < len(s); i++ {
		if s[i-1].create.After(s[i].create) {
			t.Errorf("TestDir d.load() wrong time seq [%d].create(%v) vs [%d].create(%v)", i-1, s[i-1].create, i, s[i].create)
		}
	}
}
