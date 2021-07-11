package ff

import (
	"fmt"
	"testing"

	"github.com/mmcdole/gofeed"
)

func TestConvert(t *testing.T) {
	for _, tt := range []struct {
		url string
	}{
		{url: "https://menthas.com/all/rss"},
		{url: "https://zenn.dev/feed"},
	} {
		tt := tt
		t.Run("getFeed:"+tt.url, func(t *testing.T) {
			fp := gofeed.NewParser()
			feed, err := fp.ParseURL(tt.url)
			if err != nil {
				fmt.Println("err", err)
			}
			f := Convert(feed)
			fmt.Println(len(f.Items))
		})
	}
}
