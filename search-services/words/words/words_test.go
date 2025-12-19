package words

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNorm(t *testing.T) {
	t.Run("empty phrase", func(t *testing.T) {
		result := Norm("")
		assert.Empty(t, result)
	})

	t.Run("single word", func(t *testing.T) {
		result := Norm("hello")
		assert.Equal(t, []string{"hello"}, result)
	})

	t.Run("multiple words", func(t *testing.T) {
		result := Norm("hello world")
		assert.ElementsMatch(t, []string{"hello", "world"}, result)
	})

	t.Run("words with punctuation", func(t *testing.T) {
		result := Norm("hello, world!")
		assert.ElementsMatch(t, []string{"hello", "world"}, result)
	})

	t.Run("mixed case", func(t *testing.T) {
		result := Norm("Hello WORLD")
		assert.ElementsMatch(t, []string{"hello", "world"}, result)
	})

	t.Run("stop words are filtered", func(t *testing.T) {
		result := Norm("the quick brown fox jumps over the lazy dog")
		// Stop words like "the", "over" should be filtered out
		expected := []string{"quick", "brown", "fox", "jump", "lazi", "dog"}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("stemming works", func(t *testing.T) {
		result := Norm("running runs runner")
		// "running" and "runs" stem to "run", "runner" stays "runner"
		assert.ElementsMatch(t, []string{"run", "runner"}, result)
	})

	t.Run("numbers are included", func(t *testing.T) {
		result := Norm("test123 word456")
		assert.ElementsMatch(t, []string{"test123", "word456"}, result)
	})

	t.Run("special characters are separators", func(t *testing.T) {
		result := Norm("hello@world.com test#value")
		assert.ElementsMatch(t, []string{"hello", "world", "com", "test", "valu"}, result)
	})

	t.Run("duplicates are removed", func(t *testing.T) {
		result := Norm("hello world hello")
		assert.Equal(t, []string{"hello", "world"}, result)
	})

	t.Run("complex phrase", func(t *testing.T) {
		phrase := "The quick brown fox jumps over the lazy dog! This is a test: 123@example.com"
		result := Norm(phrase)

		// Should contain stemmed words, no stop words, no duplicates
		expected := []string{
			"quick", "brown", "fox", "jump", "lazi", "dog",
			"test", "123", "exampl", "com",
		}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("only punctuation", func(t *testing.T) {
		result := Norm("!@#$%^&*()")
		assert.Empty(t, result)
	})

	t.Run("only stop words", func(t *testing.T) {
		result := Norm("the and or but")
		assert.Empty(t, result)
	})

	t.Run("empty words after processing", func(t *testing.T) {
		result := Norm("   ")
		assert.Empty(t, result)
	})
}
