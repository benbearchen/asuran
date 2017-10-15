package profile

import (
	"github.com/benbearchen/asuran/util"

	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ProfileRootDir struct {
	dir      string
	nameList []string

	lock sync.Mutex
}

func NewProfileRootDir(dir string) *ProfileRootDir {
	d := new(ProfileRootDir)
	d.dir = dir
	d.nameList = []string{"cmds.cfg", "cmds-bak1.cfg", "cmds-bak2.cfg"}

	d.mkdir()

	return d
}

const (
	tag = "#checksum#"
)

func (d *ProfileRootDir) Save(ip, content string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	dir, err := d.makeProfileDir(ip)
	if err != nil {
		return err
	}

	d.moveOldsForSave(dir, d.nameList)

	if len(content) > 0 {
		c := content[len(content)-1]
		if c != '\n' {
			content += "\n"
		}
	}

	c := []byte(content)

	hash := fmt.Sprintf("%x", md5.Sum(c))

	suffix := make([]byte, 0, len(tag)+1+len(hash))
	suffix = append(suffix, tag...)
	suffix = append(suffix, ' ')
	suffix = append(suffix, []byte(hash)...)

	c = append(c, suffix...)

	var filemode os.FileMode = 0666
	path := filepath.Join(dir, d.nameList[0])
	return ioutil.WriteFile(path, c, filemode)
}

func (d *ProfileRootDir) Load(ip string) (string, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	dir, err := d.makeProfileDir(ip)
	if err != nil {
		return "", err
	}

	for _, name := range d.nameList {
		p := filepath.Join(dir, name)
		content, err := d.checksum(p)
		if err == nil {
			return string(content), nil
		}
	}

	return "", nil
}

func (d *ProfileRootDir) LoadHistory(ip string, index int) (string, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	dir, err := d.makeProfileDir(ip)
	if err != nil {
		return "", err
	}

	if index < 0 || index >= len(d.nameList) {
		return "", fmt.Errorf("无效历史序号 %d", index)
	}

	p := filepath.Join(dir, d.nameList[index])
	content, err := d.checksum(p)
	return string(content), err
}

func (d *ProfileRootDir) mkdir() error {
	return util.MakeDir(d.dir)
}

func (d *ProfileRootDir) profileDir(ip string) string {
	ip = strings.Replace(ip, ":", "-", -1)
	return filepath.Join(d.dir, ip)
}

func (d *ProfileRootDir) makeProfileDir(ip string) (string, error) {
	dir := d.profileDir(ip)
	return dir, util.MakeDir(dir)
}

func (d *ProfileRootDir) moveOldsForSave(dir string, names []string) {
	path0 := filepath.Join(dir, names[0])
	if _, err := d.checksum(path0); err != nil {
		return
	}

	i := 1
	for ; i < len(names); i++ {
		p := filepath.Join(dir, names[i])
		fi, err := os.Stat(p)
		if err != nil {
			break
		}

		if fi.Size() == 0 {
			os.Remove(p)
			break
		}
	}

	if i == len(names) {
		p := filepath.Join(dir, names[len(names)-1])
		os.Remove(p)
		i--
	}

	for ; i > 0; i-- {
		p1 := filepath.Join(dir, names[i-1])
		p2 := filepath.Join(dir, names[i])
		os.Rename(p1, p2)
	}
}

func (d *ProfileRootDir) checksum(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return d.checksumBytes(b)
}

func (d *ProfileRootDir) checksumBytes(c []byte) ([]byte, error) {
	p := bytes.LastIndex(c, []byte(tag))
	if p < 0 {
		return nil, fmt.Errorf("missing checksum tag")
	} else if p > 0 {
		prev := c[p-1]
		if prev != '\r' && prev != '\n' {
			return nil, fmt.Errorf("invalid checksum tag")
		}
	}

	content := c[:p]
	tagHash := strings.ToLower(strings.TrimSpace(string(c[p+len(tag):])))
	hash := fmt.Sprintf("%x", md5.Sum(content))
	if tagHash != hash {
		return nil, fmt.Errorf("checksum error")
	}

	return content, nil
}
