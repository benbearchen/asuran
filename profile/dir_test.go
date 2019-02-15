package profile

import "testing"

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func TestSprintfX(t *testing.T) {
	a := fmt.Sprintf("%x", [32]byte{})
	b := fmt.Sprintf("%02x", [32]byte{})
	if a != b {
		t.Errorf("hex %s != %s", a, b)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := fmt.Sprintf(".test-%x", time.Now().UnixNano())
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		t.Fatalf("can't create test dir `%s': %v", dir, err)
	}

	defer os.RemoveAll(dir)

	d := NewProfileRootDir(dir)
	ip := "123"
	messages := []string{
		"",
		"hello world\n",
		"abcd\nkajsdfasdf\n",
		"1\n",
	}

	for _, msg := range messages {
		err = d.Save(ip, msg)
		if err != nil {
			t.Errorf("save msg `%s' fail: %v", msg, err)
			continue
		}

		c, err := d.Load(ip)
		if err != nil {
			t.Errorf("load msg for `%s' fail: %v", msg, err)
			continue
		}

		if c != msg {
			t.Errorf("save msg `%s' and load changed: `%s'", msg, c)
			continue
		}
	}

	nc := len(messages)
	if nc > len(d.nameList) {
		nc = len(d.nameList)
	}

	subdir, err := d.makeProfileDir(ip)
	if err != nil {
		t.Errorf("get profile `%s' dir fail: %v", ip, err)
	} else {
		for i := 0; i < nc; i++ {
			msg := messages[len(messages)-1-i]
			name := d.nameList[i]
			c, err := d.checksum(filepath.Join(subdir, name))
			if err != nil {
				t.Errorf("can't fetch content of `%s': %v", name, err)
			} else if string(c) != msg {
				t.Errorf("fetch content `%s' don't match `%s'", string(c), msg)
			}
		}
	}
}
