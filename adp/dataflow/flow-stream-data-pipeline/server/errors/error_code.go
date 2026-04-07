// Package errors 服务错误码
package errors

import (
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"flow-stream-data-pipeline/locale"
)

const (
	//400
	StreamDataPipeline_CountExceeded_Tags              = "StreamDataPipeline.CountExceeded.Tags"
	StreamDataPipeline_CountExceeded_Pipelines         = "StreamDataPipeline.CountExceeded.Pipelines"
	StreamDataPipeline_Duplicated_PipelineID           = "StreamDataPipeline.Duplicated.PipelineID"
	StreamDataPipeline_Duplicated_PipelineName         = "StreamDataPipeline.Duplicated.PipelineName"
	StreamDataPipeline_InvalidParameter_Builtin        = "StreamDataPipeline.InvalidParameter.Builtin"
	StreamDataPipeline_InvalidParameter_Direction      = "StreamDataPipeline.InvalidParameter.Direction"
	StreamDataPipeline_InvalidParameter_IndexBase      = "StreamDataPipeline.InvalidParameter.IndexBase"
	StreamDataPipeline_InvalidParameter_Limit          = "StreamDataPipeline.InvalidParameter.Limit"
	StreamDataPipeline_InvalidParameter_Offset         = "StreamDataPipeline.InvalidParameter.Offset"
	StreamDataPipeline_InvalidParameter_OutputType     = "StreamDataPipeline.InvalidParameter.OutputType"
	StreamDataPipeline_InvalidParameter_PipelineID     = "StreamDataPipeline.InvalidParameter.PipelineID"
	StreamDataPipeline_InvalidParameter_PipelineStatus = "StreamDataPipeline.InvalidParameter.PipelineStatus"
	StreamDataPipeline_InvalidParameter_RequestBody    = "StreamDataPipeline.InvalidParameter.RequestBody"
	StreamDataPipeline_InvalidParameter_Sort           = "StreamDataPipeline.InvalidParameter.Sort"
	StreamDataPipeline_InvalidParameter_Tag            = "StreamDataPipeline.InvalidParameter.Tag"
	StreamDataPipeline_LengthExceeded_Comment          = "StreamDataPipeline.LengthExceeded.Comment"
	StreamDataPipeline_LengthExceeded_PipelineName     = "StreamDataPipeline.LengthExceeded.PipelineName"
	StreamDataPipeline_LengthExceeded_Tag              = "StreamDataPipeline.LengthExceeded.Tag"
	StreamDataPipeline_NullParameter_IndexBaseType     = "StreamDataPipeline.NullParameter.IndexBaseType"
	StreamDataPipeline_NullParameter_PipelineID        = "StreamDataPipeline.NullParameter.PipelineID"
	StreamDataPipeline_NullParameter_PipelineName      = "StreamDataPipeline.NullParameter.PipelineName"
	StreamDataPipeline_NullParameter_Tag               = "StreamDataPipeline.NullParameter.Tag"
	StreamDataPipeline_OutOfRange_CpuLimit             = "StreamDataPipeline.OutOfRange.CpuLimit"
	StreamDataPipeline_OutOfRange_MemoryLimit          = "StreamDataPipeline.OutOfRange.MemoryLimit"
	StreamDataPipeline_UnSupported_ResourceType        = "StreamDataPipeline.UnSupported.ResourceType"

	// 404
	StreamDataPipeline_NotFound_Pipeline = "StreamDataPipeline.NotFound.Pipeline"

	// 500
	StreamDataPipeline_InternalError_CheckPipelineIfExistFailed    = "StreamDataPipeline.InternalError.CheckPipelineIfExistFailed"
	StreamDataPipeline_InternalError_CreatePipelineFailed          = "StreamDataPipeline.InternalError.CreatePipelineFailed"
	StreamDataPipeline_InternalError_CreateTopicsFailed            = "StreamDataPipeline.InternalError.CreateTopicsFailed"
	StreamDataPipeline_InternalError_DataBase                      = "StreamDataPipeline.InternalError.DataBase"
	StreamDataPipeline_InternalError_DeleteTopicsFailed            = "StreamDataPipeline.InternalError.DeleteTopicsFailed"
	StreamDataPipeline_InternalError_PauseNotRunningPipelineFailed = "StreamDataPipeline.InternalError.PauseNotRunningPipelineFailed"
	StreamDataPipeline_InternalError_GetIndexBaseByTypeFailed      = "StreamDataPipeline.InternalError.GetIndexBaseByTypeFailed"
	StreamDataPipeline_InternalError_CheckPermissionFailed         = "StreamDataPipeline.InternalError.CheckPermissionFailed"
	StreamDataPipeline_InternalError_CreateResourcesFailed         = "StreamDataPipeline.InternalError.CreateResourcesFailed"
	StreamDataPipeline_InternalError_DeleteResourcesFailed         = "StreamDataPipeline.InternalError.DeleteResourcesFailed"
	StreamDataPipeline_InternalError_FilterResourcesFailed         = "StreamDataPipeline.InternalError.FilterResourcesFailed"
	StreamDataPipeline_InternalError_UpdateResourceFailed          = "StreamDataPipeline.InternalError.UpdateResourceFailed"
	StreamDataPipeline_InternalError_MQPublishMsgFailed            = "StreamDataPipeline.InternalError.MQPublishMsgFailed"

	StreamDataPipeline_NotReady_Pipeline = "StreamDataPipeline.NotReady.Pipeline"
	StreamDataPipeline_Unmarshall_Data   = "StreamDataPipeline.Unmarshall.Data"
)

