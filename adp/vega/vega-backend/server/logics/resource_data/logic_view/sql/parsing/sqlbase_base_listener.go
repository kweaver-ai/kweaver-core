// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Code generated from SqlBase.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parsing // SqlBase
import "github.com/antlr4-go/antlr/v4"

// BaseSqlBaseListener is a complete listener for a parse tree produced by SqlBaseParser.
type BaseSqlBaseListener struct{}

var _ SqlBaseListener = &BaseSqlBaseListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseSqlBaseListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseSqlBaseListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseSqlBaseListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseSqlBaseListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterSingleStatement is called when production singleStatement is entered.
func (s *BaseSqlBaseListener) EnterSingleStatement(ctx *SingleStatementContext) {}

// ExitSingleStatement is called when production singleStatement is exited.
func (s *BaseSqlBaseListener) ExitSingleStatement(ctx *SingleStatementContext) {}

// EnterStandaloneExpression is called when production standaloneExpression is entered.
func (s *BaseSqlBaseListener) EnterStandaloneExpression(ctx *StandaloneExpressionContext) {}

// ExitStandaloneExpression is called when production standaloneExpression is exited.
func (s *BaseSqlBaseListener) ExitStandaloneExpression(ctx *StandaloneExpressionContext) {}

// EnterStandalonePathSpecification is called when production standalonePathSpecification is entered.
func (s *BaseSqlBaseListener) EnterStandalonePathSpecification(ctx *StandalonePathSpecificationContext) {
}

// ExitStandalonePathSpecification is called when production standalonePathSpecification is exited.
func (s *BaseSqlBaseListener) ExitStandalonePathSpecification(ctx *StandalonePathSpecificationContext) {
}

// EnterStandaloneRoutineBody is called when production standaloneRoutineBody is entered.
func (s *BaseSqlBaseListener) EnterStandaloneRoutineBody(ctx *StandaloneRoutineBodyContext) {}

// ExitStandaloneRoutineBody is called when production standaloneRoutineBody is exited.
func (s *BaseSqlBaseListener) ExitStandaloneRoutineBody(ctx *StandaloneRoutineBodyContext) {}

// EnterStatementDefault is called when production statementDefault is entered.
func (s *BaseSqlBaseListener) EnterStatementDefault(ctx *StatementDefaultContext) {}

// ExitStatementDefault is called when production statementDefault is exited.
func (s *BaseSqlBaseListener) ExitStatementDefault(ctx *StatementDefaultContext) {}

// EnterUse is called when production use is entered.
func (s *BaseSqlBaseListener) EnterUse(ctx *UseContext) {}

// ExitUse is called when production use is exited.
func (s *BaseSqlBaseListener) ExitUse(ctx *UseContext) {}

// EnterCreateSchema is called when production createSchema is entered.
func (s *BaseSqlBaseListener) EnterCreateSchema(ctx *CreateSchemaContext) {}

// ExitCreateSchema is called when production createSchema is exited.
func (s *BaseSqlBaseListener) ExitCreateSchema(ctx *CreateSchemaContext) {}

// EnterDropSchema is called when production dropSchema is entered.
func (s *BaseSqlBaseListener) EnterDropSchema(ctx *DropSchemaContext) {}

// ExitDropSchema is called when production dropSchema is exited.
func (s *BaseSqlBaseListener) ExitDropSchema(ctx *DropSchemaContext) {}

// EnterRenameSchema is called when production renameSchema is entered.
func (s *BaseSqlBaseListener) EnterRenameSchema(ctx *RenameSchemaContext) {}

// ExitRenameSchema is called when production renameSchema is exited.
func (s *BaseSqlBaseListener) ExitRenameSchema(ctx *RenameSchemaContext) {}

// EnterCreateTableAsSelect is called when production createTableAsSelect is entered.
func (s *BaseSqlBaseListener) EnterCreateTableAsSelect(ctx *CreateTableAsSelectContext) {}

// ExitCreateTableAsSelect is called when production createTableAsSelect is exited.
func (s *BaseSqlBaseListener) ExitCreateTableAsSelect(ctx *CreateTableAsSelectContext) {}

// EnterCreateTable is called when production createTable is entered.
func (s *BaseSqlBaseListener) EnterCreateTable(ctx *CreateTableContext) {}

// ExitCreateTable is called when production createTable is exited.
func (s *BaseSqlBaseListener) ExitCreateTable(ctx *CreateTableContext) {}

// EnterDropTable is called when production dropTable is entered.
func (s *BaseSqlBaseListener) EnterDropTable(ctx *DropTableContext) {}

// ExitDropTable is called when production dropTable is exited.
func (s *BaseSqlBaseListener) ExitDropTable(ctx *DropTableContext) {}

// EnterTruncateTable is called when production truncateTable is entered.
func (s *BaseSqlBaseListener) EnterTruncateTable(ctx *TruncateTableContext) {}

// ExitTruncateTable is called when production truncateTable is exited.
func (s *BaseSqlBaseListener) ExitTruncateTable(ctx *TruncateTableContext) {}

// EnterCacheTable is called when production cacheTable is entered.
func (s *BaseSqlBaseListener) EnterCacheTable(ctx *CacheTableContext) {}

// ExitCacheTable is called when production cacheTable is exited.
func (s *BaseSqlBaseListener) ExitCacheTable(ctx *CacheTableContext) {}

// EnterDropCache is called when production dropCache is entered.
func (s *BaseSqlBaseListener) EnterDropCache(ctx *DropCacheContext) {}

// ExitDropCache is called when production dropCache is exited.
func (s *BaseSqlBaseListener) ExitDropCache(ctx *DropCacheContext) {}

// EnterShowCache is called when production showCache is entered.
func (s *BaseSqlBaseListener) EnterShowCache(ctx *ShowCacheContext) {}

// ExitShowCache is called when production showCache is exited.
func (s *BaseSqlBaseListener) ExitShowCache(ctx *ShowCacheContext) {}

// EnterCreateCube is called when production createCube is entered.
func (s *BaseSqlBaseListener) EnterCreateCube(ctx *CreateCubeContext) {}

// ExitCreateCube is called when production createCube is exited.
func (s *BaseSqlBaseListener) ExitCreateCube(ctx *CreateCubeContext) {}

// EnterInsertCube is called when production insertCube is entered.
func (s *BaseSqlBaseListener) EnterInsertCube(ctx *InsertCubeContext) {}

// ExitInsertCube is called when production insertCube is exited.
func (s *BaseSqlBaseListener) ExitInsertCube(ctx *InsertCubeContext) {}

// EnterInsertOverwriteCube is called when production insertOverwriteCube is entered.
func (s *BaseSqlBaseListener) EnterInsertOverwriteCube(ctx *InsertOverwriteCubeContext) {}

// ExitInsertOverwriteCube is called when production insertOverwriteCube is exited.
func (s *BaseSqlBaseListener) ExitInsertOverwriteCube(ctx *InsertOverwriteCubeContext) {}

// EnterReloadCube is called when production reloadCube is entered.
func (s *BaseSqlBaseListener) EnterReloadCube(ctx *ReloadCubeContext) {}

// ExitReloadCube is called when production reloadCube is exited.
func (s *BaseSqlBaseListener) ExitReloadCube(ctx *ReloadCubeContext) {}

// EnterDropCube is called when production dropCube is entered.
func (s *BaseSqlBaseListener) EnterDropCube(ctx *DropCubeContext) {}

// ExitDropCube is called when production dropCube is exited.
func (s *BaseSqlBaseListener) ExitDropCube(ctx *DropCubeContext) {}

// EnterShowCubes is called when production showCubes is entered.
func (s *BaseSqlBaseListener) EnterShowCubes(ctx *ShowCubesContext) {}

// ExitShowCubes is called when production showCubes is exited.
func (s *BaseSqlBaseListener) ExitShowCubes(ctx *ShowCubesContext) {}

// EnterCreateIndex is called when production createIndex is entered.
func (s *BaseSqlBaseListener) EnterCreateIndex(ctx *CreateIndexContext) {}

// ExitCreateIndex is called when production createIndex is exited.
func (s *BaseSqlBaseListener) ExitCreateIndex(ctx *CreateIndexContext) {}

// EnterDropIndex is called when production dropIndex is entered.
func (s *BaseSqlBaseListener) EnterDropIndex(ctx *DropIndexContext) {}

// ExitDropIndex is called when production dropIndex is exited.
func (s *BaseSqlBaseListener) ExitDropIndex(ctx *DropIndexContext) {}

// EnterRenameIndex is called when production renameIndex is entered.
func (s *BaseSqlBaseListener) EnterRenameIndex(ctx *RenameIndexContext) {}

// ExitRenameIndex is called when production renameIndex is exited.
func (s *BaseSqlBaseListener) ExitRenameIndex(ctx *RenameIndexContext) {}

// EnterUpdateIndex is called when production updateIndex is entered.
func (s *BaseSqlBaseListener) EnterUpdateIndex(ctx *UpdateIndexContext) {}

// ExitUpdateIndex is called when production updateIndex is exited.
func (s *BaseSqlBaseListener) ExitUpdateIndex(ctx *UpdateIndexContext) {}

// EnterShowIndex is called when production showIndex is entered.
func (s *BaseSqlBaseListener) EnterShowIndex(ctx *ShowIndexContext) {}

// ExitShowIndex is called when production showIndex is exited.
func (s *BaseSqlBaseListener) ExitShowIndex(ctx *ShowIndexContext) {}

// EnterInsertInto is called when production insertInto is entered.
func (s *BaseSqlBaseListener) EnterInsertInto(ctx *InsertIntoContext) {}

// ExitInsertInto is called when production insertInto is exited.
func (s *BaseSqlBaseListener) ExitInsertInto(ctx *InsertIntoContext) {}

// EnterInsertOverwrite is called when production insertOverwrite is entered.
func (s *BaseSqlBaseListener) EnterInsertOverwrite(ctx *InsertOverwriteContext) {}

