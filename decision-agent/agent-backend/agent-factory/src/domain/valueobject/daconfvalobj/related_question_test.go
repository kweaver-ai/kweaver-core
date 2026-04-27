package daconfvalobj

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRelatedQuestion_ValObjCheck(t *testing.T) {
	t.Parallel()

	rq := &RelatedQuestion{IsEnabled: true}
	err := rq.ValObjCheck()
	assert.NoError(t, err)

	rq = &RelatedQuestion{IsEnabled: false}
	err = rq.ValObjCheck()
	assert.NoError(t, err)
}
