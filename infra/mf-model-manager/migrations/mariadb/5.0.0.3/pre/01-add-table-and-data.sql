USE model_management;


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
    f_num_type varchar(50) not null comment '1-千  2-万 3-百万 4-千万 5-亿',
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