-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：在 kweaver schema 下创建 mf-model-manager 相关表，并从 adp schema 复制数据
-- ==========================================

SET SCHEMA kweaver;

-- ==========================================
-- 1. t_llm_model
-- ==========================================
CREATE TABLE if not exists t_llm_model
(
    f_model_id     VARCHAR(50 CHAR)             not null,
    f_model_series VARCHAR(50 CHAR)             not null,
    f_model_type   VARCHAR(50 CHAR)             not null,
    f_model_name   VARCHAR(100 CHAR)            not null,
    f_model        VARCHAR(50 CHAR)             not null,
    f_model_config VARCHAR(1000 CHAR)           not null,
    f_create_by    VARCHAR(50 CHAR)             not null,
    f_create_time  datetime(6)             null,
    f_update_by    VARCHAR(50 CHAR)             null,
    f_update_time  datetime(6)             null,
    f_max_model_len        INT         null,
    f_model_parameters     INT         null,
    "f_quota" INT DEFAULT 0,
    "f_default" int DEFAULT 0,
    CLUSTER PRIMARY KEY (f_model_id)
);

INSERT INTO kweaver."t_llm_model" (
    f_model_id, f_model_series, f_model_type, f_model_name, f_model, f_model_config,
    f_create_by, f_create_time, f_update_by, f_update_time,
    f_max_model_len, f_model_parameters, "f_quota", "f_default"
)
SELECT
    f_model_id, f_model_series, f_model_type, f_model_name, f_model, f_model_config,
    f_create_by, f_create_time, f_update_by, f_update_time,
    f_max_model_len, f_model_parameters, "f_quota", "f_default"
FROM adp."t_llm_model" s;

-- ==========================================
-- 2. t_small_model
-- ==========================================
CREATE TABLE if not exists t_small_model
(
    f_model_id VARCHAR(50 CHAR) not null,
    f_model_name VARCHAR(50 CHAR) not null,
    f_model_type VARCHAR(50 CHAR) not null,
    f_model_config VARCHAR(1000 CHAR) not null,
    f_create_time datetime(6) not null,
    f_update_time datetime(6) not null,
    f_create_by    VARCHAR(50 CHAR)             not null,
    f_update_by    VARCHAR(50 CHAR)             null,
    "f_adapter" INT DEFAULT 0,
    "f_adapter_code" VARCHAR(15000 CHAR),
    "f_batch_size" int,
    "f_max_tokens" int,
    "f_embedding_dim" int,
    CLUSTER PRIMARY KEY (f_model_id)
);

INSERT INTO kweaver."t_small_model" (
    f_model_id, f_model_name, f_model_type, f_model_config,
    f_create_time, f_update_time, f_create_by, f_update_by,
    "f_adapter", "f_adapter_code", "f_batch_size", "f_max_tokens", "f_embedding_dim"
)
SELECT
    f_model_id, f_model_name, f_model_type, f_model_config,
    f_create_time, f_update_time, f_create_by, f_update_by,
    "f_adapter", "f_adapter_code", "f_batch_size", "f_max_tokens", "f_embedding_dim"
FROM adp."t_small_model" s;

-- ==========================================
-- 3. t_prompt_item_list
-- ==========================================
CREATE TABLE if not exists t_prompt_item_list
(
    f_id                  VARCHAR(50 CHAR)          not null,
    f_prompt_item_id      VARCHAR(50 CHAR)          not null,
    f_prompt_item_name    VARCHAR(50 CHAR)          not null,
    f_prompt_item_type_id VARCHAR(50 CHAR)          null,
    f_prompt_item_type    VARCHAR(50 CHAR)          null,
    f_create_by           VARCHAR(50 CHAR)          not null,
    f_create_time         datetime(6)          null,
    f_update_by           VARCHAR(50 CHAR)          null,
    f_update_time         datetime(6)          null,
    f_item_is_delete      INT default 0 not null,
    f_type_is_delete      INT default 0 not null,
    f_built_in            INT default 0        not null,
    CLUSTER PRIMARY KEY (f_id)
);

INSERT INTO kweaver."t_prompt_item_list" (
    f_id, f_prompt_item_id, f_prompt_item_name, f_prompt_item_type_id, f_prompt_item_type,
    f_create_by, f_create_time, f_update_by, f_update_time,
    f_item_is_delete, f_type_is_delete, f_built_in
)
SELECT
    f_id, f_prompt_item_id, f_prompt_item_name, f_prompt_item_type_id, f_prompt_item_type,
    f_create_by, f_create_time, f_update_by, f_update_time,
    f_item_is_delete, f_type_is_delete, f_built_in
FROM adp."t_prompt_item_list" s;

