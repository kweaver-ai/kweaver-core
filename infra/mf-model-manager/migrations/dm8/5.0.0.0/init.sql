SET SCHEMA model_management;

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
    CLUSTER PRIMARY KEY (f_model_id)
);

