SET SCHEMA adp;

CREATE TABLE if not exists t_data_agent_conversation
(
    f_id                 VARCHAR(40 CHAR)  not null,
    f_agent_app_key      VARCHAR(40 CHAR)  not null,
    f_title              VARCHAR(255 CHAR) not null,
    f_origin             VARCHAR(40 CHAR) not null default 'web_chat',
    f_message_index      INT not null default 0,
    f_read_message_index INT not null default 0,
    f_ext                text not null,
    f_create_time        BIGINT  not null default 0,
    f_update_time        BIGINT  not null default 0,
    f_create_by          VARCHAR(40 CHAR)  not null default '',
    f_update_by          VARCHAR(40 CHAR)  not null default '',
    f_is_deleted         TINYINT not null default 0,
    CLUSTER PRIMARY KEY (f_id)
);

CREATE INDEX IF NOT EXISTS t_data_agent_conversation_idx_agent_app_key ON t_data_agent_conversation(f_agent_app_key);



CREATE TABLE if not exists t_data_agent_conversation_message (
    f_id              VARCHAR(40 CHAR) not null,
    f_agent_app_key   VARCHAR(40 CHAR)  not null,
    f_conversation_id VARCHAR(40 CHAR) not null default '',
    f_agent_id        VARCHAR(40 CHAR)  not null,
    f_agent_version   VARCHAR(32 CHAR)  not null,
    f_reply_id        VARCHAR(40 CHAR) not null default '',
    f_index           INT not null,
    f_role            VARCHAR(255 CHAR) not null,
    f_content         text  not null,
    f_content_type    VARCHAR(32 CHAR),
    f_status          VARCHAR(32 CHAR),
    f_ext             text  not null,
    f_create_time        BIGINT  not null default 0,
    f_update_time        BIGINT  not null default 0,
    f_create_by          VARCHAR(40 CHAR)  not null default '',
    f_update_by          VARCHAR(40 CHAR)  not null default '',
    f_is_deleted         TINYINT not null default 0,
    CLUSTER PRIMARY KEY (f_id)
);

CREATE INDEX IF NOT EXISTS t_data_agent_conversation_message_idx_agent_app_key ON t_data_agent_conversation_message(f_agent_app_key);
CREATE INDEX IF NOT EXISTS t_data_agent_conversation_message_idx_conversation_id ON t_data_agent_conversation_message(f_conversation_id);



CREATE TABLE IF NOT EXISTS t_data_agent_temporary_area (
  f_temp_area_id VARCHAR(100 CHAR) NOT NULL,
  f_source_id VARCHAR(40 CHAR) DEFAULT NULL,
  f_conversation_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_id BIGINT NOT NULL IDENTITY(1, 1),
  f_created_at BIGINT DEFAULT 0,
  f_source_type VARCHAR(40 CHAR) DEFAULT NULL,
  f_user_id VARCHAR(40 CHAR) NOT NULL,
  CLUSTER PRIMARY KEY (f_id)
);

CREATE INDEX IF NOT EXISTS t_data_agent_temporary_area_idx_temp_area_id ON t_data_agent_temporary_area(f_temp_area_id);
CREATE INDEX IF NOT EXISTS t_data_agent_temporary_area_idx_source_id ON t_data_agent_temporary_area(f_source_id);
CREATE INDEX IF NOT EXISTS t_data_agent_temporary_area_idx_conversation_id ON t_data_agent_temporary_area(f_conversation_id);
CREATE INDEX IF NOT EXISTS t_data_agent_temporary_area_idx_created_at ON t_data_agent_temporary_area(f_created_at);

