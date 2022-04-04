package ff_test

import (
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/nakatanakatana/ff"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

type modifierFuncTest struct {
	key    string
	value  string
	expect gofeed.Item
}

func TestCreateModifierInvalidKey(t *testing.T) {
	t.Parallel()

	modifiersMap := ff.CreateModifierMap()

	f := ff.CreateModifier("invalidKey", "value", modifiersMap)
	if f != nil {
		t.Fail()
	}
}

func TestCreateModifier(t *testing.T) {
	t.Parallel()

	modifiersMap := ff.CreateModifierMap()
	testItem := createTestItem()

	expectRemoveDescription := *testItem
	expectRemoveDescription.Description = ""
	expectRemoveContent := *testItem
	expectRemoveContent.Content = ""

	for _, tt := range []modifierFuncTest{
		// remove
		{key: "rm.description", value: "", expect: expectRemoveDescription},
		{key: "rm.content", value: "", expect: expectRemoveContent},
	} {
		tt := tt

		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()

			f := ff.CreateModifier(tt.key, tt.value, modifiersMap)
			result := f(testItem)
			assert.Check(t, is.DeepEqual(tt.expect, *result))
		})
	}
}
