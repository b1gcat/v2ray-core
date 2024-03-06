package apk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ext(t *testing.T) {
	ext, err := getFileExtensionFromUrl("GET /assets/files/mypicture.jpg?width=1000&height=600 HTTP1.1")
	assert.Equal(t, nil, err)
	assert.Equal(t, "jpg", ext)
}
