//go:build !android
// +build !android

package libav

import (
	"github.com/asticode/go-astiav"
)

const (
	MediaTypeVideo = astiav.MediaTypeVideo
	MediaTypeAudio = astiav.MediaTypeAudio
)
