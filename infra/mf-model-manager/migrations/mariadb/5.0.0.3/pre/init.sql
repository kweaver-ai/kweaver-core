USE model_management;


create table if not exists t_llm_model
(
    f_model_id     varchar(50)             not null,
    f_model_series varchar(50)             not null,
    f_model_type   varchar(50)             not null,
    f_model_name   varchar(100)            not null,
    f_model        varchar(50)             not null,
    f_model_config varchar(1000)           not null,
    f_create_by    varchar(50)             not null,
    f_create_time  datetime(6)             null,
    f_update_by    varchar(50)             null,
    f_update_time  datetime(6)             null,
    f_max_model_len        int(11)         null,
    f_model_parameters     int(11)         null,
    f_quota        int default 0    not null,
    primary key (f_model_id)
);



create table if not exists t_small_model
(
    f_model_id varchar(50) not null comment '主键，使用雪花id',
    f_model_name varchar(50) not null comment '小模型名称',
    f_model_type varchar(50) not null comment '小模型类型',
    f_model_config varchar(1000) not null comment '小模型配置json',
    f_create_time datetime(6) not null comment '创建时间',
    f_update_time datetime(6) not null comment '编辑时间',
    f_create_by    varchar(50)             not null,
    f_update_by    varchar(50)             null,
    f_adapter  int default 0    null,
    f_adapter_code       text          null,
    primary key (f_model_id)
);

create table if not exists t_prompt_item_list
(
    f_id                  varchar(50)          not null,
    f_prompt_item_id      varchar(50)          not null,
    f_prompt_item_name    varchar(50)          not null,
    f_prompt_item_type_id varchar(50)          null,
    f_prompt_item_type    varchar(50)          null,
    f_create_by           varchar(50)          not null,
    f_create_time         datetime(6)          null,
    f_update_by           varchar(50)          null,
    f_update_time         datetime(6)          null,
    f_item_is_delete      int default 0 not null,
    f_type_is_delete      int default 0 not null,
    f_built_in            int default 0        not null,
    primary key (f_id)
);


create table if not exists t_prompt_list
(
    f_prompt_id           varchar(50)          not null,
    f_prompt_item_id      varchar(50)          not null,
    f_prompt_item_type_id varchar(50)          not null,
    f_prompt_service_id   varchar(50)          not null,
    f_prompt_type         varchar(50)          not null,
    f_prompt_name         varchar(50)          not null,
    f_prompt_desc         varchar(255)         null,
    f_messages            longtext             null,
    f_variables           varchar(1000)        null,
    f_icon                varchar(50)          not null,
    f_model_id            varchar(50)          not null,
    f_model_para          varchar(150)         not null,
    f_opening_remarks     varchar(150)         null,
    f_is_deploy           int default 0 not null,
    f_prompt_deploy_url   varchar(150)         null,
    f_prompt_deploy_api   varchar(150)         null,
    f_create_by           varchar(50)          not null,
    f_create_time         datetime(6)          null,
    f_update_by           varchar(50)          null,
    f_update_time         datetime(6)          null,
    f_is_delete           int default 0 not null,
    f_built_in            int default 0        not null,
    primary key (f_prompt_id),
    unique key uk_f_prompt_service_id (f_prompt_service_id)
);

create table if not exists t_prompt_template_list
(
    f_prompt_id       varchar(50)          not null,
    f_prompt_type     varchar(50)          not null,
    f_prompt_name     varchar(50)          not null,
    f_prompt_desc     varchar(255)         null,
    f_messages        longtext             null,
    f_variables       varchar(1000)        null,
    f_icon            varchar(50)          not null,
    f_opening_remarks varchar(150)         null,
    f_input           varchar(1000)        null,
    f_create_by       varchar(50)          not null,
    f_create_time     datetime(6)          null,
    f_update_by       varchar(50)          null,
    f_update_time     datetime(6)          null,
    f_is_delete       int default 0 not null,
    primary key (f_prompt_id)
);

CREATE TABLE if not exists t_model_monitor (
    f_id                  varchar(50)          not null,
    f_create_time         datetime          not null,
    f_model_name         varchar(50)         not null,
    f_model_id          varchar(50)          not null,
    f_generation_tokens_total BIGINT not null,
    f_prompt_tokens_total BIGINT not null,
    f_average_first_token_time DECIMAL(10, 2) not null,
    f_generation_token_speed  DECIMAL(10, 2) not null,
    f_total_token_speed  DECIMAL(10, 2) not null,
    primary key (f_id)
);

