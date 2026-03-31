package driveradapters

import (
	"context"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"flow-stream-data-pipeline/common"
	"flow-stream-data-pipeline/pipeline-mgmt/interfaces"
	fmock "flow-stream-data-pipeline/pipeline-mgmt/interfaces/mock"
)

func Test_Validate_ValidatePipeline(t *testing.T) {
	Convey("validate pipeline", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				CpuMax:           8,
				MemoryMax:        8096,
				MaxPipelineCount: 100,
			},
		}

		asMock := fmock.NewMockAuthService(mockCtrl)
		pmService := fmock.NewMockPipelineMgmtService(mockCtrl)
		handler := mockNewPipelineMgmtRestHandler(appSetting, asMock, pmService)
		ctx := context.Background()

		Convey("validate pipeline id failed", func() {
			pipeline := &interfaces.Pipeline{PipelineID: "_"}
			err := handler.ValidatePipeline(ctx, pipeline)
			So(err, ShouldNotBeNil)
		})

		Convey("validate pipeline name failed", func() {
			pipeline := &interfaces.Pipeline{PipelineID: "xyz", PipelineName: ""}
			err := handler.ValidatePipeline(ctx, pipeline)
			So(err, ShouldNotBeNil)
		})

		Convey("validate pipeline tag failed", func() {
			pipeline := &interfaces.Pipeline{
				PipelineID:   "xyz",
				PipelineName: "xyz",
				Tags:         []string{"1", "2", "3", "4", "5", "6"},
			}
			err := handler.ValidatePipeline(ctx, pipeline)
			So(err, ShouldNotBeNil)
		})

		Convey("validate pipeline output type failed", func() {
			pipeline := &interfaces.Pipeline{
				PipelineID:   "xyz",
				PipelineName: "xyz",
				OutputType:   "damie",
			}
			err := handler.ValidatePipeline(ctx, pipeline)
			So(err, ShouldNotBeNil)
		})

		Convey("validate pipeline resource failed", func() {
			pipeline := &interfaces.Pipeline{
				PipelineID:       "xyz",
				PipelineName:     "xyz",
				OutputType:       "index_base",
				DeploymentConfig: &interfaces.DeploymentConfig{CpuLimit: 0},
			}
			err := handler.ValidatePipeline(ctx, pipeline)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_Validate_ValidatePipelineID(t *testing.T) {
	Convey("Test ValidateViewID", t, func() {

		Convey("Validate failed", func() {
			viewID := "xxx&^xxx"
			err := validatePipelineID(testCtx, viewID)
			So(err, ShouldNotBeNil)
		})

		Convey("Validate success", func() {
			viewID := "xx-lxx"
			err := validatePipelineID(testCtx, viewID)
			So(err, ShouldBeNil)
		})

	})
}

func Test_Validate_ValidatePipelineName(t *testing.T) {
	Convey("Test ValidatePipelineName", t, func() {
		Convey("Validate failed, pipeline name is empty", func() {
			pipelineName := ""
			err := validatePipelineName(testCtx, pipelineName)
			So(err, ShouldNotBeNil)
		})

		Convey("Validate failed, pipeline name is too long", func() {
			pipelineName := strings.Repeat("x", 41)
			err := validatePipelineName(testCtx, pipelineName)
			So(err, ShouldNotBeNil)
		})

		Convey("Validate success", func() {
			pipelineName := "xx-lxx"
			err := validatePipelineName(testCtx, pipelineName)
			So(err, ShouldBeNil)
		})
	})
}

func Test_Validate_ValidateTags(t *testing.T) {
	Convey("Test validateTags", t, func() {

		Convey("Validate failed, because the number of tags exceeds the upper limit", func() {
			err := validateTags(testCtx, []string{"a", "b", "c", "d", "e", "f"})
			So(err, ShouldNotBeNil)
		})

		Convey("Validate failed, because some tags are null", func() {
			err := validateTags(testCtx, []string{"a", "", "c"})
			So(err, ShouldNotBeNil)
		})

		Convey("Validate failed, because some tags exceeds the upper length limit", func() {
			err := validateTags(testCtx, []string{"0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRST"})
			So(err, ShouldNotBeNil)
		})

		Convey("Validate failed, because some tags have special character", func() {
			err := validateTags(testCtx, []string{"%[]#"})
			So(err, ShouldNotBeNil)
		})

		Convey("Validate succeed", func() {
			err := validateTags(testCtx, []string{"test"})
			So(err, ShouldBeNil)
		})
	})
}

func Test_Validate_ValidateComment(t *testing.T) {
	Convey("Test validateComment", t, func() {
		Convey("Validate failed, because the comment exceeds the upper length limit", func() {
			str := ""
			for i := 0; i < interfaces.COMMENT_MAX_LENGTH+10; i++ {
				str += "a"
			}

			err := validateComment(testCtx, str)
			So(err, ShouldNotBeNil)
		})

		Convey("Validate succeed", func() {
			err := validateComment(testCtx, "test")
			So(err, ShouldBeNil)
		})
	})
}

func Test_Validate_ValidateOutputType(t *testing.T) {
	Convey("Test validateOutputType", t, func() {
		Convey("Validate failed, because the output type is not supported", func() {
			err := validateOutputType(testCtx, "damie")
			So(err, ShouldNotBeNil)
		})

		Convey("Validate succeed", func() {
			err := validateOutputType(testCtx, "index_base")
			So(err, ShouldBeNil)
		})
	})
}

func Test_Validate_ValidateListPipelinesQuery(t *testing.T) {
	Convey("Test validatePaginationQueryParameters", t, func() {

		Convey("Validate failed, because the offset cannot be converted to int", func() {
			offset := "a"
			limit := "1000"
			sort := "update_time"
			direction := interfaces.ASC_DIRECTION

			_, err := ValidateListPipelinesQuery(testCtx, offset, limit, sort, direction)
			So(err, ShouldNotBeNil)
		})

		Convey("Validate failed, because the offset is not greater than MIN_OFFSET", func() {
			offset := "-1"
			limit := "1000"
			sort := "update_time"
			direction := interfaces.ASC_DIRECTION

			_, err := ValidateListPipelinesQuery(testCtx, offset, limit, sort, direction)
			So(err, ShouldNotBeNil)
		})

		Convey("Validate failed, because the limit cannot be converted to int", func() {
			offset := "0"
			limit := "a"
			sort := "update_time"
			direction := interfaces.ASC_DIRECTION

			_, err := ValidateListPipelinesQuery(testCtx, offset, limit, sort, direction)
			So(err, ShouldNotBeNil)
		})

		Convey("Validate failed, because the limit is not in the range of [MIN_LIMIT,MAX_LIMIT]", func() {
			offset := "0"
			limit := "1100"
			sort := "update_time"
			direction := interfaces.ASC_DIRECTION

			_, err := ValidateListPipelinesQuery(testCtx, offset, limit, sort, direction)

			So(err, ShouldNotBeNil)
		})

		Convey("Validate failed, because the sort type does not belong to any item in set METRIC_MODEL_SORT", func() {
			offset := "0"
			limit := "800"
			sort := "update_time1"
			direction := interfaces.ASC_DIRECTION

			_, err := ValidateListPipelinesQuery(testCtx, offset, limit, sort, direction)

			So(err, ShouldNotBeNil)
		})

		Convey("Validate failed, because the sort direction is not DESC or ASC", func() {
			offset := "0"
			limit := "800"
			sort := "update_time"
			direction := "abc"

			_, err := ValidateListPipelinesQuery(testCtx, offset, limit, sort, direction)

			So(err, ShouldNotBeNil)
		})

		Convey("Validate succeed", func() {
			offset := "0"
			limit := "800"
			sort := "update_time"
			direction := interfaces.ASC_DIRECTION

			_, err := ValidateListPipelinesQuery(testCtx, offset, limit, sort, direction)
			So(err, ShouldBeNil)
		})
	})
}

func Test_Validate_ValidatePipelineStatusInfo(t *testing.T) {
	Convey("Test validatePipelineStatusInfo", t, func() {
		Convey("Validate failed, because the status is not supported", func() {
			err := ValidatePipelineStatusInfo(testCtx, interfaces.PipelineStatusParamter{Status: "abc"})
			So(err, ShouldNotBeNil)
		})

		Convey("Validate succeed", func() {
			err := ValidatePipelineStatusInfo(testCtx, interfaces.PipelineStatusParamter{Status: interfaces.PipelineStatus_Close})
			So(err, ShouldBeNil)
		})
	})
}
