SET SCHEMA model_management;

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
    CLUSTER PRIMARY KEY (f_id)
);

