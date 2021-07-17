package ff

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

func TestCreateFilter(t *testing.T) {

	t.Run("if invalid key, return nil", func(t *testing.T) {
		f := CreateFilter("invalidkey", "value")
		if f != nil {
			t.Fail()
		}
	})

	testUpdated := time.Date(2021, time.July, 11, 0, 0, 0, 0, time.UTC)
	testPublished := time.Date(2021, time.July, 1, 0, 0, 0, 0, time.UTC)
	testData := &gofeed.Item{
		Title:           "title",
		Description:     "description",
		Link:            "https://github.com/nakatanakatana/ff",
		Author:          &gofeed.Person{Name: "aname", Email: "aname@nakatanakatana.dev"},
		UpdatedParsed:   &testUpdated,
		PublishedParsed: &testPublished,
	}

	for _, tt := range []struct {
		key    string
		value  string
		expect bool
	}{
		// equal
		{key: "title.equal", value: "title", expect: true},
		{key: "title.equal", value: "other_title", expect: false},
		{key: "description.equal", value: "description", expect: true},
		{key: "description.equal", value: "other_description", expect: false},
		{key: "link.equal", value: "https://github.com/nakatanakatana/ff", expect: true},
		{key: "link.equal", value: "https://github.com/nakatanakatana/other", expect: false},
		{key: "author.equal", value: "aname", expect: true},
		{key: "author.equal", value: "other_author_name", expect: false},
		// not_equal
		{key: "title.not_equal", value: "title", expect: false},
		{key: "title.not_equal", value: "other_title", expect: true},
		{key: "description.not_equal", value: "description", expect: false},
		{key: "description.not_equal", value: "other_description", expect: true},
		{key: "link.not_equal", value: "https://github.com/nakatanakatana/ff", expect: false},
		{key: "link.not_equal", value: "https://github.com/nakatanakatana/other", expect: true},
		{key: "author.not_equal", value: "aname", expect: false},
		{key: "author.not_equal", value: "other_author_name", expect: true},
		// contains
		{key: "title.contains", value: "t", expect: true},
		{key: "title.contains", value: "titles", expect: false},
		{key: "description.contains", value: "c", expect: true},
		{key: "description.contains", value: "descriptions", expect: false},
		{key: "link.contains", value: "github.com/nakatanakatana", expect: true},
		{key: "link.contains", value: "github.com/nakatanakatana/other", expect: false},
		{key: "author.contains", value: "name", expect: true},
		{key: "author.contains", value: "names", expect: false},
		// not_contains
		{key: "title.not_contains", value: "t", expect: false},
		{key: "title.not_contains", value: "titles", expect: true},
		{key: "description.not_contains", value: "c", expect: false},
		{key: "description.not_contains", value: "descriptions", expect: true},
		{key: "link.not_contains", value: "github.com/nakatanakatana", expect: false},
		{key: "link.not_contains", value: "github.com/nakatanakatana/other", expect: true},
		{key: "author.not_contains", value: "name", expect: false},
		{key: "author.not_contains", value: "names", expect: true},
		// from
		{key: "updated_at.from", value: "invalid date", expect: true},
		{key: "updated_at.from", value: "2021-07-07T12:00:00+09:00", expect: true},
		{key: "updated_at.from", value: "2021-07-14T12:00:00+09:00", expect: false},
		{key: "published_at.from", value: "invalid date", expect: true},
		{key: "published_at.from", value: "2021-06-30T12:00:00+09:00", expect: true},
		{key: "published_at.from", value: "2021-07-07T12:00:00+09:00", expect: false},
	} {
		tt := tt
		t.Run(tt.key+"="+tt.value+":"+strconv.FormatBool(tt.expect), func(t *testing.T) {
			f := CreateFilter(tt.key, tt.value)
			if f(testData) != tt.expect {
				t.Fail()
			}
		})
	}

	testDataHasNil := &gofeed.Item{
		Title:       "title",
		Description: "description",
		Link:        "https://github.com/nakatanakatana/ff",
	}

	for _, tt := range []struct {
		key    string
		value  string
		expect bool
	}{
		// equal
		{key: "author.equal", value: "aname", expect: false},
		{key: "author.equal", value: "other_author_name", expect: false},
		// not_equal
		{key: "author.not_equal", value: "aname", expect: true},
		{key: "author.not_equal", value: "other_author_name", expect: true},
		// contains
		{key: "author.contains", value: "name", expect: false},
		{key: "author.contains", value: "names", expect: false},
		// not_contains
		{key: "author.not_contains", value: "name", expect: true},
		{key: "author.not_contains", value: "names", expect: true},
		// from
		{key: "updated_at.from", value: "invalid date", expect: true},
		{key: "updated_at.from", value: "2021-07-07T12:00:00+09:00", expect: true},
		{key: "updated_at.from", value: "2021-07-14T12:00:00+09:00", expect: true},
		{key: "published_at.from", value: "invalid date", expect: true},
		{key: "published_at.from", value: "2021-06-30T12:00:00+09:00", expect: true},
		{key: "published_at.from", value: "2021-07-07T12:00:00+09:00", expect: true},
	} {
		tt := tt
		t.Run("hasNil: "+tt.key+"="+tt.value+":"+strconv.FormatBool(tt.expect), func(t *testing.T) {
			f := CreateFilter(tt.key, tt.value)
			if f(testDataHasNil) != tt.expect {
				t.Fail()
			}
		})
	}
}