// ExitInsertOverwrite is called when production insertOverwrite is exited.
func (s *BaseSqlBaseListener) ExitInsertOverwrite(ctx *InsertOverwriteContext) {}

// EnterDelete is called when production delete is entered.
func (s *BaseSqlBaseListener) EnterDelete(ctx *DeleteContext) {}

// ExitDelete is called when production delete is exited.
func (s *BaseSqlBaseListener) ExitDelete(ctx *DeleteContext) {}

// EnterUpdateTable is called when production updateTable is entered.
func (s *BaseSqlBaseListener) EnterUpdateTable(ctx *UpdateTableContext) {}

// ExitUpdateTable is called when production updateTable is exited.
func (s *BaseSqlBaseListener) ExitUpdateTable(ctx *UpdateTableContext) {}

// EnterRenameTable is called when production renameTable is entered.
func (s *BaseSqlBaseListener) EnterRenameTable(ctx *RenameTableContext) {}

// ExitRenameTable is called when production renameTable is exited.
func (s *BaseSqlBaseListener) ExitRenameTable(ctx *RenameTableContext) {}

// EnterCommentTable is called when production commentTable is entered.
func (s *BaseSqlBaseListener) EnterCommentTable(ctx *CommentTableContext) {}

// ExitCommentTable is called when production commentTable is exited.
func (s *BaseSqlBaseListener) ExitCommentTable(ctx *CommentTableContext) {}

// EnterRenameColumn is called when production renameColumn is entered.
func (s *BaseSqlBaseListener) EnterRenameColumn(ctx *RenameColumnContext) {}

// ExitRenameColumn is called when production renameColumn is exited.
func (s *BaseSqlBaseListener) ExitRenameColumn(ctx *RenameColumnContext) {}

// EnterDropColumn is called when production dropColumn is entered.
func (s *BaseSqlBaseListener) EnterDropColumn(ctx *DropColumnContext) {}

// ExitDropColumn is called when production dropColumn is exited.
func (s *BaseSqlBaseListener) ExitDropColumn(ctx *DropColumnContext) {}

// EnterAddColumn is called when production addColumn is entered.
func (s *BaseSqlBaseListener) EnterAddColumn(ctx *AddColumnContext) {}

// ExitAddColumn is called when production addColumn is exited.
func (s *BaseSqlBaseListener) ExitAddColumn(ctx *AddColumnContext) {}

// EnterSetTableProperties is called when production setTableProperties is entered.
func (s *BaseSqlBaseListener) EnterSetTableProperties(ctx *SetTablePropertiesContext) {}

// ExitSetTableProperties is called when production setTableProperties is exited.
func (s *BaseSqlBaseListener) ExitSetTableProperties(ctx *SetTablePropertiesContext) {}

// EnterAnalyze is called when production analyze is entered.
func (s *BaseSqlBaseListener) EnterAnalyze(ctx *AnalyzeContext) {}

// ExitAnalyze is called when production analyze is exited.
func (s *BaseSqlBaseListener) ExitAnalyze(ctx *AnalyzeContext) {}

// EnterCreateView is called when production createView is entered.
func (s *BaseSqlBaseListener) EnterCreateView(ctx *CreateViewContext) {}

// ExitCreateView is called when production createView is exited.
func (s *BaseSqlBaseListener) ExitCreateView(ctx *CreateViewContext) {}

// EnterSetMaterializedViewProperties is called when production setMaterializedViewProperties is entered.
func (s *BaseSqlBaseListener) EnterSetMaterializedViewProperties(ctx *SetMaterializedViewPropertiesContext) {
}

// ExitSetMaterializedViewProperties is called when production setMaterializedViewProperties is exited.
func (s *BaseSqlBaseListener) ExitSetMaterializedViewProperties(ctx *SetMaterializedViewPropertiesContext) {
}

// EnterDropView is called when production dropView is entered.
func (s *BaseSqlBaseListener) EnterDropView(ctx *DropViewContext) {}

// ExitDropView is called when production dropView is exited.
func (s *BaseSqlBaseListener) ExitDropView(ctx *DropViewContext) {}

// EnterCall is called when production call is entered.
func (s *BaseSqlBaseListener) EnterCall(ctx *CallContext) {}

// ExitCall is called when production call is exited.
func (s *BaseSqlBaseListener) ExitCall(ctx *CallContext) {}

// EnterCreateRole is called when production createRole is entered.
func (s *BaseSqlBaseListener) EnterCreateRole(ctx *CreateRoleContext) {}

// ExitCreateRole is called when production createRole is exited.
func (s *BaseSqlBaseListener) ExitCreateRole(ctx *CreateRoleContext) {}

// EnterDropRole is called when production dropRole is entered.
func (s *BaseSqlBaseListener) EnterDropRole(ctx *DropRoleContext) {}

// ExitDropRole is called when production dropRole is exited.
func (s *BaseSqlBaseListener) ExitDropRole(ctx *DropRoleContext) {}

// EnterGrantRoles is called when production grantRoles is entered.
func (s *BaseSqlBaseListener) EnterGrantRoles(ctx *GrantRolesContext) {}

// ExitGrantRoles is called when production grantRoles is exited.
func (s *BaseSqlBaseListener) ExitGrantRoles(ctx *GrantRolesContext) {}

// EnterRevokeRoles is called when production revokeRoles is entered.
func (s *BaseSqlBaseListener) EnterRevokeRoles(ctx *RevokeRolesContext) {}

// ExitRevokeRoles is called when production revokeRoles is exited.
func (s *BaseSqlBaseListener) ExitRevokeRoles(ctx *RevokeRolesContext) {}

// EnterSetRole is called when production setRole is entered.
func (s *BaseSqlBaseListener) EnterSetRole(ctx *SetRoleContext) {}

// ExitSetRole is called when production setRole is exited.
func (s *BaseSqlBaseListener) ExitSetRole(ctx *SetRoleContext) {}

// EnterGrant is called when production grant is entered.
func (s *BaseSqlBaseListener) EnterGrant(ctx *GrantContext) {}

// ExitGrant is called when production grant is exited.
func (s *BaseSqlBaseListener) ExitGrant(ctx *GrantContext) {}

// EnterRevoke is called when production revoke is entered.
func (s *BaseSqlBaseListener) EnterRevoke(ctx *RevokeContext) {}

// ExitRevoke is called when production revoke is exited.
func (s *BaseSqlBaseListener) ExitRevoke(ctx *RevokeContext) {}

// EnterShowGrants is called when production showGrants is entered.
func (s *BaseSqlBaseListener) EnterShowGrants(ctx *ShowGrantsContext) {}

// ExitShowGrants is called when production showGrants is exited.
func (s *BaseSqlBaseListener) ExitShowGrants(ctx *ShowGrantsContext) {}

// EnterExplain is called when production explain is entered.
func (s *BaseSqlBaseListener) EnterExplain(ctx *ExplainContext) {}

// ExitExplain is called when production explain is exited.
func (s *BaseSqlBaseListener) ExitExplain(ctx *ExplainContext) {}

// EnterShowExternalFunction is called when production showExternalFunction is entered.
func (s *BaseSqlBaseListener) EnterShowExternalFunction(ctx *ShowExternalFunctionContext) {}

// ExitShowExternalFunction is called when production showExternalFunction is exited.
func (s *BaseSqlBaseListener) ExitShowExternalFunction(ctx *ShowExternalFunctionContext) {}

// EnterShowCreateTable is called when production showCreateTable is entered.
func (s *BaseSqlBaseListener) EnterShowCreateTable(ctx *ShowCreateTableContext) {}

// ExitShowCreateTable is called when production showCreateTable is exited.
func (s *BaseSqlBaseListener) ExitShowCreateTable(ctx *ShowCreateTableContext) {}

// EnterShowCreateView is called when production showCreateView is entered.
func (s *BaseSqlBaseListener) EnterShowCreateView(ctx *ShowCreateViewContext) {}

// ExitShowCreateView is called when production showCreateView is exited.
func (s *BaseSqlBaseListener) ExitShowCreateView(ctx *ShowCreateViewContext) {}

// EnterShowCreateCube is called when production showCreateCube is entered.
func (s *BaseSqlBaseListener) EnterShowCreateCube(ctx *ShowCreateCubeContext) {}

// ExitShowCreateCube is called when production showCreateCube is exited.
func (s *BaseSqlBaseListener) ExitShowCreateCube(ctx *ShowCreateCubeContext) {}

// EnterShowTables is called when production showTables is entered.
func (s *BaseSqlBaseListener) EnterShowTables(ctx *ShowTablesContext) {}

// ExitShowTables is called when production showTables is exited.
func (s *BaseSqlBaseListener) ExitShowTables(ctx *ShowTablesContext) {}

// EnterShowSchemas is called when production showSchemas is entered.
func (s *BaseSqlBaseListener) EnterShowSchemas(ctx *ShowSchemasContext) {}

// ExitShowSchemas is called when production showSchemas is exited.
func (s *BaseSqlBaseListener) ExitShowSchemas(ctx *ShowSchemasContext) {}

// EnterShowCatalogs is called when production showCatalogs is entered.
func (s *BaseSqlBaseListener) EnterShowCatalogs(ctx *ShowCatalogsContext) {}

// ExitShowCatalogs is called when production showCatalogs is exited.
func (s *BaseSqlBaseListener) ExitShowCatalogs(ctx *ShowCatalogsContext) {}

// EnterShowColumns is called when production showColumns is entered.
func (s *BaseSqlBaseListener) EnterShowColumns(ctx *ShowColumnsContext) {}

// ExitShowColumns is called when production showColumns is exited.
func (s *BaseSqlBaseListener) ExitShowColumns(ctx *ShowColumnsContext) {}

// EnterShowStats is called when production showStats is entered.
func (s *BaseSqlBaseListener) EnterShowStats(ctx *ShowStatsContext) {}

