package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/dlclark/regexp2"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	serrors "flow-stream-data-pipeline/errors"
	"flow-stream-data-pipeline/pipeline-mgmt/interfaces"
)

// 校验任务信息
func (r *restHandler) ValidatePipeline(ctx context.Context, pipeline *interfaces.Pipeline) error {
	err := validatePipelineID(ctx, pipeline.PipelineID)
	if err != nil {
		return err
	}

	err = validatePipelineName(ctx, pipeline.PipelineName)
	if err != nil {
		return err
	}

	err = validateTags(ctx, pipeline.Tags)
	if err != nil {
		return err
	}

	// 去掉tag前后空格以及数组去重
	pipeline.Tags = libCommon.TagSliceTransform(pipeline.Tags)

	err = validateComment(ctx, pipeline.Comment)
	if err != nil {
		return err
	}

	err = validateOutputType(ctx, pipeline.OutputType)
	if err != nil {
		return err
	}

	// 校验cpu限额和内存限额
	err = r.ValidateResource(ctx, pipeline.DeploymentConfig)
	if err != nil {
		return err
	}
	return nil
}

func validatePipelineID(ctx context.Context, pipelineID string) error {
	if pipelineID != "" {
		// 管道 id 不能超过40个字符, 只包含小写英文字母和数字和连字符(-)
		// 不能包含下划线，因为k8s deploy 名称不允许包含下划线
		re := regexp2.MustCompile(interfaces.RegexPattern_ID, regexp2.RE2)
		match, err := re.MatchString(pipelineID)
		if err != nil || !match {
			errDetails := `The ID can contain lowercase letters, digits and hyphen(-), and it cannot start with a hyphen and exceed 40 characters`
			return rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_InvalidParameter_PipelineID).
				WithErrorDetails(errDetails)
		}
	}

	return nil
}

// 对象名称合法性校验
func validatePipelineName(ctx context.Context, name string) error {
	if name == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_NullParameter_PipelineName)
	}

	if utf8.RuneCountInString(name) > interfaces.OBJECT_NAME_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_LengthExceeded_PipelineName).
			WithErrorDetails(fmt.Sprintf("The length of pipeline name '%s' exceeds %v", name, interfaces.OBJECT_NAME_MAX_LENGTH))
	}

	return nil
}

// 标签数组合法性校验
func validateTags(ctx context.Context, tags []string) error {
	if len(tags) > interfaces.TAGS_MAX_NUMBER {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_CountExceeded_Tags).
			WithErrorDetails(fmt.Sprintf("The length of the tag array exceeds %v", interfaces.TAGS_MAX_NUMBER))
	}

	for _, tag := range tags {
		// 去除tag的左右空格
		tag = strings.Trim(tag, " ")

		if tag == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_NullParameter_Tag)
		}

		if utf8.RuneCountInString(tag) > interfaces.OBJECT_NAME_MAX_LENGTH {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_LengthExceeded_Tag).
				WithErrorDetails(fmt.Sprintf("The length of some tags in the tag array exceeds %d", interfaces.OBJECT_NAME_MAX_LENGTH))
		}

		if isInvalid := strings.ContainsAny(interfaces.TAG_NAME_INVALID_CHARACTER, tag); isInvalid {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_InvalidParameter_Tag).
				WithErrorDetails(fmt.Sprintf("The tag contains special characters, such as %s", interfaces.TAG_NAME_INVALID_CHARACTER))
		}
	}
	return nil
}

// 备注合法性校验
func validateComment(ctx context.Context, comment string) error {
	if utf8.RuneCountInString(comment) > interfaces.COMMENT_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_LengthExceeded_Comment).
			WithErrorDetails(fmt.Sprintf("The length of the comment exceeds %v", interfaces.COMMENT_MAX_LENGTH))
	}
	return nil
}

// 校验数据输出类型
func validateOutputType(ctx context.Context, outputType string) error {
	if _, ok := interfaces.OutputTypeMap[outputType]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_InvalidParameter_OutputType).
			WithErrorDetails(fmt.Sprintf("Unsupport output type '%s'", outputType))
	}

	return nil
}