var (
	errorCodeList = []string{
		StreamDataPipeline_CountExceeded_Tags,
		StreamDataPipeline_CountExceeded_Pipelines,
		StreamDataPipeline_Duplicated_PipelineID,
		StreamDataPipeline_Duplicated_PipelineName,
		StreamDataPipeline_InvalidParameter_Builtin,
		StreamDataPipeline_InvalidParameter_Direction,
		StreamDataPipeline_InvalidParameter_IndexBase,
		StreamDataPipeline_InvalidParameter_Limit,
		StreamDataPipeline_InvalidParameter_Offset,
		StreamDataPipeline_InvalidParameter_OutputType,
		StreamDataPipeline_InvalidParameter_PipelineID,
		StreamDataPipeline_InvalidParameter_PipelineStatus,
		StreamDataPipeline_InvalidParameter_RequestBody,
		StreamDataPipeline_InvalidParameter_Sort,
		StreamDataPipeline_InvalidParameter_Tag,
		StreamDataPipeline_LengthExceeded_Comment,
		StreamDataPipeline_LengthExceeded_PipelineName,
		StreamDataPipeline_LengthExceeded_Tag,
		StreamDataPipeline_NullParameter_IndexBaseType,
		StreamDataPipeline_NullParameter_PipelineID,
		StreamDataPipeline_NullParameter_PipelineName,
		StreamDataPipeline_NullParameter_Tag,
		StreamDataPipeline_OutOfRange_CpuLimit,
		StreamDataPipeline_OutOfRange_MemoryLimit,
		StreamDataPipeline_UnSupported_ResourceType,

		// 404
		StreamDataPipeline_NotFound_Pipeline,

		// 500
		StreamDataPipeline_InternalError_CheckPipelineIfExistFailed,
		StreamDataPipeline_InternalError_CreatePipelineFailed,
		StreamDataPipeline_InternalError_CreateTopicsFailed,
		StreamDataPipeline_InternalError_DataBase,
		StreamDataPipeline_InternalError_DeleteTopicsFailed,
		StreamDataPipeline_InternalError_PauseNotRunningPipelineFailed,
		StreamDataPipeline_InternalError_GetIndexBaseByTypeFailed,
		StreamDataPipeline_InternalError_CheckPermissionFailed,
		StreamDataPipeline_InternalError_CreateResourcesFailed,
		StreamDataPipeline_InternalError_DeleteResourcesFailed,
		StreamDataPipeline_InternalError_FilterResourcesFailed,
		StreamDataPipeline_InternalError_UpdateResourceFailed,
		StreamDataPipeline_InternalError_MQPublishMsgFailed,

		StreamDataPipeline_NotReady_Pipeline,
		StreamDataPipeline_Unmarshall_Data,
	}
)

func init() {
	locale.Register()
	rest.Register(errorCodeList)
}
