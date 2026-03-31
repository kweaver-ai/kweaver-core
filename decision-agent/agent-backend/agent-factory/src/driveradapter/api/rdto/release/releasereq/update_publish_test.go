package releasereq

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
	"github.com/stretchr/testify/assert"
)

func TestUpdatePublishInfoReq_StructFields(t *testing.T) {
	t.Parallel()

	req := &UpdatePublishInfoReq{}
	req.CategoryIDs = []string{"cat-1", "cat-2"}
	req.Description = "Test description"
	req.PublishToWhere = []daenum.PublishToWhere{
		daenum.PublishToWhereSquare,
	}
	req.PublishToBes = []cdaenum.PublishToBe{
		cdaenum.PublishToBeAPIAgent,
	}

	assert.Len(t, req.CategoryIDs, 2)
	assert.Equal(t, "Test description", req.Description)
	assert.Len(t, req.PublishToWhere, 1)
	assert.Len(t, req.PublishToBes, 1)
}

func TestUpdatePublishInfoReq_Empty(t *testing.T) {
	t.Parallel()

	req := &UpdatePublishInfoReq{}

	assert.Nil(t, req.CategoryIDs)
	assert.Empty(t, req.Description)
	assert.Nil(t, req.PublishToWhere)
	assert.Nil(t, req.PublishToBes)
}

func TestUpdatePublishInfoReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := &UpdatePublishInfoReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Empty(t, errMsgMap)
}

func TestUpdatePublishInfoReq_CustomCheck_Valid(t *testing.T) {
	t.Parallel()

	req := &UpdatePublishInfoReq{}
	req.PublishToWhere = []daenum.PublishToWhere{
		daenum.PublishToWhereSquare,
		daenum.PublishToWhereCustomSpace,
	}
	req.PublishToBes = []cdaenum.PublishToBe{
		cdaenum.PublishToBeAPIAgent,
		cdaenum.PublishToBeWebSDKAgent,
	}

	err := req.CustomCheck()

	assert.NoError(t, err)
}

func TestUpdatePublishInfoReq_CustomCheck_Empty(t *testing.T) {
	t.Parallel()

	req := &UpdatePublishInfoReq{}

	err := req.CustomCheck()

	assert.NoError(t, err)
}

func TestUpdatePublishInfoReq_CustomCheck_InvalidPublishToWhere(t *testing.T) {
	t.Parallel()

	req := &UpdatePublishInfoReq{}
	req.PublishToWhere = []daenum.PublishToWhere{
		daenum.PublishToWhereSquare,
		daenum.PublishToWhere("invalid"),
	}

	err := req.CustomCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "publish_to_where is invalid")
}

func TestUpdatePublishInfoReq_CustomCheck_InvalidPublishToBes(t *testing.T) {
	t.Parallel()

	req := &UpdatePublishInfoReq{}
	req.PublishToBes = []cdaenum.PublishToBe{
		cdaenum.PublishToBeAPIAgent,
		cdaenum.PublishToBe("invalid"),
	}

	err := req.CustomCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "publish_to_bes is invalid")
}

func TestUpdatePublishInfoReq_CustomCheck_AllValidPublishToWhere(t *testing.T) {
	t.Parallel()

	validTargets := []daenum.PublishToWhere{
		daenum.PublishToWhereSquare,
		daenum.PublishToWhereCustomSpace,
	}

	for _, target := range validTargets {
		req := &UpdatePublishInfoReq{}
		req.PublishToWhere = []daenum.PublishToWhere{target}

		err := req.CustomCheck()
		assert.NoError(t, err, "PublishToWhere %s should be valid", target)
	}
}

func TestUpdatePublishInfoReq_CustomCheck_AllValidPublishToBes(t *testing.T) {
	t.Parallel()

	validTargets := []cdaenum.PublishToBe{
		cdaenum.PublishToBeAPIAgent,
		cdaenum.PublishToBeWebSDKAgent,
		cdaenum.PublishToBeSkillAgent,
		cdaenum.PublishToBeDataFlowAgent,
	}

	for _, target := range validTargets {
		req := &UpdatePublishInfoReq{}
		req.PublishToBes = []cdaenum.PublishToBe{target}

		err := req.CustomCheck()
		assert.NoError(t, err, "PublishToBe %s should be valid", target)
	}
}

func TestUpdatePublishInfoReq_WithCategoryIDs(t *testing.T) {
	t.Parallel()

	req := &UpdatePublishInfoReq{}
	req.CategoryIDs = []string{"cat-1", "cat-2", "cat-3"}

	assert.Len(t, req.CategoryIDs, 3)
	assert.Equal(t, "cat-1", req.CategoryIDs[0])
	assert.Equal(t, "cat-2", req.CategoryIDs[1])
	assert.Equal(t, "cat-3", req.CategoryIDs[2])
}

func TestUpdatePublishInfoReq_WithDescription(t *testing.T) {
	t.Parallel()

	descriptions := []string{
		"Short description",
		"This is a longer description with more details",
		"包含中文的描述",
		"Description with numbers 12345 and special chars !@#$%",
	}

	for _, desc := range descriptions {
		req := &UpdatePublishInfoReq{}
		req.Description = desc

		assert.Equal(t, desc, req.Description)
	}
}

func TestUpdatePublishInfoReq_WithMultiplePublishTargets(t *testing.T) {
	t.Parallel()

	req := &UpdatePublishInfoReq{}
	req.PublishToWhere = []daenum.PublishToWhere{
		daenum.PublishToWhereSquare,
		daenum.PublishToWhereCustomSpace,
	}
	req.PublishToBes = []cdaenum.PublishToBe{
		cdaenum.PublishToBeAPIAgent,
		cdaenum.PublishToBeWebSDKAgent,
		cdaenum.PublishToBeSkillAgent,
	}

	err := req.CustomCheck()
	assert.NoError(t, err)
	assert.Len(t, req.PublishToWhere, 2)
	assert.Len(t, req.PublishToBes, 3)
}

func TestUpdatePublishInfoReq_WithDuplicatePublishTargets(t *testing.T) {
	t.Parallel()

	req := &UpdatePublishInfoReq{}
	req.PublishToWhere = []daenum.PublishToWhere{
		daenum.PublishToWhereSquare,
		daenum.PublishToWhereSquare,
	}
	req.PublishToBes = []cdaenum.PublishToBe{
		cdaenum.PublishToBeAPIAgent,
		cdaenum.PublishToBeAPIAgent,
	}

	err := req.CustomCheck()
	assert.NoError(t, err)
	assert.Len(t, req.PublishToWhere, 2)
	assert.Len(t, req.PublishToBes, 2)
}

func TestUpdatePublishInfoReq_WithNilPmsControl(t *testing.T) {
	t.Parallel()

	req := &UpdatePublishInfoReq{}
	req.PmsControl = nil

	assert.Nil(t, req.PmsControl)
	err := req.CustomCheck()
	assert.NoError(t, err)
}