// ExitShowStats is called when production showStats is exited.
func (s *BaseSqlBaseListener) ExitShowStats(ctx *ShowStatsContext) {}

// EnterShowStatsForQuery is called when production showStatsForQuery is entered.
func (s *BaseSqlBaseListener) EnterShowStatsForQuery(ctx *ShowStatsForQueryContext) {}

// ExitShowStatsForQuery is called when production showStatsForQuery is exited.
func (s *BaseSqlBaseListener) ExitShowStatsForQuery(ctx *ShowStatsForQueryContext) {}

// EnterShowRoles is called when production showRoles is entered.
func (s *BaseSqlBaseListener) EnterShowRoles(ctx *ShowRolesContext) {}

// ExitShowRoles is called when production showRoles is exited.
func (s *BaseSqlBaseListener) ExitShowRoles(ctx *ShowRolesContext) {}

// EnterShowRoleGrants is called when production showRoleGrants is entered.
func (s *BaseSqlBaseListener) EnterShowRoleGrants(ctx *ShowRoleGrantsContext) {}

// ExitShowRoleGrants is called when production showRoleGrants is exited.
func (s *BaseSqlBaseListener) ExitShowRoleGrants(ctx *ShowRoleGrantsContext) {}

// EnterShowFunctions is called when production showFunctions is entered.
func (s *BaseSqlBaseListener) EnterShowFunctions(ctx *ShowFunctionsContext) {}

// ExitShowFunctions is called when production showFunctions is exited.
func (s *BaseSqlBaseListener) ExitShowFunctions(ctx *ShowFunctionsContext) {}

// EnterShowSession is called when production showSession is entered.
func (s *BaseSqlBaseListener) EnterShowSession(ctx *ShowSessionContext) {}

// ExitShowSession is called when production showSession is exited.
func (s *BaseSqlBaseListener) ExitShowSession(ctx *ShowSessionContext) {}

// EnterSetSession is called when production setSession is entered.
func (s *BaseSqlBaseListener) EnterSetSession(ctx *SetSessionContext) {}

// ExitSetSession is called when production setSession is exited.
func (s *BaseSqlBaseListener) ExitSetSession(ctx *SetSessionContext) {}

// EnterResetSession is called when production resetSession is entered.
func (s *BaseSqlBaseListener) EnterResetSession(ctx *ResetSessionContext) {}

// ExitResetSession is called when production resetSession is exited.
func (s *BaseSqlBaseListener) ExitResetSession(ctx *ResetSessionContext) {}

// EnterStartTransaction is called when production startTransaction is entered.
func (s *BaseSqlBaseListener) EnterStartTransaction(ctx *StartTransactionContext) {}

// ExitStartTransaction is called when production startTransaction is exited.
func (s *BaseSqlBaseListener) ExitStartTransaction(ctx *StartTransactionContext) {}

// EnterCommit is called when production commit is entered.
func (s *BaseSqlBaseListener) EnterCommit(ctx *CommitContext) {}

// ExitCommit is called when production commit is exited.
func (s *BaseSqlBaseListener) ExitCommit(ctx *CommitContext) {}

// EnterRollback is called when production rollback is entered.
func (s *BaseSqlBaseListener) EnterRollback(ctx *RollbackContext) {}

// ExitRollback is called when production rollback is exited.
func (s *BaseSqlBaseListener) ExitRollback(ctx *RollbackContext) {}

// EnterPrepare is called when production prepare is entered.
func (s *BaseSqlBaseListener) EnterPrepare(ctx *PrepareContext) {}

// ExitPrepare is called when production prepare is exited.
func (s *BaseSqlBaseListener) ExitPrepare(ctx *PrepareContext) {}

// EnterDeallocate is called when production deallocate is entered.
func (s *BaseSqlBaseListener) EnterDeallocate(ctx *DeallocateContext) {}

// ExitDeallocate is called when production deallocate is exited.
func (s *BaseSqlBaseListener) ExitDeallocate(ctx *DeallocateContext) {}

// EnterExecute is called when production execute is entered.
func (s *BaseSqlBaseListener) EnterExecute(ctx *ExecuteContext) {}

// ExitExecute is called when production execute is exited.
func (s *BaseSqlBaseListener) ExitExecute(ctx *ExecuteContext) {}

// EnterTableExecute is called when production tableExecute is entered.
func (s *BaseSqlBaseListener) EnterTableExecute(ctx *TableExecuteContext) {}

// ExitTableExecute is called when production tableExecute is exited.
func (s *BaseSqlBaseListener) ExitTableExecute(ctx *TableExecuteContext) {}

// EnterDescribeInput is called when production describeInput is entered.
func (s *BaseSqlBaseListener) EnterDescribeInput(ctx *DescribeInputContext) {}

// ExitDescribeInput is called when production describeInput is exited.
func (s *BaseSqlBaseListener) ExitDescribeInput(ctx *DescribeInputContext) {}

// EnterDescribeOutput is called when production describeOutput is entered.
func (s *BaseSqlBaseListener) EnterDescribeOutput(ctx *DescribeOutputContext) {}

// ExitDescribeOutput is called when production describeOutput is exited.
func (s *BaseSqlBaseListener) ExitDescribeOutput(ctx *DescribeOutputContext) {}

// EnterSetPath is called when production setPath is entered.
func (s *BaseSqlBaseListener) EnterSetPath(ctx *SetPathContext) {}

// ExitSetPath is called when production setPath is exited.
func (s *BaseSqlBaseListener) ExitSetPath(ctx *SetPathContext) {}

// EnterVacuumTable is called when production vacuumTable is entered.
func (s *BaseSqlBaseListener) EnterVacuumTable(ctx *VacuumTableContext) {}

// ExitVacuumTable is called when production vacuumTable is exited.
func (s *BaseSqlBaseListener) ExitVacuumTable(ctx *VacuumTableContext) {}

// EnterRefreshMetadataCache is called when production refreshMetadataCache is entered.
func (s *BaseSqlBaseListener) EnterRefreshMetadataCache(ctx *RefreshMetadataCacheContext) {}

// ExitRefreshMetadataCache is called when production refreshMetadataCache is exited.
func (s *BaseSqlBaseListener) ExitRefreshMetadataCache(ctx *RefreshMetadataCacheContext) {}

// EnterShowViews is called when production showViews is entered.
func (s *BaseSqlBaseListener) EnterShowViews(ctx *ShowViewsContext) {}

// ExitShowViews is called when production showViews is exited.
func (s *BaseSqlBaseListener) ExitShowViews(ctx *ShowViewsContext) {}

// EnterAssignmentList is called when production assignmentList is entered.
func (s *BaseSqlBaseListener) EnterAssignmentList(ctx *AssignmentListContext) {}

// ExitAssignmentList is called when production assignmentList is exited.
func (s *BaseSqlBaseListener) ExitAssignmentList(ctx *AssignmentListContext) {}

// EnterAssignmentItem is called when production assignmentItem is entered.
func (s *BaseSqlBaseListener) EnterAssignmentItem(ctx *AssignmentItemContext) {}

// ExitAssignmentItem is called when production assignmentItem is exited.
func (s *BaseSqlBaseListener) ExitAssignmentItem(ctx *AssignmentItemContext) {}

// EnterQuery is called when production query is entered.
func (s *BaseSqlBaseListener) EnterQuery(ctx *QueryContext) {}

// ExitQuery is called when production query is exited.
func (s *BaseSqlBaseListener) ExitQuery(ctx *QueryContext) {}

// EnterWith is called when production with is entered.
func (s *BaseSqlBaseListener) EnterWith(ctx *WithContext) {}

// ExitWith is called when production with is exited.
func (s *BaseSqlBaseListener) ExitWith(ctx *WithContext) {}

// EnterTableElement is called when production tableElement is entered.
func (s *BaseSqlBaseListener) EnterTableElement(ctx *TableElementContext) {}

// ExitTableElement is called when production tableElement is exited.
func (s *BaseSqlBaseListener) ExitTableElement(ctx *TableElementContext) {}

// EnterColumnDefinition is called when production columnDefinition is entered.
func (s *BaseSqlBaseListener) EnterColumnDefinition(ctx *ColumnDefinitionContext) {}

// ExitColumnDefinition is called when production columnDefinition is exited.
func (s *BaseSqlBaseListener) ExitColumnDefinition(ctx *ColumnDefinitionContext) {}

// EnterLikeClause is called when production likeClause is entered.
func (s *BaseSqlBaseListener) EnterLikeClause(ctx *LikeClauseContext) {}

// ExitLikeClause is called when production likeClause is exited.
func (s *BaseSqlBaseListener) ExitLikeClause(ctx *LikeClauseContext) {}

// EnterCubeProperties is called when production cubeProperties is entered.
func (s *BaseSqlBaseListener) EnterCubeProperties(ctx *CubePropertiesContext) {}

// ExitCubeProperties is called when production cubeProperties is exited.
func (s *BaseSqlBaseListener) ExitCubeProperties(ctx *CubePropertiesContext) {}

// EnterCubeProperty is called when production cubeProperty is entered.
func (s *BaseSqlBaseListener) EnterCubeProperty(ctx *CubePropertyContext) {}

// ExitCubeProperty is called when production cubeProperty is exited.
func (s *BaseSqlBaseListener) ExitCubeProperty(ctx *CubePropertyContext) {}

// EnterProperties is called when production properties is entered.
func (s *BaseSqlBaseListener) EnterProperties(ctx *PropertiesContext) {}

// ExitProperties is called when production properties is exited.
func (s *BaseSqlBaseListener) ExitProperties(ctx *PropertiesContext) {}

// EnterPropertyAssignments is called when production propertyAssignments is entered.
func (s *BaseSqlBaseListener) EnterPropertyAssignments(ctx *PropertyAssignmentsContext) {}

