/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

grammar SqlBase;

tokens {
    DELIMITER
}

singleStatement
    : statement EOF
    ;

standaloneExpression
    : expression EOF
    ;

standalonePathSpecification
    : pathSpecification EOF
    ;

standaloneRoutineBody
    : routineBody EOF
    ;

statement
    : query                                                            #statementDefault
    | USE schema=identifier                                            #use
    | USE catalog=identifier '.' schema=identifier                     #use
    | USE cluster=identifier '.' catalog=identifier '.' schema=identifier                     #use
    | CREATE (SCHEMA | DATABASE) (IF NOT EXISTS)? qualifiedName
        (WITH properties)?                                             #createSchema
    | DROP (SCHEMA | DATABASE) (IF EXISTS)? qualifiedName (CASCADE | RESTRICT)?     #dropSchema
    | ALTER (SCHEMA | DATABASE) qualifiedName RENAME TO identifier                  #renameSchema
    | CREATE TABLE (IF NOT EXISTS)? qualifiedName columnAliases?
        (COMMENT string)?
        (WITH properties)? AS (query | '('query')')
        (WITH (NO)? DATA)?                                             #createTableAsSelect
    | CREATE TABLE (IF NOT EXISTS)? qualifiedName
        '(' tableElement (',' tableElement)* ')'
         (COMMENT string)?
         (WITH properties)?                                            #createTable
    | DROP TABLE (IF EXISTS)? qualifiedName                            #dropTable
    | TRUNCATE TABLE qualifiedName                                     #truncateTable
    | CACHE TABLE tableName=qualifiedName
        (WHERE booleanExpression)?
        (WITH properties)?                #cacheTable
    | DROP CACHE (IF EXISTS)? qualifiedName
        (WHERE booleanExpression)?                                     #dropCache
    | SHOW CACHE qualifiedName?                            #showCache
    | CREATE CUBE (IF NOT EXISTS)? cubeName=qualifiedName
        ON tableName=qualifiedName
        WITH cubeProperties
        (WHERE expression)? #createCube
    | INSERT INTO CUBE cubeName=qualifiedName (WHERE expression)?     #insertCube
    | INSERT OVERWRITE CUBE cubeName=qualifiedName (WHERE expression)?     #insertOverwriteCube
    | RELOAD CUBE (IF EXISTS)? cubeName=qualifiedName                    #reloadCube
    | DROP CUBE (IF EXISTS)? cubeName=qualifiedName                    #dropCube
    | SHOW CUBES (FOR tableName=qualifiedName)?                                 #showCubes
    | CREATE INDEX (IF NOT EXISTS)? indexName=qualifiedName
        USING indexType
        ON tableName=qualifiedName columnAliases
        (WITH properties)?
        (WHERE expression)?                                                     #createIndex
    | DROP INDEX (IF EXISTS)? indexName=qualifiedName
        (WHERE expression)?                                                     #dropIndex
    | ALTER INDEX (IF EXISTS)? from=qualifiedName RENAME TO to=qualifiedName    #renameIndex
    | UPDATE INDEX (IF EXISTS)? qualifiedName                                   #updateIndex
    | SHOW INDEX (IF EXISTS)? qualifiedName?                                    #showIndex
    | INSERT INTO qualifiedName columnAliases? query                   #insertInto
    | INSERT OVERWRITE (TABLE)? qualifiedName columnAliases? query                   #insertOverwrite
    | DELETE FROM qualifiedName (WHERE booleanExpression)?             #delete
    | UPDATE qualifiedName SET assignmentList (WHERE booleanExpression)?             #updateTable
    | ALTER TABLE from=qualifiedName RENAME TO to=qualifiedName        #renameTable
    | COMMENT ON TABLE qualifiedName IS (string | NULL)                #commentTable
    | ALTER TABLE tableName=qualifiedName
        RENAME COLUMN from=identifier TO to=identifier                 #renameColumn
    | ALTER TABLE tableName=qualifiedName
        DROP COLUMN column=qualifiedName                               #dropColumn
    | ALTER TABLE tableName=qualifiedName
        ADD COLUMN column=columnDefinition                             #addColumn
    | ALTER TABLE tableName=qualifiedName
        SET PROPERTIES propertyAssignments                             #setTableProperties
    | ANALYZE qualifiedName (WITH properties)?                         #analyze
    | CREATE (OR REPLACE)? VIEW qualifiedName
        (SECURITY (DEFINER | INVOKER))? AS query                       #createView
    | ALTER MATERIALIZED VIEW qualifiedName
        SET PROPERTIES propertyAssignments                             #setMaterializedViewProperties
    | DROP VIEW (IF EXISTS)? qualifiedName                             #dropView
    | CALL qualifiedName '(' (callArgument (',' callArgument)*)? ')'   #call
    | CREATE ROLE name=identifier
        (WITH ADMIN grantor)?                                          #createRole
    | DROP ROLE name=identifier                                        #dropRole
    | GRANT
        roles
        TO principal (',' principal)*
        (WITH ADMIN OPTION)?
        (GRANTED BY grantor)?                                          #grantRoles
    | REVOKE
        (ADMIN OPTION FOR)?
        roles
        FROM principal (',' principal)*
        (GRANTED BY grantor)?                                          #revokeRoles
    | SET ROLE (ALL | NONE | role=identifier)                          #setRole
    | GRANT
        (privilege (',' privilege)* | ALL PRIVILEGES)
        ON TABLE? qualifiedName TO grantee=principal
        (WITH GRANT OPTION)?                                           #grant
    | REVOKE
        (GRANT OPTION FOR)?
        (privilege (',' privilege)* | ALL PRIVILEGES)
        ON TABLE? qualifiedName FROM grantee=principal                #revoke
    | SHOW GRANTS
        (ON TABLE? qualifiedName)?                                     #showGrants
    | EXPLAIN ANALYZE? VERBOSE?
        ('(' explainOption (',' explainOption)* ')')? statement        #explain
    | SHOW EXTERNAL FUNCTION qualifiedName types?                        #showExternalFunction
    | SHOW CREATE TABLE qualifiedName                                  #showCreateTable
    | SHOW CREATE VIEW qualifiedName                                   #showCreateView
    | SHOW CREATE CUBE qualifiedName                                   #showCreateCube
    | SHOW TABLES ((FROM | IN) qualifiedName)?
        (LIKE pattern=string (ESCAPE escape=string)?)?                 #showTables
    | SHOW (SCHEMAS | DATABASES) ((FROM | IN) (cluster=identifier '.')? catalog=identifier)?
        (LIKE pattern=string (ESCAPE escape=string)?)?                 #showSchemas
    | SHOW CATALOGS (LIKE pattern=string)?                             #showCatalogs
    | SHOW COLUMNS (FROM | IN) qualifiedName                           #showColumns
    | SHOW FULL COLUMNS (FROM | IN) qualifiedName                           #showColumns
    | SHOW STATS FOR qualifiedName                                     #showStats
    | SHOW STATS FOR '(' querySpecification ')'                        #showStatsForQuery
    | SHOW CURRENT? ROLES ((FROM | IN) identifier)?                    #showRoles
    | SHOW ROLE GRANTS ((FROM | IN) identifier)?                       #showRoleGrants
    | DESCRIBE qualifiedName                                           #showColumns
    | DESC qualifiedName                                               #showColumns
    | SHOW FUNCTIONS
        (LIKE pattern=string (ESCAPE escape=string)?)?                 #showFunctions
    | SHOW SESSION                                                     #showSession
    | SET SESSION qualifiedName EQ expression                          #setSession
    | RESET SESSION qualifiedName                                      #resetSession
    | START TRANSACTION (transactionMode (',' transactionMode)*)?      #startTransaction
    | COMMIT WORK?                                                     #commit
    | ROLLBACK WORK?                                                   #rollback
    | PREPARE identifier FROM statement                                #prepare
    | DEALLOCATE PREPARE identifier                                    #deallocate
    | EXECUTE identifier (USING expression (',' expression)*)?         #execute
    | ALTER TABLE tableName=qualifiedName
            EXECUTE procedureName=identifier
            ('(' (callArgument (',' callArgument)*)? ')')?
            (WHERE where=booleanExpression)?                           #tableExecute
    | EXECUTE identifier (USING expression (',' expression)*)?         #execute
    | DESCRIBE INPUT identifier                                        #describeInput
    | DESCRIBE OUTPUT identifier                                       #describeOutput
    | SET PATH pathSpecification                                       #setPath
    | VACUUM TABLE qualifiedName (FULL (UNIFY)?)? (PARTITION partition=string)?
        (AND_WAIT)?                                                    #vacuumTable
    | REFRESH META CACHE (FOR (cluster=identifier '.')? catalog=identifier)?                   #refreshMetadataCache
    | SHOW VIEWS ((FROM | IN) qualifiedName)?
        (LIKE? pattern=string (ESCAPE escape=string)?)?                #showViews
    ;

