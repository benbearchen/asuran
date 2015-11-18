package pack

import (
	"os"
	"path/filepath"
)

type packList []*Pack
type packMap map[string]packList

type Dir struct {
	dir   string
	packs packMap
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

	return names
}

func (d *Dir) mkdir() error {
	di, err := os.Stat(d.dir)
	create := false
	if err != nil {
		create = true
	} else if !di.IsDir() {
		os.Rename(d.dir, d.dir+".bak")
		create = true
	}

	if !create {
		return nil
	}

	return os.MkdirAll(d.dir, os.ModePerm)
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
