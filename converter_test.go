package ff_test

import (
	"log"
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/nakatanakatana/ff"
	"gotest.tools/v3/assert"
)

func TestConvert(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		url string
	}{
		{url: "https://menthas.com/all/rss"},
		{url: "https://zenn.dev/feed"},
	} {
		tt := tt
		t.Run(tt.url, func(t *testing.T) {
			t.Parallel()

			fp := gofeed.NewParser()

			feed, err := fp.ParseURL(tt.url)
			if err != nil {
				log.Println("err", err)
			}

			f := ff.Convert(feed)
			log.Println(len(f.Items))
		})
	}
}

func TestConvertWithNil(t *testing.T) {
	t.Parallel()

	t.Run("nil feed", func(t *testing.T) {
		t.Parallel()

		converted := ff.Convert(nil)
		assert.Assert(t, converted != nil)
		assert.Equal(t, 0, len(converted.Items))
	})

	t.Run("feed with nil item", func(t *testing.T) {
		t.Parallel()

		feed := &gofeed.Feed{
			Items: []*gofeed.Item{nil},
		}

		converted := ff.Convert(feed)
		assert.Equal(t, 0, len(converted.Items))
	})
}
