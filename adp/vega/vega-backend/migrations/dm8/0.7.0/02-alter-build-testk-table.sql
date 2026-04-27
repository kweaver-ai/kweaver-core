-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：修改 t_build_task 表的索引结构
-- ==========================================

SET SCHEMA kweaver;

ALTER TABLE t_build_task ADD COLUMN IF NOT EXISTS f_embedding_fields VARCHAR(255 CHAR) NOT NULL DEFAULT '';
ALTER TABLE t_build_task ADD COLUMN IF NOT EXISTS f_build_key_fields VARCHAR(255 CHAR) NOT NULL DEFAULT '';
ALTER TABLE t_build_task ADD COLUMN IF NOT EXISTS f_embedding_model VARCHAR(40 CHAR) NOT NULL DEFAULT '';
ALTER TABLE t_build_task ADD COLUMN IF NOT EXISTS f_model_dimensions INT NOT NULL DEFAULT 0;
