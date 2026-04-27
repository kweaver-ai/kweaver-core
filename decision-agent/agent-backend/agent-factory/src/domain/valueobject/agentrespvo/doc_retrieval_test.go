package agentrespvo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocRetrievalRes_AnswerAndCites(t *testing.T) {
	t.Parallel()

	res := &DocRetrievalRes{
		Answer: DocRetrievalAnswer{
			Result: "test answer",
			FullResult: FullResult{
				Text: "full answer text",
				References: []*Reference{
					{
						Content:            "reference content 1",
						RetrieveSourceType: "doc",
						Score:              0.95,
						Meta: &Meta{
							DocID:   "doc1",
							DocName: "Document 1",
						},
					},
					{
						Content:            "reference content 2",
						RetrieveSourceType: "web",
						Score:              0.87,
						Meta: &Meta{
							DocID:   "doc2",
							DocName: "Document 2",
						},
					},
				},
			},
		},
	}

	answer, cites := res.AnswerAndCites()

	assert.Equal(t, "full answer text", answer)
	assert.Len(t, cites, 2)

	// Check first cite
	assert.Equal(t, "reference content 1", cites[0].Content)
	assert.Equal(t, "doc", cites[0].CiteType)
	assert.Equal(t, 0.95, cites[0].Score)
	assert.NotNil(t, cites[0].Meta)

	// Check second cite
	assert.Equal(t, "reference content 2", cites[1].Content)
	assert.Equal(t, "web", cites[1].CiteType)
	assert.Equal(t, 0.87, cites[1].Score)
}

func TestDocRetrievalRes_AnswerAndCites_NoReferences(t *testing.T) {
	t.Parallel()

	res := &DocRetrievalRes{
		Answer: DocRetrievalAnswer{
			Result: "test answer",
			FullResult: FullResult{
				Text:       "full answer text",
				References: []*Reference{},
			},
		},
	}

	answer, cites := res.AnswerAndCites()

	assert.Equal(t, "full answer text", answer)
	assert.Empty(t, cites)
}

func TestDocRetrievalRes_AnswerAndCites_NilReferences(t *testing.T) {
	t.Parallel()

	res := &DocRetrievalRes{
		Answer: DocRetrievalAnswer{
			Result: "test answer",
			FullResult: FullResult{
				Text:       "full answer text",
				References: nil,
			},
		},
	}

	answer, cites := res.AnswerAndCites()

	assert.Equal(t, "full answer text", answer)
	assert.Nil(t, cites)
}

func TestDocRetrievalRes_AnswerAndCites_WithSlices(t *testing.T) {
	t.Parallel()

	res := &DocRetrievalRes{
		Answer: DocRetrievalAnswer{
			Result: "test answer",
			FullResult: FullResult{
				Text: "full answer text",
				References: []*Reference{
					{
						Content:            "reference with slices",
						RetrieveSourceType: "doc",
						Score:              0.90,
						Meta: &Meta{
							DocID:   "doc1",
							DocName: "Document 1",
							Slices: []*Slice{
								{
									Score:   0.92,
									ID:      "slice1",
									No:      1,
									Content: "slice content 1",
									Pages:   []int{1, 2},
								},
							},
						},
					},
				},
			},
		},
	}

	answer, cites := res.AnswerAndCites()

	assert.Equal(t, "full answer text", answer)
	assert.Len(t, cites, 1)
	assert.Len(t, cites[0].Meta.(*Meta).Slices, 1)
	assert.Equal(t, "slice1", cites[0].Meta.(*Meta).Slices[0].ID)
}
