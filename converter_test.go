package ff_test

import (
	"log"
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/nakatanakatana/ff"
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
