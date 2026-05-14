-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- Issue #382 方案 B：Catalog / Resource 可检索业务 KV（extensions）副表
-- Vega 0.8.0 版本线（#426：自 0.9.0 目录迁入，与发布版本一致）

USE kweaver;

CREATE TABLE IF NOT EXISTS t_entity_extension (
    f_entity_id VARCHAR(40) NOT NULL COMMENT '等于 t_catalog.f_id 或 t_resource.f_id（全局唯一）',
    f_key       VARCHAR(128) NOT NULL COMMENT '扩展键，建议 dip: 等业务前缀',
    f_value     VARCHAR(512) NOT NULL COMMENT '扩展值，可为 JSON 文本',
    f_create_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '创建时间（毫秒）',
    f_update_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '更新时间（毫秒）',
    PRIMARY KEY (f_entity_id, f_key),
    KEY idx_entity (f_entity_id),
    KEY idx_entity_key_value (f_entity_id, f_key, f_value(191))
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT='实体级 extensions 行存储（与 schema_definition 内 Property.extensions 分离）';