assignmentList
    :  assignmentItem (',' assignmentItem)*
    ;

assignmentItem
    :  qualifiedName EQ expression
    ;

query
    :  with? queryNoWith
    ;

with
    : WITH RECURSIVE? namedQuery (',' namedQuery)*
    ;

tableElement
    : columnDefinition
    | likeClause
    ;

columnDefinition
    : identifier type (NOT NULL)? (PRIMARY KEY)? (COMMENT string)? (WITH properties)?
    ;

likeClause
    : LIKE qualifiedName (optionType=(INCLUDING | EXCLUDING) PROPERTIES)?
    ;

cubeProperties
    : '(' (cubeProperty (',' cubeProperty)*)? ')'
    ;

cubeProperty
    : AGGREGATIONS EQ '(' aggregations ')'
    | GROUP EQ '(' cubeGroup ')'
    | FILTER EQ '(' sourceFilter ')'
    | property
    ;

properties
    : '(' property (',' property)* ')'
    | '(' propertyAssignments ')'
    ;

propertyAssignments
    : property (',' property)*
    ;

property
    : identifier EQ expression
    ;

functionProperties
    : '(' functionProperty (',' functionProperty)* ')'
    ;

functionProperty
    : identifier EQ string
    ;

sqlParameterDeclaration
    : identifier type
    ;

