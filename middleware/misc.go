package middleware

import (
	"encoding/base32"
	"encoding/base64"
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

// os.CreateFile, If there is no parent directory, create it.
func CreateFile(dir string, file string) (*os.File, error) {
	f, err := os.Create(path.Join(dir, file))
	if err != nil && errors.Is(err, syscall.ENOENT) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err == nil {
			f, err = os.Create(path.Join(dir, file))
		}
	}
	return f, err
}

// base32.StdEncoding + NoPadding
var Base32 = base32.StdEncoding.WithPadding(base32.NoPadding)

// base64.URLEncoding + NoPadding
var Base64 = base64.URLEncoding.WithPadding(base64.NoPadding)
