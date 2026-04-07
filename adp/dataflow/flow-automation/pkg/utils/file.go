package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// DagReader
type DagReader interface {
	ReadPathsFromDir(dir string) ([]string, error)
	ReadDag(path string) ([]byte, error)
}

var (
	DefaultReader DagReader = &FileDagReader{}
)

// FileDagReader
type FileDagReader struct {
}

// ReadPathsFromDir
func (r FileDagReader) ReadPathsFromDir(dir string) (dagFiles []string, err error) {
	if err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		ext := filepath.Ext(path)
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		dagFiles = append(dagFiles, path)
		return nil
	}); err != nil {
		return
	}

	return
}

// ReadDag
func (r FileDagReader) ReadDag(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}
