SET SEARCH_PATH TO adp;

-- 迁移数据从 dip_data_operator_hub 到 adp
-- 注意：使用 WHERE NOT EXISTS 避免重复插入数据

-- 1. 迁移t_metadata_api表数据
INSERT INTO adp.t_metadata_api (
    f_summary,
    f_version,
    f_svc_url,
    f_description,
    f_path,
    f_method,
    f_api_spec,
    f_create_user,
    f_update_user,
    f_create_time,
    f_update_time
)
SELECT
    f_summary,
    f_version,
    f_svc_url,
    f_description,
    f_path,
    f_method,
    f_api_spec,
    f_create_user,
    f_update_user,
    f_create_time,
    f_update_time
FROM dip_data_operator_hub.t_metadata_api
WHERE NOT EXISTS (
    SELECT 1
    FROM adp.t_metadata_api
    WHERE adp.t_metadata_api.f_version = dip_data_operator_hub.t_metadata_api.f_version
);

-- 2. 迁移t_op_registry表数据
INSERT INTO adp.t_op_registry (
    f_op_id,
    f_name,
    f_metadata_version,
    f_metadata_type,
    f_status,
    f_operator_type,
    f_execution_mode,
    f_category,
    f_source,
    f_execute_control,
    f_extend_info,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time,
    f_is_deleted,
    f_is_internal,
    f_is_data_source
)
SELECT
    f_op_id,
    f_name,
    f_metadata_version,
    f_metadata_type,
    f_status,
    f_operator_type,
    f_execution_mode,
    f_category,
    f_source,
    f_execute_control,
    f_extend_info,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time,
    f_is_deleted,
    f_is_internal,
    f_is_data_source
FROM dip_data_operator_hub.t_op_registry
WHERE NOT EXISTS (
    SELECT 1
    FROM adp.t_op_registry
    WHERE adp.t_op_registry.f_op_id = dip_data_operator_hub.t_op_registry.f_op_id
    AND adp.t_op_registry.f_metadata_version = dip_data_operator_hub.t_op_registry.f_metadata_version
);

-- 3. 迁移t_toolbox表数据
INSERT INTO adp.t_toolbox (
    f_box_id,
    f_name,
    f_description,
    f_svc_url,
    f_status,
    f_is_internal,
    f_source,
    f_category,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time,
    f_release_user,
    f_release_time
)
SELECT
    f_box_id,
    f_name,
    f_description,
    f_svc_url,
    f_status,
    f_is_internal,
    f_source,
    f_category,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time,
    f_release_user,
    f_release_time
FROM dip_data_operator_hub.t_toolbox
WHERE NOT EXISTS (
    SELECT 1
    FROM adp.t_toolbox
    WHERE adp.t_toolbox.f_box_id = dip_data_operator_hub.t_toolbox.f_box_id
);

-- 4. 迁移t_tool表数据
INSERT INTO adp.t_tool (
    f_tool_id,
    f_box_id,
    f_name,
    f_description,
    f_source_type,
    f_source_id,
    f_status,
    f_use_count,
    f_use_rule,
    f_parameters,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time,
    f_extend_info,
    f_is_deleted
)
SELECT
    f_tool_id,
    f_box_id,
    f_name,
    f_description,
    f_source_type,
    f_source_id,
    f_status,
    f_use_count,
    f_use_rule,
    f_parameters,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time,
    f_extend_info,
    f_is_deleted
FROM dip_data_operator_hub.t_tool
WHERE NOT EXISTS (
    SELECT 1
    FROM adp.t_tool
    WHERE adp.t_tool.f_tool_id = dip_data_operator_hub.t_tool.f_tool_id
);

-- 5. 迁移t_mcp_server_config表数据
INSERT INTO adp.t_mcp_server_config (
    f_mcp_id,
    f_creation_type,
    f_name,
    f_description,
    f_mode,
    f_url,
    f_headers,
    f_command,
    f_env,
    f_args,
    f_status,
    f_is_internal,
    f_source,
    f_category,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time,
    f_version
)
SELECT
    f_mcp_id,
    f_creation_type,
    f_name,
    f_description,
    f_mode,
    f_url,
    f_headers,
    f_command,
    f_env,
    f_args,
    f_status,
    f_is_internal,
    f_source,
    f_category,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time,
    f_version
FROM dip_data_operator_hub.t_mcp_server_config
WHERE NOT EXISTS (
    SELECT 1
    FROM adp.t_mcp_server_config
    WHERE adp.t_mcp_server_config.f_mcp_id = dip_data_operator_hub.t_mcp_server_config.f_mcp_id
);

-- 6. 迁移t_mcp_tool表数据
INSERT INTO adp.t_mcp_tool (
    f_mcp_tool_id,
    f_mcp_id,
    f_mcp_version,
    f_box_id,
    f_box_name,
    f_tool_id,
    f_name,
    f_description,
    f_use_rule,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time
)
SELECT
    f_mcp_tool_id,
    f_mcp_id,
    f_mcp_version,
    f_box_id,
    f_box_name,
    f_tool_id,
    f_name,
    f_description,
    f_use_rule,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time
FROM dip_data_operator_hub.t_mcp_tool
WHERE NOT EXISTS (
    SELECT 1
    FROM adp.t_mcp_tool
    WHERE adp.t_mcp_tool.f_mcp_tool_id = dip_data_operator_hub.t_mcp_tool.f_mcp_tool_id
);

