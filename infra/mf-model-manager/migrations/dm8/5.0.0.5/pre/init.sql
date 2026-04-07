SET SCHEMA adp;

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

INSERT INTO t_prompt_item_list(f_create_by, f_create_time, f_id, f_item_is_delete, f_prompt_item_id, f_prompt_item_name,

                    f_prompt_item_type, f_prompt_item_type_id, f_type_is_delete, f_update_by, f_update_time, f_built_in)

                select 'admin', current_timestamp, '1500000000000000001', 0, '1510000000000000001', '内置提示词',

                    'chat', '1520000000000000001', 0, 'admin', current_timestamp, 1

                from DUAL where not exists(select f_id from t_prompt_item_list where f_id = '1500000000000000001');

INSERT INTO t_prompt_list(f_create_by, f_create_time, f_icon, f_is_delete, f_is_deploy, f_messages, f_model_id,

                    f_model_para, f_opening_remarks, f_prompt_deploy_api, f_prompt_deploy_url, f_prompt_desc,

                    f_prompt_id, f_prompt_item_id, f_prompt_item_type_id, f_prompt_name, f_prompt_service_id,

                    f_prompt_type, f_update_by, f_update_time, f_variables, f_built_in)

                select 'admin', current_timestamp, 5, 0, 0, '你可以重新组织和输出混乱复杂的会议记录，并根据当前状态、遇到的问题和提出的解决方案撰写会议纪要。你只负责会议记录方面的问题，不回答其他。

', '', '{}', '', null, null, '帮你重新组织和输出混乱复杂的会议纪要',

                    '1100000000000000030', '1510000000000000001', '1520000000000000001', '会议纪要', '1200000000000000030',

                    'chat', 'admin', current_timestamp,

                    '[]', 1

                from DUAL where not exists(select f_prompt_id from t_prompt_list where f_prompt_id = '1100000000000000030');

