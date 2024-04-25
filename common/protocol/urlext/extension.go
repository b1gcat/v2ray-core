package apk

import (
	"bytes"
	"errors"
	"net/url"
	"strings"
)

var (
	URLExtension = "URL-extension"
)

type SniffHeader struct {
	Ext string
}

func (h *SniffHeader) Protocol() string {
	return URLExtension
}

func (h *SniffHeader) Domain() string {
	return ""
}

var errNotExt = errors.New("not URL Extension signature")

func SniffURLExtension(b []byte) (*SniffHeader, error) {
	if len(b) < 12 {
		return nil, errNotExt
	}

	if !bytes.Equal(b[:5], []byte("GET /")) {
		return nil, errNotExt
	}

	ext, err := getFileExtensionFromUrl(string(b))
	if err != nil {
		return nil, errNotExt
	}

	return &SniffHeader{Ext: ext}, nil
}

func getFileExtensionFromUrl(raw string) (string, error) {
	start := 4 //GET /
	fakeUrl := ""

	for k, v := range raw {
		if v == ' ' && k > 3 {
			fakeUrl = "http://1.2.3.4" + raw[start:k]
			break
		}
	}

	u, err := url.Parse(fakeUrl)
	if err != nil {
		return "", err
	}
	pos := strings.LastIndex(u.Path, ".")
	if pos == -1 {
		return "", errors.New("couldn't find a period to indicate a file extension")
	}
	return u.Path[pos+1 : len(u.Path)], nil
}