routineCharacteristics
    : routineCharacteristic*
    ;

routineCharacteristic
    : LANGUAGE language
    | determinism
    | nullCallClause
    ;

routineBody
    : returnStatement
    | externalBodyReference
    ;

returnStatement
    : RETURN expression
    ;

externalBodyReference
    : EXTERNAL (NAME externalRoutineName)?
    ;

language
    : identifier
    ;

determinism
    : DETERMINISTIC
    | NOT DETERMINISTIC;

nullCallClause
    : RETURNS NULL ON NULL INPUT
    | CALLED ON NULL INPUT
    ;

externalRoutineName
    : identifier
    ;

queryNoWith:
      queryTerm
      (ORDER BY sortItem (',' sortItem)*)?
      (OFFSET offset=INTEGER_VALUE (ROW | ROWS)?)?
      ((LIMIT limit=(INTEGER_VALUE | ALL)) | (FETCH (FIRST | NEXT) (fetchFirst=INTEGER_VALUE)? (ROW | ROWS) (ONLY | WITH TIES)))?
    ;

queryTerm
    : queryPrimary                                                             #queryTermDefault
    | left=queryTerm operator=INTERSECT setQuantifier? right=queryTerm         #setOperation
    | left=queryTerm operator=(UNION | EXCEPT) setQuantifier? right=queryTerm  #setOperation
    ;

queryPrimary
    : querySpecification                   #queryPrimaryDefault
    | TABLE qualifiedName                  #table
    | VALUES expression (',' expression)*  #inlineTable
    | '(' queryNoWith  ')'                 #subquery
    ;

sortItem
    : expression ordering=(ASC | DESC)? (NULLS nullOrdering=(FIRST | LAST))?
    ;

querySpecification
    : SELECT setQuantifier? selectItem (',' selectItem)*
      (FROM relation (',' relation)*)?
      (WHERE where=booleanExpression)?
      (GROUP BY groupBy)?
      (HAVING having=booleanExpression)?
    ;

groupBy
    : setQuantifier? groupingElement (',' groupingElement)*
    ;

groupingElement
    : groupingSet                                            #singleGroupingSet
    | ROLLUP '(' (expression (',' expression)*)? ')'         #rollup
    | CUBE '(' (expression (',' expression)*)? ')'           #cube
    | GROUPING SETS '(' groupingSet (',' groupingSet)* ')'   #multipleGroupingSets
    ;

groupingSet
    : '(' (expression (',' expression)*)? ')'
    | expression
    ;

cubeGroup
    : (identifier (',' identifier)*)?
    ;

sourceFilter
    : expression
    ;

namedQuery
    : name=identifier (columnAliases)? AS '(' query ')'
    ;

setQuantifier
    : DISTINCT
    | ALL
    ;

selectItem
    : expression (AS? identifier)?  #selectSingle
    | qualifiedName '.' ASTERISK    #selectAll
    | ASTERISK                      #selectAll
    ;

relation
    : left=relation
      ( CROSS JOIN right=sampledRelation
      | joinType JOIN rightRelation=relation joinCriteria
      | NATURAL joinType JOIN right=sampledRelation
      )                                           #joinRelation
    | sampledRelation                             #relationDefault
    ;

joinType
    : INNER?
    | LEFT OUTER?
    | RIGHT OUTER?
    | FULL OUTER?
    ;

joinCriteria
    : ON booleanExpression
    | USING '(' identifier (',' identifier)* ')'
    ;

