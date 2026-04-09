USE model_management;

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