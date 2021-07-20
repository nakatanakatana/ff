package ff

import (
	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
)

func Convert(f *gofeed.Feed) *feeds.Feed {
	feed := &feeds.Feed{
		Title:       f.Title,
		Link:        &feeds.Link{Href: f.Link},
		Description: f.Description,
		Copyright:   f.Copyright,
	}
	if f.Author != nil {
		feed.Author = &feeds.Author{Name: f.Author.Name, Email: f.Author.Email}
	}

	if f.UpdatedParsed != nil {
		feed.Updated = *f.UpdatedParsed
	}

	if f.PublishedParsed != nil {
		feed.Created = *f.PublishedParsed
	}

	if f.Image != nil {
		feed.Image = &feeds.Image{Url: f.Image.URL, Title: f.Image.Title}
	}

	items := make([]*feeds.Item, len(f.Items))
	for index, i := range f.Items {
		items[index] = &feeds.Item{
			Title:       i.Title,
			Link:        &feeds.Link{Href: i.Link},
			Description: i.Description,
			Id:          i.GUID,
			Content:     i.Content,
		}
		if i.Author != nil {
			items[index].Author = &feeds.Author{Name: i.Author.Name, Email: i.Author.Email}
		}

		if i.UpdatedParsed != nil {
			items[index].Updated = *i.UpdatedParsed
		}

		if i.PublishedParsed != nil {
			items[index].Created = *i.PublishedParsed
		}
	}

	feed.Items = items

	return feed
}