// ExitPropertyAssignments is called when production propertyAssignments is exited.
func (s *BaseSqlBaseListener) ExitPropertyAssignments(ctx *PropertyAssignmentsContext) {}

// EnterProperty is called when production property is entered.
func (s *BaseSqlBaseListener) EnterProperty(ctx *PropertyContext) {}

// ExitProperty is called when production property is exited.
func (s *BaseSqlBaseListener) ExitProperty(ctx *PropertyContext) {}

// EnterFunctionProperties is called when production functionProperties is entered.
func (s *BaseSqlBaseListener) EnterFunctionProperties(ctx *FunctionPropertiesContext) {}

// ExitFunctionProperties is called when production functionProperties is exited.
func (s *BaseSqlBaseListener) ExitFunctionProperties(ctx *FunctionPropertiesContext) {}

// EnterFunctionProperty is called when production functionProperty is entered.
func (s *BaseSqlBaseListener) EnterFunctionProperty(ctx *FunctionPropertyContext) {}

// ExitFunctionProperty is called when production functionProperty is exited.
func (s *BaseSqlBaseListener) ExitFunctionProperty(ctx *FunctionPropertyContext) {}

// EnterSqlParameterDeclaration is called when production sqlParameterDeclaration is entered.
func (s *BaseSqlBaseListener) EnterSqlParameterDeclaration(ctx *SqlParameterDeclarationContext) {}

// ExitSqlParameterDeclaration is called when production sqlParameterDeclaration is exited.
func (s *BaseSqlBaseListener) ExitSqlParameterDeclaration(ctx *SqlParameterDeclarationContext) {}

// EnterRoutineCharacteristics is called when production routineCharacteristics is entered.
func (s *BaseSqlBaseListener) EnterRoutineCharacteristics(ctx *RoutineCharacteristicsContext) {}

// ExitRoutineCharacteristics is called when production routineCharacteristics is exited.
func (s *BaseSqlBaseListener) ExitRoutineCharacteristics(ctx *RoutineCharacteristicsContext) {}

// EnterRoutineCharacteristic is called when production routineCharacteristic is entered.
func (s *BaseSqlBaseListener) EnterRoutineCharacteristic(ctx *RoutineCharacteristicContext) {}

// ExitRoutineCharacteristic is called when production routineCharacteristic is exited.
func (s *BaseSqlBaseListener) ExitRoutineCharacteristic(ctx *RoutineCharacteristicContext) {}

// EnterRoutineBody is called when production routineBody is entered.
func (s *BaseSqlBaseListener) EnterRoutineBody(ctx *RoutineBodyContext) {}

// ExitRoutineBody is called when production routineBody is exited.
func (s *BaseSqlBaseListener) ExitRoutineBody(ctx *RoutineBodyContext) {}

// EnterReturnStatement is called when production returnStatement is entered.
func (s *BaseSqlBaseListener) EnterReturnStatement(ctx *ReturnStatementContext) {}

// ExitReturnStatement is called when production returnStatement is exited.
func (s *BaseSqlBaseListener) ExitReturnStatement(ctx *ReturnStatementContext) {}

// EnterExternalBodyReference is called when production externalBodyReference is entered.
func (s *BaseSqlBaseListener) EnterExternalBodyReference(ctx *ExternalBodyReferenceContext) {}

// ExitExternalBodyReference is called when production externalBodyReference is exited.
func (s *BaseSqlBaseListener) ExitExternalBodyReference(ctx *ExternalBodyReferenceContext) {}

// EnterLanguage is called when production language is entered.
func (s *BaseSqlBaseListener) EnterLanguage(ctx *LanguageContext) {}

// ExitLanguage is called when production language is exited.
func (s *BaseSqlBaseListener) ExitLanguage(ctx *LanguageContext) {}

// EnterDeterminism is called when production determinism is entered.
func (s *BaseSqlBaseListener) EnterDeterminism(ctx *DeterminismContext) {}

// ExitDeterminism is called when production determinism is exited.
func (s *BaseSqlBaseListener) ExitDeterminism(ctx *DeterminismContext) {}

// EnterNullCallClause is called when production nullCallClause is entered.
func (s *BaseSqlBaseListener) EnterNullCallClause(ctx *NullCallClauseContext) {}

// ExitNullCallClause is called when production nullCallClause is exited.
func (s *BaseSqlBaseListener) ExitNullCallClause(ctx *NullCallClauseContext) {}

// EnterExternalRoutineName is called when production externalRoutineName is entered.
func (s *BaseSqlBaseListener) EnterExternalRoutineName(ctx *ExternalRoutineNameContext) {}

// ExitExternalRoutineName is called when production externalRoutineName is exited.
func (s *BaseSqlBaseListener) ExitExternalRoutineName(ctx *ExternalRoutineNameContext) {}

// EnterQueryNoWith is called when production queryNoWith is entered.
func (s *BaseSqlBaseListener) EnterQueryNoWith(ctx *QueryNoWithContext) {}

// ExitQueryNoWith is called when production queryNoWith is exited.
func (s *BaseSqlBaseListener) ExitQueryNoWith(ctx *QueryNoWithContext) {}

// EnterQueryTermDefault is called when production queryTermDefault is entered.
func (s *BaseSqlBaseListener) EnterQueryTermDefault(ctx *QueryTermDefaultContext) {}

// ExitQueryTermDefault is called when production queryTermDefault is exited.
func (s *BaseSqlBaseListener) ExitQueryTermDefault(ctx *QueryTermDefaultContext) {}

// EnterSetOperation is called when production setOperation is entered.
func (s *BaseSqlBaseListener) EnterSetOperation(ctx *SetOperationContext) {}

// ExitSetOperation is called when production setOperation is exited.
func (s *BaseSqlBaseListener) ExitSetOperation(ctx *SetOperationContext) {}

// EnterQueryPrimaryDefault is called when production queryPrimaryDefault is entered.
func (s *BaseSqlBaseListener) EnterQueryPrimaryDefault(ctx *QueryPrimaryDefaultContext) {}

// ExitQueryPrimaryDefault is called when production queryPrimaryDefault is exited.
func (s *BaseSqlBaseListener) ExitQueryPrimaryDefault(ctx *QueryPrimaryDefaultContext) {}

// EnterTable is called when production table is entered.
func (s *BaseSqlBaseListener) EnterTable(ctx *TableContext) {}

// ExitTable is called when production table is exited.
func (s *BaseSqlBaseListener) ExitTable(ctx *TableContext) {}

// EnterInlineTable is called when production inlineTable is entered.
func (s *BaseSqlBaseListener) EnterInlineTable(ctx *InlineTableContext) {}

// ExitInlineTable is called when production inlineTable is exited.
func (s *BaseSqlBaseListener) ExitInlineTable(ctx *InlineTableContext) {}

// EnterSubquery is called when production subquery is entered.
func (s *BaseSqlBaseListener) EnterSubquery(ctx *SubqueryContext) {}

// ExitSubquery is called when production subquery is exited.
func (s *BaseSqlBaseListener) ExitSubquery(ctx *SubqueryContext) {}

// EnterSortItem is called when production sortItem is entered.
func (s *BaseSqlBaseListener) EnterSortItem(ctx *SortItemContext) {}

// ExitSortItem is called when production sortItem is exited.
func (s *BaseSqlBaseListener) ExitSortItem(ctx *SortItemContext) {}

// EnterQuerySpecification is called when production querySpecification is entered.
func (s *BaseSqlBaseListener) EnterQuerySpecification(ctx *QuerySpecificationContext) {}

// ExitQuerySpecification is called when production querySpecification is exited.
func (s *BaseSqlBaseListener) ExitQuerySpecification(ctx *QuerySpecificationContext) {}

// EnterGroupBy is called when production groupBy is entered.
func (s *BaseSqlBaseListener) EnterGroupBy(ctx *GroupByContext) {}

// ExitGroupBy is called when production groupBy is exited.
func (s *BaseSqlBaseListener) ExitGroupBy(ctx *GroupByContext) {}

// EnterSingleGroupingSet is called when production singleGroupingSet is entered.
func (s *BaseSqlBaseListener) EnterSingleGroupingSet(ctx *SingleGroupingSetContext) {}

// ExitSingleGroupingSet is called when production singleGroupingSet is exited.
func (s *BaseSqlBaseListener) ExitSingleGroupingSet(ctx *SingleGroupingSetContext) {}

// EnterRollup is called when production rollup is entered.
func (s *BaseSqlBaseListener) EnterRollup(ctx *RollupContext) {}

// ExitRollup is called when production rollup is exited.
func (s *BaseSqlBaseListener) ExitRollup(ctx *RollupContext) {}

// EnterCube is called when production cube is entered.
func (s *BaseSqlBaseListener) EnterCube(ctx *CubeContext) {}

// ExitCube is called when production cube is exited.
func (s *BaseSqlBaseListener) ExitCube(ctx *CubeContext) {}

// EnterMultipleGroupingSets is called when production multipleGroupingSets is entered.
func (s *BaseSqlBaseListener) EnterMultipleGroupingSets(ctx *MultipleGroupingSetsContext) {}

// ExitMultipleGroupingSets is called when production multipleGroupingSets is exited.
func (s *BaseSqlBaseListener) ExitMultipleGroupingSets(ctx *MultipleGroupingSetsContext) {}

// EnterGroupingSet is called when production groupingSet is entered.
func (s *BaseSqlBaseListener) EnterGroupingSet(ctx *GroupingSetContext) {}

// ExitGroupingSet is called when production groupingSet is exited.
func (s *BaseSqlBaseListener) ExitGroupingSet(ctx *GroupingSetContext) {}

// EnterCubeGroup is called when production cubeGroup is entered.
func (s *BaseSqlBaseListener) EnterCubeGroup(ctx *CubeGroupContext) {}

// ExitCubeGroup is called when production cubeGroup is exited.
func (s *BaseSqlBaseListener) ExitCubeGroup(ctx *CubeGroupContext) {}

