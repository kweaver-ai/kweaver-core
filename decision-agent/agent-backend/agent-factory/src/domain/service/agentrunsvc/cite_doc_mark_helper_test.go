package agentsvc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkInDocIndex(t *testing.T) {
	t.Parallel()

	t.Run("returns unchanged text when no patterns match", func(t *testing.T) {
		t.Parallel()

		text := "This is a simple text with no references"
		has, docIndexs, newText := markInDocIndex(text, docRefPatternList)

		assert.False(t, has)
		assert.Empty(t, docIndexs)
		assert.Equal(t, text, newText)
	})

	t.Run("marks reference to single document", func(t *testing.T) {
		t.Parallel()

		text := "第1个参考信息"
		has, docIndexs, _ := markInDocIndex(text, docRefPatternList)

		assert.True(t, has)
		assert.Contains(t, docIndexs, "1")
	})

	t.Run("marks reference to multiple documents", func(t *testing.T) {
		t.Parallel()

		text := "第1个和第4个参考信息"
		has, docIndexs, _ := markInDocIndex(text, docRefPatternList)

		assert.True(t, has)
		assert.Contains(t, docIndexs, "1")
		assert.Contains(t, docIndexs, "4")
	})

	t.Run("marks reference in parentheses format", func(t *testing.T) {
		t.Parallel()

		text := "（参考信息1）"
		has, docIndexs, newText := markInDocIndex(text, docRefPatternList)

		assert.True(t, has)
		assert.Contains(t, docIndexs, "1")
		assert.Contains(t, newText, "p='r'")
	})

	t.Run("marks reference with comma separator", func(t *testing.T) {
		t.Parallel()

		text := "（参考信息1, 4）"
		has, docIndexs, _ := markInDocIndex(text, docRefPatternList)

		assert.True(t, has)
		assert.Contains(t, docIndexs, "1")
		assert.Contains(t, docIndexs, "4")
	})

	t.Run("marks reference document ID format", func(t *testing.T) {
		t.Parallel()

		text := "（参考文档ID：第2个）"
		has, docIndexs, newText := markInDocIndex(text, docRefPatternList)

		assert.True(t, has)
		assert.Contains(t, docIndexs, "2")
		assert.Contains(t, newText, "p='r'")
	})

	t.Run("marks reference document without prefix", func(t *testing.T) {
		t.Parallel()

		text := "参考文档第3个"
		has, docIndexs, newText := markInDocIndex(text, docRefPatternList)

		assert.True(t, has)
		assert.Contains(t, docIndexs, "3")
		assert.Contains(t, newText, "p='r'")
	})

	t.Run("marks multiple references with dunhao separator", func(t *testing.T) {
		t.Parallel()

		text := "参考文档第1个、第4个"
		has, docIndexs, _ := markInDocIndex(text, docRefPatternList)

		assert.True(t, has)
		assert.Contains(t, docIndexs, "1")
		assert.Contains(t, docIndexs, "4")
	})

	t.Run("handles text with mixed content and references", func(t *testing.T) {
		t.Parallel()

		text := "根据第1个参考信息，我们得出结论。"
		has, docIndexs, newText := markInDocIndex(text, docRefPatternList)

		assert.True(t, has)
		assert.Contains(t, docIndexs, "1")
		assert.Contains(t, newText, "根据")
		assert.Contains(t, newText, "结论")
	})

	t.Run("handles empty text", func(t *testing.T) {
		t.Parallel()

		text := ""
		has, docIndexs, newText := markInDocIndex(text, docRefPatternList)

		assert.False(t, has)
		assert.Empty(t, docIndexs)
		assert.Equal(t, "", newText)
	})
}

func TestStringIndexInfoKey(t *testing.T) {
	t.Parallel()

	t.Run("generates unique key for start and end positions", func(t *testing.T) {
		t.Parallel()

		info := &stringIndexInfo{
			Start: 10,
			End:   20,
			Value: "test",
		}

		key := info.key()
		expectedKey := "10:20"

		assert.Equal(t, expectedKey, key)
	})

	t.Run("generates different keys for different positions", func(t *testing.T) {
		t.Parallel()

		info1 := &stringIndexInfo{Start: 0, End: 5, Value: "first"}
		info2 := &stringIndexInfo{Start: 10, End: 15, Value: "second"}

		key1 := info1.key()
		key2 := info2.key()

		assert.NotEqual(t, key1, key2)
	})

	t.Run("generates same key for same positions", func(t *testing.T) {
		t.Parallel()

		info1 := &stringIndexInfo{Start: 5, End: 10, Value: "first"}
		info2 := &stringIndexInfo{Start: 5, End: 10, Value: "second"}

		key1 := info1.key()
		key2 := info2.key()

		assert.Equal(t, key1, key2)
	})
}

func TestSentenceInfo(t *testing.T) {
	t.Parallel()

	t.Run("creates sentence info with default values", func(t *testing.T) {
		t.Parallel()

		info := &sentenceInfo{
			MaxScoreMap: make(map[string]*maxScoreSlice),
		}

		assert.NotNil(t, info.MaxScoreMap)
		assert.Empty(t, info.MaxScoreMap)
		assert.Empty(t, info.DocIndexs)
		assert.Equal(t, 0.0, info.AvgScore)
		assert.False(t, info.HasRefrence)
		assert.Empty(t, info.Text)
	})

	t.Run("creates sentence info with provided text", func(t *testing.T) {
		t.Parallel()

		info := &sentenceInfo{
			Text: "test sentence",
		}

		assert.Equal(t, "test sentence", info.Text)
	})

	t.Run("creates sentence info with scores", func(t *testing.T) {
		t.Parallel()

		info := &sentenceInfo{
			MaxScoreMap: map[string]*maxScoreSlice{
				"1": {DocIndex: 1, Score: 0.8, SliceIndex: 0},
				"2": {DocIndex: 2, Score: 0.9, SliceIndex: 1},
			},
			AvgScore: 0.85,
		}

		assert.Len(t, info.MaxScoreMap, 2)
		assert.Equal(t, 0.85, info.AvgScore)
	})
}

func TestMaxScoreSlice(t *testing.T) {
	t.Parallel()

	t.Run("creates max score slice with all fields", func(t *testing.T) {
		t.Parallel()

		mss := &maxScoreSlice{
			DocIndex:   1,
			Score:      0.95,
			SliceIndex: 2,
		}

		assert.Equal(t, 1, mss.DocIndex)
		assert.Equal(t, 0.95, mss.Score)
		assert.Equal(t, 2, mss.SliceIndex)
	})
}
