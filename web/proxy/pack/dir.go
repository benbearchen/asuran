package pack

import (
	"github.com/benbearchen/asuran/util"

	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

type packList []*Pack
type packMap map[string]packList

type Dir struct {
	dir   string
	packs packMap

	tick int64
}

func New(dir string) *Dir {
	d := new(Dir)
	d.dir = dir
	d.packs = make(packMap)

	d.mkdir()
	d.load()

	return d
}

func (d *Dir) ListNames() []string {
	names := make([]string, 0, len(d.packs))
	for k, _ := range d.packs {
		names = append(names, k)
	}

	sort.Strings(names)
	return names
}

func (d *Dir) mkdir() error {
	return util.MakeDir(d.dir)
}

func (d *Dir) load() error {
	dir, err := os.Open(d.dir)
	if err != nil {
		return err
	}

	files, err := dir.Readdir(0)
	if len(files) == 0 {
		return err
	}

	packs := make(packMap)
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		path := filepath.Join(d.dir, f.Name())
		p, err := loadPack(path)
		if err == nil {
			packs.Add(p)
		}
	}

	d.packs = packs
	return nil // or, err?
}

func (d *Dir) Save(name, author, comment string, cmd string) error {
	if !VerifyName(name) {
		return fmt.Errorf("name(%q) contains invalid chars(%q)", name, INVALID_NAME_CHARS)
	}

	if !VerifyName(author) {
		return fmt.Errorf("author(%q) contains invalid chars(%q)", author, INVALID_NAME_CHARS)
	}

	if !VerifyComment(comment) {
		return fmt.Errorf("comment(%q) contains invalid chars(%q)", comment, INVALID_COMMENT_CHARS)
	}

	p := newPack(name, author, comment, cmd)
	d.packs.Add(p)

	filename := d.NameByTime(p.CreateTime().UnixNano())
	return p.WriteTo(filepath.Join(d.dir, filename), true)
}

func (d *Dir) NameByTime(nano int64) string {
	var minInterval int64 = 50000 // 50us
	if d.tick+minInterval < nano {
		d.tick = nano
	} else {
		d.tick += minInterval
	}

	return strconv.FormatInt(d.tick, 10)
}

func (d *Dir) GetPack(name string) *Pack {
	pl, ok := d.packs[name]
	if ok {
		return pl[len(pl)-1]
	}

	return nil
}

func (d *Dir) Get(name string) string {
	p := d.GetPack(name)
	if p != nil {
		return p.Command()
	}

	return ""
}

func (d *Dir) GetHistoryPack(name string, time int64) *Pack {
	pl, ok := d.packs[name]
	if !ok {
		return nil
	}

	for _, p := range pl {
		if p.CreateTime().UnixNano() == time {
			return p
		}
	}

	return nil
}

func (d *Dir) GetHistory(name string, time int64) string {
	p := d.GetHistoryPack(name, time)
	if p != nil {
		return p.Command()
	}

	return ""
}

func (m packMap) Add(p *Pack) {
	s, ok := m[p.Name()]
	if !ok {
		s = packList{p}
	} else {
		if p.CreateTime().After(s[len(s)-1].CreateTime()) {
			s = append(s, p)
		} else {
			i := len(s) - 1
			for i > 0 {
				if p.CreateTime().Before(s[i-1].CreateTime()) {
					i--
				} else {
					break
				}
			}

			c := make(packList, 0, len(s)+1)
			c = append(c, s[:i]...)
			c = append(c, p)
			c = append(c, s[i:]...)
			s = c
		}
	}

	m[p.Name()] = s
}
