package ff

import (
	"github.com/mmcdole/gofeed"
)

type (
	FilterFunc        = func(i *gofeed.Item) bool
	FilterFuncCreator = func(param string) FilterFunc
	FilterFuncMap     = map[string]FilterFuncCreator
)

func CreateFiltersMap(muteAuthors, muteURLs []string) FilterFuncMap {
	return map[string]FilterFuncCreator{
		"title.equal":              TitleEqual,
		"description.equal":        DescriptionEqual,
		"link.equal":               LinkEqual,
		"author.equal":             AuthorEqual,
		"title.not_equal":          TitleNotEqual,
		"description.not_equal":    DescriptionNotEqual,
		"link.not_equal":           LinkNotEqual,
		"author.not_equal":         AuthorNotEqual,
		"title.contains":           TitleContains,
		"description.contains":     DescriptionContains,
		"link.contains":            LinkContains,
		"author.contains":          AuthorContains,
		"title.not_contains":       TitleNotContains,
		"description.not_contains": DescriptionNotContains,
		"link.not_contains":        LinkNotContains,
		"author.not_contains":      AuthorNotContains,
		"updated_at.from":          UpdateAtFrom,
		"published_at.from":        PublishedAtFrom,
		"updated_at.latest":        UpdateAtLatest,
		"published_at.latest":      PublishedAtLatest,
		"latest":                   DateLatest,
		"mute_authors":             CreateAuthorMute(muteAuthors),
		"mute_urls":                CreateLinkMute(muteURLs),
	}
}

func CreateFilter(key string, value string, filters map[string]FilterFuncCreator) FilterFunc {
	f, ok := filters[key]
	if !ok {
		return nil
	}

	return f(value)
}

func apply(i *gofeed.Item, ff ...FilterFunc) bool {
	for _, f := range ff {
		if !f(i) {
			return false
		}
	}

	return true
}

func Filter(f *gofeed.Feed, ff ...FilterFunc) (*gofeed.Feed, error) {
	items := make([]*gofeed.Item, len(f.Items))
	count := 0

	for _, i := range f.Items {
		if apply(i, ff...) {
			items[count] = i
			count++
		}
	}

	f.Items = items[:count]

	return f, nil
}
