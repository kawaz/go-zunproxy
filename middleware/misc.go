package middleware

import (
	"errors"
	"io"
	"os"
	"path"
	"syscall"
)

func Close(c io.Closer) {
	if c != nil {
		c.Close()
	}
}

func CreateFile(dir string, file string) (*os.File, error) {
	bodyFile, err := os.Create(path.Join(dir, file))
	if err != nil && errors.Is(err, syscall.ENOENT) {

		err = os.MkdirAll(dir, os.ModePerm)
		if err == nil {
			bodyFile, err = os.Create(path.Join(dir, file))
		}
	}
	return bodyFile, err
}
