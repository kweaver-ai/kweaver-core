-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

USE adp;

DELETE FROM t_catalog WHERE f_id='adp_bkn_catalog';
DELETE FROM t_resource WHERE f_catalog_id='adp_bkn_catalog';

ALTER TABLE t_resource DROP COLUMN IF EXISTS f_logic_definition_type;