-- 7. 迁移t_mcp_server_release表数据
INSERT INTO adp.t_mcp_server_release (
    f_mcp_id,
    f_creation_type,
    f_name,
    f_description,
    f_mode,
    f_url,
    f_headers,
    f_command,
    f_env,
    f_args,
    f_status,
    f_is_internal,
    f_source,
    f_category,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time,
    f_version,
    f_release_desc,
    f_release_user,
    f_release_time
)
SELECT
    f_mcp_id,
    f_creation_type,
    f_name,
    f_description,
    f_mode,
    f_url,
    f_headers,
    f_command,
    f_env,
    f_args,
    f_status,
    f_is_internal,
    f_source,
    f_category,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time,
    f_version,
    f_release_desc,
    f_release_user,
    f_release_time
FROM dip_data_operator_hub.t_mcp_server_release
WHERE NOT EXISTS (
    SELECT 1
    FROM adp.t_mcp_server_release
    WHERE adp.t_mcp_server_release.f_mcp_id = dip_data_operator_hub.t_mcp_server_release.f_mcp_id
    AND adp.t_mcp_server_release.f_version = dip_data_operator_hub.t_mcp_server_release.f_version
);

-- 8. 迁移t_mcp_server_release_history表数据
INSERT INTO adp.t_mcp_server_release_history (
    f_mcp_id,
    f_mcp_release,
    f_version,
    f_release_desc,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time
)
SELECT
    f_mcp_id,
    f_mcp_release,
    f_version,
    f_release_desc,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time
FROM dip_data_operator_hub.t_mcp_server_release_history
WHERE NOT EXISTS (
    SELECT 1
    FROM adp.t_mcp_server_release_history
    WHERE adp.t_mcp_server_release_history.f_mcp_id = dip_data_operator_hub.t_mcp_server_release_history.f_mcp_id
    AND adp.t_mcp_server_release_history.f_version = dip_data_operator_hub.t_mcp_server_release_history.f_version
);

-- 9. 迁移t_internal_component_config表数据
INSERT INTO adp.t_internal_component_config (
    f_component_type,
    f_component_id,
    f_config_version,
    f_config_source,
    f_protected_flag
)
SELECT
    f_component_type,
    f_component_id,
    f_config_version,
    f_config_source,
    f_protected_flag
FROM dip_data_operator_hub.t_internal_component_config
WHERE NOT EXISTS (
    SELECT 1
    FROM adp.t_internal_component_config
    WHERE adp.t_internal_component_config.f_component_type = dip_data_operator_hub.t_internal_component_config.f_component_type
    AND adp.t_internal_component_config.f_component_id = dip_data_operator_hub.t_internal_component_config.f_component_id
);

-- 10. 迁移t_operator_release表数据
INSERT INTO adp.t_operator_release (
    f_op_id,
    f_name,
    f_metadata_version,
    f_metadata_type,
    f_status,
    f_operator_type,
    f_execution_mode,
    f_category,
    f_source,
    f_execute_control,
    f_extend_info,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time,
    f_tag,
    f_release_user,
    f_release_time,
    f_is_internal,
    f_is_data_source
)
SELECT
    f_op_id,
    f_name,
    f_metadata_version,
    f_metadata_type,
    f_status,
    f_operator_type,
    f_execution_mode,
    f_category,
    f_source,
    f_execute_control,
    f_extend_info,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time,
    f_tag,
    f_release_user,
    f_release_time,
    f_is_internal,
    f_is_data_source
FROM dip_data_operator_hub.t_operator_release
WHERE NOT EXISTS (
    SELECT 1
    FROM adp.t_operator_release
    WHERE adp.t_operator_release.f_op_id = dip_data_operator_hub.t_operator_release.f_op_id
    AND adp.t_operator_release.f_tag = dip_data_operator_hub.t_operator_release.f_tag
);

-- 11. 迁移t_operator_release_history表数据
INSERT INTO adp.t_operator_release_history (
    f_op_id,
    f_op_release,
    f_metadata_version,
    f_metadata_type,
    f_tag,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time
)
SELECT
    f_op_id,
    f_op_release,
    f_metadata_version,
    f_metadata_type,
    f_tag,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time
FROM dip_data_operator_hub.t_operator_release_history
WHERE NOT EXISTS (
    SELECT 1
    FROM adp.t_operator_release_history
    WHERE adp.t_operator_release_history.f_op_id = dip_data_operator_hub.t_operator_release_history.f_op_id
    AND adp.t_operator_release_history.f_tag = dip_data_operator_hub.t_operator_release_history.f_tag
);

-- 12. 迁移t_category表数据
INSERT INTO adp.t_category (
    f_category_id,
    f_category_name,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time
)
SELECT
    f_category_id,
    f_category_name,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time
FROM dip_data_operator_hub.t_category
WHERE NOT EXISTS (
    SELECT 1
    FROM adp.t_category
    WHERE adp.t_category.f_category_id = dip_data_operator_hub.t_category.f_category_id
);

-- 13. 迁移t_outbox_message表数据
INSERT INTO adp.t_outbox_message (
    f_event_id,
    f_event_type,
    f_topic,
    f_payload,
    f_status,
    f_created_at,
    f_updated_at,
    f_next_retry_at,
    f_retry_count
)
SELECT
    f_event_id,
    f_event_type,
    f_topic,
    f_payload,
    f_status,
    f_created_at,
    f_updated_at,
    f_next_retry_at,
    f_retry_count
FROM dip_data_operator_hub.t_outbox_message
WHERE NOT EXISTS (
    SELECT 1
    FROM adp.t_outbox_message
    WHERE adp.t_outbox_message.f_event_id = dip_data_operator_hub.t_outbox_message.f_event_id
);
