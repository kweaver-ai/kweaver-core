// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"ontology-query/interfaces"
)

var (
	AA  interfaces.AuthAccess
	AOA interfaces.AgentOperatorAccess
	MFA interfaces.ModelFactoryAccess
	OMA interfaces.OntologyManagerAccess
	OSA interfaces.OpenSearchAccess
	UA  interfaces.UniqueryAccess
	VBA interfaces.VegaBackendAccess
)

func SetAuthAccess(aa interfaces.AuthAccess) {
	AA = aa
}

func SetAgentOperatorAccess(aoa interfaces.AgentOperatorAccess) {
	AOA = aoa
}

func SetModelFactoryAccess(mfa interfaces.ModelFactoryAccess) {
	MFA = mfa
}

func SetOntologyManagerAccess(ota interfaces.OntologyManagerAccess) {
	OMA = ota
}

func SetOpenSearchAccess(osa interfaces.OpenSearchAccess) {
	OSA = osa
}

func SetUniqueryAccess(ua interfaces.UniqueryAccess) {
	UA = ua
}

func SetVegaBackendAccess(v interfaces.VegaBackendAccess) {
	VBA = v
}
