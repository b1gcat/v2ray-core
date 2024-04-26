package apk

import (
	"bytes"
	"errors"
	"net/url"
	"strings"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/strmatcher"
)

var (
	URLExtension = "URL-extension"
)

type SniffHeader struct{}

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

	var path string
	var err error

	if !bytes.Equal(b[:5], []byte("GET /")) {
		if !bytes.Equal(b[:6], []byte("POST /")) {
			return nil, errNotExt
		}

		path, err = getPathFromUrl(string(b), 5)
		if err != nil {
			return nil, errNotExt
		}
	} else {
		path, err = getPathFromUrl(string(b), 4)
		if err != nil {
			return nil, errNotExt
		}
	}

	if !findmatcherUrlPath(path) {
		return nil, errNotExt
	}

	return &SniffHeader{}, nil
}

func getPathFromUrl(raw string, idx int) (string, error) {
	fakeUrl := "http://1.2.3.4" + strings.Split(raw[idx:], " ")[0]

	u, err := url.Parse(fakeUrl)
	if err != nil {
		return "", err
	}
	path := u.Path
	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}
	return path, nil
}

var (
	matcherURlPath = strmatcher.NewMphIndexMatcher()
)

func CompileMatcherUrlPath(list []string) {
	type urlPath struct {
		Type   strmatcher.Type
		Domain string
	}
	rules := make([]urlPath, 0)
	for _, v := range list {
		tv := strings.Split(v, ":")
		if len(tv) != 2 {
			continue
		}
		switch tv[0] {
		case "full":
			rules = append(rules, urlPath{Type: strmatcher.Full, Domain: tv[1]})

		case "regex":
			rules = append(rules, urlPath{Type: strmatcher.Regex, Domain: tv[1]})

		case "keyword":
			rules = append(rules, urlPath{Type: strmatcher.Substr, Domain: tv[1]})

		case "suffix":
			rules = append(rules, urlPath{Type: strmatcher.Domain, Domain: tv[1]})
		}
	}

	for _, rule := range rules {
		matcher, err := rule.Type.New(rule.Domain)
		common.Must(err)
		matcherURlPath.Add(matcher)
	}

	matcherURlPath.Build()
}

func findmatcherUrlPath(path string) bool {
	return matcherURlPath.MatchAny(path)
}