// 校验 cpu 限额和内存限额
func (r *restHandler) ValidateResource(ctx context.Context, dcfg *interfaces.DeploymentConfig) error {
	// 校验 cpu 限额参数是否超过范围
	if dcfg.CpuLimit < interfaces.CPU_LIMIT_MIN || dcfg.CpuLimit > r.appSetting.ServerSetting.CpuMax {
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			serrors.StreamDataPipeline_OutOfRange_CpuLimit).WithErrorDetails("The cpu limit parameter is out of range")
	}

	// 校验内存限额参数是否超过范围
	if dcfg.MemoryLimit < interfaces.MEMORY_LIMIT_MIN || dcfg.MemoryLimit > r.appSetting.ServerSetting.MemoryMax {
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			serrors.StreamDataPipeline_OutOfRange_MemoryLimit).WithErrorDetails("The memory limit parameter is out of range")
	}

	return nil
}

// // 对象名称精确值和模糊值的校验
// func validateNameandNamePattern(ctx context.Context, name, namePattern string) error {
// 	// name_pattern和name不能同时存在
// 	if namePattern != "" && name != "" {
// 		return rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_ConflictParameter_NameAndNamePatternCoexist).
// 			WithErrorDetails("Parameters name_pattern and name are passed in at the same time")
// 	}

// 	return nil
// }

// 校验分页获取参数
func ValidateListPipelinesQuery(ctx context.Context, offset, limit, sort, direction string) (pq interfaces.PaginationQueryParameters, err error) {
	off, err := strconv.Atoi(offset)
	if err != nil {
		return pq, rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_InvalidParameter_Offset).
			WithErrorDetails(err.Error())
	}
	if off < interfaces.MIN_OFFSET {
		return pq, rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_InvalidParameter_Offset).
			WithErrorDetails("The offset is not greater than 0")
	}

	lim, err := strconv.Atoi(limit)
	if err != nil {
		return pq, rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_InvalidParameter_Limit).
			WithErrorDetails(err.Error())
	}
	if !(lim == interfaces.NO_LIMIT || (lim >= interfaces.MIN_LIMIT && lim <= interfaces.MAX_LIMIT)) {
		return pq, rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_InvalidParameter_Limit).
			WithErrorDetails("Number per page is not [1,1000]")
	}

	_, ok := interfaces.TABLE_SORT[sort]
	if !ok {
		types := make([]string, 0)
		for t := range interfaces.TABLE_SORT {
			types = append(types, t)
		}
		return pq, rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_InvalidParameter_Sort).
			WithErrorDetails(fmt.Sprintf("Wrong sort type, does not belong to any item in set %v ", types))
	}

	if direction != interfaces.DESC_DIRECTION && direction != interfaces.ASC_DIRECTION {
		return pq, rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_InvalidParameter_Direction).
			WithErrorDetails("The sort direction is not DESC or ASC")
	}

	pq.Direction = direction
	pq.Limit = lim
	pq.Offset = off
	pq.Sort = interfaces.TABLE_SORT[sort]
	return pq, nil
}

// 校验任务状态和任务详情，开启或关闭任务时，只有 close，running; 当更新任务状态为 failed 时，details必须存在
func ValidatePipelineStatusInfo(ctx context.Context, status interfaces.PipelineStatusParamter) error {
	if status.Status != interfaces.PipelineStatus_Close && status.Status != interfaces.PipelineStatus_Running &&
		status.Status != interfaces.PipelineStatus_Error {

		return rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_InvalidParameter_PipelineStatus).
			WithErrorDetails("the value of pipeline status is not close, running or error")
	}

	// if status.Status == interfaces.PipelineStatus_Error && status.Details == "" {
	// 	return rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_InvalidParameter_PipelineStatus).
	// 		WithErrorDetails("when pipeline status is failed, pipeline status details cannot be empty")
	// }

	return nil
}