// EnterSourceFilter is called when production sourceFilter is entered.
func (s *BaseSqlBaseListener) EnterSourceFilter(ctx *SourceFilterContext) {}

// ExitSourceFilter is called when production sourceFilter is exited.
func (s *BaseSqlBaseListener) ExitSourceFilter(ctx *SourceFilterContext) {}

// EnterNamedQuery is called when production namedQuery is entered.
func (s *BaseSqlBaseListener) EnterNamedQuery(ctx *NamedQueryContext) {}

// ExitNamedQuery is called when production namedQuery is exited.
func (s *BaseSqlBaseListener) ExitNamedQuery(ctx *NamedQueryContext) {}

// EnterSetQuantifier is called when production setQuantifier is entered.
func (s *BaseSqlBaseListener) EnterSetQuantifier(ctx *SetQuantifierContext) {}

// ExitSetQuantifier is called when production setQuantifier is exited.
func (s *BaseSqlBaseListener) ExitSetQuantifier(ctx *SetQuantifierContext) {}

// EnterSelectSingle is called when production selectSingle is entered.
func (s *BaseSqlBaseListener) EnterSelectSingle(ctx *SelectSingleContext) {}

// ExitSelectSingle is called when production selectSingle is exited.
func (s *BaseSqlBaseListener) ExitSelectSingle(ctx *SelectSingleContext) {}

// EnterSelectAll is called when production selectAll is entered.
func (s *BaseSqlBaseListener) EnterSelectAll(ctx *SelectAllContext) {}

// ExitSelectAll is called when production selectAll is exited.
func (s *BaseSqlBaseListener) ExitSelectAll(ctx *SelectAllContext) {}

// EnterRelationDefault is called when production relationDefault is entered.
func (s *BaseSqlBaseListener) EnterRelationDefault(ctx *RelationDefaultContext) {}

// ExitRelationDefault is called when production relationDefault is exited.
func (s *BaseSqlBaseListener) ExitRelationDefault(ctx *RelationDefaultContext) {}

// EnterJoinRelation is called when production joinRelation is entered.
func (s *BaseSqlBaseListener) EnterJoinRelation(ctx *JoinRelationContext) {}

// ExitJoinRelation is called when production joinRelation is exited.
func (s *BaseSqlBaseListener) ExitJoinRelation(ctx *JoinRelationContext) {}

// EnterJoinType is called when production joinType is entered.
func (s *BaseSqlBaseListener) EnterJoinType(ctx *JoinTypeContext) {}

// ExitJoinType is called when production joinType is exited.
func (s *BaseSqlBaseListener) ExitJoinType(ctx *JoinTypeContext) {}

// EnterJoinCriteria is called when production joinCriteria is entered.
func (s *BaseSqlBaseListener) EnterJoinCriteria(ctx *JoinCriteriaContext) {}

// ExitJoinCriteria is called when production joinCriteria is exited.
func (s *BaseSqlBaseListener) ExitJoinCriteria(ctx *JoinCriteriaContext) {}

// EnterSampledRelation is called when production sampledRelation is entered.
func (s *BaseSqlBaseListener) EnterSampledRelation(ctx *SampledRelationContext) {}

// ExitSampledRelation is called when production sampledRelation is exited.
func (s *BaseSqlBaseListener) ExitSampledRelation(ctx *SampledRelationContext) {}

// EnterSampleType is called when production sampleType is entered.
func (s *BaseSqlBaseListener) EnterSampleType(ctx *SampleTypeContext) {}

// ExitSampleType is called when production sampleType is exited.
func (s *BaseSqlBaseListener) ExitSampleType(ctx *SampleTypeContext) {}

// EnterAliasedRelation is called when production aliasedRelation is entered.
func (s *BaseSqlBaseListener) EnterAliasedRelation(ctx *AliasedRelationContext) {}

// ExitAliasedRelation is called when production aliasedRelation is exited.
func (s *BaseSqlBaseListener) ExitAliasedRelation(ctx *AliasedRelationContext) {}

// EnterColumnAliases is called when production columnAliases is entered.
func (s *BaseSqlBaseListener) EnterColumnAliases(ctx *ColumnAliasesContext) {}

// ExitColumnAliases is called when production columnAliases is exited.
func (s *BaseSqlBaseListener) ExitColumnAliases(ctx *ColumnAliasesContext) {}

// EnterTableName is called when production tableName is entered.
func (s *BaseSqlBaseListener) EnterTableName(ctx *TableNameContext) {}

// ExitTableName is called when production tableName is exited.
func (s *BaseSqlBaseListener) ExitTableName(ctx *TableNameContext) {}

// EnterSubqueryRelation is called when production subqueryRelation is entered.
func (s *BaseSqlBaseListener) EnterSubqueryRelation(ctx *SubqueryRelationContext) {}

// ExitSubqueryRelation is called when production subqueryRelation is exited.
func (s *BaseSqlBaseListener) ExitSubqueryRelation(ctx *SubqueryRelationContext) {}

// EnterUnnest is called when production unnest is entered.
func (s *BaseSqlBaseListener) EnterUnnest(ctx *UnnestContext) {}

// ExitUnnest is called when production unnest is exited.
func (s *BaseSqlBaseListener) ExitUnnest(ctx *UnnestContext) {}

// EnterLateral is called when production lateral is entered.
func (s *BaseSqlBaseListener) EnterLateral(ctx *LateralContext) {}

// ExitLateral is called when production lateral is exited.
func (s *BaseSqlBaseListener) ExitLateral(ctx *LateralContext) {}

// EnterParenthesizedRelation is called when production parenthesizedRelation is entered.
func (s *BaseSqlBaseListener) EnterParenthesizedRelation(ctx *ParenthesizedRelationContext) {}

// ExitParenthesizedRelation is called when production parenthesizedRelation is exited.
func (s *BaseSqlBaseListener) ExitParenthesizedRelation(ctx *ParenthesizedRelationContext) {}

// EnterExpression is called when production expression is entered.
func (s *BaseSqlBaseListener) EnterExpression(ctx *ExpressionContext) {}

// ExitExpression is called when production expression is exited.
func (s *BaseSqlBaseListener) ExitExpression(ctx *ExpressionContext) {}

// EnterLogicalNot is called when production logicalNot is entered.
func (s *BaseSqlBaseListener) EnterLogicalNot(ctx *LogicalNotContext) {}

// ExitLogicalNot is called when production logicalNot is exited.
func (s *BaseSqlBaseListener) ExitLogicalNot(ctx *LogicalNotContext) {}

// EnterPredicated is called when production predicated is entered.
func (s *BaseSqlBaseListener) EnterPredicated(ctx *PredicatedContext) {}

// ExitPredicated is called when production predicated is exited.
func (s *BaseSqlBaseListener) ExitPredicated(ctx *PredicatedContext) {}

// EnterLogicalBinary is called when production logicalBinary is entered.
func (s *BaseSqlBaseListener) EnterLogicalBinary(ctx *LogicalBinaryContext) {}

// ExitLogicalBinary is called when production logicalBinary is exited.
func (s *BaseSqlBaseListener) ExitLogicalBinary(ctx *LogicalBinaryContext) {}

// EnterComparison is called when production comparison is entered.
func (s *BaseSqlBaseListener) EnterComparison(ctx *ComparisonContext) {}

// ExitComparison is called when production comparison is exited.
func (s *BaseSqlBaseListener) ExitComparison(ctx *ComparisonContext) {}

// EnterQuantifiedComparison is called when production quantifiedComparison is entered.
func (s *BaseSqlBaseListener) EnterQuantifiedComparison(ctx *QuantifiedComparisonContext) {}

// ExitQuantifiedComparison is called when production quantifiedComparison is exited.
func (s *BaseSqlBaseListener) ExitQuantifiedComparison(ctx *QuantifiedComparisonContext) {}

// EnterBetween is called when production between is entered.
func (s *BaseSqlBaseListener) EnterBetween(ctx *BetweenContext) {}

// ExitBetween is called when production between is exited.
func (s *BaseSqlBaseListener) ExitBetween(ctx *BetweenContext) {}

// EnterInList is called when production inList is entered.
func (s *BaseSqlBaseListener) EnterInList(ctx *InListContext) {}

// ExitInList is called when production inList is exited.
func (s *BaseSqlBaseListener) ExitInList(ctx *InListContext) {}

// EnterInSubquery is called when production inSubquery is entered.
func (s *BaseSqlBaseListener) EnterInSubquery(ctx *InSubqueryContext) {}

// ExitInSubquery is called when production inSubquery is exited.
func (s *BaseSqlBaseListener) ExitInSubquery(ctx *InSubqueryContext) {}

// EnterLike is called when production like is entered.
func (s *BaseSqlBaseListener) EnterLike(ctx *LikeContext) {}

// ExitLike is called when production like is exited.
func (s *BaseSqlBaseListener) ExitLike(ctx *LikeContext) {}

// EnterNullPredicate is called when production nullPredicate is entered.
func (s *BaseSqlBaseListener) EnterNullPredicate(ctx *NullPredicateContext) {}

// ExitNullPredicate is called when production nullPredicate is exited.
func (s *BaseSqlBaseListener) ExitNullPredicate(ctx *NullPredicateContext) {}

// EnterDistinctFrom is called when production distinctFrom is entered.
func (s *BaseSqlBaseListener) EnterDistinctFrom(ctx *DistinctFromContext) {}

// ExitDistinctFrom is called when production distinctFrom is exited.
func (s *BaseSqlBaseListener) ExitDistinctFrom(ctx *DistinctFromContext) {}

// EnterValueExpressionDefault is called when production valueExpressionDefault is entered.
func (s *BaseSqlBaseListener) EnterValueExpressionDefault(ctx *ValueExpressionDefaultContext) {}

