package gormcase

import (
	"testing"

	"github.com/ing-bank/gormtestutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestDeepGorm_Initialize_TriggersCaseInsensitivityCorrectly(t *testing.T) {
	t.Parallel()

	type ObjectA struct {
		ID    uint
		Name  string
		Age   int
		Other string
	}

	defaultQuery := func(db *gorm.DB) *gorm.DB { return db }

	tests := map[string]struct {
		filter   map[string]any
		query    func(*gorm.DB) *gorm.DB
		existing []ObjectA
		expected []ObjectA
	}{
		"nothing": {
			expected: []ObjectA{},
			query:    defaultQuery,
		},

		// Check if everything still works
		"simple where query": {
			filter: map[string]any{
				"name": "jessica",
			},
			existing: []ObjectA{{ID: 1, Name: "jessica", Age: 46}, {ID: 2, Name: "amy", Age: 35}},
			expected: []ObjectA{{ID: 1, Name: "jessica", Age: 46}},
			query:    defaultQuery,
		},
		"more complex where query": {
			filter: map[string]any{
				"name": "jessica",
				"age":  53,
			},
			existing: []ObjectA{{ID: 1, Name: "jessica", Age: 53}, {ID: 2, Name: "jessica", Age: 20}},
			expected: []ObjectA{{ID: 1, Name: "jessica", Age: 53}},
			query:    defaultQuery,
		},
		"multi-value where query": {
			filter: map[string]any{
				"name": []string{"jessica", "amy"},
			},
			existing: []ObjectA{{ID: 1, Name: "jessica", Age: 53}, {ID: 2, Name: "amy", Age: 20}},
			expected: []ObjectA{{ID: 1, Name: "jessica", Age: 53}, {ID: 2, Name: "amy", Age: 20}},
			query:    defaultQuery,
		},
		"more complex multi-value where query": {
			filter: map[string]any{
				"name": []string{"jessica", "amy"},
				"age":  []int{53, 20},
			},
			existing: []ObjectA{{ID: 1, Name: "jessica", Age: 53}, {ID: 2, Name: "amy", Age: 20}},
			expected: []ObjectA{{ID: 1, Name: "jessica", Age: 53}, {ID: 2, Name: "amy", Age: 20}},
			query:    defaultQuery,
		},

		// Real tests
		"simple insensitivity query": {
			filter: map[string]any{
				"name": "JESSICA",
			},
			existing: []ObjectA{{ID: 1, Name: "jessica", Age: 53}},
			expected: []ObjectA{{ID: 1, Name: "jessica", Age: 53}},
			query:    defaultQuery,
		},
		"more complex insensitivity query": {
			filter: map[string]any{
				"name": "JESSICA",
				"age":  20,
			},
			existing: []ObjectA{{ID: 1, Name: "jessica", Age: 20}},
			expected: []ObjectA{{ID: 1, Name: "jessica", Age: 20}},
			query:    defaultQuery,
		},
		"multi-value, all insensitivity queries": {
			filter: map[string]any{
				"name": []string{"JeSSICA", "AmY"},
			},
			existing: []ObjectA{{ID: 1, Name: "jessica", Age: 53}, {ID: 2, Name: "amy", Age: 20}, {ID: 3, Name: "John", Age: 25}},
			expected: []ObjectA{{ID: 1, Name: "jessica", Age: 53}, {ID: 2, Name: "amy", Age: 20}},
			query:    defaultQuery,
		},
		"more complex multi-value, all insensitivity queries": {
			filter: map[string]any{
				"name":  []string{"JESSICA", "AMY"},
				"other": []string{"AAAooo", "AAAooo"},
			},
			existing: []ObjectA{{ID: 1, Name: "jessica", Age: 53, Other: "aaaooo"}, {ID: 2, Name: "amy", Age: 20, Other: "aaaooo"}, {ID: 3, Name: "John", Age: 25, Other: "aaaooo"}},
			expected: []ObjectA{{ID: 1, Name: "jessica", Age: 53, Other: "aaaooo"}, {ID: 2, Name: "amy", Age: 20, Other: "aaaooo"}, {ID: 3, Name: "John", Age: 25, Other: "aaaooo"}},
			query:    defaultQuery,
		},
		"multi-value, some insensitivity queries": {
			filter: map[string]any{
				"name": []string{"JESSICA", "JOHN"},
			},
			existing: []ObjectA{{ID: 1, Name: "JESSICA", Age: 53}, {ID: 2, Name: "amy", Age: 20}, {ID: 3, Name: "John", Age: 25}},
			expected: []ObjectA{{ID: 1, Name: "JESSICA", Age: 53}, {ID: 3, Name: "John", Age: 25}},
			query:    defaultQuery,
		},
		"more complex multi-value, some insensitivity queries": {
			filter: map[string]any{
				"name":  []string{"jessica", "JOHN"},
				"other": []string{"AAB", "Bb"},
			},
			existing: []ObjectA{{ID: 1, Name: "jessica", Age: 53, Other: "aab"}, {ID: 2, Name: "amy", Age: 20}, {ID: 3, Name: "John", Age: 25, Other: "bb"}},
			expected: []ObjectA{{ID: 1, Name: "jessica", Age: 53, Other: "aab"}, {ID: 3, Name: "John", Age: 25, Other: "bb"}},
			query:    defaultQuery,
		},
		"explicitly disable liking in query": {
			filter: map[string]any{
				"name": []string{"JESSICA", "AMY"},
				"age":  []int{53, 20},
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Set(tagName, false)
			},
			existing: []ObjectA{{ID: 1, Name: "jessica", Age: 53}, {ID: 2, Name: "amy", Age: 20}},
			expected: []ObjectA{},
		},
	}

	for name, testData := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Arrange
			db := gormtestutil.NewMemoryDatabase(t, gormtestutil.WithName(t.Name()))
			require.NoError(t, db.AutoMigrate(&ObjectA{}))
			plugin := New()

			require.NoError(t, db.CreateInBatches(testData.existing, 10).Error)

			db = testData.query(db)

			// Act
			err := db.Use(plugin)

			// Assert
			require.NoError(t, err)

			var actual []ObjectA
			err = db.Where(testData.filter).Find(&actual).Error
			require.NoError(t, err)

			assert.Equal(t, testData.expected, actual)
		})
	}
}

