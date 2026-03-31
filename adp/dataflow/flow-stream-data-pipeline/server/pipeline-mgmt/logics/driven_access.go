package logics

import "flow-stream-data-pipeline/pipeline-mgmt/interfaces"

var (
	IBAccess interfaces.IndexBaseAccess
	MQAccess interfaces.MQAccess
	PMAccess interfaces.PipelineMgmtAccess
	PA       interfaces.PermissionAccess
	AA       interfaces.AuthAccess
)

func SetIndexBaseAccess(ibAccess interfaces.IndexBaseAccess) {
	IBAccess = ibAccess
}

func SetMQAccess(mqAccess interfaces.MQAccess) {
	MQAccess = mqAccess
}

func SetPipelineMgmtAccess(pmAccess interfaces.PipelineMgmtAccess) {
	PMAccess = pmAccess
}

func SetPermissionAccess(pa interfaces.PermissionAccess) {
	PA = pa
}

func SetAuthAccess(aa interfaces.AuthAccess) {
	AA = aa
}