// ExitValueExpressionDefault is called when production valueExpressionDefault is exited.
func (s *BaseSqlBaseListener) ExitValueExpressionDefault(ctx *ValueExpressionDefaultContext) {}

// EnterConcatenation is called when production concatenation is entered.
func (s *BaseSqlBaseListener) EnterConcatenation(ctx *ConcatenationContext) {}

// ExitConcatenation is called when production concatenation is exited.
func (s *BaseSqlBaseListener) ExitConcatenation(ctx *ConcatenationContext) {}

// EnterArithmeticBinary is called when production arithmeticBinary is entered.
func (s *BaseSqlBaseListener) EnterArithmeticBinary(ctx *ArithmeticBinaryContext) {}

// ExitArithmeticBinary is called when production arithmeticBinary is exited.
func (s *BaseSqlBaseListener) ExitArithmeticBinary(ctx *ArithmeticBinaryContext) {}

// EnterArithmeticUnary is called when production arithmeticUnary is entered.
func (s *BaseSqlBaseListener) EnterArithmeticUnary(ctx *ArithmeticUnaryContext) {}

// ExitArithmeticUnary is called when production arithmeticUnary is exited.
func (s *BaseSqlBaseListener) ExitArithmeticUnary(ctx *ArithmeticUnaryContext) {}

// EnterAtTimeZone is called when production atTimeZone is entered.
func (s *BaseSqlBaseListener) EnterAtTimeZone(ctx *AtTimeZoneContext) {}

// ExitAtTimeZone is called when production atTimeZone is exited.
func (s *BaseSqlBaseListener) ExitAtTimeZone(ctx *AtTimeZoneContext) {}

// EnterDereference is called when production dereference is entered.
func (s *BaseSqlBaseListener) EnterDereference(ctx *DereferenceContext) {}

// ExitDereference is called when production dereference is exited.
func (s *BaseSqlBaseListener) ExitDereference(ctx *DereferenceContext) {}

// EnterTypeConstructor is called when production typeConstructor is entered.
func (s *BaseSqlBaseListener) EnterTypeConstructor(ctx *TypeConstructorContext) {}

// ExitTypeConstructor is called when production typeConstructor is exited.
func (s *BaseSqlBaseListener) ExitTypeConstructor(ctx *TypeConstructorContext) {}

// EnterSpecialDateTimeFunction is called when production specialDateTimeFunction is entered.
func (s *BaseSqlBaseListener) EnterSpecialDateTimeFunction(ctx *SpecialDateTimeFunctionContext) {}

// ExitSpecialDateTimeFunction is called when production specialDateTimeFunction is exited.
func (s *BaseSqlBaseListener) ExitSpecialDateTimeFunction(ctx *SpecialDateTimeFunctionContext) {}

// EnterSubstring is called when production substring is entered.
func (s *BaseSqlBaseListener) EnterSubstring(ctx *SubstringContext) {}

// ExitSubstring is called when production substring is exited.
func (s *BaseSqlBaseListener) ExitSubstring(ctx *SubstringContext) {}

// EnterCast is called when production cast is entered.
func (s *BaseSqlBaseListener) EnterCast(ctx *CastContext) {}

// ExitCast is called when production cast is exited.
func (s *BaseSqlBaseListener) ExitCast(ctx *CastContext) {}

// EnterLambda is called when production lambda is entered.
func (s *BaseSqlBaseListener) EnterLambda(ctx *LambdaContext) {}

// ExitLambda is called when production lambda is exited.
func (s *BaseSqlBaseListener) ExitLambda(ctx *LambdaContext) {}

// EnterParenthesizedExpression is called when production parenthesizedExpression is entered.
func (s *BaseSqlBaseListener) EnterParenthesizedExpression(ctx *ParenthesizedExpressionContext) {}

// ExitParenthesizedExpression is called when production parenthesizedExpression is exited.
func (s *BaseSqlBaseListener) ExitParenthesizedExpression(ctx *ParenthesizedExpressionContext) {}

// EnterParameter is called when production parameter is entered.
func (s *BaseSqlBaseListener) EnterParameter(ctx *ParameterContext) {}

// ExitParameter is called when production parameter is exited.
func (s *BaseSqlBaseListener) ExitParameter(ctx *ParameterContext) {}

// EnterNormalize is called when production normalize is entered.
func (s *BaseSqlBaseListener) EnterNormalize(ctx *NormalizeContext) {}

// ExitNormalize is called when production normalize is exited.
func (s *BaseSqlBaseListener) ExitNormalize(ctx *NormalizeContext) {}

// EnterIntervalLiteral is called when production intervalLiteral is entered.
func (s *BaseSqlBaseListener) EnterIntervalLiteral(ctx *IntervalLiteralContext) {}

// ExitIntervalLiteral is called when production intervalLiteral is exited.
func (s *BaseSqlBaseListener) ExitIntervalLiteral(ctx *IntervalLiteralContext) {}

// EnterNumericLiteral is called when production numericLiteral is entered.
func (s *BaseSqlBaseListener) EnterNumericLiteral(ctx *NumericLiteralContext) {}

// ExitNumericLiteral is called when production numericLiteral is exited.
func (s *BaseSqlBaseListener) ExitNumericLiteral(ctx *NumericLiteralContext) {}

// EnterBooleanLiteral is called when production booleanLiteral is entered.
func (s *BaseSqlBaseListener) EnterBooleanLiteral(ctx *BooleanLiteralContext) {}

// ExitBooleanLiteral is called when production booleanLiteral is exited.
func (s *BaseSqlBaseListener) ExitBooleanLiteral(ctx *BooleanLiteralContext) {}

// EnterSimpleCase is called when production simpleCase is entered.
func (s *BaseSqlBaseListener) EnterSimpleCase(ctx *SimpleCaseContext) {}

// ExitSimpleCase is called when production simpleCase is exited.
func (s *BaseSqlBaseListener) ExitSimpleCase(ctx *SimpleCaseContext) {}

// EnterColumnReference is called when production columnReference is entered.
func (s *BaseSqlBaseListener) EnterColumnReference(ctx *ColumnReferenceContext) {}

// ExitColumnReference is called when production columnReference is exited.
func (s *BaseSqlBaseListener) ExitColumnReference(ctx *ColumnReferenceContext) {}

// EnterNullLiteral is called when production nullLiteral is entered.
func (s *BaseSqlBaseListener) EnterNullLiteral(ctx *NullLiteralContext) {}

// ExitNullLiteral is called when production nullLiteral is exited.
func (s *BaseSqlBaseListener) ExitNullLiteral(ctx *NullLiteralContext) {}

// EnterRowConstructor is called when production rowConstructor is entered.
func (s *BaseSqlBaseListener) EnterRowConstructor(ctx *RowConstructorContext) {}

// ExitRowConstructor is called when production rowConstructor is exited.
func (s *BaseSqlBaseListener) ExitRowConstructor(ctx *RowConstructorContext) {}

// EnterSubscript is called when production subscript is entered.
func (s *BaseSqlBaseListener) EnterSubscript(ctx *SubscriptContext) {}

// ExitSubscript is called when production subscript is exited.
func (s *BaseSqlBaseListener) ExitSubscript(ctx *SubscriptContext) {}

// EnterCurrentPath is called when production currentPath is entered.
func (s *BaseSqlBaseListener) EnterCurrentPath(ctx *CurrentPathContext) {}

// ExitCurrentPath is called when production currentPath is exited.
func (s *BaseSqlBaseListener) ExitCurrentPath(ctx *CurrentPathContext) {}

// EnterSubqueryExpression is called when production subqueryExpression is entered.
func (s *BaseSqlBaseListener) EnterSubqueryExpression(ctx *SubqueryExpressionContext) {}

// ExitSubqueryExpression is called when production subqueryExpression is exited.
func (s *BaseSqlBaseListener) ExitSubqueryExpression(ctx *SubqueryExpressionContext) {}

// EnterBinaryLiteral is called when production binaryLiteral is entered.
func (s *BaseSqlBaseListener) EnterBinaryLiteral(ctx *BinaryLiteralContext) {}

// ExitBinaryLiteral is called when production binaryLiteral is exited.
func (s *BaseSqlBaseListener) ExitBinaryLiteral(ctx *BinaryLiteralContext) {}

// EnterCurrentUser is called when production currentUser is entered.
func (s *BaseSqlBaseListener) EnterCurrentUser(ctx *CurrentUserContext) {}

// ExitCurrentUser is called when production currentUser is exited.
func (s *BaseSqlBaseListener) ExitCurrentUser(ctx *CurrentUserContext) {}

// EnterExtract is called when production extract is entered.
func (s *BaseSqlBaseListener) EnterExtract(ctx *ExtractContext) {}

// ExitExtract is called when production extract is exited.
func (s *BaseSqlBaseListener) ExitExtract(ctx *ExtractContext) {}

// EnterStringLiteral is called when production stringLiteral is entered.
func (s *BaseSqlBaseListener) EnterStringLiteral(ctx *StringLiteralContext) {}

// ExitStringLiteral is called when production stringLiteral is exited.
func (s *BaseSqlBaseListener) ExitStringLiteral(ctx *StringLiteralContext) {}

// EnterArrayConstructor is called when production arrayConstructor is entered.
func (s *BaseSqlBaseListener) EnterArrayConstructor(ctx *ArrayConstructorContext) {}

// ExitArrayConstructor is called when production arrayConstructor is exited.
func (s *BaseSqlBaseListener) ExitArrayConstructor(ctx *ArrayConstructorContext) {}

// EnterFunctionCall is called when production functionCall is entered.
func (s *BaseSqlBaseListener) EnterFunctionCall(ctx *FunctionCallContext) {}

// ExitFunctionCall is called when production functionCall is exited.
func (s *BaseSqlBaseListener) ExitFunctionCall(ctx *FunctionCallContext) {}

// EnterExists is called when production exists is entered.
func (s *BaseSqlBaseListener) EnterExists(ctx *ExistsContext) {}

