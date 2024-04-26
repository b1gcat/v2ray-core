package apk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ext(t *testing.T) {
	ext, err := getPathFromUrl(`GET /qqqqq.mieee.mi.com/ashuewqihu HTTP1.1`, 4)
	assert.Equal(t, nil, err)
	assert.Equal(t, `/qqqqq.mieee.mi.com/ashuewqihu`, ext)
}
