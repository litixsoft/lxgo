package lxCollection_test

import (
	lxCollection "github.com/litixsoft/lxgo/collection"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

// TestFirstIndex, tests for FirstIndex collection func
func TestFirstIndex(t *testing.T) {

	// Slice for test
	strs := []string{"swift", "go", "rust", "rust", "go"}

	// Find first index
	assert.Equal(t, 1, lxCollection.FirstIndex(strs, "go"), "should be returns the first index of the target string.")
	assert.Equal(t, 2, lxCollection.FirstIndex(strs, "rust"), "should be returns the first index of the target string.")

	// Not find first index
	assert.Equal(t, -1, lxCollection.FirstIndex(strs, "java"), "should be not matched target string.")
	assert.Equal(t, -1, lxCollection.FirstIndex(strs, "php"), "should be not matched target string.")
}

// TestInclude, tests for Include collection func
func TestInclude(t *testing.T) {

	// Slice for test
	strs := []string{"swift", "go", "rust", "rust", "go"}

	// The target string is in the slice.
	assert.True(t, lxCollection.Include(strs, "go"), "go should be found in slice.")
	assert.True(t, lxCollection.Include(strs, "rust"), "rust should be found in slice.")

	// The target string is not in the slice.
	assert.False(t, lxCollection.Include(strs, "java"), "java should not be found in slice.")
	assert.False(t, lxCollection.Include(strs, "php"), "php should not be found in slice.")
}

// TestFindAny, should be returns true if one of the strings in the slice match with predicate.
func TestFindAny(t *testing.T) {

	// Slice for test
	strs := []string{"swift", "go", "rust", "rust", "go"}

	// Check find swift value
	swift := lxCollection.FindAny(strs, func(v string) bool {
		return strings.HasPrefix(v, "swift")
	})

	assert.True(t, swift, "swift should be found and returns true.")

	// Check find rust value
	rust := lxCollection.FindAny(strs, func(v string) bool {
		return strings.HasPrefix(v, "rust")
	})

	assert.True(t, rust, "rust should be found and returns true.")

	// Check not found java value
	java := lxCollection.FindAny(strs, func(v string) bool {
		return strings.HasPrefix(v, "java")
	})

	assert.False(t, java, "java should be not found and returns false.")

	// Check not found java value
	php := lxCollection.FindAny(strs, func(v string) bool {
		return strings.HasPrefix(v, "php")
	})

	assert.False(t, php, "php should be not found and returns false.")
}

// TestFindAll, should be returns true if all of the strings in the slice match with predicate.
func TestFindAll(t *testing.T) {

	// Slice for test
	strs := []string{"swiftlang", "golang", "rustlang"}

	// Check if all have a lang suffix, they should be true
	lang := lxCollection.FindAll(strs, func(v string) bool {
		return strings.HasSuffix(v, "lang")
	})

	assert.True(t, lang, "should be true, all have lang suffix.")

	// Check if all have a swift prefix, they should be false
	swift := lxCollection.FindAll(strs, func(v string) bool {
		return strings.HasPrefix(v, "swift")
	})

	assert.False(t, swift, "should be false, all have not swift suffix.")
}

// TestFilter, should be returns a new slice containing all strings in the slice that match the predicate.
func TestFilter(t *testing.T) {

	// Slice for test
	strs := []string{"swift-darwin", "swift-linux", "go-windows", "go-linux", "go-darwin"}

	// Filter all values, that contains linux in name, in a new slice
	actual := lxCollection.Filter(strs, func(v string) bool {
		return strings.Contains(v, "linux")
	})

	// Slice for check
	check := []string{"swift-linux", "go-linux"}

	assert.Equal(t, check, actual, "should be equal slices.")
}

// TestMap,  should be returns a new slice with the results of applying function to each string in the original slice.
func TestMap(t *testing.T) {

	// Slice for test
	strs := []string{"fight", "fight", "fight", "bite", "bite", "bite"}

	// Map all values to upper in a new slice
	actual := lxCollection.Map(strs, strings.ToUpper)

	// Slice for check
	check := []string{"FIGHT", "FIGHT", "FIGHT", "BITE", "BITE", "BITE"}

	// Check the new slice
	assert.Equal(t, check, actual, "should be equal slices.")
	assert.NotContains(t, actual, "fight", "should be not contains fight in lower case.")
	assert.NotContains(t, actual, "bite", "should be not contains bite in lower case.")
}