// ExitExists is called when production exists is exited.
func (s *BaseSqlBaseListener) ExitExists(ctx *ExistsContext) {}

// EnterPosition is called when production position is entered.
func (s *BaseSqlBaseListener) EnterPosition(ctx *PositionContext) {}

// ExitPosition is called when production position is exited.
func (s *BaseSqlBaseListener) ExitPosition(ctx *PositionContext) {}

// EnterSearchedCase is called when production searchedCase is entered.
func (s *BaseSqlBaseListener) EnterSearchedCase(ctx *SearchedCaseContext) {}

// ExitSearchedCase is called when production searchedCase is exited.
func (s *BaseSqlBaseListener) ExitSearchedCase(ctx *SearchedCaseContext) {}

// EnterGroupingOperation is called when production groupingOperation is entered.
func (s *BaseSqlBaseListener) EnterGroupingOperation(ctx *GroupingOperationContext) {}

// ExitGroupingOperation is called when production groupingOperation is exited.
func (s *BaseSqlBaseListener) ExitGroupingOperation(ctx *GroupingOperationContext) {}

// EnterBasicStringLiteral is called when production basicStringLiteral is entered.
func (s *BaseSqlBaseListener) EnterBasicStringLiteral(ctx *BasicStringLiteralContext) {}

// ExitBasicStringLiteral is called when production basicStringLiteral is exited.
func (s *BaseSqlBaseListener) ExitBasicStringLiteral(ctx *BasicStringLiteralContext) {}

// EnterUnicodeStringLiteral is called when production unicodeStringLiteral is entered.
func (s *BaseSqlBaseListener) EnterUnicodeStringLiteral(ctx *UnicodeStringLiteralContext) {}

// ExitUnicodeStringLiteral is called when production unicodeStringLiteral is exited.
func (s *BaseSqlBaseListener) ExitUnicodeStringLiteral(ctx *UnicodeStringLiteralContext) {}

// EnterNullTreatment is called when production nullTreatment is entered.
func (s *BaseSqlBaseListener) EnterNullTreatment(ctx *NullTreatmentContext) {}

// ExitNullTreatment is called when production nullTreatment is exited.
func (s *BaseSqlBaseListener) ExitNullTreatment(ctx *NullTreatmentContext) {}

// EnterTimeZoneInterval is called when production timeZoneInterval is entered.
func (s *BaseSqlBaseListener) EnterTimeZoneInterval(ctx *TimeZoneIntervalContext) {}

// ExitTimeZoneInterval is called when production timeZoneInterval is exited.
func (s *BaseSqlBaseListener) ExitTimeZoneInterval(ctx *TimeZoneIntervalContext) {}

// EnterTimeZoneString is called when production timeZoneString is entered.
func (s *BaseSqlBaseListener) EnterTimeZoneString(ctx *TimeZoneStringContext) {}

// ExitTimeZoneString is called when production timeZoneString is exited.
func (s *BaseSqlBaseListener) ExitTimeZoneString(ctx *TimeZoneStringContext) {}

// EnterComparisonOperator is called when production comparisonOperator is entered.
func (s *BaseSqlBaseListener) EnterComparisonOperator(ctx *ComparisonOperatorContext) {}

// ExitComparisonOperator is called when production comparisonOperator is exited.
func (s *BaseSqlBaseListener) ExitComparisonOperator(ctx *ComparisonOperatorContext) {}

// EnterComparisonQuantifier is called when production comparisonQuantifier is entered.
func (s *BaseSqlBaseListener) EnterComparisonQuantifier(ctx *ComparisonQuantifierContext) {}

// ExitComparisonQuantifier is called when production comparisonQuantifier is exited.
func (s *BaseSqlBaseListener) ExitComparisonQuantifier(ctx *ComparisonQuantifierContext) {}

// EnterAggregations is called when production aggregations is entered.
func (s *BaseSqlBaseListener) EnterAggregations(ctx *AggregationsContext) {}

// ExitAggregations is called when production aggregations is exited.
func (s *BaseSqlBaseListener) ExitAggregations(ctx *AggregationsContext) {}

// EnterBooleanValue is called when production booleanValue is entered.
func (s *BaseSqlBaseListener) EnterBooleanValue(ctx *BooleanValueContext) {}

// ExitBooleanValue is called when production booleanValue is exited.
func (s *BaseSqlBaseListener) ExitBooleanValue(ctx *BooleanValueContext) {}

// EnterIndexType is called when production indexType is entered.
func (s *BaseSqlBaseListener) EnterIndexType(ctx *IndexTypeContext) {}

// ExitIndexType is called when production indexType is exited.
func (s *BaseSqlBaseListener) ExitIndexType(ctx *IndexTypeContext) {}

// EnterInterval is called when production interval is entered.
func (s *BaseSqlBaseListener) EnterInterval(ctx *IntervalContext) {}

// ExitInterval is called when production interval is exited.
func (s *BaseSqlBaseListener) ExitInterval(ctx *IntervalContext) {}

// EnterIntervalField is called when production intervalField is entered.
func (s *BaseSqlBaseListener) EnterIntervalField(ctx *IntervalFieldContext) {}

// ExitIntervalField is called when production intervalField is exited.
func (s *BaseSqlBaseListener) ExitIntervalField(ctx *IntervalFieldContext) {}

// EnterNormalForm is called when production normalForm is entered.
func (s *BaseSqlBaseListener) EnterNormalForm(ctx *NormalFormContext) {}

// ExitNormalForm is called when production normalForm is exited.
func (s *BaseSqlBaseListener) ExitNormalForm(ctx *NormalFormContext) {}

// EnterTypes is called when production types is entered.
func (s *BaseSqlBaseListener) EnterTypes(ctx *TypesContext) {}

// ExitTypes is called when production types is exited.
func (s *BaseSqlBaseListener) ExitTypes(ctx *TypesContext) {}

// EnterType is called when production type is entered.
func (s *BaseSqlBaseListener) EnterType(ctx *TypeContext) {}

// ExitType is called when production type is exited.
func (s *BaseSqlBaseListener) ExitType(ctx *TypeContext) {}

// EnterTypeParameter is called when production typeParameter is entered.
func (s *BaseSqlBaseListener) EnterTypeParameter(ctx *TypeParameterContext) {}

// ExitTypeParameter is called when production typeParameter is exited.
func (s *BaseSqlBaseListener) ExitTypeParameter(ctx *TypeParameterContext) {}

// EnterBaseType is called when production baseType is entered.
func (s *BaseSqlBaseListener) EnterBaseType(ctx *BaseTypeContext) {}

// ExitBaseType is called when production baseType is exited.
func (s *BaseSqlBaseListener) ExitBaseType(ctx *BaseTypeContext) {}

// EnterWhenClause is called when production whenClause is entered.
func (s *BaseSqlBaseListener) EnterWhenClause(ctx *WhenClauseContext) {}

// ExitWhenClause is called when production whenClause is exited.
func (s *BaseSqlBaseListener) ExitWhenClause(ctx *WhenClauseContext) {}

// EnterFilter is called when production filter is entered.
func (s *BaseSqlBaseListener) EnterFilter(ctx *FilterContext) {}

// ExitFilter is called when production filter is exited.
func (s *BaseSqlBaseListener) ExitFilter(ctx *FilterContext) {}

// EnterOver is called when production over is entered.
func (s *BaseSqlBaseListener) EnterOver(ctx *OverContext) {}

// ExitOver is called when production over is exited.
func (s *BaseSqlBaseListener) ExitOver(ctx *OverContext) {}

// EnterWindowFrame is called when production windowFrame is entered.
func (s *BaseSqlBaseListener) EnterWindowFrame(ctx *WindowFrameContext) {}

// ExitWindowFrame is called when production windowFrame is exited.
func (s *BaseSqlBaseListener) ExitWindowFrame(ctx *WindowFrameContext) {}

// EnterUnboundedFrame is called when production unboundedFrame is entered.
func (s *BaseSqlBaseListener) EnterUnboundedFrame(ctx *UnboundedFrameContext) {}

// ExitUnboundedFrame is called when production unboundedFrame is exited.
func (s *BaseSqlBaseListener) ExitUnboundedFrame(ctx *UnboundedFrameContext) {}

// EnterCurrentRowBound is called when production currentRowBound is entered.
func (s *BaseSqlBaseListener) EnterCurrentRowBound(ctx *CurrentRowBoundContext) {}

// ExitCurrentRowBound is called when production currentRowBound is exited.
func (s *BaseSqlBaseListener) ExitCurrentRowBound(ctx *CurrentRowBoundContext) {}

// EnterBoundedFrame is called when production boundedFrame is entered.
func (s *BaseSqlBaseListener) EnterBoundedFrame(ctx *BoundedFrameContext) {}

// ExitBoundedFrame is called when production boundedFrame is exited.
func (s *BaseSqlBaseListener) ExitBoundedFrame(ctx *BoundedFrameContext) {}

// EnterExplainFormat is called when production explainFormat is entered.
func (s *BaseSqlBaseListener) EnterExplainFormat(ctx *ExplainFormatContext) {}

// ExitExplainFormat is called when production explainFormat is exited.
func (s *BaseSqlBaseListener) ExitExplainFormat(ctx *ExplainFormatContext) {}

// EnterExplainType is called when production explainType is entered.
func (s *BaseSqlBaseListener) EnterExplainType(ctx *ExplainTypeContext) {}

// ExitExplainType is called when production explainType is exited.
func (s *BaseSqlBaseListener) ExitExplainType(ctx *ExplainTypeContext) {}

// EnterIsolationLevel is called when production isolationLevel is entered.
func (s *BaseSqlBaseListener) EnterIsolationLevel(ctx *IsolationLevelContext) {}

// ExitIsolationLevel is called when production isolationLevel is exited.
func (s *BaseSqlBaseListener) ExitIsolationLevel(ctx *IsolationLevelContext) {}

