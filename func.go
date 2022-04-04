package ff

import (
	"github.com/mmcdole/gofeed"
)

func Apply(f *gofeed.Feed, ff ...FilterFunc) (*gofeed.Feed, error) {
	items := make([]*gofeed.Item, len(f.Items))
	count := 0

	for _, i := range f.Items {
		if filterApply(i, ff...) {
			items[count] = i
			count++
		}
	}

	f.Items = items[:count]

	return f, nil
}
