package utils

import (
	"io"
	"os"
)

// AutoCleanTempFile wraps a temp file that auto-deletes on Close
type AutoCleanTempFile struct {
	*os.File
}

// Close closes the file and removes it from disk
func (f *AutoCleanTempFile) Close() error {
	name := f.Name()
	err := f.File.Close()
	_ = os.Remove(name)
	return err
}

// BufferToTempFile copies reader content to a temp file with given prefix.
// Returns a ReadCloser that auto-deletes the file on Close, along with the
// number of bytes written.
func BufferToTempFile(r io.Reader, prefix string) (io.ReadCloser, int64, error) {
	tmp, err := os.CreateTemp("", prefix+"-*")
	if err != nil {
		return nil, 0, err
	}

	written, err := io.Copy(tmp, r)
	if err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return nil, 0, err
	}

	if _, err = tmp.Seek(0, io.SeekStart); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return nil, 0, err
	}

	return &AutoCleanTempFile{File: tmp}, written, nil
}