sampledRelation
    : aliasedRelation (
        TABLESAMPLE sampleType '(' percentage=expression ')'
      )?
    ;

sampleType
    : BERNOULLI
    | SYSTEM
    ;

aliasedRelation
    : relationPrimary (AS? identifier columnAliases?)?
    ;

columnAliases
    : '(' identifier (',' identifier)* ')'
    ;

relationPrimary
    : qualifiedName                                                   #tableName
    | '(' query ')'                                                   #subqueryRelation
    | UNNEST '(' expression (',' expression)* ')' (WITH ORDINALITY)?  #unnest
    | LATERAL '(' query ')'                                           #lateral
    | '(' relation ')'                                                #parenthesizedRelation
    ;

expression
    : booleanExpression
    ;

booleanExpression
    : valueExpression predicate[$valueExpression.ctx]?             #predicated
    | NOT booleanExpression                                        #logicalNot
    | left=booleanExpression operator=AND right=booleanExpression  #logicalBinary
    | left=booleanExpression operator=OR right=booleanExpression   #logicalBinary
    ;

// workaround for https://github.com/antlr/antlr4/issues/780
predicate[ParserRuleContext value]
    : comparisonOperator right=valueExpression                            #comparison
    | comparisonOperator comparisonQuantifier '(' query ')'               #quantifiedComparison
    | NOT? BETWEEN lower=valueExpression AND upper=valueExpression        #between
    | NOT? IN '(' expression (',' expression)* ')'                        #inList
    | NOT? IN '(' query ')'                                               #inSubquery
    | NOT? LIKE pattern=valueExpression (ESCAPE escape=valueExpression)?  #like
    | IS NOT? NULL                                                        #nullPredicate
    | IS NOT? DISTINCT FROM right=valueExpression                         #distinctFrom
    ;

valueExpression
    : primaryExpression                                                                 #valueExpressionDefault
    | valueExpression AT timeZoneSpecifier                                              #atTimeZone
    | operator=(MINUS | PLUS) valueExpression                                           #arithmeticUnary
    | left=valueExpression operator=(ASTERISK | SLASH | PERCENT) right=valueExpression  #arithmeticBinary
    | left=valueExpression operator=(PLUS | MINUS) right=valueExpression                #arithmeticBinary
    | left=valueExpression CONCAT right=valueExpression                                 #concatenation
    ;

primaryExpression
    : NULL                                                                                #nullLiteral
    | interval                                                                            #intervalLiteral
    | identifier string                                                                   #typeConstructor
    | DOUBLE_PRECISION string                                                             #typeConstructor
    | number                                                                              #numericLiteral
    | booleanValue                                                                        #booleanLiteral
    | string                                                                              #stringLiteral
    | BINARY_LITERAL                                                                      #binaryLiteral
    | '?'                                                                                 #parameter
    | POSITION '(' valueExpression IN valueExpression ')'                                 #position
    | '(' expression (',' expression)+ ')'                                                #rowConstructor
    | ROW '(' expression (',' expression)* ')'                                            #rowConstructor
    | qualifiedName '(' ASTERISK ')' filter? over?                                        #functionCall
    | qualifiedName '(' (setQuantifier? expression (',' expression)*)?
        (ORDER BY sortItem (',' sortItem)*)? ')' filter? (nullTreatment? over)?           #functionCall
    | identifier '->' expression                                                          #lambda
    | '(' (identifier (',' identifier)*)? ')' '->' expression                             #lambda
    | '(' query ')'                                                                       #subqueryExpression
    // This is an extension to ANSI SQL, which considers EXISTS to be a <boolean expression>
    | EXISTS '(' query ')'                                                                #exists
    | CASE valueExpression whenClause+ (ELSE elseExpression=expression)? END              #simpleCase
    | CASE whenClause+ (ELSE elseExpression=expression)? END                              #searchedCase
    | CAST '(' expression AS type ')'                                                     #cast
    | TRY_CAST '(' expression AS type ')'                                                 #cast
    | ARRAY '[' (expression (',' expression)*)? ']'                                       #arrayConstructor
    | value=primaryExpression '[' index=valueExpression ']'                               #subscript
    | identifier                                                                          #columnReference
    | base=primaryExpression '.' fieldName=identifier                                     #dereference
    | name=CURRENT_DATE                                                                   #specialDateTimeFunction
    | name=CURRENT_TIME ('(' precision=INTEGER_VALUE ')')?                                #specialDateTimeFunction
    | name=CURRENT_TIMESTAMP ('(' precision=INTEGER_VALUE ')')?                           #specialDateTimeFunction
    | name=LOCALTIME ('(' precision=INTEGER_VALUE ')')?                                   #specialDateTimeFunction
    | name=LOCALTIMESTAMP ('(' precision=INTEGER_VALUE ')')?                              #specialDateTimeFunction
    | name=CURRENT_USER                                                                   #currentUser
    | name=CURRENT_PATH                                                                   #currentPath
    | SUBSTRING '(' valueExpression FROM valueExpression (FOR valueExpression)? ')'       #substring
    | NORMALIZE '(' valueExpression (',' normalForm)? ')'                                 #normalize
    | EXTRACT '(' identifier FROM valueExpression ')'                                     #extract
    | '(' expression ')'                                                                  #parenthesizedExpression
    | GROUPING '(' (qualifiedName (',' qualifiedName)*)? ')'                              #groupingOperation
    ;