// EnterTransactionAccessMode is called when production transactionAccessMode is entered.
func (s *BaseSqlBaseListener) EnterTransactionAccessMode(ctx *TransactionAccessModeContext) {}

// ExitTransactionAccessMode is called when production transactionAccessMode is exited.
func (s *BaseSqlBaseListener) ExitTransactionAccessMode(ctx *TransactionAccessModeContext) {}

// EnterReadUncommitted is called when production readUncommitted is entered.
func (s *BaseSqlBaseListener) EnterReadUncommitted(ctx *ReadUncommittedContext) {}

// ExitReadUncommitted is called when production readUncommitted is exited.
func (s *BaseSqlBaseListener) ExitReadUncommitted(ctx *ReadUncommittedContext) {}

// EnterReadCommitted is called when production readCommitted is entered.
func (s *BaseSqlBaseListener) EnterReadCommitted(ctx *ReadCommittedContext) {}

// ExitReadCommitted is called when production readCommitted is exited.
func (s *BaseSqlBaseListener) ExitReadCommitted(ctx *ReadCommittedContext) {}

// EnterRepeatableRead is called when production repeatableRead is entered.
func (s *BaseSqlBaseListener) EnterRepeatableRead(ctx *RepeatableReadContext) {}

// ExitRepeatableRead is called when production repeatableRead is exited.
func (s *BaseSqlBaseListener) ExitRepeatableRead(ctx *RepeatableReadContext) {}

// EnterSerializable is called when production serializable is entered.
func (s *BaseSqlBaseListener) EnterSerializable(ctx *SerializableContext) {}

// ExitSerializable is called when production serializable is exited.
func (s *BaseSqlBaseListener) ExitSerializable(ctx *SerializableContext) {}

// EnterPositionalArgument is called when production positionalArgument is entered.
func (s *BaseSqlBaseListener) EnterPositionalArgument(ctx *PositionalArgumentContext) {}

// ExitPositionalArgument is called when production positionalArgument is exited.
func (s *BaseSqlBaseListener) ExitPositionalArgument(ctx *PositionalArgumentContext) {}

// EnterNamedArgument is called when production namedArgument is entered.
func (s *BaseSqlBaseListener) EnterNamedArgument(ctx *NamedArgumentContext) {}

// ExitNamedArgument is called when production namedArgument is exited.
func (s *BaseSqlBaseListener) ExitNamedArgument(ctx *NamedArgumentContext) {}

// EnterQualifiedArgument is called when production qualifiedArgument is entered.
func (s *BaseSqlBaseListener) EnterQualifiedArgument(ctx *QualifiedArgumentContext) {}

// ExitQualifiedArgument is called when production qualifiedArgument is exited.
func (s *BaseSqlBaseListener) ExitQualifiedArgument(ctx *QualifiedArgumentContext) {}

// EnterUnqualifiedArgument is called when production unqualifiedArgument is entered.
func (s *BaseSqlBaseListener) EnterUnqualifiedArgument(ctx *UnqualifiedArgumentContext) {}

// ExitUnqualifiedArgument is called when production unqualifiedArgument is exited.
func (s *BaseSqlBaseListener) ExitUnqualifiedArgument(ctx *UnqualifiedArgumentContext) {}

// EnterPathSpecification is called when production pathSpecification is entered.
func (s *BaseSqlBaseListener) EnterPathSpecification(ctx *PathSpecificationContext) {}

// ExitPathSpecification is called when production pathSpecification is exited.
func (s *BaseSqlBaseListener) ExitPathSpecification(ctx *PathSpecificationContext) {}

// EnterPrivilege is called when production privilege is entered.
func (s *BaseSqlBaseListener) EnterPrivilege(ctx *PrivilegeContext) {}

// ExitPrivilege is called when production privilege is exited.
func (s *BaseSqlBaseListener) ExitPrivilege(ctx *PrivilegeContext) {}

// EnterQualifiedName is called when production qualifiedName is entered.
func (s *BaseSqlBaseListener) EnterQualifiedName(ctx *QualifiedNameContext) {}

// ExitQualifiedName is called when production qualifiedName is exited.
func (s *BaseSqlBaseListener) ExitQualifiedName(ctx *QualifiedNameContext) {}

// EnterSpecifiedPrincipal is called when production specifiedPrincipal is entered.
func (s *BaseSqlBaseListener) EnterSpecifiedPrincipal(ctx *SpecifiedPrincipalContext) {}

// ExitSpecifiedPrincipal is called when production specifiedPrincipal is exited.
func (s *BaseSqlBaseListener) ExitSpecifiedPrincipal(ctx *SpecifiedPrincipalContext) {}

// EnterCurrentUserGrantor is called when production currentUserGrantor is entered.
func (s *BaseSqlBaseListener) EnterCurrentUserGrantor(ctx *CurrentUserGrantorContext) {}

// ExitCurrentUserGrantor is called when production currentUserGrantor is exited.
func (s *BaseSqlBaseListener) ExitCurrentUserGrantor(ctx *CurrentUserGrantorContext) {}

// EnterCurrentRoleGrantor is called when production currentRoleGrantor is entered.
func (s *BaseSqlBaseListener) EnterCurrentRoleGrantor(ctx *CurrentRoleGrantorContext) {}

// ExitCurrentRoleGrantor is called when production currentRoleGrantor is exited.
func (s *BaseSqlBaseListener) ExitCurrentRoleGrantor(ctx *CurrentRoleGrantorContext) {}

// EnterUnspecifiedPrincipal is called when production unspecifiedPrincipal is entered.
func (s *BaseSqlBaseListener) EnterUnspecifiedPrincipal(ctx *UnspecifiedPrincipalContext) {}

// ExitUnspecifiedPrincipal is called when production unspecifiedPrincipal is exited.
func (s *BaseSqlBaseListener) ExitUnspecifiedPrincipal(ctx *UnspecifiedPrincipalContext) {}

// EnterUserPrincipal is called when production userPrincipal is entered.
func (s *BaseSqlBaseListener) EnterUserPrincipal(ctx *UserPrincipalContext) {}

// ExitUserPrincipal is called when production userPrincipal is exited.
func (s *BaseSqlBaseListener) ExitUserPrincipal(ctx *UserPrincipalContext) {}

// EnterRolePrincipal is called when production rolePrincipal is entered.
func (s *BaseSqlBaseListener) EnterRolePrincipal(ctx *RolePrincipalContext) {}

// ExitRolePrincipal is called when production rolePrincipal is exited.
func (s *BaseSqlBaseListener) ExitRolePrincipal(ctx *RolePrincipalContext) {}

// EnterRoles is called when production roles is entered.
func (s *BaseSqlBaseListener) EnterRoles(ctx *RolesContext) {}

// ExitRoles is called when production roles is exited.
func (s *BaseSqlBaseListener) ExitRoles(ctx *RolesContext) {}

// EnterUnquotedIdentifier is called when production unquotedIdentifier is entered.
func (s *BaseSqlBaseListener) EnterUnquotedIdentifier(ctx *UnquotedIdentifierContext) {}

// ExitUnquotedIdentifier is called when production unquotedIdentifier is exited.
func (s *BaseSqlBaseListener) ExitUnquotedIdentifier(ctx *UnquotedIdentifierContext) {}

// EnterQuotedIdentifier is called when production quotedIdentifier is entered.
func (s *BaseSqlBaseListener) EnterQuotedIdentifier(ctx *QuotedIdentifierContext) {}

// ExitQuotedIdentifier is called when production quotedIdentifier is exited.
func (s *BaseSqlBaseListener) ExitQuotedIdentifier(ctx *QuotedIdentifierContext) {}

// EnterBackQuotedIdentifier is called when production backQuotedIdentifier is entered.
func (s *BaseSqlBaseListener) EnterBackQuotedIdentifier(ctx *BackQuotedIdentifierContext) {}

// ExitBackQuotedIdentifier is called when production backQuotedIdentifier is exited.
func (s *BaseSqlBaseListener) ExitBackQuotedIdentifier(ctx *BackQuotedIdentifierContext) {}

// EnterDigitIdentifier is called when production digitIdentifier is entered.
func (s *BaseSqlBaseListener) EnterDigitIdentifier(ctx *DigitIdentifierContext) {}

// ExitDigitIdentifier is called when production digitIdentifier is exited.
func (s *BaseSqlBaseListener) ExitDigitIdentifier(ctx *DigitIdentifierContext) {}

// EnterDecimalLiteral is called when production decimalLiteral is entered.
func (s *BaseSqlBaseListener) EnterDecimalLiteral(ctx *DecimalLiteralContext) {}

// ExitDecimalLiteral is called when production decimalLiteral is exited.
func (s *BaseSqlBaseListener) ExitDecimalLiteral(ctx *DecimalLiteralContext) {}

// EnterDoubleLiteral is called when production doubleLiteral is entered.
func (s *BaseSqlBaseListener) EnterDoubleLiteral(ctx *DoubleLiteralContext) {}

// ExitDoubleLiteral is called when production doubleLiteral is exited.
func (s *BaseSqlBaseListener) ExitDoubleLiteral(ctx *DoubleLiteralContext) {}

// EnterIntegerLiteral is called when production integerLiteral is entered.
func (s *BaseSqlBaseListener) EnterIntegerLiteral(ctx *IntegerLiteralContext) {}

// ExitIntegerLiteral is called when production integerLiteral is exited.
func (s *BaseSqlBaseListener) ExitIntegerLiteral(ctx *IntegerLiteralContext) {}

// EnterNonReserved is called when production nonReserved is entered.
func (s *BaseSqlBaseListener) EnterNonReserved(ctx *NonReservedContext) {}

// ExitNonReserved is called when production nonReserved is exited.
func (s *BaseSqlBaseListener) ExitNonReserved(ctx *NonReservedContext) {}
