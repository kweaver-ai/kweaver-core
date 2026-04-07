package agentrespvo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocRetrievalField_NewInstance(t *testing.T) {
	t.Parallel()

	field := &DocRetrievalField{}

	assert.NotNil(t, field)
	assert.Equal(t, "", field.Text)
	assert.Nil(t, field.Cites)
}

func TestDocRetrievalField_WithValues(t *testing.T) {
	t.Parallel()

	cites := []*CiteDoc{
		{DocID: "doc1", DocName: "Document 1"},
		{DocID: "doc2", DocName: "Document 2"},
	}

	field := &DocRetrievalField{
		Text:  "Sample text",
		Cites: cites,
	}

	assert.Equal(t, "Sample text", field.Text)
	assert.Len(t, field.Cites, 2)
	assert.Equal(t, "doc1", field.Cites[0].DocID)
}

func TestGraphRetrievalField_NewInstance(t *testing.T) {
	t.Parallel()

	field := &GraphRetrievalField{}

	assert.NotNil(t, field)
	assert.Equal(t, "", field.KGID)
	assert.Equal(t, "", field.KGName)
	assert.Nil(t, field.Subgraph)
	assert.Equal(t, "", field.Text)
	assert.Nil(t, field.Cites)
}

func TestGraphRetrievalField_WithValues(t *testing.T) {
	t.Parallel()

	field := &GraphRetrievalField{
		KGID:     "kg_123",
		KGName:   "Test Knowledge Graph",
		Subgraph: map[string]string{"node1": "value1"},
		Text:     "Graph retrieval text",
		Cites:    []string{"cite1", "cite2"},
	}

	assert.Equal(t, "kg_123", field.KGID)
	assert.Equal(t, "Test Knowledge Graph", field.KGName)
	assert.NotNil(t, field.Subgraph)
	assert.Equal(t, "Graph retrieval text", field.Text)
	assert.NotNil(t, field.Cites)
}

func TestCiteDoc_NewInstance(t *testing.T) {
	t.Parallel()

	cite := &CiteDoc{}

	assert.NotNil(t, cite)
	assert.Equal(t, "", cite.Content)
	assert.Equal(t, "", cite.ExtType)
	assert.Equal(t, "", cite.DocID)
	assert.Equal(t, "", cite.DocName)
	assert.Equal(t, "", cite.ObjectID)
	assert.Equal(t, "", cite.ParentPath)
	assert.Equal(t, int64(0), cite.Size)
	assert.Equal(t, "", cite.Type)
	assert.Nil(t, cite.Slices)
	assert.Equal(t, "", cite.SpaceID)
	assert.Equal(t, "", cite.DocLibType)
}

func TestCiteDoc_WithValues(t *testing.T) {
	t.Parallel()

	slices := []*V1Slice{
		{ID: "slice1", No: 1, Content: "Content 1"},
	}

	cite := &CiteDoc{
		Content:    "Document content",
		ExtType:    ".pdf",
		DocID:      "gns://abc123",
		DocName:    "Test Document.pdf",
		ObjectID:   "abc123",
		ParentPath: "/folder/",
		Size:       1024,
		Type:       "document",
		Slices:     slices,
		SpaceID:    "space_123",
		DocLibType: "knowledge_doc_lib",
	}

	assert.Equal(t, "Document content", cite.Content)
	assert.Equal(t, ".pdf", cite.ExtType)
	assert.Equal(t, "gns://abc123", cite.DocID)
	assert.Equal(t, "Test Document.pdf", cite.DocName)
	assert.Equal(t, "abc123", cite.ObjectID)
	assert.Equal(t, "/folder/", cite.ParentPath)
	assert.Equal(t, int64(1024), cite.Size)
	assert.Equal(t, "document", cite.Type)
	assert.Len(t, cite.Slices, 1)
	assert.Equal(t, "space_123", cite.SpaceID)
	assert.Equal(t, "knowledge_doc_lib", cite.DocLibType)
}

func TestV1Slice_NewInstance(t *testing.T) {
	t.Parallel()

	slice := &V1Slice{}

	assert.NotNil(t, slice)
	assert.Equal(t, "", slice.ID)
	assert.Equal(t, 0, slice.No)
	assert.Equal(t, "", slice.Content)
	assert.Nil(t, slice.Pages)
}

func TestV1Slice_WithValues(t *testing.T) {
	t.Parallel()

	slice := &V1Slice{
		ID:      "slice_123",
		No:      5,
		Content: "Slice content",
		Pages:   []int{1, 2, 3},
	}

	assert.Equal(t, "slice_123", slice.ID)
	assert.Equal(t, 5, slice.No)
	assert.Equal(t, "Slice content", slice.Content)
	assert.Len(t, slice.Pages, 3)
}
