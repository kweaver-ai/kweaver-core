package agentsvc

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/stretchr/testify/assert"
)

func TestAddCiteDocMark(t *testing.T) {
	t.Parallel()

	t.Run("returns unchanged answer when no cites provided", func(t *testing.T) {
		t.Parallel()

		svc := &agentSvc{}
		answer := "This is a simple answer"
		cites := []*agentrespvo.CiteDoc{}

		result := svc.addCiteDocMark(answer, cites)

		assert.Equal(t, answer, result)
	})

	t.Run("returns unchanged answer when cites is nil", func(t *testing.T) {
		t.Parallel()

		svc := &agentSvc{}
		answer := "This is a simple answer"

		result := svc.addCiteDocMark(answer, nil)

		assert.Equal(t, answer, result)
	})

	t.Run("returns empty string when answer is empty", func(t *testing.T) {
		t.Parallel()

		svc := &agentSvc{}
		answer := ""
		cites := []*agentrespvo.CiteDoc{
			{Content: "Test content", Slices: []*agentrespvo.V1Slice{}},
		}

		result := svc.addCiteDocMark(answer, cites)

		assert.Equal(t, "", result)
	})

	t.Run("processes single cite with empty slices", func(t *testing.T) {
		t.Parallel()

		svc := &agentSvc{}
		answer := "This is an answer。"
		cites := []*agentrespvo.CiteDoc{
			{Content: "Test content", Slices: []*agentrespvo.V1Slice{}},
		}

		result := svc.addCiteDocMark(answer, cites)

		assert.NotEmpty(t, result)
		assert.Contains(t, result, "This is an answer")
	})

	t.Run("processes multiple cites with content", func(t *testing.T) {
		t.Parallel()

		svc := &agentSvc{}
		answer := "First sentence。Second sentence。Third sentence。"
		cites := []*agentrespvo.CiteDoc{
			{Content: "First cite content", Slices: []*agentrespvo.V1Slice{{Content: "slice1"}}},
			{Content: "Second cite content", Slices: []*agentrespvo.V1Slice{{Content: "slice2"}}},
		}

		result := svc.addCiteDocMark(answer, cites)

		assert.NotEmpty(t, result)
	})
}

func TestGetSentenceDocScore(t *testing.T) {
	t.Parallel()

	t.Run("returns sentence info with empty doc cites", func(t *testing.T) {
		t.Parallel()

		svc := &agentSvc{}
		sentence := "This is a test sentence"
		docCites := map[int]*docCite{}

		result := svc.getSentenceDocScore(sentence, docCites)

		assert.NotNil(t, result)
		assert.Equal(t, sentence, result.Text)
		assert.Empty(t, result.DocIndexs)
		// NaN is expected when dividing by zero (empty docCites)
		assert.True(t, result.AvgScore != result.AvgScore || result.AvgScore >= 0.0) // NaN check
	})

	t.Run("returns sentence info with nil slices", func(t *testing.T) {
		t.Parallel()

		svc := &agentSvc{}
		sentence := "Test sentence"
		docCites := map[int]*docCite{
			1: {Content: "Cite content", Index: 1, Slices: nil},
		}

		result := svc.getSentenceDocScore(sentence, docCites)

		assert.NotNil(t, result)
		assert.Equal(t, sentence, result.Text)
	})

	t.Run("processes single doc cite with single slice", func(t *testing.T) {
		t.Parallel()

		svc := &agentSvc{}
		sentence := "This is a test sentence about something"
		docCites := map[int]*docCite{
			1: {
				Content: "Test document content",
				Index:   1,
				Slices:  []*agentrespvo.V1Slice{{Content: "test sentence"}},
			},
		}

		result := svc.getSentenceDocScore(sentence, docCites)

		assert.NotNil(t, result)
		assert.NotEmpty(t, result.MaxScoreMap)
	})

	t.Run("processes multiple doc cites", func(t *testing.T) {
		t.Parallel()

		svc := &agentSvc{}
		sentence := "Test sentence"
		docCites := map[int]*docCite{
			1: {Content: "First", Index: 1, Slices: []*agentrespvo.V1Slice{{Content: "content1"}}},
			2: {Content: "Second", Index: 2, Slices: []*agentrespvo.V1Slice{{Content: "content2"}}},
		}

		result := svc.getSentenceDocScore(sentence, docCites)

		assert.NotNil(t, result)
		assert.NotEmpty(t, result.MaxScoreMap)
	})

	t.Run("calculates average score across all cites", func(t *testing.T) {
		t.Parallel()

		svc := &agentSvc{}
		sentence := "Test sentence"
		docCites := map[int]*docCite{
			1: {Content: "First", Index: 1, Slices: []*agentrespvo.V1Slice{{Content: "test"}}},
			2: {Content: "Second", Index: 2, Slices: []*agentrespvo.V1Slice{{Content: "sentence"}}},
		}

		result := svc.getSentenceDocScore(sentence, docCites)

		assert.NotNil(t, result)
		assert.GreaterOrEqual(t, result.AvgScore, 0.0)
	})
}