string
    : STRING                                #basicStringLiteral
    | UNICODE_STRING (UESCAPE STRING)?      #unicodeStringLiteral
    ;

nullTreatment
    : IGNORE NULLS
    | RESPECT NULLS
    ;

timeZoneSpecifier
    : TIME ZONE interval  #timeZoneInterval
    | TIME ZONE string    #timeZoneString
    ;

comparisonOperator
    : EQ | NEQ | LT | LTE | GT | GTE
    ;

comparisonQuantifier
    : ALL | SOME | ANY
    ;

aggregations
    :  expression (',' expression)*
    ;

booleanValue
    : TRUE | FALSE
    ;

indexType
    : BITMAP | BLOOM | MINMAX | BTREE
    ;

interval
    : INTERVAL sign=(PLUS | MINUS)? string from=intervalField (TO to=intervalField)?
    ;

intervalField
    : YEAR | MONTH | DAY | HOUR | MINUTE | SECOND
    ;

normalForm
    : NFD | NFC | NFKD | NFKC
    ;

types
    : '(' (type (',' type)*)? ')'
    ;

type
    : type ARRAY
    | ARRAY '<' type '>'
    | MAP '<' type ',' type '>'
    | ROW '(' identifier type (',' identifier type)* ')'
    | baseType ('(' typeParameter (',' typeParameter)* ')')?
    | INTERVAL from=intervalField TO to=intervalField
    ;

typeParameter
    : INTEGER_VALUE | type
    ;

baseType
    : TIME_WITH_TIME_ZONE
    | TIMESTAMP_WITH_TIME_ZONE
    | DOUBLE_PRECISION
    | identifier
    ;

whenClause
    : WHEN condition=expression THEN result=expression
    ;

filter
    : FILTER '(' WHERE booleanExpression ')'
    ;

over
    : OVER '('
        (PARTITION BY partition+=expression (',' partition+=expression)*)?
        (ORDER BY sortItem (',' sortItem)*)?
        windowFrame?
      ')'
    ;

windowFrame
    : frameType=RANGE start=frameBound
    | frameType=ROWS start=frameBound
    | frameType=RANGE BETWEEN start=frameBound AND end=frameBound
    | frameType=ROWS BETWEEN start=frameBound AND end=frameBound
    ;

frameBound
    : UNBOUNDED boundType=PRECEDING                 #unboundedFrame
    | UNBOUNDED boundType=FOLLOWING                 #unboundedFrame
    | CURRENT ROW                                   #currentRowBound
    | expression boundType=(PRECEDING | FOLLOWING)  #boundedFrame
    ;


explainOption
    : FORMAT value=(TEXT | GRAPHVIZ | JSON)                 #explainFormat
    | TYPE value=(LOGICAL | DISTRIBUTED | VALIDATE | IO)    #explainType
    ;

transactionMode
    : ISOLATION LEVEL levelOfIsolation    #isolationLevel
    | READ accessMode=(ONLY | WRITE)      #transactionAccessMode
    ;

levelOfIsolation
    : READ UNCOMMITTED                    #readUncommitted
    | READ COMMITTED                      #readCommitted
    | REPEATABLE READ                     #repeatableRead
    | SERIALIZABLE                        #serializable
    ;

callArgument
    : expression                    #positionalArgument
    | identifier '=>' expression    #namedArgument
    ;

pathElement
    : identifier '.' identifier     #qualifiedArgument
    | identifier                    #unqualifiedArgument
    ;

pathSpecification
    : pathElement (',' pathElement)*
    ;

privilege
    : SELECT | DELETE | UPDATE | INSERT | identifier
    ;

qualifiedName
    : identifier ('.' identifier)*
    ;

grantor
    : principal             #specifiedPrincipal
    | CURRENT_USER          #currentUserGrantor
    | CURRENT_ROLE          #currentRoleGrantor
    ;