-- ==========================================
-- 4. t_prompt_list
-- ==========================================
CREATE TABLE if not exists t_prompt_list
(
    f_prompt_id           VARCHAR(50 CHAR)          not null,
    f_prompt_item_id      VARCHAR(50 CHAR)          not null,
    f_prompt_item_type_id VARCHAR(50 CHAR)          not null,
    f_prompt_service_id   VARCHAR(50 CHAR)          not null,
    f_prompt_type         VARCHAR(50 CHAR)          not null,
    f_prompt_name         VARCHAR(50 CHAR)          not null,
    f_prompt_desc         VARCHAR(255 CHAR)         null,
    f_messages            text             null,
    f_variables           VARCHAR(1000 CHAR)        null,
    f_icon                VARCHAR(50 CHAR)          not null,
    f_model_id            VARCHAR(50 CHAR)          not null,
    f_model_para          VARCHAR(150 CHAR)         not null,
    f_opening_remarks     VARCHAR(150 CHAR)         null,
    f_is_deploy           INT default 0 not null,
    f_prompt_deploy_url   VARCHAR(150 CHAR)         null,
    f_prompt_deploy_api   VARCHAR(150 CHAR)         null,
    f_create_by           VARCHAR(50 CHAR)          not null,
    f_create_time         datetime(6)          null,
    f_update_by           VARCHAR(50 CHAR)          null,
    f_update_time         datetime(6)          null,
    f_is_delete           INT default 0 not null,
    f_built_in            INT default 0        not null,
    CLUSTER PRIMARY KEY (f_prompt_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_prompt_list_uk_f_prompt_service_id ON t_prompt_list(f_prompt_service_id);

INSERT INTO kweaver."t_prompt_list" (
    f_prompt_id, f_prompt_item_id, f_prompt_item_type_id, f_prompt_service_id,
    f_prompt_type, f_prompt_name, f_prompt_desc, f_messages, f_variables,
    f_icon, f_model_id, f_model_para, f_opening_remarks,
    f_is_deploy, f_prompt_deploy_url, f_prompt_deploy_api,
    f_create_by, f_create_time, f_update_by, f_update_time,
    f_is_delete, f_built_in
)
SELECT
    f_prompt_id, f_prompt_item_id, f_prompt_item_type_id, f_prompt_service_id,
    f_prompt_type, f_prompt_name, f_prompt_desc, f_messages, f_variables,
    f_icon, f_model_id, f_model_para, f_opening_remarks,
    f_is_deploy, f_prompt_deploy_url, f_prompt_deploy_api,
    f_create_by, f_create_time, f_update_by, f_update_time,
    f_is_delete, f_built_in
FROM adp."t_prompt_list" s;

-- ==========================================
-- 5. t_prompt_template_list
-- ==========================================
CREATE TABLE if not exists t_prompt_template_list
(
    f_prompt_id       VARCHAR(50 CHAR)          not null,
    f_prompt_type     VARCHAR(50 CHAR)          not null,
    f_prompt_name     VARCHAR(50 CHAR)          not null,
    f_prompt_desc     VARCHAR(255 CHAR)         null,
    f_messages        text             null,
    f_variables       VARCHAR(1000 CHAR)        null,
    f_icon            VARCHAR(50 CHAR)          not null,
    f_opening_remarks VARCHAR(150 CHAR)         null,
    f_input           VARCHAR(1000 CHAR)        null,
    f_create_by       VARCHAR(50 CHAR)          not null,
    f_create_time     datetime(6)          null,
    f_update_by       VARCHAR(50 CHAR)          null,
    f_update_time     datetime(6)          null,
    f_is_delete       INT default 0 not null,
    CLUSTER PRIMARY KEY (f_prompt_id)
);

INSERT INTO kweaver."t_prompt_template_list" (
    f_prompt_id, f_prompt_type, f_prompt_name, f_prompt_desc,
    f_messages, f_variables, f_icon, f_opening_remarks, f_input,
    f_create_by, f_create_time, f_update_by, f_update_time, f_is_delete
)
SELECT
    f_prompt_id, f_prompt_type, f_prompt_name, f_prompt_desc,
    f_messages, f_variables, f_icon, f_opening_remarks, f_input,
    f_create_by, f_create_time, f_update_by, f_update_time, f_is_delete
FROM adp."t_prompt_template_list" s;

-- ==========================================
-- 6. t_model_monitor
-- ==========================================
CREATE TABLE if not exists t_model_monitor (
    f_id                  VARCHAR(50 CHAR)          not null,
    f_create_time         datetime(0) not null,
    f_model_name         VARCHAR(50 CHAR)         not null,
    f_model_id          VARCHAR(50 CHAR)          not null,
    f_generation_tokens_total BIGINT not null,
    f_prompt_tokens_total BIGINT not null,
    f_average_first_token_time DECIMAL(10, 2) not null,
    f_generation_token_speed  DECIMAL(10, 2) not null,
    f_total_token_speed  DECIMAL(10, 2) not null,
    CLUSTER PRIMARY KEY (f_id)
);

INSERT INTO kweaver."t_model_monitor" (
    f_id, f_create_time, f_model_name, f_model_id,
    f_generation_tokens_total, f_prompt_tokens_total,
    f_average_first_token_time, f_generation_token_speed, f_total_token_speed
)
SELECT
    f_id, f_create_time, f_model_name, f_model_id,
    f_generation_tokens_total, f_prompt_tokens_total,
    f_average_first_token_time, f_generation_token_speed, f_total_token_speed
FROM adp."t_model_monitor" s;

-- ==========================================
-- 7. t_model_quota_config
-- ==========================================
CREATE TABLE if not exists t_model_quota_config
(
    f_id VARCHAR(50 CHAR) not null,
    f_model_id VARCHAR(50 CHAR) not null,
    f_billing_type INT not null,
    f_input_tokens FLOAT not null,
    f_output_tokens FLOAT not null,
    f_referprice_in FLOAT not null,
    f_referprice_out FLOAT not null,
    f_currency_type BIGINT not null,
    f_create_time datetime(6) not null,
    f_update_time datetime(6) not null,
    f_num_type VARCHAR(50 CHAR) not null,
    f_price_type VARCHAR(50 CHAR) not null default '["thousand", "thousand"]',
    CLUSTER PRIMARY KEY (f_id)
);

INSERT INTO kweaver."t_model_quota_config" (
    f_id, f_model_id, f_billing_type,
    f_input_tokens, f_output_tokens, f_referprice_in, f_referprice_out,
    f_currency_type, f_create_time, f_update_time, f_num_type, f_price_type
)
SELECT
    f_id, f_model_id, f_billing_type,
    f_input_tokens, f_output_tokens, f_referprice_in, f_referprice_out,
    f_currency_type, f_create_time, f_update_time, f_num_type, f_price_type
FROM adp."t_model_quota_config" s;

-- ==========================================
-- 8. t_user_quota_config
-- ==========================================
CREATE TABLE if not exists t_user_quota_config
(
    f_id VARCHAR(50 CHAR) not null,
    f_model_conf VARCHAR(50 CHAR) not null,
    f_user_id VARCHAR(50 CHAR) not null,
    f_input_tokens FLOAT not null,
    f_output_tokens FLOAT not null,
    f_create_time datetime(6) not null,
    f_update_time datetime(6) not null,
    f_num_type VARCHAR(50 CHAR) not null,
    CLUSTER PRIMARY KEY (f_id)
);

INSERT INTO kweaver."t_user_quota_config" (
    f_id, f_model_conf, f_user_id,
    f_input_tokens, f_output_tokens,
    f_create_time, f_update_time, f_num_type
)
SELECT
    f_id, f_model_conf, f_user_id,
    f_input_tokens, f_output_tokens,
    f_create_time, f_update_time, f_num_type
FROM adp."t_user_quota_config" s;

-- ==========================================
-- 9. t_model_op_detail
-- ==========================================
CREATE TABLE if not exists t_model_op_detail
(
    f_id VARCHAR(50 CHAR) not null,
    f_model_id VARCHAR(50 CHAR) not null,
    f_user_id VARCHAR(50 CHAR) not null,
    f_input_tokens BIGINT not null,
    f_output_tokens BIGINT not null,
    f_referprice_in FLOAT not null,
    f_referprice_out FLOAT not null,
    f_total_price DECIMAL(38,10) not null,
    f_create_time datetime(6) not null,
    f_currency_type BIGINT not null,
    f_price_type VARCHAR(50 CHAR) not null default '["thousand", "thousand"]',
    f_total_count int default 0 not null,
    f_failed_count int default 0 not null,
    f_average_total_time FLOAT default 0.0,
    f_average_first_time FLOAT default 0.0,
    CLUSTER PRIMARY KEY (f_id)
);

INSERT INTO kweaver."t_model_op_detail" (
    f_id, f_model_id, f_user_id,
    f_input_tokens, f_output_tokens, f_referprice_in, f_referprice_out,
    f_total_price, f_create_time, f_currency_type, f_price_type,
    f_total_count, f_failed_count, f_average_total_time, f_average_first_time
)
SELECT
    f_id, f_model_id, f_user_id,
    f_input_tokens, f_output_tokens, f_referprice_in, f_referprice_out,
    f_total_price, f_create_time, f_currency_type, f_price_type,
    f_total_count, f_failed_count, f_average_total_time, f_average_first_time
FROM adp."t_model_op_detail" s;
