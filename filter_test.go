package ff_test

import (
	"context"
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/nakatanakatana/ff"
)

type filterFuncTest struct {
	key    string
	value  string
	expect bool
}

func TestCreateFilterInvalidKey(t *testing.T) {
	t.Parallel()

	filtersMap := ff.CreateFiltersMap([]string{}, []string{})

	f := ff.CreateFilter("invalidkey", "value", filtersMap)
	if f != nil {
		t.Fail()
	}
}

func TestCreateFilter(t *testing.T) {
	t.Parallel()

	filtersMap := ff.CreateFiltersMap([]string{}, []string{})
	testItem := createTestItem()

	for _, tt := range []filterFuncTest{
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
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			f := ff.CreateFilter(tt.key, tt.value, filtersMap)
			if f(ctx, testItem) != tt.expect {
				t.Fail()
			}
		})
	}
}

func TestCreateFilterItemHasNil(t *testing.T) {
	t.Parallel()

	filtersMap := ff.CreateFiltersMap([]string{}, []string{})

	testItemHasNil := &gofeed.Item{
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
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			f := ff.CreateFilter(tt.key, tt.value, filtersMap)
			if f(ctx, testItemHasNil) != tt.expect {
				t.Fail()
			}
		})
	}
}

func TestAuthorMute(t *testing.T) {
	t.Parallel()

	testItem := createTestItem()

	for _, tt := range []struct {
		name    string
		targets []string
		expect  bool
	}{
		{"empty", []string{}, true},
		{"equal title", []string{"title"}, false},
		{"equal description", []string{"description"}, false},
		{"contains link", []string{"github"}, false},
		{"contains author", []string{"name"}, false},
		{"contains description", []string{"desc"}, false},
		{"contains title", []string{"hoge", "fuga", "title"}, false},
		{"contains multiple", []string{"hoge", "name", "title"}, false},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			f := ff.CreateAuthorMute(tt.targets)("")
			if f(ctx, testItem) != tt.expect {
				t.Fail()
			}
		})
	}
}

func TestAuthorMuteItemHasNil(t *testing.T) {
	t.Parallel()

	testItemHasNil := &gofeed.Item{
		Title:       "title",
		Description: "description",
		Link:        "https://github.com/nakatanakatana/ff",
	}

	for _, tt := range []struct {
		name    string
		targets []string
		expect  bool
	}{
		{"empty", []string{}, true},
		{"contains title", []string{"title"}, false},
		{"contains description", []string{"description"}, false},
		{"contains link", []string{"github"}, false},
		{"empty author ignore", []string{"name"}, true},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			f := ff.CreateAuthorMute(tt.targets)("")
			if f(ctx, testItemHasNil) != tt.expect {
				t.Fail()
			}
		})
	}
}

func TestLinkMute(t *testing.T) {
	t.Parallel()

	testItem := createTestItem()

	for _, tt := range []struct {
		name    string
		targets []string
		expect  bool
	}{
		{"empty", []string{}, true},
		{"contains link", []string{"git"}, false},
		{"contains link", []string{"github.com"}, false},
		{"all targets don't contain link", []string{"abc", "def", "ghi"}, true},
		{"partial targets contains link", []string{"abc", "def", "git"}, false},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			f := ff.CreateLinkMute(tt.targets)("")
			if f(ctx, testItem) != tt.expect {
				t.Fail()
			}
		})
	}
}

func TestFilterDoAllFilterFuncAndCondition(t *testing.T) {
	t.Parallel()

	testItem := createTestItem()

	for _, tt := range []struct {
		name      string
		filters   []ff.FilterFunc
		expectLen int
	}{
		{"empty", []ff.FilterFunc{}, 1},
		{"nilFilter only", []ff.FilterFunc{ff.NilFilter("")}, 1},
		{"all match", []ff.FilterFunc{ff.TitleEqual("title"), ff.DescriptionEqual("description")}, 1},
		{"partial unmatch", []ff.FilterFunc{ff.TitleEqual("ti"), ff.TitleEqual("title")}, 0},
	} {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testFeed := &gofeed.Feed{
				Items: []*gofeed.Item{testItem},
			}

			result, err := ff.Apply(testFeed, tt.filters, []ff.ModifierFunc{})
			if err != nil {
				t.Fail()
			}

			if len(result.Items) != tt.expectLen {
				t.Fail()
			}
		})
	}
}