func TestHlRef(t *testing.T) {
	t.Parallel()

	t.Run("handles sentence with no scores", func(t *testing.T) {
		t.Parallel()

		svc := &agentSvc{}
		sentenceInfo := &sentenceInfo{
			MaxScoreMap: map[string]*maxScoreSlice{},
			AvgScore:    0.0,
			Text:        "Test sentence。",
		}
		sentence := "Test sentence。"

		svc.hlRef(sentenceInfo, sentence)

		assert.NotEmpty(t, sentenceInfo.Text)
	})

	t.Run("handles sentence with scores below cut score", func(t *testing.T) {
		t.Parallel()

		svc := &agentSvc{}
		sentenceInfo := &sentenceInfo{
			MaxScoreMap: map[string]*maxScoreSlice{
				"1": {DocIndex: 1, Score: 0.5, SliceIndex: 0},
			},
			AvgScore: 0.5,
			Text:     "Test sentence。",
		}
		sentence := "Test sentence。"

		svc.hlRef(sentenceInfo, sentence)

		assert.NotEmpty(t, sentenceInfo.Text)
	})

	t.Run("handles sentence with scores above cut score", func(t *testing.T) {
		t.Parallel()

		svc := &agentSvc{}
		sentenceInfo := &sentenceInfo{
			MaxScoreMap: map[string]*maxScoreSlice{
				"1": {DocIndex: 1, Score: 0.9, SliceIndex: 0},
			},
			AvgScore: 0.9,
			Text:     "Test sentence。",
		}
		sentence := "Test sentence。"

		svc.hlRef(sentenceInfo, sentence)

		assert.NotEmpty(t, sentenceInfo.Text)
		assert.NotEmpty(t, sentenceInfo.DocIndexs)
	})

	t.Run("handles sentence with multiple scores above cut score", func(t *testing.T) {
		t.Parallel()

		svc := &agentSvc{}
		sentenceInfo := &sentenceInfo{
			MaxScoreMap: map[string]*maxScoreSlice{
				"1": {DocIndex: 1, Score: 0.9, SliceIndex: 0},
				"2": {DocIndex: 2, Score: 0.85, SliceIndex: 0},
			},
			AvgScore: 0.87,
			Text:     "Test sentence。",
		}
		sentence := "Test sentence。"

		svc.hlRef(sentenceInfo, sentence)

		assert.NotEmpty(t, sentenceInfo.Text)
		assert.Len(t, sentenceInfo.DocIndexs, 2)
	})

	t.Run("handles sentence with avg score plus cap score above cut score", func(t *testing.T) {
		t.Parallel()

		svc := &agentSvc{}
		sentenceInfo := &sentenceInfo{
			MaxScoreMap: map[string]*maxScoreSlice{
				"1": {DocIndex: 1, Score: 0.7, SliceIndex: 0},
			},
			AvgScore: 0.7,
			Text:     "Test sentence。",
		}
		sentence := "Test sentence。"

		svc.hlRef(sentenceInfo, sentence)

		assert.NotEmpty(t, sentenceInfo.Text)
	})

	t.Run("handles sentence with no max score map entries", func(t *testing.T) {
		t.Parallel()

		svc := &agentSvc{}
		sentenceInfo := &sentenceInfo{
			MaxScoreMap: map[string]*maxScoreSlice{},
			AvgScore:    0.0,
			Text:        "Test sentence。",
		}
		sentence := "Test sentence。"

		svc.hlRef(sentenceInfo, sentence)

		assert.NotEmpty(t, sentenceInfo.Text)
		assert.Empty(t, sentenceInfo.DocIndexs)
	})
}

func TestSameWordsPercentage(t *testing.T) {
	t.Parallel()

	t.Run("returns 0.0 for same words percentage", func(t *testing.T) {
		t.Parallel()

		svc := &agentSvc{}
		sentence := "test sentence"
		sliceContent := "test content"

		result := svc.sameWordsPercentage(sentence, sliceContent)

		assert.Equal(t, 0.0, result)
	})

	t.Run("returns 0.0 for empty inputs", func(t *testing.T) {
		t.Parallel()

		svc := &agentSvc{}

		result := svc.sameWordsPercentage("", "")

		assert.Equal(t, 0.0, result)
	})

	t.Run("returns 0.0 for different inputs", func(t *testing.T) {
		t.Parallel()

		svc := &agentSvc{}
		sentence := "completely different text"
		sliceContent := "another unrelated content"

		result := svc.sameWordsPercentage(sentence, sliceContent)

		assert.Equal(t, 0.0, result)
	})
}
