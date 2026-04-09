// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package factory

import (
	"vega-backend/logics/connectors/local/fileset/anyshare"
	"vega-backend/logics/connectors/local/index/opensearch"
	"vega-backend/logics/connectors/local/table/mariadb"
	"vega-backend/logics/connectors/local/table/postgresql"
)

// InitLocalConnectors 初始化本地 connector
func (cf *ConnectorFactory) InitLocalConnectors() {
	cf.connectors["mysql"] = mariadb.NewMariaDBConnector()
	cf.connectors["opensearch"] = opensearch.NewOpenSearchConnector()
	//cf.connectors["oracle"] = oracle.NewOracleConnector()
	cf.connectors["mariadb"] = mariadb.NewMariaDBConnector()
	cf.connectors["postgresql"] = postgresql.NewPostgresqlConnector()
	cf.connectors["anyshare"] = anyshare.NewAnyShareConnector()
}
