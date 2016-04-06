package pack

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	FIELD_NAME        = "name"
	FIELD_AUTHOR      = "author"
	FIELD_COMMENT     = "comment"
	FIELD_CREATE_TIME = "create"
)

type Pack struct {
	cmd string

	name    string
	author  string
	comment string
	create  time.Time
}

func newPack(name, author, comment string, cmd string) *Pack {
	p := new(Pack)
	p.cmd = cmd
	p.name = name
	p.author = author
	p.comment = comment
	p.create = time.Now()
	return p
}

func loadPack(path string) (*Pack, error) {
	p := new(Pack)
	err := p.load(path)
	return p, err
}

func (p *Pack) load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	file := string(b)
	lines := strings.Split(file, "\n")
	if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}

	cmd := ""
	titled := false
	titleOver := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) <= 3 || titleOver || line[:2] != "##" {
			if titled && !titleOver {
				titleOver = true // an empty line is needed after titles
			} else {
				cmd += line + "\n"
			}

			continue
		}

		kv := strings.SplitN(line[2:], ":", 2)
		if len(kv) != 2 {
			continue
		}

		titled = true

		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		switch k {
		case FIELD_NAME:
			p.name = v
		case FIELD_AUTHOR:
			p.author = v
		case FIELD_COMMENT:
			p.comment = v
		case FIELD_CREATE_TIME:
			u, err := strconv.ParseInt(v, 10, 64)
			if err == nil {
				p.create = time.Unix(0, u)
			}
		}
	}

	p.cmd = cmd
	return nil
}

func (p *Pack) Name() string {
	return p.name
}

func (p *Pack) Author() string {
	return p.author
}

func (p *Pack) Comment() string {
	return p.comment
}

func (p *Pack) CreateTime() time.Time {
	return p.create
}

func (p *Pack) Command() string {
	return p.cmd
}

func (p *Pack) File() string {
	header := ""
	addHeader := func(h, v string) {
		header += "## " + h + ": " + v + "\n"
	}

	addHeader(FIELD_NAME, p.name)
	addHeader(FIELD_AUTHOR, p.author)
	addHeader(FIELD_COMMENT, p.comment)

	addHeader(FIELD_CREATE_TIME, strconv.FormatInt(p.create.UnixNano(), 10))

	header += "\n"

	return header + p.cmd
}

func (p *Pack) WriteTo(path string, readonly bool) error {
	var filemode os.FileMode = 0666
	if readonly {
		filemode = 0444
	}

	return ioutil.WriteFile(path, []byte(p.File()), filemode)
}

const INVALID_NAME_CHARS = "'\"&=?#:%/\\\r\n \t"

func VerifyName(name string) bool {
	p := strings.IndexAny(name, INVALID_NAME_CHARS)
	return p < 0
}

const INVALID_COMMENT_CHARS = "\r\n\t"

func VerifyComment(comment string) bool {
	p := strings.IndexAny(comment, INVALID_COMMENT_CHARS)
	return p < 0
}
