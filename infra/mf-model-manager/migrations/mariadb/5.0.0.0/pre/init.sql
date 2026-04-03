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
    primary key (f_model_id)
);
