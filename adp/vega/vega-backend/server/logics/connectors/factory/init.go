// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package factory

import (
	"vega-backend/interfaces"
	"vega-backend/logics/connectors/local/fileset/anyshare"
	"vega-backend/logics/connectors/local/index/opensearch"
	"vega-backend/logics/connectors/local/table/mariadb"
	"vega-backend/logics/connectors/local/table/postgresql"
)

// InitLocalConnectors 初始化本地 connector
func (cf *ConnectorFactory) InitLocalConnectors() {
	cf.connectors[interfaces.ConnectorTypeMySQL] = mariadb.NewMariaDBConnector()
	cf.connectors[interfaces.ConnectorTypeOpenSearch] = opensearch.NewOpenSearchConnector()
	//cf.connectors[interfaces.ConnectorTypeOracle] = oracle.NewOracleConnector()
	cf.connectors[interfaces.ConnectorTypeMariaDB] = mariadb.NewMariaDBConnector()
	cf.connectors[interfaces.ConnectorTypePostgreSQL] = postgresql.NewPostgresqlConnector()
	cf.connectors[interfaces.ConnectorTypeAnyShare] = anyshare.NewAnyShareConnector()
}
