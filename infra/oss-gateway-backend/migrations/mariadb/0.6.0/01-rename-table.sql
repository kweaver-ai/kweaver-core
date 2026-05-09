-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：将 oss-gateway-backend 相关表从 adp 库迁移至 kweaver 库
-- ==========================================
USE kweaver;

RENAME TABLE adp.t_storage_config TO kweaver.t_storage_config;
RENAME TABLE adp.t_multipart_upload_task TO kweaver.t_multipart_upload_task;
