-- 迁移数据从5.3.0 dip_data_agent schema 到 5.4.0 adp schema
-- 注意：使用WHERE NOT EXISTS避免重复插入数据

SET SCHEMA adp;

-- 迁移 t_data_agent_conversation 表数据
INSERT INTO adp.t_data_agent_conversation (
    f_id,
    f_agent_app_key,
    f_title,
    f_origin,
    f_message_index,
    f_read_message_index,
    f_ext,
    f_create_time,
    f_update_time,
    f_create_by,
    f_update_by,
    f_is_deleted
)
SELECT
    f_id,
    f_agent_app_key,
    f_title,
    f_origin,
    f_message_index,
    f_read_message_index,
    f_ext,
    f_create_time,
    f_update_time,
    f_create_by,
    f_update_by,
    f_is_deleted
FROM dip_data_agent.t_data_agent_conversation
WHERE NOT EXISTS (
    SELECT 1 FROM adp.t_data_agent_conversation t
    WHERE t.f_id = dip_data_agent.t_data_agent_conversation.f_id
);

-- 迁移 t_data_agent_conversation_message 表数据
INSERT INTO adp.t_data_agent_conversation_message (
    f_id,
    f_agent_app_key,
    f_conversation_id,
    f_agent_id,
    f_agent_version,
    f_reply_id,
    f_index,
    f_role,
    f_content,
    f_content_type,
    f_status,
    f_ext,
    f_create_time,
    f_update_time,
    f_create_by,
    f_update_by,
    f_is_deleted
)
SELECT
    f_id,
    f_agent_app_key,
    f_conversation_id,
    f_agent_id,
    f_agent_version,
    f_reply_id,
    f_index,
    f_role,
    f_content,
    f_content_type,
    f_status,
    f_ext,
    f_create_time,
    f_update_time,
    f_create_by,
    f_update_by,
    f_is_deleted
FROM dip_data_agent.t_data_agent_conversation_message
WHERE NOT EXISTS (
    SELECT 1 FROM adp.t_data_agent_conversation_message t
    WHERE t.f_id = dip_data_agent.t_data_agent_conversation_message.f_id
);

-- t_data_agent_temporary_area 表
INSERT INTO adp.t_data_agent_temporary_area (
    f_temp_area_id,
    f_source_id,
    f_conversation_id,
    f_id,
    f_created_at,
    f_source_type,
    f_user_id
)
SELECT
    f_temp_area_id,
    f_source_id,
    f_conversation_id,
    f_id,
    f_created_at,
    f_source_type,
    f_user_id
FROM dip_data_agent.t_temporary_area
WHERE NOT EXISTS (
    SELECT 1 FROM adp.t_data_agent_temporary_area t
    WHERE t.f_id = dip_data_agent.t_temporary_area.f_id
);