principal
    : identifier            #unspecifiedPrincipal
    | USER identifier       #userPrincipal
    | ROLE identifier       #rolePrincipal
    ;

roles
    : identifier (',' identifier)*
    ;

identifier
    : IDENTIFIER             #unquotedIdentifier
    | QUOTED_IDENTIFIER      #quotedIdentifier
    | nonReserved            #unquotedIdentifier
    | BACKQUOTED_IDENTIFIER  #backQuotedIdentifier
    | DIGIT_IDENTIFIER       #digitIdentifier
    ;

number
    : MINUS? DECIMAL_VALUE  #decimalLiteral
    | MINUS? DOUBLE_VALUE   #doubleLiteral
    | MINUS? INTEGER_VALUE  #integerLiteral
    ;

nonReserved
    // IMPORTANT: this rule must only contain tokens. Nested rules are not supported. See SqlParser.exitNonReserved
    : ADD | ADMIN | ALL | ANALYZE | ANY | ARRAY | ASC | AT
    | BERNOULLI
    | CALL | CALLED | CASCADE | CATALOGS | COLUMN | COLUMNS | COMMENT | COMMIT | COMMITTED | CURRENT
    | DATA | DATABASE | DATABASES | DATE | DAY | DEFINER | DESC | DISTRIBUTED
    | EXCLUDING | EXPLAIN | EXTERNAL
    | FETCH | FILTER | FIRST | FOLLOWING | FORMAT | FUNCTION | FUNCTIONS
    | GRANT | GRANTED | GRANTS | GRAPHVIZ
    | HOUR
    | IF | INCLUDING | INPUT | INTERVAL | INVOKER | IO | ISOLATION
    | JSON
    | LAST | LATERAL | LEVEL | LIMIT | LOGICAL
    | MAP | MINUTE | MONTH
    | NAME | NEXT | NFC | NFD | NFKC | NFKD | NO | NONE | NULLIF | NULLS
    | OFFSET | ONLY | OPTION | ORDINALITY | OUTPUT | OVER
    | PARTITION | PARTITIONS | PATH | POSITION | PRECEDING | PRIVILEGES | PROPERTIES
    | RANGE | READ | RENAME | REPEATABLE | REPLACE | RESET | RESTRICT | RETURN | RETURNS | REVOKE | ROLE | ROLES | ROLLBACK | ROW | ROWS
    | SCHEMA | SCHEMAS | SECOND | SECURITY | SERIALIZABLE | SESSION | SET | SETS
    | SHOW | SOME | START | STATS | SUBSTRING | SYSTEM
    | TABLES | TABLESAMPLE | TEXT | TIES | TIME | TIMESTAMP | TO | TRANSACTION | TRY_CAST | TYPE
    | UNBOUNDED | UNCOMMITTED | USE | USER
    | VALIDATE | VERBOSE | VIEW | VIEWS
    | WORK | WRITE
    | YEAR
    | ZONE
    | INDEX
    ;

