package apk

import (
	"bytes"
	"errors"
	"regexp"
)

var (
	APKDownload = "APK-download"
)

type SniffHeader struct{}

func (h *SniffHeader) Protocol() string {
	return APKDownload
}

func (h *SniffHeader) Domain() string {
	return ""
}

var errNotAPK = errors.New("not APK signature")

func SniffApkDownload(b []byte) (*SniffHeader, error) {
	if len(b) < 12 {
		return nil, errNotAPK
	}

	if !bytes.Equal(b[:5], []byte("GET /")) {
		return nil, errNotAPK
	}

	var re = regexp.MustCompile(`(?m)^GET /.*.apk HTTP`)

	match := re.FindAll(b, -1)
	if match != nil {
		return &SniffHeader{}, nil
	}

	return nil, errNotAPK
}