func TestAuthorMute(t *testing.T) {
	testUpdated := time.Date(2021, time.July, 11, 0, 0, 0, 0, time.UTC)
	testPublished := time.Date(2021, time.July, 1, 0, 0, 0, 0, time.UTC)
	testData := &gofeed.Item{
		Title:           "title",
		Description:     "description",
		Link:            "https://github.com/nakatanakatana/ff",
		Author:          &gofeed.Person{Name: "aname", Email: "aname@nakatanakatana.dev"},
		UpdatedParsed:   &testUpdated,
		PublishedParsed: &testPublished,
	}

	for _, tt := range []struct {
		targets []string
		expect  bool
	}{
		{[]string{}, true},
		{[]string{"title"}, false},
		{[]string{"description"}, false},
		{[]string{"github"}, false},
		{[]string{"name"}, false},
		{[]string{"desc"}, false},
		{[]string{"hoge", "fuga", "title"}, false},
		{[]string{"hoge", "name", "title"}, false},
	} {
		tt := tt
		t.Run(strings.Join(tt.targets, ",")+":"+strconv.FormatBool(tt.expect), func(t *testing.T) {
			f := CreateAuthorMute(tt.targets)("")
			if f(testData) != tt.expect {
				t.Fail()
			}
		})
	}

	testDataHasNil := &gofeed.Item{
		Title:       "title",
		Description: "description",
		Link:        "https://github.com/nakatanakatana/ff",
	}

	for _, tt := range []struct {
		targets []string
		expect  bool
	}{
		{[]string{}, true},
		{[]string{"title"}, false},
		{[]string{"description"}, false},
		{[]string{"github"}, false},
		{[]string{"name"}, true},
		{[]string{"desc"}, false},
		{[]string{"hoge", "fuga", "title"}, false},
		{[]string{"hoge", "name", "title"}, false},
	} {
		tt := tt
		t.Run(strings.Join(tt.targets, ",")+":"+strconv.FormatBool(tt.expect), func(t *testing.T) {
			f := CreateAuthorMute(tt.targets)("")
			if f(testDataHasNil) != tt.expect {
				t.Fail()
			}
		})
	}
}

func TestLinkMute(t *testing.T) {
	testUpdated := time.Date(2021, time.July, 11, 0, 0, 0, 0, time.UTC)
	testPublished := time.Date(2021, time.July, 1, 0, 0, 0, 0, time.UTC)
	testData := &gofeed.Item{
		Title:           "title",
		Description:     "description",
		Link:            "https://github.com/nakatanakatana/ff",
		Author:          &gofeed.Person{Name: "aname", Email: "aname@nakatanakatana.dev"},
		UpdatedParsed:   &testUpdated,
		PublishedParsed: &testPublished,
	}

	for _, tt := range []struct {
		targets []string
		expect  bool
	}{
		{[]string{}, true},
		{[]string{"git"}, false},
		{[]string{"github.com"}, false},
		{[]string{"abc", "def", "ghi"}, true},
		{[]string{"abc", "def", "git"}, false},
	} {
		tt := tt
		t.Run(strings.Join(tt.targets, ",")+":"+strconv.FormatBool(tt.expect), func(t *testing.T) {
			f := CreateLinkMute(tt.targets)("")
			if f(testData) != tt.expect {
				t.Fail()
			}
		})
	}
}