create table if not exists t_model_quota_config
(
    f_id varchar(50) not null comment '主键，使用雪花id',
    f_model_id varchar(50) not null comment '模型id',
    f_billing_type int not null comment '0 统一计费 ， 1 input output 单独计费',
    f_input_tokens float not null comment 'input tokens配额',
    f_output_tokens float not null comment 'output tokens配额',
    f_referprice_in float not null comment 'input tokens参考单价',
    f_referprice_out float not null comment 'output tokens参考单价',
    f_currency_type bigint not null comment '货币类型 0:RMB/人民币 1:$/美元',
    f_create_time datetime(6) not null comment '创建时间',
    f_update_time datetime(6) not null comment '编辑时间',
    f_num_type varchar(50) not null comment '1-千  2-万 3-百万 4-千万',
    f_price_type varchar(50) not null default '["thousand", "thousand"]' comment '列表，计费单价显示单位, thousand-/千tokens million-/百万tokens',
    primary key (f_id)
);

create table if not exists t_user_quota_config
(
    f_id varchar(50) not null comment '主键，使用雪花id',
    f_model_conf varchar(50) not null comment '模型配额配置id（基于哪个模型配额）',
    f_user_id varchar(50) not null comment '用户id',
    f_input_tokens float not null comment 'input tokens配额',
    f_output_tokens float not null comment 'output tokens配额',
    f_create_time datetime(6) not null comment '创建时间',
    f_update_time datetime(6) not null comment '编辑时间',
    f_num_type varchar(50) not null comment '1-千  2-万 3-百万 4-千万',
    primary key (f_id)
);

create table if not exists t_model_op_detail
(
    f_id varchar(50) not null comment '主键，使用雪花id',
    f_model_id varchar(50) not null comment '模型id',
    f_user_id varchar(50) not null comment '用户id',
    f_input_tokens bigint not null comment 'input tokens消费 ',
    f_output_tokens bigint not null comment 'output tokens消费',
    f_referprice_in float not null comment 'input tokens参考单价',
    f_referprice_out float not null comment 'output tokens参考单价',
    f_total_price DECIMAL(38,10) not null comment '消费总金额',
    f_create_time datetime(6) not null comment '创建时间',
    f_currency_type bigint not null comment '货币类型 0:RMB/人民币 1:$/美元',
    f_price_type varchar(50) not null default '["thousand", "thousand"]' comment '列表，计费单价显示单位, thousand-/千tokens million-/百万tokens',
    primary key (f_id)
);

insert into t_prompt_item_list(f_create_by, f_create_time, f_id, f_item_is_delete, f_prompt_item_id, f_prompt_item_name,
                    f_prompt_item_type, f_prompt_item_type_id, f_type_is_delete, f_update_by, f_update_time, f_built_in)
                select 'admin', current_timestamp, '1500000000000000001', 0, '1510000000000000001', '内置提示词',
                    'chat', '1520000000000000001', 0, 'admin', current_timestamp, 1
                from DUAL where not exists(select f_id from t_prompt_item_list where f_id = '1500000000000000001');


insert into t_prompt_list(f_create_by, f_create_time, f_icon, f_is_delete, f_is_deploy, f_messages, f_model_id,
                    f_model_para, f_opening_remarks, f_prompt_deploy_api, f_prompt_deploy_url, f_prompt_desc,
                    f_prompt_id, f_prompt_item_id, f_prompt_item_type_id, f_prompt_name, f_prompt_service_id,
                    f_prompt_type, f_update_by, f_update_time, f_variables, f_built_in)
                select 'admin', current_timestamp, 5, 0, 0, '你可以重新组织和输出混乱复杂的会议记录，并根据当前状态、遇到的问题和提出的解决方案撰写会议纪要。你只负责会议记录方面的问题，不回答其他。
', '', '{}', '', null, null, '帮你重新组织和输出混乱复杂的会议纪要',
                    '1100000000000000030', '1510000000000000001', '1520000000000000001', '会议纪要', '1200000000000000030',
                    'chat', 'admin', current_timestamp,
                    '[]', 1
                from DUAL where not exists(select f_prompt_id from t_prompt_list where f_prompt_id = '1100000000000000030');