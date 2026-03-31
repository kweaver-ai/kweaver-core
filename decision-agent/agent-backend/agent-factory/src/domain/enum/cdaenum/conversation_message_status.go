package cdaenum

import "github.com/pkg/errors"

// 消息类型
// 消息状态，随Role类型变化，Role为 User时：Received :  已接收(消息成功接收并持久化， 初始状态)Processed: 处理完成（成功触发后续的Agent Call）；
// Role为Assistant时：Processing： 生成中（消息正在生成中 ， 初始状态）Succeded： 生成成功（消息处理完成，返回成功）Failed： 生成失败（消息生成失败）Cancelled: 取消生成（用户、系统终止会话）
type ConversationMsgStatus string

func (t ConversationMsgStatus) EnumCheck() (err error) {
	if t != MsgStatusReceived && t != MsgStatusProcessed && t != MsgStatusSucceded && t != MsgStatusFailed && t != MsgStatusCancelled && t != MsgStatusProcessing {
		err = errors.New("消息状态不合法")
		return
	}

	return
}

const (
	// 已接收
	MsgStatusReceived ConversationMsgStatus = "received"
	// 处理完成
	MsgStatusProcessed ConversationMsgStatus = "processed"
	// 生成中
	MsgStatusProcessing ConversationMsgStatus = "processing"
	// 生成成功
	MsgStatusSucceded ConversationMsgStatus = "succeded"
	// 生成失败
	MsgStatusFailed ConversationMsgStatus = "failed"
	// 已取消
	MsgStatusCancelled ConversationMsgStatus = "cancelled"
)