func TestDeepGorm_Initialize_TriggersCaseInsensitivityCorrectlyWithConditionalTag(t *testing.T) {
	t.Parallel()

	type ObjectB struct {
		Name  string `gormcase:"true"`
		Other string
	}

	tests := map[string]struct {
		filter   map[string]any
		existing []ObjectB
		expected []ObjectB
	}{
		"simple filter on allowed fields": {
			filter: map[string]any{
				"name": "JESSICA",
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}, {Name: "amy", Other: "def"}},
			expected: []ObjectB{{Name: "jessica", Other: "abc"}},
		},
		"simple filter on disallowed fields": {
			filter: map[string]any{
				"other": "ABC",
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}, {Name: "jessica", Other: "abc"}},
			expected: []ObjectB{},
		},
		"multi-filter on allowed fields": {
			filter: map[string]any{
				"name": []string{"JESSICA", "AMY"},
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}, {Name: "amy", Other: "def"}},
			expected: []ObjectB{{Name: "jessica", Other: "abc"}, {Name: "amy", Other: "def"}},
		},
		"multi-filter on disallowed fields": {
			filter: map[string]any{
				"other": []string{"ABC", "abC"},
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}, {Name: "jessica", Other: "abc"}},
			expected: []ObjectB{},
		},
	}

	for name, testData := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Arrange
			db := gormtestutil.NewMemoryDatabase(t, gormtestutil.WithName(t.Name()))
			_ = db.AutoMigrate(&ObjectB{})
			plugin := New(TaggedOnly())

			if err := db.CreateInBatches(testData.existing, 10).Error; err != nil {
				t.Error(err)
				t.FailNow()
			}

			// Act
			err := db.Use(plugin)

			// Assert
			require.NoError(t, err)

			var actual []ObjectB
			err = db.Where(testData.filter).Find(&actual).Error
			require.NoError(t, err)

			assert.Equal(t, testData.expected, actual)
		})
	}
}

func TestDeepGorm_Initialize_TriggersCaseInsensitivityCorrectlyWithSetting(t *testing.T) {
	t.Parallel()

	type ObjectB struct {
		Name  string
		Other string
	}

	tests := map[string]struct {
		filter   map[string]any
		query    func(*gorm.DB) *gorm.DB
		existing []ObjectB
		expected []ObjectB
	}{
		"case-insensitive with query set to true": {
			filter: map[string]any{
				"name": "jEssIca",
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Set(tagName, true)
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}},
			expected: []ObjectB{{Name: "jessica", Other: "abc"}},
		},
		"case-sensitive with query set to false": {
			filter: map[string]any{
				"name": "jEssIca",
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Set(tagName, false)
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}},
			expected: []ObjectB{},
		},
		"case-sensitive with query set to random value": {
			filter: map[string]any{
				"name": "jEssIca",
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Set(tagName, "yes")
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}},
			expected: []ObjectB{},
		},
		"case-sensitive with query unset": {
			filter: map[string]any{
				"name": "jEssIca",
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}},
			expected: []ObjectB{},
		},
	}

	for name, testData := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Arrange
			db := gormtestutil.NewMemoryDatabase(t, gormtestutil.WithName(t.Name()))
			_ = db.AutoMigrate(&ObjectB{})
			plugin := New(SettingOnly())

			if err := db.CreateInBatches(testData.existing, 10).Error; err != nil {
				t.Error(err)
				t.FailNow()
			}

			db = testData.query(db)

			// Act
			err := db.Use(plugin)

			// Assert
			require.NoError(t, err)

			var actual []ObjectB
			err = db.Where(testData.filter).Find(&actual).Error
			require.NoError(t, err)

			assert.Equal(t, testData.expected, actual)
		})
	}
}
