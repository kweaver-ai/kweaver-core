// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Code generated from SqlBase.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parsing // SqlBase
import "github.com/antlr4-go/antlr/v4"

// SqlBaseListener is a complete listener for a parse tree produced by SqlBaseParser.
type SqlBaseListener interface {
	antlr.ParseTreeListener

	// EnterSingleStatement is called when entering the singleStatement production.
	EnterSingleStatement(c *SingleStatementContext)

	// EnterStandaloneExpression is called when entering the standaloneExpression production.
	EnterStandaloneExpression(c *StandaloneExpressionContext)

	// EnterStandalonePathSpecification is called when entering the standalonePathSpecification production.
	EnterStandalonePathSpecification(c *StandalonePathSpecificationContext)

	// EnterStandaloneRoutineBody is called when entering the standaloneRoutineBody production.
	EnterStandaloneRoutineBody(c *StandaloneRoutineBodyContext)

	// EnterStatementDefault is called when entering the statementDefault production.
	EnterStatementDefault(c *StatementDefaultContext)

	// EnterUse is called when entering the use production.
	EnterUse(c *UseContext)

	// EnterCreateSchema is called when entering the createSchema production.
	EnterCreateSchema(c *CreateSchemaContext)

	// EnterDropSchema is called when entering the dropSchema production.
	EnterDropSchema(c *DropSchemaContext)

	// EnterRenameSchema is called when entering the renameSchema production.
	EnterRenameSchema(c *RenameSchemaContext)

	// EnterCreateTableAsSelect is called when entering the createTableAsSelect production.
	EnterCreateTableAsSelect(c *CreateTableAsSelectContext)

	// EnterCreateTable is called when entering the createTable production.
	EnterCreateTable(c *CreateTableContext)

	// EnterDropTable is called when entering the dropTable production.
	EnterDropTable(c *DropTableContext)

	// EnterTruncateTable is called when entering the truncateTable production.
	EnterTruncateTable(c *TruncateTableContext)

	// EnterCacheTable is called when entering the cacheTable production.
	EnterCacheTable(c *CacheTableContext)

	// EnterDropCache is called when entering the dropCache production.
	EnterDropCache(c *DropCacheContext)

	// EnterShowCache is called when entering the showCache production.
	EnterShowCache(c *ShowCacheContext)

	// EnterCreateCube is called when entering the createCube production.
	EnterCreateCube(c *CreateCubeContext)

	// EnterInsertCube is called when entering the insertCube production.
	EnterInsertCube(c *InsertCubeContext)

	// EnterInsertOverwriteCube is called when entering the insertOverwriteCube production.
	EnterInsertOverwriteCube(c *InsertOverwriteCubeContext)

	// EnterReloadCube is called when entering the reloadCube production.
	EnterReloadCube(c *ReloadCubeContext)

	// EnterDropCube is called when entering the dropCube production.
	EnterDropCube(c *DropCubeContext)

	// EnterShowCubes is called when entering the showCubes production.
	EnterShowCubes(c *ShowCubesContext)

	// EnterCreateIndex is called when entering the createIndex production.
	EnterCreateIndex(c *CreateIndexContext)

	// EnterDropIndex is called when entering the dropIndex production.
	EnterDropIndex(c *DropIndexContext)

	// EnterRenameIndex is called when entering the renameIndex production.
	EnterRenameIndex(c *RenameIndexContext)

	// EnterUpdateIndex is called when entering the updateIndex production.
	EnterUpdateIndex(c *UpdateIndexContext)

	// EnterShowIndex is called when entering the showIndex production.
	EnterShowIndex(c *ShowIndexContext)

	// EnterInsertInto is called when entering the insertInto production.
	EnterInsertInto(c *InsertIntoContext)

	// EnterInsertOverwrite is called when entering the insertOverwrite production.
	EnterInsertOverwrite(c *InsertOverwriteContext)

	// EnterDelete is called when entering the delete production.
	EnterDelete(c *DeleteContext)

	// EnterUpdateTable is called when entering the updateTable production.
	EnterUpdateTable(c *UpdateTableContext)

	// EnterRenameTable is called when entering the renameTable production.
	EnterRenameTable(c *RenameTableContext)

	// EnterCommentTable is called when entering the commentTable production.
	EnterCommentTable(c *CommentTableContext)

	// EnterRenameColumn is called when entering the renameColumn production.
	EnterRenameColumn(c *RenameColumnContext)

	// EnterDropColumn is called when entering the dropColumn production.
	EnterDropColumn(c *DropColumnContext)

	// EnterAddColumn is called when entering the addColumn production.
	EnterAddColumn(c *AddColumnContext)

	// EnterSetTableProperties is called when entering the setTableProperties production.
	EnterSetTableProperties(c *SetTablePropertiesContext)

	// EnterAnalyze is called when entering the analyze production.
	EnterAnalyze(c *AnalyzeContext)

	// EnterCreateView is called when entering the createView production.
	EnterCreateView(c *CreateViewContext)

	// EnterSetMaterializedViewProperties is called when entering the setMaterializedViewProperties production.
	EnterSetMaterializedViewProperties(c *SetMaterializedViewPropertiesContext)

	// EnterDropView is called when entering the dropView production.
	EnterDropView(c *DropViewContext)

	// EnterCall is called when entering the call production.
	EnterCall(c *CallContext)

	// EnterCreateRole is called when entering the createRole production.
	EnterCreateRole(c *CreateRoleContext)

	// EnterDropRole is called when entering the dropRole production.
	EnterDropRole(c *DropRoleContext)

	// EnterGrantRoles is called when entering the grantRoles production.
	EnterGrantRoles(c *GrantRolesContext)

	// EnterRevokeRoles is called when entering the revokeRoles production.
	EnterRevokeRoles(c *RevokeRolesContext)

	// EnterSetRole is called when entering the setRole production.
	EnterSetRole(c *SetRoleContext)

	// EnterGrant is called when entering the grant production.
	EnterGrant(c *GrantContext)

	// EnterRevoke is called when entering the revoke production.
	EnterRevoke(c *RevokeContext)

	// EnterShowGrants is called when entering the showGrants production.
	EnterShowGrants(c *ShowGrantsContext)

	// EnterExplain is called when entering the explain production.
	EnterExplain(c *ExplainContext)

	// EnterShowExternalFunction is called when entering the showExternalFunction production.
	EnterShowExternalFunction(c *ShowExternalFunctionContext)

	// EnterShowCreateTable is called when entering the showCreateTable production.
	EnterShowCreateTable(c *ShowCreateTableContext)

	// EnterShowCreateView is called when entering the showCreateView production.
	EnterShowCreateView(c *ShowCreateViewContext)

	// EnterShowCreateCube is called when entering the showCreateCube production.
	EnterShowCreateCube(c *ShowCreateCubeContext)

	// EnterShowTables is called when entering the showTables production.
	EnterShowTables(c *ShowTablesContext)

	// EnterShowSchemas is called when entering the showSchemas production.
	EnterShowSchemas(c *ShowSchemasContext)

	// EnterShowCatalogs is called when entering the showCatalogs production.
	EnterShowCatalogs(c *ShowCatalogsContext)

	// EnterShowColumns is called when entering the showColumns production.
	EnterShowColumns(c *ShowColumnsContext)

	// EnterShowStats is called when entering the showStats production.
	EnterShowStats(c *ShowStatsContext)

	// EnterShowStatsForQuery is called when entering the showStatsForQuery production.
	EnterShowStatsForQuery(c *ShowStatsForQueryContext)

	// EnterShowRoles is called when entering the showRoles production.
	EnterShowRoles(c *ShowRolesContext)

	// EnterShowRoleGrants is called when entering the showRoleGrants production.
	EnterShowRoleGrants(c *ShowRoleGrantsContext)

	// EnterShowFunctions is called when entering the showFunctions production.
	EnterShowFunctions(c *ShowFunctionsContext)

	// EnterShowSession is called when entering the showSession production.
	EnterShowSession(c *ShowSessionContext)

	// EnterSetSession is called when entering the setSession production.
	EnterSetSession(c *SetSessionContext)

	// EnterResetSession is called when entering the resetSession production.
	EnterResetSession(c *ResetSessionContext)

	// EnterStartTransaction is called when entering the startTransaction production.
	EnterStartTransaction(c *StartTransactionContext)

	// EnterCommit is called when entering the commit production.
	EnterCommit(c *CommitContext)

	// EnterRollback is called when entering the rollback production.
	EnterRollback(c *RollbackContext)

	// EnterPrepare is called when entering the prepare production.
	EnterPrepare(c *PrepareContext)

	// EnterDeallocate is called when entering the deallocate production.
	EnterDeallocate(c *DeallocateContext)

	// EnterExecute is called when entering the execute production.
	EnterExecute(c *ExecuteContext)

	// EnterTableExecute is called when entering the tableExecute production.
	EnterTableExecute(c *TableExecuteContext)

	// EnterDescribeInput is called when entering the describeInput production.
	EnterDescribeInput(c *DescribeInputContext)

	// EnterDescribeOutput is called when entering the describeOutput production.
	EnterDescribeOutput(c *DescribeOutputContext)

	// EnterSetPath is called when entering the setPath production.
	EnterSetPath(c *SetPathContext)

	// EnterVacuumTable is called when entering the vacuumTable production.
	EnterVacuumTable(c *VacuumTableContext)

	// EnterRefreshMetadataCache is called when entering the refreshMetadataCache production.
	EnterRefreshMetadataCache(c *RefreshMetadataCacheContext)

	// EnterShowViews is called when entering the showViews production.
	EnterShowViews(c *ShowViewsContext)

	// EnterAssignmentList is called when entering the assignmentList production.
	EnterAssignmentList(c *AssignmentListContext)

	// EnterAssignmentItem is called when entering the assignmentItem production.
	EnterAssignmentItem(c *AssignmentItemContext)

	// EnterQuery is called when entering the query production.
	EnterQuery(c *QueryContext)

	// EnterWith is called when entering the with production.
	EnterWith(c *WithContext)

	// EnterTableElement is called when entering the tableElement production.
	EnterTableElement(c *TableElementContext)

	// EnterColumnDefinition is called when entering the columnDefinition production.
	EnterColumnDefinition(c *ColumnDefinitionContext)

	// EnterLikeClause is called when entering the likeClause production.
	EnterLikeClause(c *LikeClauseContext)

	// EnterCubeProperties is called when entering the cubeProperties production.
	EnterCubeProperties(c *CubePropertiesContext)

	// EnterCubeProperty is called when entering the cubeProperty production.
	EnterCubeProperty(c *CubePropertyContext)

	// EnterProperties is called when entering the properties production.
	EnterProperties(c *PropertiesContext)

	// EnterPropertyAssignments is called when entering the propertyAssignments production.
	EnterPropertyAssignments(c *PropertyAssignmentsContext)

	// EnterProperty is called when entering the property production.
	EnterProperty(c *PropertyContext)

	// EnterFunctionProperties is called when entering the functionProperties production.
	EnterFunctionProperties(c *FunctionPropertiesContext)

	// EnterFunctionProperty is called when entering the functionProperty production.
	EnterFunctionProperty(c *FunctionPropertyContext)

	// EnterSqlParameterDeclaration is called when entering the sqlParameterDeclaration production.
	EnterSqlParameterDeclaration(c *SqlParameterDeclarationContext)

	// EnterRoutineCharacteristics is called when entering the routineCharacteristics production.
	EnterRoutineCharacteristics(c *RoutineCharacteristicsContext)

	// EnterRoutineCharacteristic is called when entering the routineCharacteristic production.
	EnterRoutineCharacteristic(c *RoutineCharacteristicContext)

	// EnterRoutineBody is called when entering the routineBody production.
	EnterRoutineBody(c *RoutineBodyContext)

	// EnterReturnStatement is called when entering the returnStatement production.
	EnterReturnStatement(c *ReturnStatementContext)

	// EnterExternalBodyReference is called when entering the externalBodyReference production.
	EnterExternalBodyReference(c *ExternalBodyReferenceContext)

	// EnterLanguage is called when entering the language production.
	EnterLanguage(c *LanguageContext)

	// EnterDeterminism is called when entering the determinism production.
	EnterDeterminism(c *DeterminismContext)

	// EnterNullCallClause is called when entering the nullCallClause production.
	EnterNullCallClause(c *NullCallClauseContext)

	// EnterExternalRoutineName is called when entering the externalRoutineName production.
	EnterExternalRoutineName(c *ExternalRoutineNameContext)

	// EnterQueryNoWith is called when entering the queryNoWith production.
	EnterQueryNoWith(c *QueryNoWithContext)

	// EnterQueryTermDefault is called when entering the queryTermDefault production.
	EnterQueryTermDefault(c *QueryTermDefaultContext)

	// EnterSetOperation is called when entering the setOperation production.
	EnterSetOperation(c *SetOperationContext)

	// EnterQueryPrimaryDefault is called when entering the queryPrimaryDefault production.
	EnterQueryPrimaryDefault(c *QueryPrimaryDefaultContext)

	// EnterTable is called when entering the table production.
	EnterTable(c *TableContext)

	// EnterInlineTable is called when entering the inlineTable production.
	EnterInlineTable(c *InlineTableContext)

	// EnterSubquery is called when entering the subquery production.
	EnterSubquery(c *SubqueryContext)

	// EnterSortItem is called when entering the sortItem production.
	EnterSortItem(c *SortItemContext)

	// EnterQuerySpecification is called when entering the querySpecification production.
	EnterQuerySpecification(c *QuerySpecificationContext)

	// EnterGroupBy is called when entering the groupBy production.
	EnterGroupBy(c *GroupByContext)

	// EnterSingleGroupingSet is called when entering the singleGroupingSet production.
	EnterSingleGroupingSet(c *SingleGroupingSetContext)

	// EnterRollup is called when entering the rollup production.
	EnterRollup(c *RollupContext)

	// EnterCube is called when entering the cube production.
	EnterCube(c *CubeContext)

	// EnterMultipleGroupingSets is called when entering the multipleGroupingSets production.
	EnterMultipleGroupingSets(c *MultipleGroupingSetsContext)

	// EnterGroupingSet is called when entering the groupingSet production.
	EnterGroupingSet(c *GroupingSetContext)

	// EnterCubeGroup is called when entering the cubeGroup production.
	EnterCubeGroup(c *CubeGroupContext)

	// EnterSourceFilter is called when entering the sourceFilter production.
	EnterSourceFilter(c *SourceFilterContext)

	// EnterNamedQuery is called when entering the namedQuery production.
	EnterNamedQuery(c *NamedQueryContext)

	// EnterSetQuantifier is called when entering the setQuantifier production.
	EnterSetQuantifier(c *SetQuantifierContext)

	// EnterSelectSingle is called when entering the selectSingle production.
	EnterSelectSingle(c *SelectSingleContext)

	// EnterSelectAll is called when entering the selectAll production.
	EnterSelectAll(c *SelectAllContext)

	// EnterRelationDefault is called when entering the relationDefault production.
	EnterRelationDefault(c *RelationDefaultContext)

	// EnterJoinRelation is called when entering the joinRelation production.
	EnterJoinRelation(c *JoinRelationContext)

	// EnterJoinType is called when entering the joinType production.
	EnterJoinType(c *JoinTypeContext)

	// EnterJoinCriteria is called when entering the joinCriteria production.
	EnterJoinCriteria(c *JoinCriteriaContext)

	// EnterSampledRelation is called when entering the sampledRelation production.
	EnterSampledRelation(c *SampledRelationContext)

	// EnterSampleType is called when entering the sampleType production.
	EnterSampleType(c *SampleTypeContext)

	// EnterAliasedRelation is called when entering the aliasedRelation production.
	EnterAliasedRelation(c *AliasedRelationContext)

	// EnterColumnAliases is called when entering the columnAliases production.
	EnterColumnAliases(c *ColumnAliasesContext)

	// EnterTableName is called when entering the tableName production.
	EnterTableName(c *TableNameContext)

	// EnterSubqueryRelation is called when entering the subqueryRelation production.
	EnterSubqueryRelation(c *SubqueryRelationContext)

	// EnterUnnest is called when entering the unnest production.
	EnterUnnest(c *UnnestContext)

	// EnterLateral is called when entering the lateral production.
	EnterLateral(c *LateralContext)

	// EnterParenthesizedRelation is called when entering the parenthesizedRelation production.
	EnterParenthesizedRelation(c *ParenthesizedRelationContext)

	// EnterExpression is called when entering the expression production.
	EnterExpression(c *ExpressionContext)

	// EnterLogicalNot is called when entering the logicalNot production.
	EnterLogicalNot(c *LogicalNotContext)

	// EnterPredicated is called when entering the predicated production.
	EnterPredicated(c *PredicatedContext)

	// EnterLogicalBinary is called when entering the logicalBinary production.
	EnterLogicalBinary(c *LogicalBinaryContext)

	// EnterComparison is called when entering the comparison production.
	EnterComparison(c *ComparisonContext)

	// EnterQuantifiedComparison is called when entering the quantifiedComparison production.
	EnterQuantifiedComparison(c *QuantifiedComparisonContext)

	// EnterBetween is called when entering the between production.
	EnterBetween(c *BetweenContext)

	// EnterInList is called when entering the inList production.
	EnterInList(c *InListContext)

	// EnterInSubquery is called when entering the inSubquery production.
	EnterInSubquery(c *InSubqueryContext)

	// EnterLike is called when entering the like production.
	EnterLike(c *LikeContext)

	// EnterNullPredicate is called when entering the nullPredicate production.
	EnterNullPredicate(c *NullPredicateContext)

	// EnterDistinctFrom is called when entering the distinctFrom production.
	EnterDistinctFrom(c *DistinctFromContext)

	// EnterValueExpressionDefault is called when entering the valueExpressionDefault production.
	EnterValueExpressionDefault(c *ValueExpressionDefaultContext)

	// EnterConcatenation is called when entering the concatenation production.
	EnterConcatenation(c *ConcatenationContext)

	// EnterArithmeticBinary is called when entering the arithmeticBinary production.
	EnterArithmeticBinary(c *ArithmeticBinaryContext)

	// EnterArithmeticUnary is called when entering the arithmeticUnary production.
	EnterArithmeticUnary(c *ArithmeticUnaryContext)

	// EnterAtTimeZone is called when entering the atTimeZone production.
	EnterAtTimeZone(c *AtTimeZoneContext)

	// EnterDereference is called when entering the dereference production.
	EnterDereference(c *DereferenceContext)

	// EnterTypeConstructor is called when entering the typeConstructor production.
	EnterTypeConstructor(c *TypeConstructorContext)

	// EnterSpecialDateTimeFunction is called when entering the specialDateTimeFunction production.
	EnterSpecialDateTimeFunction(c *SpecialDateTimeFunctionContext)

	// EnterSubstring is called when entering the substring production.
	EnterSubstring(c *SubstringContext)

	// EnterCast is called when entering the cast production.
	EnterCast(c *CastContext)

	// EnterLambda is called when entering the lambda production.
	EnterLambda(c *LambdaContext)

	// EnterParenthesizedExpression is called when entering the parenthesizedExpression production.
	EnterParenthesizedExpression(c *ParenthesizedExpressionContext)

	// EnterParameter is called when entering the parameter production.
	EnterParameter(c *ParameterContext)

	// EnterNormalize is called when entering the normalize production.
	EnterNormalize(c *NormalizeContext)

	// EnterIntervalLiteral is called when entering the intervalLiteral production.
	EnterIntervalLiteral(c *IntervalLiteralContext)

	// EnterNumericLiteral is called when entering the numericLiteral production.
	EnterNumericLiteral(c *NumericLiteralContext)

	// EnterBooleanLiteral is called when entering the booleanLiteral production.
	EnterBooleanLiteral(c *BooleanLiteralContext)

	// EnterSimpleCase is called when entering the simpleCase production.
	EnterSimpleCase(c *SimpleCaseContext)

	// EnterColumnReference is called when entering the columnReference production.
	EnterColumnReference(c *ColumnReferenceContext)

	// EnterNullLiteral is called when entering the nullLiteral production.
	EnterNullLiteral(c *NullLiteralContext)

	// EnterRowConstructor is called when entering the rowConstructor production.
	EnterRowConstructor(c *RowConstructorContext)

	// EnterSubscript is called when entering the subscript production.
	EnterSubscript(c *SubscriptContext)

	// EnterCurrentPath is called when entering the currentPath production.
	EnterCurrentPath(c *CurrentPathContext)

	// EnterSubqueryExpression is called when entering the subqueryExpression production.
	EnterSubqueryExpression(c *SubqueryExpressionContext)

	// EnterBinaryLiteral is called when entering the binaryLiteral production.
	EnterBinaryLiteral(c *BinaryLiteralContext)

	// EnterCurrentUser is called when entering the currentUser production.
	EnterCurrentUser(c *CurrentUserContext)

	// EnterExtract is called when entering the extract production.
	EnterExtract(c *ExtractContext)

	// EnterStringLiteral is called when entering the stringLiteral production.
	EnterStringLiteral(c *StringLiteralContext)

	// EnterArrayConstructor is called when entering the arrayConstructor production.
	EnterArrayConstructor(c *ArrayConstructorContext)

	// EnterFunctionCall is called when entering the functionCall production.
	EnterFunctionCall(c *FunctionCallContext)

	// EnterExists is called when entering the exists production.
	EnterExists(c *ExistsContext)

	// EnterPosition is called when entering the position production.
	EnterPosition(c *PositionContext)

	// EnterSearchedCase is called when entering the searchedCase production.
	EnterSearchedCase(c *SearchedCaseContext)

	// EnterGroupingOperation is called when entering the groupingOperation production.
	EnterGroupingOperation(c *GroupingOperationContext)

	// EnterBasicStringLiteral is called when entering the basicStringLiteral production.
	EnterBasicStringLiteral(c *BasicStringLiteralContext)

	// EnterUnicodeStringLiteral is called when entering the unicodeStringLiteral production.
	EnterUnicodeStringLiteral(c *UnicodeStringLiteralContext)

	// EnterNullTreatment is called when entering the nullTreatment production.
	EnterNullTreatment(c *NullTreatmentContext)

	// EnterTimeZoneInterval is called when entering the timeZoneInterval production.
	EnterTimeZoneInterval(c *TimeZoneIntervalContext)

	// EnterTimeZoneString is called when entering the timeZoneString production.
	EnterTimeZoneString(c *TimeZoneStringContext)

	// EnterComparisonOperator is called when entering the comparisonOperator production.
	EnterComparisonOperator(c *ComparisonOperatorContext)

	// EnterComparisonQuantifier is called when entering the comparisonQuantifier production.
	EnterComparisonQuantifier(c *ComparisonQuantifierContext)

	// EnterAggregations is called when entering the aggregations production.
	EnterAggregations(c *AggregationsContext)

	// EnterBooleanValue is called when entering the booleanValue production.
	EnterBooleanValue(c *BooleanValueContext)

	// EnterIndexType is called when entering the indexType production.
	EnterIndexType(c *IndexTypeContext)

	// EnterInterval is called when entering the interval production.
	EnterInterval(c *IntervalContext)

	// EnterIntervalField is called when entering the intervalField production.
	EnterIntervalField(c *IntervalFieldContext)

	// EnterNormalForm is called when entering the normalForm production.
	EnterNormalForm(c *NormalFormContext)

	// EnterTypes is called when entering the types production.
	EnterTypes(c *TypesContext)

	// EnterType is called when entering the type production.
	EnterType(c *TypeContext)

	// EnterTypeParameter is called when entering the typeParameter production.
	EnterTypeParameter(c *TypeParameterContext)

	// EnterBaseType is called when entering the baseType production.
	EnterBaseType(c *BaseTypeContext)

	// EnterWhenClause is called when entering the whenClause production.
	EnterWhenClause(c *WhenClauseContext)

	// EnterFilter is called when entering the filter production.
	EnterFilter(c *FilterContext)

	// EnterOver is called when entering the over production.
	EnterOver(c *OverContext)

	// EnterWindowFrame is called when entering the windowFrame production.
	EnterWindowFrame(c *WindowFrameContext)

	// EnterUnboundedFrame is called when entering the unboundedFrame production.
	EnterUnboundedFrame(c *UnboundedFrameContext)

	// EnterCurrentRowBound is called when entering the currentRowBound production.
	EnterCurrentRowBound(c *CurrentRowBoundContext)

	// EnterBoundedFrame is called when entering the boundedFrame production.
	EnterBoundedFrame(c *BoundedFrameContext)

	// EnterExplainFormat is called when entering the explainFormat production.
	EnterExplainFormat(c *ExplainFormatContext)

	// EnterExplainType is called when entering the explainType production.
	EnterExplainType(c *ExplainTypeContext)

	// EnterIsolationLevel is called when entering the isolationLevel production.
	EnterIsolationLevel(c *IsolationLevelContext)

	// EnterTransactionAccessMode is called when entering the transactionAccessMode production.
	EnterTransactionAccessMode(c *TransactionAccessModeContext)

	// EnterReadUncommitted is called when entering the readUncommitted production.
	EnterReadUncommitted(c *ReadUncommittedContext)

	// EnterReadCommitted is called when entering the readCommitted production.
	EnterReadCommitted(c *ReadCommittedContext)

	// EnterRepeatableRead is called when entering the repeatableRead production.
	EnterRepeatableRead(c *RepeatableReadContext)

	// EnterSerializable is called when entering the serializable production.
	EnterSerializable(c *SerializableContext)

	// EnterPositionalArgument is called when entering the positionalArgument production.
	EnterPositionalArgument(c *PositionalArgumentContext)

	// EnterNamedArgument is called when entering the namedArgument production.
	EnterNamedArgument(c *NamedArgumentContext)

	// EnterQualifiedArgument is called when entering the qualifiedArgument production.
	EnterQualifiedArgument(c *QualifiedArgumentContext)

	// EnterUnqualifiedArgument is called when entering the unqualifiedArgument production.
	EnterUnqualifiedArgument(c *UnqualifiedArgumentContext)

	// EnterPathSpecification is called when entering the pathSpecification production.
	EnterPathSpecification(c *PathSpecificationContext)

	// EnterPrivilege is called when entering the privilege production.
	EnterPrivilege(c *PrivilegeContext)

	// EnterQualifiedName is called when entering the qualifiedName production.
	EnterQualifiedName(c *QualifiedNameContext)

	// EnterSpecifiedPrincipal is called when entering the specifiedPrincipal production.
	EnterSpecifiedPrincipal(c *SpecifiedPrincipalContext)

	// EnterCurrentUserGrantor is called when entering the currentUserGrantor production.
	EnterCurrentUserGrantor(c *CurrentUserGrantorContext)

	// EnterCurrentRoleGrantor is called when entering the currentRoleGrantor production.
	EnterCurrentRoleGrantor(c *CurrentRoleGrantorContext)

	// EnterUnspecifiedPrincipal is called when entering the unspecifiedPrincipal production.
	EnterUnspecifiedPrincipal(c *UnspecifiedPrincipalContext)

	// EnterUserPrincipal is called when entering the userPrincipal production.
	EnterUserPrincipal(c *UserPrincipalContext)

	// EnterRolePrincipal is called when entering the rolePrincipal production.
	EnterRolePrincipal(c *RolePrincipalContext)

	// EnterRoles is called when entering the roles production.
	EnterRoles(c *RolesContext)

	// EnterUnquotedIdentifier is called when entering the unquotedIdentifier production.
	EnterUnquotedIdentifier(c *UnquotedIdentifierContext)

	// EnterQuotedIdentifier is called when entering the quotedIdentifier production.
	EnterQuotedIdentifier(c *QuotedIdentifierContext)

	// EnterBackQuotedIdentifier is called when entering the backQuotedIdentifier production.
	EnterBackQuotedIdentifier(c *BackQuotedIdentifierContext)

	// EnterDigitIdentifier is called when entering the digitIdentifier production.
	EnterDigitIdentifier(c *DigitIdentifierContext)

	// EnterDecimalLiteral is called when entering the decimalLiteral production.
	EnterDecimalLiteral(c *DecimalLiteralContext)

	// EnterDoubleLiteral is called when entering the doubleLiteral production.
	EnterDoubleLiteral(c *DoubleLiteralContext)

	// EnterIntegerLiteral is called when entering the integerLiteral production.
	EnterIntegerLiteral(c *IntegerLiteralContext)

	// EnterNonReserved is called when entering the nonReserved production.
	EnterNonReserved(c *NonReservedContext)

	// ExitSingleStatement is called when exiting the singleStatement production.
	ExitSingleStatement(c *SingleStatementContext)

	// ExitStandaloneExpression is called when exiting the standaloneExpression production.
	ExitStandaloneExpression(c *StandaloneExpressionContext)

	// ExitStandalonePathSpecification is called when exiting the standalonePathSpecification production.
	ExitStandalonePathSpecification(c *StandalonePathSpecificationContext)

	// ExitStandaloneRoutineBody is called when exiting the standaloneRoutineBody production.
	ExitStandaloneRoutineBody(c *StandaloneRoutineBodyContext)

	// ExitStatementDefault is called when exiting the statementDefault production.
	ExitStatementDefault(c *StatementDefaultContext)

	// ExitUse is called when exiting the use production.
	ExitUse(c *UseContext)

	// ExitCreateSchema is called when exiting the createSchema production.
	ExitCreateSchema(c *CreateSchemaContext)

	// ExitDropSchema is called when exiting the dropSchema production.
	ExitDropSchema(c *DropSchemaContext)

	// ExitRenameSchema is called when exiting the renameSchema production.
	ExitRenameSchema(c *RenameSchemaContext)

	// ExitCreateTableAsSelect is called when exiting the createTableAsSelect production.
	ExitCreateTableAsSelect(c *CreateTableAsSelectContext)

	// ExitCreateTable is called when exiting the createTable production.
	ExitCreateTable(c *CreateTableContext)

	// ExitDropTable is called when exiting the dropTable production.
	ExitDropTable(c *DropTableContext)

	// ExitTruncateTable is called when exiting the truncateTable production.
	ExitTruncateTable(c *TruncateTableContext)

	// ExitCacheTable is called when exiting the cacheTable production.
	ExitCacheTable(c *CacheTableContext)

	// ExitDropCache is called when exiting the dropCache production.
	ExitDropCache(c *DropCacheContext)

	// ExitShowCache is called when exiting the showCache production.
	ExitShowCache(c *ShowCacheContext)

	// ExitCreateCube is called when exiting the createCube production.
	ExitCreateCube(c *CreateCubeContext)

	// ExitInsertCube is called when exiting the insertCube production.
	ExitInsertCube(c *InsertCubeContext)

	// ExitInsertOverwriteCube is called when exiting the insertOverwriteCube production.
	ExitInsertOverwriteCube(c *InsertOverwriteCubeContext)

	// ExitReloadCube is called when exiting the reloadCube production.
	ExitReloadCube(c *ReloadCubeContext)

	// ExitDropCube is called when exiting the dropCube production.
	ExitDropCube(c *DropCubeContext)

	// ExitShowCubes is called when exiting the showCubes production.
	ExitShowCubes(c *ShowCubesContext)

	// ExitCreateIndex is called when exiting the createIndex production.
	ExitCreateIndex(c *CreateIndexContext)

	// ExitDropIndex is called when exiting the dropIndex production.
	ExitDropIndex(c *DropIndexContext)

	// ExitRenameIndex is called when exiting the renameIndex production.
	ExitRenameIndex(c *RenameIndexContext)

	// ExitUpdateIndex is called when exiting the updateIndex production.
	ExitUpdateIndex(c *UpdateIndexContext)

	// ExitShowIndex is called when exiting the showIndex production.
	ExitShowIndex(c *ShowIndexContext)

	// ExitInsertInto is called when exiting the insertInto production.
	ExitInsertInto(c *InsertIntoContext)

	// ExitInsertOverwrite is called when exiting the insertOverwrite production.
	ExitInsertOverwrite(c *InsertOverwriteContext)

	// ExitDelete is called when exiting the delete production.
	ExitDelete(c *DeleteContext)

	// ExitUpdateTable is called when exiting the updateTable production.
	ExitUpdateTable(c *UpdateTableContext)

	// ExitRenameTable is called when exiting the renameTable production.
	ExitRenameTable(c *RenameTableContext)

	// ExitCommentTable is called when exiting the commentTable production.
	ExitCommentTable(c *CommentTableContext)

	// ExitRenameColumn is called when exiting the renameColumn production.
	ExitRenameColumn(c *RenameColumnContext)

	// ExitDropColumn is called when exiting the dropColumn production.
	ExitDropColumn(c *DropColumnContext)

	// ExitAddColumn is called when exiting the addColumn production.
	ExitAddColumn(c *AddColumnContext)

	// ExitSetTableProperties is called when exiting the setTableProperties production.
	ExitSetTableProperties(c *SetTablePropertiesContext)

	// ExitAnalyze is called when exiting the analyze production.
	ExitAnalyze(c *AnalyzeContext)

	// ExitCreateView is called when exiting the createView production.
	ExitCreateView(c *CreateViewContext)

	// ExitSetMaterializedViewProperties is called when exiting the setMaterializedViewProperties production.
	ExitSetMaterializedViewProperties(c *SetMaterializedViewPropertiesContext)

	// ExitDropView is called when exiting the dropView production.
	ExitDropView(c *DropViewContext)

	// ExitCall is called when exiting the call production.
	ExitCall(c *CallContext)

	// ExitCreateRole is called when exiting the createRole production.
	ExitCreateRole(c *CreateRoleContext)

	// ExitDropRole is called when exiting the dropRole production.
	ExitDropRole(c *DropRoleContext)

	// ExitGrantRoles is called when exiting the grantRoles production.
	ExitGrantRoles(c *GrantRolesContext)

	// ExitRevokeRoles is called when exiting the revokeRoles production.
	ExitRevokeRoles(c *RevokeRolesContext)

	// ExitSetRole is called when exiting the setRole production.
	ExitSetRole(c *SetRoleContext)

	// ExitGrant is called when exiting the grant production.
	ExitGrant(c *GrantContext)

	// ExitRevoke is called when exiting the revoke production.
	ExitRevoke(c *RevokeContext)

	// ExitShowGrants is called when exiting the showGrants production.
	ExitShowGrants(c *ShowGrantsContext)

	// ExitExplain is called when exiting the explain production.
	ExitExplain(c *ExplainContext)

	// ExitShowExternalFunction is called when exiting the showExternalFunction production.
	ExitShowExternalFunction(c *ShowExternalFunctionContext)

	// ExitShowCreateTable is called when exiting the showCreateTable production.
	ExitShowCreateTable(c *ShowCreateTableContext)

	// ExitShowCreateView is called when exiting the showCreateView production.
	ExitShowCreateView(c *ShowCreateViewContext)

	// ExitShowCreateCube is called when exiting the showCreateCube production.
	ExitShowCreateCube(c *ShowCreateCubeContext)

	// ExitShowTables is called when exiting the showTables production.
	ExitShowTables(c *ShowTablesContext)

	// ExitShowSchemas is called when exiting the showSchemas production.
	ExitShowSchemas(c *ShowSchemasContext)

	// ExitShowCatalogs is called when exiting the showCatalogs production.
	ExitShowCatalogs(c *ShowCatalogsContext)

	// ExitShowColumns is called when exiting the showColumns production.
	ExitShowColumns(c *ShowColumnsContext)

	// ExitShowStats is called when exiting the showStats production.
	ExitShowStats(c *ShowStatsContext)

	// ExitShowStatsForQuery is called when exiting the showStatsForQuery production.
	ExitShowStatsForQuery(c *ShowStatsForQueryContext)

	// ExitShowRoles is called when exiting the showRoles production.
	ExitShowRoles(c *ShowRolesContext)

	// ExitShowRoleGrants is called when exiting the showRoleGrants production.
	ExitShowRoleGrants(c *ShowRoleGrantsContext)

	// ExitShowFunctions is called when exiting the showFunctions production.
	ExitShowFunctions(c *ShowFunctionsContext)

	// ExitShowSession is called when exiting the showSession production.
	ExitShowSession(c *ShowSessionContext)

	// ExitSetSession is called when exiting the setSession production.
	ExitSetSession(c *SetSessionContext)

	// ExitResetSession is called when exiting the resetSession production.
	ExitResetSession(c *ResetSessionContext)

	// ExitStartTransaction is called when exiting the startTransaction production.
	ExitStartTransaction(c *StartTransactionContext)

	// ExitCommit is called when exiting the commit production.
	ExitCommit(c *CommitContext)

	// ExitRollback is called when exiting the rollback production.
	ExitRollback(c *RollbackContext)

	// ExitPrepare is called when exiting the prepare production.
	ExitPrepare(c *PrepareContext)

	// ExitDeallocate is called when exiting the deallocate production.
	ExitDeallocate(c *DeallocateContext)

	// ExitExecute is called when exiting the execute production.
	ExitExecute(c *ExecuteContext)

	// ExitTableExecute is called when exiting the tableExecute production.
	ExitTableExecute(c *TableExecuteContext)

	// ExitDescribeInput is called when exiting the describeInput production.
	ExitDescribeInput(c *DescribeInputContext)

	// ExitDescribeOutput is called when exiting the describeOutput production.
	ExitDescribeOutput(c *DescribeOutputContext)

	// ExitSetPath is called when exiting the setPath production.
	ExitSetPath(c *SetPathContext)

	// ExitVacuumTable is called when exiting the vacuumTable production.
	ExitVacuumTable(c *VacuumTableContext)

	// ExitRefreshMetadataCache is called when exiting the refreshMetadataCache production.
	ExitRefreshMetadataCache(c *RefreshMetadataCacheContext)

	// ExitShowViews is called when exiting the showViews production.
	ExitShowViews(c *ShowViewsContext)

	// ExitAssignmentList is called when exiting the assignmentList production.
	ExitAssignmentList(c *AssignmentListContext)

	// ExitAssignmentItem is called when exiting the assignmentItem production.
	ExitAssignmentItem(c *AssignmentItemContext)

	// ExitQuery is called when exiting the query production.
	ExitQuery(c *QueryContext)

	// ExitWith is called when exiting the with production.
	ExitWith(c *WithContext)

	// ExitTableElement is called when exiting the tableElement production.
	ExitTableElement(c *TableElementContext)

	// ExitColumnDefinition is called when exiting the columnDefinition production.
	ExitColumnDefinition(c *ColumnDefinitionContext)

	// ExitLikeClause is called when exiting the likeClause production.
	ExitLikeClause(c *LikeClauseContext)

	// ExitCubeProperties is called when exiting the cubeProperties production.
	ExitCubeProperties(c *CubePropertiesContext)

	// ExitCubeProperty is called when exiting the cubeProperty production.
	ExitCubeProperty(c *CubePropertyContext)

	// ExitProperties is called when exiting the properties production.
	ExitProperties(c *PropertiesContext)

	// ExitPropertyAssignments is called when exiting the propertyAssignments production.
	ExitPropertyAssignments(c *PropertyAssignmentsContext)

	// ExitProperty is called when exiting the property production.
	ExitProperty(c *PropertyContext)

	// ExitFunctionProperties is called when exiting the functionProperties production.
	ExitFunctionProperties(c *FunctionPropertiesContext)

	// ExitFunctionProperty is called when exiting the functionProperty production.
	ExitFunctionProperty(c *FunctionPropertyContext)

	// ExitSqlParameterDeclaration is called when exiting the sqlParameterDeclaration production.
	ExitSqlParameterDeclaration(c *SqlParameterDeclarationContext)

	// ExitRoutineCharacteristics is called when exiting the routineCharacteristics production.
	ExitRoutineCharacteristics(c *RoutineCharacteristicsContext)

	// ExitRoutineCharacteristic is called when exiting the routineCharacteristic production.
	ExitRoutineCharacteristic(c *RoutineCharacteristicContext)

	// ExitRoutineBody is called when exiting the routineBody production.
	ExitRoutineBody(c *RoutineBodyContext)

	// ExitReturnStatement is called when exiting the returnStatement production.
	ExitReturnStatement(c *ReturnStatementContext)

	// ExitExternalBodyReference is called when exiting the externalBodyReference production.
	ExitExternalBodyReference(c *ExternalBodyReferenceContext)

	// ExitLanguage is called when exiting the language production.
	ExitLanguage(c *LanguageContext)

	// ExitDeterminism is called when exiting the determinism production.
	ExitDeterminism(c *DeterminismContext)

	// ExitNullCallClause is called when exiting the nullCallClause production.
	ExitNullCallClause(c *NullCallClauseContext)

	// ExitExternalRoutineName is called when exiting the externalRoutineName production.
	ExitExternalRoutineName(c *ExternalRoutineNameContext)

	// ExitQueryNoWith is called when exiting the queryNoWith production.
	ExitQueryNoWith(c *QueryNoWithContext)

	// ExitQueryTermDefault is called when exiting the queryTermDefault production.
	ExitQueryTermDefault(c *QueryTermDefaultContext)

	// ExitSetOperation is called when exiting the setOperation production.
	ExitSetOperation(c *SetOperationContext)

	// ExitQueryPrimaryDefault is called when exiting the queryPrimaryDefault production.
	ExitQueryPrimaryDefault(c *QueryPrimaryDefaultContext)

	// ExitTable is called when exiting the table production.
	ExitTable(c *TableContext)

	// ExitInlineTable is called when exiting the inlineTable production.
	ExitInlineTable(c *InlineTableContext)

	// ExitSubquery is called when exiting the subquery production.
	ExitSubquery(c *SubqueryContext)

	// ExitSortItem is called when exiting the sortItem production.
	ExitSortItem(c *SortItemContext)

	// ExitQuerySpecification is called when exiting the querySpecification production.
	ExitQuerySpecification(c *QuerySpecificationContext)

	// ExitGroupBy is called when exiting the groupBy production.
	ExitGroupBy(c *GroupByContext)

	// ExitSingleGroupingSet is called when exiting the singleGroupingSet production.
	ExitSingleGroupingSet(c *SingleGroupingSetContext)

	// ExitRollup is called when exiting the rollup production.
	ExitRollup(c *RollupContext)

	// ExitCube is called when exiting the cube production.
	ExitCube(c *CubeContext)

	// ExitMultipleGroupingSets is called when exiting the multipleGroupingSets production.
	ExitMultipleGroupingSets(c *MultipleGroupingSetsContext)

	// ExitGroupingSet is called when exiting the groupingSet production.
	ExitGroupingSet(c *GroupingSetContext)

	// ExitCubeGroup is called when exiting the cubeGroup production.
	ExitCubeGroup(c *CubeGroupContext)

	// ExitSourceFilter is called when exiting the sourceFilter production.
	ExitSourceFilter(c *SourceFilterContext)

	// ExitNamedQuery is called when exiting the namedQuery production.
	ExitNamedQuery(c *NamedQueryContext)

	// ExitSetQuantifier is called when exiting the setQuantifier production.
	ExitSetQuantifier(c *SetQuantifierContext)

	// ExitSelectSingle is called when exiting the selectSingle production.
	ExitSelectSingle(c *SelectSingleContext)

	// ExitSelectAll is called when exiting the selectAll production.
	ExitSelectAll(c *SelectAllContext)

	// ExitRelationDefault is called when exiting the relationDefault production.
	ExitRelationDefault(c *RelationDefaultContext)

	// ExitJoinRelation is called when exiting the joinRelation production.
	ExitJoinRelation(c *JoinRelationContext)

	// ExitJoinType is called when exiting the joinType production.
	ExitJoinType(c *JoinTypeContext)

	// ExitJoinCriteria is called when exiting the joinCriteria production.
	ExitJoinCriteria(c *JoinCriteriaContext)

	// ExitSampledRelation is called when exiting the sampledRelation production.
	ExitSampledRelation(c *SampledRelationContext)

	// ExitSampleType is called when exiting the sampleType production.
	ExitSampleType(c *SampleTypeContext)

	// ExitAliasedRelation is called when exiting the aliasedRelation production.
	ExitAliasedRelation(c *AliasedRelationContext)

	// ExitColumnAliases is called when exiting the columnAliases production.
	ExitColumnAliases(c *ColumnAliasesContext)

	// ExitTableName is called when exiting the tableName production.
	ExitTableName(c *TableNameContext)

	// ExitSubqueryRelation is called when exiting the subqueryRelation production.
	ExitSubqueryRelation(c *SubqueryRelationContext)

	// ExitUnnest is called when exiting the unnest production.
	ExitUnnest(c *UnnestContext)

	// ExitLateral is called when exiting the lateral production.
	ExitLateral(c *LateralContext)

	// ExitParenthesizedRelation is called when exiting the parenthesizedRelation production.
	ExitParenthesizedRelation(c *ParenthesizedRelationContext)

	// ExitExpression is called when exiting the expression production.
	ExitExpression(c *ExpressionContext)

	// ExitLogicalNot is called when exiting the logicalNot production.
	ExitLogicalNot(c *LogicalNotContext)

	// ExitPredicated is called when exiting the predicated production.
	ExitPredicated(c *PredicatedContext)

	// ExitLogicalBinary is called when exiting the logicalBinary production.
	ExitLogicalBinary(c *LogicalBinaryContext)

	// ExitComparison is called when exiting the comparison production.
	ExitComparison(c *ComparisonContext)

	// ExitQuantifiedComparison is called when exiting the quantifiedComparison production.
	ExitQuantifiedComparison(c *QuantifiedComparisonContext)

	// ExitBetween is called when exiting the between production.
	ExitBetween(c *BetweenContext)

	// ExitInList is called when exiting the inList production.
	ExitInList(c *InListContext)

	// ExitInSubquery is called when exiting the inSubquery production.
	ExitInSubquery(c *InSubqueryContext)

	// ExitLike is called when exiting the like production.
	ExitLike(c *LikeContext)

	// ExitNullPredicate is called when exiting the nullPredicate production.
	ExitNullPredicate(c *NullPredicateContext)

	// ExitDistinctFrom is called when exiting the distinctFrom production.
	ExitDistinctFrom(c *DistinctFromContext)

	// ExitValueExpressionDefault is called when exiting the valueExpressionDefault production.
	ExitValueExpressionDefault(c *ValueExpressionDefaultContext)

	// ExitConcatenation is called when exiting the concatenation production.
	ExitConcatenation(c *ConcatenationContext)

	// ExitArithmeticBinary is called when exiting the arithmeticBinary production.
	ExitArithmeticBinary(c *ArithmeticBinaryContext)

	// ExitArithmeticUnary is called when exiting the arithmeticUnary production.
	ExitArithmeticUnary(c *ArithmeticUnaryContext)

	// ExitAtTimeZone is called when exiting the atTimeZone production.
	ExitAtTimeZone(c *AtTimeZoneContext)

	// ExitDereference is called when exiting the dereference production.
	ExitDereference(c *DereferenceContext)

	// ExitTypeConstructor is called when exiting the typeConstructor production.
	ExitTypeConstructor(c *TypeConstructorContext)

	// ExitSpecialDateTimeFunction is called when exiting the specialDateTimeFunction production.
	ExitSpecialDateTimeFunction(c *SpecialDateTimeFunctionContext)

	// ExitSubstring is called when exiting the substring production.
	ExitSubstring(c *SubstringContext)

	// ExitCast is called when exiting the cast production.
	ExitCast(c *CastContext)

	// ExitLambda is called when exiting the lambda production.
	ExitLambda(c *LambdaContext)

	// ExitParenthesizedExpression is called when exiting the parenthesizedExpression production.
	ExitParenthesizedExpression(c *ParenthesizedExpressionContext)

	// ExitParameter is called when exiting the parameter production.
	ExitParameter(c *ParameterContext)

	// ExitNormalize is called when exiting the normalize production.
	ExitNormalize(c *NormalizeContext)

	// ExitIntervalLiteral is called when exiting the intervalLiteral production.
	ExitIntervalLiteral(c *IntervalLiteralContext)

	// ExitNumericLiteral is called when exiting the numericLiteral production.
	ExitNumericLiteral(c *NumericLiteralContext)

	// ExitBooleanLiteral is called when exiting the booleanLiteral production.
	ExitBooleanLiteral(c *BooleanLiteralContext)

	// ExitSimpleCase is called when exiting the simpleCase production.
	ExitSimpleCase(c *SimpleCaseContext)

	// ExitColumnReference is called when exiting the columnReference production.
	ExitColumnReference(c *ColumnReferenceContext)

	// ExitNullLiteral is called when exiting the nullLiteral production.
	ExitNullLiteral(c *NullLiteralContext)

	// ExitRowConstructor is called when exiting the rowConstructor production.
	ExitRowConstructor(c *RowConstructorContext)

	// ExitSubscript is called when exiting the subscript production.
	ExitSubscript(c *SubscriptContext)

	// ExitCurrentPath is called when exiting the currentPath production.
	ExitCurrentPath(c *CurrentPathContext)

	// ExitSubqueryExpression is called when exiting the subqueryExpression production.
	ExitSubqueryExpression(c *SubqueryExpressionContext)

	// ExitBinaryLiteral is called when exiting the binaryLiteral production.
	ExitBinaryLiteral(c *BinaryLiteralContext)

	// ExitCurrentUser is called when exiting the currentUser production.
	ExitCurrentUser(c *CurrentUserContext)

	// ExitExtract is called when exiting the extract production.
	ExitExtract(c *ExtractContext)

	// ExitStringLiteral is called when exiting the stringLiteral production.
	ExitStringLiteral(c *StringLiteralContext)

	// ExitArrayConstructor is called when exiting the arrayConstructor production.
	ExitArrayConstructor(c *ArrayConstructorContext)

	// ExitFunctionCall is called when exiting the functionCall production.
	ExitFunctionCall(c *FunctionCallContext)

	// ExitExists is called when exiting the exists production.
	ExitExists(c *ExistsContext)

	// ExitPosition is called when exiting the position production.
	ExitPosition(c *PositionContext)

	// ExitSearchedCase is called when exiting the searchedCase production.
	ExitSearchedCase(c *SearchedCaseContext)

	// ExitGroupingOperation is called when exiting the groupingOperation production.
	ExitGroupingOperation(c *GroupingOperationContext)

	// ExitBasicStringLiteral is called when exiting the basicStringLiteral production.
	ExitBasicStringLiteral(c *BasicStringLiteralContext)

	// ExitUnicodeStringLiteral is called when exiting the unicodeStringLiteral production.
	ExitUnicodeStringLiteral(c *UnicodeStringLiteralContext)

	// ExitNullTreatment is called when exiting the nullTreatment production.
	ExitNullTreatment(c *NullTreatmentContext)

	// ExitTimeZoneInterval is called when exiting the timeZoneInterval production.
	ExitTimeZoneInterval(c *TimeZoneIntervalContext)

	// ExitTimeZoneString is called when exiting the timeZoneString production.
	ExitTimeZoneString(c *TimeZoneStringContext)

	// ExitComparisonOperator is called when exiting the comparisonOperator production.
	ExitComparisonOperator(c *ComparisonOperatorContext)

	// ExitComparisonQuantifier is called when exiting the comparisonQuantifier production.
	ExitComparisonQuantifier(c *ComparisonQuantifierContext)

	// ExitAggregations is called when exiting the aggregations production.
	ExitAggregations(c *AggregationsContext)

	// ExitBooleanValue is called when exiting the booleanValue production.
	ExitBooleanValue(c *BooleanValueContext)

	// ExitIndexType is called when exiting the indexType production.
	ExitIndexType(c *IndexTypeContext)

	// ExitInterval is called when exiting the interval production.
	ExitInterval(c *IntervalContext)

	// ExitIntervalField is called when exiting the intervalField production.
	ExitIntervalField(c *IntervalFieldContext)

	// ExitNormalForm is called when exiting the normalForm production.
	ExitNormalForm(c *NormalFormContext)

	// ExitTypes is called when exiting the types production.
	ExitTypes(c *TypesContext)

	// ExitType is called when exiting the type production.
	ExitType(c *TypeContext)

	// ExitTypeParameter is called when exiting the typeParameter production.
	ExitTypeParameter(c *TypeParameterContext)

	// ExitBaseType is called when exiting the baseType production.
	ExitBaseType(c *BaseTypeContext)

	// ExitWhenClause is called when exiting the whenClause production.
	ExitWhenClause(c *WhenClauseContext)

	// ExitFilter is called when exiting the filter production.
	ExitFilter(c *FilterContext)

	// ExitOver is called when exiting the over production.
	ExitOver(c *OverContext)

	// ExitWindowFrame is called when exiting the windowFrame production.
	ExitWindowFrame(c *WindowFrameContext)

	// ExitUnboundedFrame is called when exiting the unboundedFrame production.
	ExitUnboundedFrame(c *UnboundedFrameContext)

	// ExitCurrentRowBound is called when exiting the currentRowBound production.
	ExitCurrentRowBound(c *CurrentRowBoundContext)

	// ExitBoundedFrame is called when exiting the boundedFrame production.
	ExitBoundedFrame(c *BoundedFrameContext)

	// ExitExplainFormat is called when exiting the explainFormat production.
	ExitExplainFormat(c *ExplainFormatContext)

	// ExitExplainType is called when exiting the explainType production.
	ExitExplainType(c *ExplainTypeContext)

	// ExitIsolationLevel is called when exiting the isolationLevel production.
	ExitIsolationLevel(c *IsolationLevelContext)

	// ExitTransactionAccessMode is called when exiting the transactionAccessMode production.
	ExitTransactionAccessMode(c *TransactionAccessModeContext)

	// ExitReadUncommitted is called when exiting the readUncommitted production.
	ExitReadUncommitted(c *ReadUncommittedContext)

	// ExitReadCommitted is called when exiting the readCommitted production.
	ExitReadCommitted(c *ReadCommittedContext)

	// ExitRepeatableRead is called when exiting the repeatableRead production.
	ExitRepeatableRead(c *RepeatableReadContext)

	// ExitSerializable is called when exiting the serializable production.
	ExitSerializable(c *SerializableContext)

	// ExitPositionalArgument is called when exiting the positionalArgument production.
	ExitPositionalArgument(c *PositionalArgumentContext)

	// ExitNamedArgument is called when exiting the namedArgument production.
	ExitNamedArgument(c *NamedArgumentContext)

	// ExitQualifiedArgument is called when exiting the qualifiedArgument production.
	ExitQualifiedArgument(c *QualifiedArgumentContext)

	// ExitUnqualifiedArgument is called when exiting the unqualifiedArgument production.
	ExitUnqualifiedArgument(c *UnqualifiedArgumentContext)

	// ExitPathSpecification is called when exiting the pathSpecification production.
	ExitPathSpecification(c *PathSpecificationContext)

	// ExitPrivilege is called when exiting the privilege production.
	ExitPrivilege(c *PrivilegeContext)

	// ExitQualifiedName is called when exiting the qualifiedName production.
	ExitQualifiedName(c *QualifiedNameContext)

	// ExitSpecifiedPrincipal is called when exiting the specifiedPrincipal production.
	ExitSpecifiedPrincipal(c *SpecifiedPrincipalContext)

	// ExitCurrentUserGrantor is called when exiting the currentUserGrantor production.
	ExitCurrentUserGrantor(c *CurrentUserGrantorContext)

	// ExitCurrentRoleGrantor is called when exiting the currentRoleGrantor production.
	ExitCurrentRoleGrantor(c *CurrentRoleGrantorContext)

	// ExitUnspecifiedPrincipal is called when exiting the unspecifiedPrincipal production.
	ExitUnspecifiedPrincipal(c *UnspecifiedPrincipalContext)

	// ExitUserPrincipal is called when exiting the userPrincipal production.
	ExitUserPrincipal(c *UserPrincipalContext)

	// ExitRolePrincipal is called when exiting the rolePrincipal production.
	ExitRolePrincipal(c *RolePrincipalContext)

	// ExitRoles is called when exiting the roles production.
	ExitRoles(c *RolesContext)

	// ExitUnquotedIdentifier is called when exiting the unquotedIdentifier production.
	ExitUnquotedIdentifier(c *UnquotedIdentifierContext)

	// ExitQuotedIdentifier is called when exiting the quotedIdentifier production.
	ExitQuotedIdentifier(c *QuotedIdentifierContext)

	// ExitBackQuotedIdentifier is called when exiting the backQuotedIdentifier production.
	ExitBackQuotedIdentifier(c *BackQuotedIdentifierContext)

	// ExitDigitIdentifier is called when exiting the digitIdentifier production.
	ExitDigitIdentifier(c *DigitIdentifierContext)

	// ExitDecimalLiteral is called when exiting the decimalLiteral production.
	ExitDecimalLiteral(c *DecimalLiteralContext)

	// ExitDoubleLiteral is called when exiting the doubleLiteral production.
	ExitDoubleLiteral(c *DoubleLiteralContext)

	// ExitIntegerLiteral is called when exiting the integerLiteral production.
	ExitIntegerLiteral(c *IntegerLiteralContext)

	// ExitNonReserved is called when exiting the nonReserved production.
	ExitNonReserved(c *NonReservedContext)
}