ADD: 'ADD';
ADMIN: 'ADMIN';
AGGREGATIONS: 'AGGREGATIONS';
ALL: 'ALL';
ALTER: 'ALTER';
ANALYZE: 'ANALYZE';
AND: 'AND';
AND_WAIT: 'AND WAIT';
ANY: 'ANY';
ARRAY: 'ARRAY';
AS: 'AS';
ASC: 'ASC';
AT: 'AT';
BERNOULLI: 'BERNOULLI';
BETWEEN: 'BETWEEN';
BY: 'BY';
CACHE: 'CACHE';
CALL: 'CALL';
CALLED: 'CALLED';
CASCADE: 'CASCADE';
CASE: 'CASE';
CAST: 'CAST';
CATALOGS: 'CATALOGS';
COLUMN: 'COLUMN';
COLUMNS: 'COLUMNS';
COMMENT: 'COMMENT';
COMMIT: 'COMMIT';
COMMITTED: 'COMMITTED';
CONSTRAINT: 'CONSTRAINT';
CREATE: 'CREATE';
CROSS: 'CROSS';
CUBE: 'CUBE';
CUBES: 'CUBES';
CURRENT: 'CURRENT';
CURRENT_DATE: 'CURRENT_DATE';
CURRENT_PATH: 'CURRENT_PATH';
CURRENT_ROLE: 'CURRENT_ROLE';
CURRENT_TIME: 'CURRENT_TIME';
CURRENT_TIMESTAMP: 'CURRENT_TIMESTAMP';
CURRENT_USER: 'CURRENT_USER';
DATA: 'DATA';
DATABASE: 'DATABASE';
DATABASES: 'DATABASES';
DATE: 'DATE';
DAY: 'DAY';
DEALLOCATE: 'DEALLOCATE';
DEFINER: 'DEFINER';
DELETE: 'DELETE';
DETERMINISTIC: 'DETERMINISTIC';
UPDATE: 'UPDATE';
DESC: 'DESC';
DESCRIBE: 'DESCRIBE';
DISTINCT: 'DISTINCT';
DISTRIBUTED: 'DISTRIBUTED';
DROP: 'DROP';
ELSE: 'ELSE';
END: 'END';
ESCAPE: 'ESCAPE';
EXCEPT: 'EXCEPT';
EXCLUDING: 'EXCLUDING';
EXECUTE: 'EXECUTE';
EXISTS: 'EXISTS';
EXPLAIN: 'EXPLAIN';
EXTRACT: 'EXTRACT';
EXTERNAL: 'EXTERNAL';
FALSE: 'FALSE';
FETCH: 'FETCH';
FILTER: 'FILTER';
FIRST: 'FIRST';
FOLLOWING: 'FOLLOWING';
FOR: 'FOR';
FORMAT: 'FORMAT';
FROM: 'FROM';
FULL: 'FULL';
FUNCTION: 'FUNCTION';
FUNCPROPERTIES: 'FUNCPROPERTIES';
FUNCTIONS: 'FUNCTIONS';
GRANT: 'GRANT';
GRANTED: 'GRANTED';
GRANTS: 'GRANTS';
GRAPHVIZ: 'GRAPHVIZ';
GROUP: 'GROUP';
GROUPING: 'GROUPING';
HAVING: 'HAVING';
HOUR: 'HOUR';
IF: 'IF';
IN: 'IN';
INCLUDING: 'INCLUDING';
INNER: 'INNER';
INPUT: 'INPUT';
INSERT: 'INSERT';
INTERSECT: 'INTERSECT';
INTERVAL: 'INTERVAL';
INTO: 'INTO';
INVOKER: 'INVOKER';
IO: 'IO';
IS: 'IS';
ISOLATION: 'ISOLATION';
JSON: 'JSON';
JOIN: 'JOIN';
LANGUAGE: 'LANGUAGE';
LAST: 'LAST';
LATERAL: 'LATERAL';
LEFT: 'LEFT';
LEVEL: 'LEVEL';
LIKE: 'LIKE';
LIMIT: 'LIMIT';
LOCALTIME: 'LOCALTIME';
LOCALTIMESTAMP: 'LOCALTIMESTAMP';
LOGICAL: 'LOGICAL';
MAP: 'MAP';
UNIFY: 'UNIFY';
MINUTE: 'MINUTE';
MONTH: 'MONTH';
NAME: 'NAME';
NATURAL: 'NATURAL';
NEXT: 'NEXT';
NFC : 'NFC';
NFD : 'NFD';
NFKC : 'NFKC';
NFKD : 'NFKD';
NO: 'NO';
NONE: 'NONE';
NORMALIZE: 'NORMALIZE';
NOT: 'NOT';
NULL: 'NULL';
NULLIF: 'NULLIF';
NULLS: 'NULLS';
OFFSET: 'OFFSET';
ON: 'ON';
ONLY: 'ONLY';
OPTION: 'OPTION';
OR: 'OR';
ORDER: 'ORDER';
ORDINALITY: 'ORDINALITY';
OUTER: 'OUTER';
OUTPUT: 'OUTPUT';
OVER: 'OVER';
OVERWRITE: 'OVERWRITE';
PARTITION: 'PARTITION';
PARTITIONS: 'PARTITIONS';
PATH: 'PATH';
POSITION: 'POSITION';
PRECEDING: 'PRECEDING';
PREPARE: 'PREPARE';
PRIVILEGES: 'PRIVILEGES';
PROPERTIES: 'PROPERTIES';
RANGE: 'RANGE';
READ: 'READ';
RECURSIVE: 'RECURSIVE';
RENAME: 'RENAME';
REPEATABLE: 'REPEATABLE';
REPLACE: 'REPLACE';
RESET: 'RESET';
RESTRICT: 'RESTRICT';
RETURN: 'RETURN';
RETURNS: 'RETURNS';
REVOKE: 'REVOKE';
RIGHT: 'RIGHT';
ROLE: 'ROLE';
ROLES: 'ROLES';
ROLLBACK: 'ROLLBACK';
ROLLUP: 'ROLLUP';
ROW: 'ROW';
ROWS: 'ROWS';
SCHEMA: 'SCHEMA';
SCHEMAS: 'SCHEMAS';
SECOND: 'SECOND';
SECURITY: 'SECURITY';
SELECT: 'SELECT';
SERIALIZABLE: 'SERIALIZABLE';
SESSION: 'SESSION';
SET: 'SET';
SETS: 'SETS';
SHOW: 'SHOW';
SOME: 'SOME';
START: 'START';
STATS: 'STATS';
SUBSTRING: 'SUBSTRING';
SYSTEM: 'SYSTEM';
TABLE: 'TABLE';
TABLES: 'TABLES';
TABLESAMPLE: 'TABLESAMPLE';
TEXT: 'TEXT';
THEN: 'THEN';
TIES: 'TIES';
TIME: 'TIME';
TIMESTAMP: 'TIMESTAMP';
TO: 'TO';
TRANSACTION: 'TRANSACTION';
TRUE: 'TRUE';
TRY_CAST: 'TRY_CAST';
TYPE: 'TYPE';
TRUNCATE: 'TRUNCATE';
UESCAPE: 'UESCAPE';
UNBOUNDED: 'UNBOUNDED';
UNCOMMITTED: 'UNCOMMITTED';
UNION: 'UNION';
UNNEST: 'UNNEST';
USE: 'USE';
USER: 'USER';
USING: 'USING';
VACUUM: 'VACUUM';
VALIDATE: 'VALIDATE';
VALUES: 'VALUES';
VERBOSE: 'VERBOSE';
VIEW: 'VIEW';
VIEWS: 'VIEWS';
WHEN: 'WHEN';
WHERE: 'WHERE';
WITH: 'WITH';
WORK: 'WORK';
WRITE: 'WRITE';
YEAR: 'YEAR';
ZONE: 'ZONE';
INDEX: 'INDEX';
BITMAP: 'BITMAP';
BLOOM: 'BLOOM';
MINMAX: 'MINMAX';
BTREE: 'BTREE';
REFRESH: 'REFRESH';
META: 'META';
RELOAD: 'RELOAD';
PRIMARY: 'PRIMARY';
KEY: 'KEY';

