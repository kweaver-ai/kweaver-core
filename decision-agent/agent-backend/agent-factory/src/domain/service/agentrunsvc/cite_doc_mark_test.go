package agentsvc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitSentences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		text           string
		minLength      int
		expectedLength int
		checkContents  []string
	}{
		{
			name:           "empty text",
			text:           "",
			minLength:      10,
			expectedLength: 0,
			checkContents:  []string{},
		},
		{
			name:           "single short sentence",
			text:           "Hello。",
			minLength:      10,
			expectedLength: 1,
			checkContents:  []string{"Hello。"},
		},
		{
			name:           "single long sentence",
			text:           "This is a very long sentence that exceeds the minimum length requirement for splitting。",
			minLength:      10,
			expectedLength: 1,
			checkContents:  []string{"This is a very long sentence that exceeds the minimum length requirement for splitting。"},
		},
		{
			name:           "multiple short sentences combined",
			text:           "Hi。Hello。Hey。",
			minLength:      10,
			expectedLength: 2, // "Hi。Hello。" (9 chars) + "Hey。" (4 chars) - first is < 10 so it continues
			checkContents:  []string{"Hi。Hello。", "Hey。"},
		},
		{
			name:           "multiple sentences with different punctuation",
			text:           "First sentence。Second sentence！Third sentence？",
			minLength:      5,
			expectedLength: 3,
			checkContents: []string{
				"First sentence。",
				"Second sentence！",
				"Third sentence？",
			},
		},
		{
			name:           "text without sentence endings",
			text:           "This is just a fragment",
			minLength:      10,
			expectedLength: 1,
			checkContents:  []string{"This is just a fragment"},
		},
		{
			name:           "long text split into multiple sentences",
			text:           "This is the first sentence that is quite long。And here is the second sentence that is also long。Finally the third sentence completes the text。",
			minLength:      20,
			expectedLength: 3,
			checkContents: []string{
				"This is the first sentence that is quite long。",
				"And here is the second sentence that is also long。",
				"Finally the third sentence completes the text。",
			},
		},
		{
			name:           "mixed punctuation - question marks",
			text:           "How are you? I am fine。Thank you!",
			minLength:      5,
			expectedLength: 2, // Combines "How are you? I am fine。" then "Thank you!"
			checkContents: []string{
				"How are you? I am fine。",
				"Thank you!",
			},
		},
		{
			name:           "very long text combined",
			text:           "Short。Another short。One more short but combined into a longer segment that meets minimum。",
			minLength:      30,
			expectedLength: 1,
			checkContents: []string{
				"Short。Another short。One more short but combined into a longer segment that meets minimum。",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := splitSentences(tt.text, tt.minLength)

			assert.Equal(t, tt.expectedLength, len(result), "Number of sentences should match")

			if len(tt.checkContents) > 0 {
				for i, expectedContent := range tt.checkContents {
					if i < len(result) {
						assert.Contains(t, result[i], expectedContent, "Sentence %d should contain expected content", i)
					}
				}
			}
		})
	}
}
