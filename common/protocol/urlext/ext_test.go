package apk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ext(t *testing.T) {
	ext, err := getPathFromUrl(`POST /d-ot.tapimg.com/f/202401/13/a/792c230fee7e309e4abab9687d0174c6?sign=3ff969b8ff1a0b0349fc0464404ea841&t=662cc792&cb=https%3A%2F%2Fd-al.tapimg.com%2F202404271738%2Fc61058b6f563050faddfe202f513c7b8%2Ff%2F202401%2F13%2Fa%2F792c230fee7e309e4abab9687d0174c6&xyip=125.117.75.250&xy_nosvr=true&xy_rid=zYS029LN97HotQ&xy_oid=47143&xy_nc=0FH8I1== HTTP1.1`, 5)
	assert.Equal(t, nil, err)
	assert.Equal(t, `/d-ot.tapimg.com/f/202401/13/a/792c230fee7e309e4abab9687d0174c6?sign=3ff969b8ff1a0b0349fc0464404ea841&t=662cc792&cb=https%3A%2F%2Fd-al.tapimg.com%2F202404271738%2Fc61058b6f563050faddfe202f513c7b8%2Ff%2F202401%2F13%2Fa%2F792c230fee7e309e4abab9687d0174c6&xyip=125.117.75.250&xy_nosvr=true&xy_rid=zYS029LN97HotQ&xy_oid=47143&xy_nc=0FH8I1==`, ext)
}
