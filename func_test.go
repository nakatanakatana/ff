package ff_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/nakatanakatana/ff"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
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

func TestParseQueries(t *testing.T) {
	t.Parallel()

	filtersMap := ff.CreateFiltersMap([]string{}, []string{})
	modifierMap := ff.CreateModifierMap()

	for _, tt := range []struct {
		name              string
		urlString         string
		expectFilterLen   int
		expectModifierLen int
	}{
		{"parameter is empty", "https://t.io/", 0, 0},
		{"filterOnly", "https://t.io/?link.contains=t.io", 1, 0},
		{"modifierOnly", "https://t.io/?rm.description", 0, 1},
		{"filterAndModifier multiple", "https://t.io/?title.equal=title&latest&rm.content", 2, 1},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.urlString)
			assert.NilError(t, err)

			f, m := ff.ParseQueries(u.Query(), filtersMap, modifierMap)
			assert.Check(t, len(f) == tt.expectFilterLen)
			assert.Check(t, len(m) == tt.expectModifierLen)
		})
	}
}

func TestFilterAndModifier(t *testing.T) {
	t.Parallel()

	testItem := createTestItem()

	expectSameItem := *testItem
	expectRemoveDescription := *testItem
	expectRemoveDescription.Description = ""

	for _, tt := range []struct {
		name          string
		filters       []ff.FilterFunc
		modifiers     []ff.ModifierFunc
		expectItemLen int
		expectItem    *gofeed.Item
	}{
		{
			"empty",
			[]ff.FilterFunc{},
			[]ff.ModifierFunc{},
			1,
			&expectSameItem,
		},
		{
			"filterOnly matched",
			[]ff.FilterFunc{ff.TitleEqual("title")},
			[]ff.ModifierFunc{},
			1,
			&expectSameItem,
		},
		{
			"filterOnly unmatch",
			[]ff.FilterFunc{ff.TitleEqual("ti")},
			[]ff.ModifierFunc{},
			0,
			nil,
		},
		{
			"modifierOnly",
			[]ff.FilterFunc{},
			[]ff.ModifierFunc{ff.RemoveDescription("")},
			1,
			&expectRemoveDescription,
		},
		{
			"filterAndModifier matched",
			[]ff.FilterFunc{ff.TitleEqual("title")},
			[]ff.ModifierFunc{ff.RemoveDescription("")},
			1,
			&expectRemoveDescription,
		},
		{
			"filterAndModifier unmatch",
			[]ff.FilterFunc{ff.TitleEqual("ti")},
			[]ff.ModifierFunc{ff.RemoveDescription("")},
			0,
			nil,
		},
	} {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			testFeed := &gofeed.Feed{
				Items: []*gofeed.Item{testItem},
			}
			result, err := ff.Apply(testFeed, tt.filters, tt.modifiers)
			assert.NilError(t, err)
			assert.Check(t, is.Len(result.Items, tt.expectItemLen))
			if tt.expectItemLen > 0 {
				assert.Check(t, is.DeepEqual(tt.expectItem, result.Items[0]))
			}
		})
	}
}