EQ  : '=';
NEQ : '<>' | '!=';
LT  : '<';
LTE : '<=';
GT  : '>';
GTE : '>=';

PLUS: '+';
MINUS: '-';
ASTERISK: '*';
SLASH: '/';
PERCENT: '%';
CONCAT: '||';

STRING
    : '\'' ( ~'\'' | '\'\'' )* '\''
    ;

UNICODE_STRING
    : 'U&\'' ( ~'\'' | '\'\'' )* '\''
    ;

// Note: we allow any character inside the binary literal and validate
// its a correct literal when the AST is being constructed. This
// allows us to provide more meaningful error messages to the user
BINARY_LITERAL
    :  'X\'' (~'\'')* '\''
    ;

INTEGER_VALUE
    : DIGIT+
    ;

DECIMAL_VALUE
    : DIGIT+ '.' DIGIT*
    | '.' DIGIT+
    ;

DOUBLE_VALUE
    : DIGIT+ ('.' DIGIT*)? EXPONENT
    | '.' DIGIT+ EXPONENT
    ;

IDENTIFIER
    : (LETTER | '_') (LETTER | DIGIT | '_' | '@' | ':')*
    ;

DIGIT_IDENTIFIER
    : DIGIT (LETTER | DIGIT | '_' | '@' | ':')+
    ;

QUOTED_IDENTIFIER
    : '"' ( ~'"' | '""' )* '"'
    ;

BACKQUOTED_IDENTIFIER
    : '`' ( ~'`' | '``' )* '`'
    ;

TIME_WITH_TIME_ZONE
    : 'TIME' WS 'WITH' WS 'TIME' WS 'ZONE'
    ;

TIMESTAMP_WITH_TIME_ZONE
    : 'TIMESTAMP' WS 'WITH' WS 'TIME' WS 'ZONE'
    ;

DOUBLE_PRECISION
    : 'DOUBLE' WS 'PRECISION'
    ;

fragment EXPONENT
    : 'E' [+-]? DIGIT+
    ;

fragment DIGIT
    : [0-9]
    ;

fragment LETTER
    : [A-Z]
    ;

SIMPLE_COMMENT
    : '--' ~[\r\n]* '\r'? '\n'? -> channel(HIDDEN)
    ;

BRACKETED_COMMENT
    : '/*' .*? '*/' -> channel(HIDDEN)
    ;

WS
    : [ \r\n\t]+ -> channel(HIDDEN)
    ;

// Catch-all for anything we can't recognize.
// We use this to be able to ignore and recover all the text
// when splitting statements with DelimiterLexer
UNRECOGNIZED
    : .
    ;
