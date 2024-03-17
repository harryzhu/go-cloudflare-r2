package cfr2

import (
	"crypto/md5"
	"encoding/hex"

	//"errors"
	"io"

	//"io/ioutil"

	"os"
	//"os"
)

func MD5File(p string) (s string, err error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	m := md5.New()
	io.Copy(m, f)
	return hex.EncodeToString(m.Sum(nil)), nil
}

func GetEnv(s string, d string) string {
	v := os.Getenv(s)
	if v != "" {
		return v
	} else {
		return d
	}
}
