package f1tv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_assbmleURL(t *testing.T) {
	cases := []struct {
		name    string
		urlPath string
		format  StreamType
		args    []interface{}
		result  string
	}{
		{
			name:    "playback URL",
			urlPath: playbackRequestPath,
			format:  BIG_SCREEN_HLS,
			args:    []interface{}{1000003910},
			result:  "https://f1tv.formula1.com/1.0/R/ENG/BIG_SCREEN_HLS/ALL/CONTENT/PLAY?contentId=1000003910",
		},
		{
			name:    "perspective playback URL",
			urlPath: playbackPerspectiveRequestPath,
			format:  BIG_SCREEN_HLS,
			args:    []interface{}{"CONTENT/PLAY?channelId=1014&contentId=1000003912"},
			result:  "https://f1tv.formula1.com/1.0/R/ENG/BIG_SCREEN_HLS/ALL/CONTENT/PLAY?channelId=1014&contentId=1000003912",
		},
		{
			name:    "content details URL",
			urlPath: contentDetailsPath,
			format:  WEB_DASH,
			args:    []interface{}{1000003910},
			result:  "https://f1tv.formula1.com/2.0/R/ENG/WEB_DASH/ALL/CONTENT/VIDEO/1000003910/F1_TV_Pro_Annual/14",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resURL, err := assembleURL(c.urlPath, c.format, c.args...)
			assert.NoError(t, err)
			assert.Equal(t, c.result, resURL.String())
		})
	}
}
