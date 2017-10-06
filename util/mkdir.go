package util

import (
	"os"
)

func MakeDir(dir string) error {
	di, err := os.Stat(dir)
	create := false
	if err != nil {
		create = true
	} else if !di.IsDir() {
		os.Rename(dir, dir+".bak")
		create = true
	}

	if !create {
		return nil
	}

	return os.MkdirAll(dir, os.ModePerm)
}
