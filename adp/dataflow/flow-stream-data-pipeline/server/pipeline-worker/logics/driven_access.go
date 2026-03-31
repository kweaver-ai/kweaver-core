package logics

import "flow-stream-data-pipeline/pipeline-worker/interfaces"

var (
	IBAccess interfaces.IndexBaseAccess
	MQAccess interfaces.MQAccess
	OSAccess interfaces.OpenSearchAccess
	PMAccess interfaces.PipelineMgmtAccess
)

func SetIndexMgmtAccess(imAccess interfaces.IndexBaseAccess) {
	IBAccess = imAccess
}

func SetMQAccess(mqAccess interfaces.MQAccess) {
	MQAccess = mqAccess
}

func SetOpensearchAccess(osAccess interfaces.OpenSearchAccess) {
	OSAccess = osAccess
}

func SetPipelineMgmtAccess(pmAccess interfaces.PipelineMgmtAccess) {
	PMAccess = pmAccess
}
