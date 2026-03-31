// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package logics contains business logic implementations.
package logics

import (
	"database/sql"

	"vega-backend/interfaces"
)

var (
	DB  *sql.DB
	AA  interfaces.AuthAccess
	PA  interfaces.PermissionAccess
	UMA interfaces.UserMgmtAccess
)

func SetDB(db *sql.DB) {
	DB = db
}

func SetAuthAccess(aa interfaces.AuthAccess) {
	AA = aa
}

func SetPermissionAccess(pa interfaces.PermissionAccess) {
	PA = pa
}

func SetUserMgmtAccess(uma interfaces.UserMgmtAccess) {
	UMA = uma
}
