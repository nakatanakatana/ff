package ff_test

import (
	"time"

	"github.com/mmcdole/gofeed"
)

func createTestItem() *gofeed.Item {
	testUpdated := time.Date(2021, time.July, 11, 0, 0, 0, 0, time.UTC)
	testPublished := time.Date(2021, time.July, 1, 0, 0, 0, 0, time.UTC)
	testItem := &gofeed.Item{
		Title:           "title",
		Description:     "description",
		Link:            "https://github.com/nakatanakatana/ff",
		Author:          &gofeed.Person{Name: "aname", Email: "aname@nakatanakatana.dev"},
		UpdatedParsed:   &testUpdated,
		PublishedParsed: &testPublished,
	}

	return testItem
}
