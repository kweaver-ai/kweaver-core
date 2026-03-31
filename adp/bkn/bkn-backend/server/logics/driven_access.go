// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"database/sql"

	"bkn-backend/interfaces"
)

var (
	DB             *sql.DB
	AA             interfaces.AuthAccess
	ASA            interfaces.ActionScheduleAccess
	ATA            interfaces.ActionTypeAccess
	BSA            interfaces.BusinessSystemAccess
	CGA            interfaces.ConceptGroupAccess
	DDA            interfaces.DataModelAccess
	DVA            interfaces.DataViewAccess
	JA             interfaces.JobAccess
	KNA            interfaces.KNAccess
	MFA            interfaces.ModelFactoryAccess
	OSA            interfaces.OpenSearchAccess
	OTA            interfaces.ObjectTypeAccess
	PA             interfaces.PermissionAccess
	RTA            interfaces.RelationTypeAccess
	RiskTypeAccess interfaces.RiskTypeAccess
	UMA            interfaces.UserMgmtAccess
	VBA            interfaces.VegaBackendAccess
)

func SetDB(db *sql.DB) {
	DB = db
}

func SetAuthAccess(aa interfaces.AuthAccess) {
	AA = aa
}

func SetActionScheduleAccess(asa interfaces.ActionScheduleAccess) {
	ASA = asa
}

func SetActionTypeAccess(ata interfaces.ActionTypeAccess) {
	ATA = ata
}

func SetBusinessSystemAccess(bsa interfaces.BusinessSystemAccess) {
	BSA = bsa
}

func SetConceptGroupAccess(cga interfaces.ConceptGroupAccess) {
	CGA = cga
}

func SetDataModelAccess(dda interfaces.DataModelAccess) {
	DDA = dda
}

func SetDataViewAccess(dva interfaces.DataViewAccess) {
	DVA = dva
}

func SetJobAccess(ja interfaces.JobAccess) {
	JA = ja
}

func SetKNAccess(kna interfaces.KNAccess) {
	KNA = kna
}

func SetModelFactoryAccess(mfa interfaces.ModelFactoryAccess) {
	MFA = mfa
}

func SetOpenSearchAccess(osa interfaces.OpenSearchAccess) {
	OSA = osa
}

func SetObjectTypeAccess(ota interfaces.ObjectTypeAccess) {
	OTA = ota
}

func SetPermissionAccess(pa interfaces.PermissionAccess) {
	PA = pa
}

func SetRelationTypeAccess(rta interfaces.RelationTypeAccess) {
	RTA = rta
}

func SetRiskTypeAccess(rta interfaces.RiskTypeAccess) {
	RiskTypeAccess = rta
}

func SetUserMgmtAccess(uma interfaces.UserMgmtAccess) {
	UMA = uma
}

func SetVegaBackendAccess(vba interfaces.VegaBackendAccess) {
	VBA = vba
}
