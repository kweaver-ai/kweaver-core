-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- Issue #382 方案 B：实体级 extensions 副表
-- Vega 0.8.0 版本线（#426：自 0.9.0 目录迁入，与发布版本一致）

SET SCHEMA kweaver;

CREATE TABLE IF NOT EXISTS t_entity_extension (
    f_entity_kind VARCHAR(20 CHAR) NOT NULL,
    f_entity_id   VARCHAR(40 CHAR) NOT NULL,
    f_key         VARCHAR(128 CHAR) NOT NULL,
    f_value       VARCHAR(512 CHAR) NOT NULL,
    f_create_time BIGINT NOT NULL DEFAULT 0,
    f_update_time BIGINT NOT NULL DEFAULT 0,
    CLUSTER PRIMARY KEY (f_entity_kind, f_entity_id, f_key)
);

CREATE INDEX IF NOT EXISTS idx_t_entity_extension_entity ON t_entity_extension(f_entity_kind, f_entity_id);
CREATE INDEX IF NOT EXISTS idx_t_entity_extension_entity_key_value ON t_entity_extension(f_entity_kind, f_entity_id, f_key, f_value);
