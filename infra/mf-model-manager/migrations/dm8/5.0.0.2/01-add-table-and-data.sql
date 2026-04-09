SET SCHEMA model_management;

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

