-- 设置目标schema为adp
SET SCHEMA adp;

-- 迁移t_metadata_api表数据
INSERT INTO "t_metadata_api" (
    "f_summary", "f_version", "f_svc_url", "f_description", "f_path", "f_method",
    "f_api_spec", "f_create_user", "f_update_user", "f_create_time", "f_update_time"
)
SELECT
    "f_summary", "f_version", "f_svc_url", "f_description", "f_path", "f_method",
    "f_api_spec", "f_create_user", "f_update_user", "f_create_time", "f_update_time"
FROM dip_data_operator_hub."t_metadata_api" src
WHERE NOT EXISTS (
    SELECT 1 FROM "t_metadata_api" dest
    WHERE dest."f_summary" = src."f_summary" AND dest."f_version" = src."f_version"
);

-- 迁移t_op_registry表数据
INSERT INTO "t_op_registry" (
    "f_op_id", "f_name", "f_metadata_version", "f_metadata_type", "f_status", "f_operator_type",
    "f_execution_mode", "f_category", "f_source", "f_execute_control", "f_extend_info",
    "f_create_user", "f_create_time", "f_update_user", "f_update_time", "f_is_deleted",
    "f_is_internal", "f_is_data_source"
)
SELECT
    "f_op_id", "f_name", "f_metadata_version", "f_metadata_type", "f_status", "f_operator_type",
    "f_execution_mode", "f_category", "f_source", "f_execute_control", "f_extend_info",
    "f_create_user", "f_create_time", "f_update_user", "f_update_time", "f_is_deleted",
    "f_is_internal", "f_is_data_source"
FROM dip_data_operator_hub."t_op_registry" src
WHERE NOT EXISTS (
    SELECT 1 FROM "t_op_registry" dest
    WHERE dest."f_op_id" = src."f_op_id" AND dest."f_metadata_version" = src."f_metadata_version"
);

-- 迁移t_toolbox表数据
INSERT INTO "t_toolbox" (
    "f_box_id", "f_name", "f_description", "f_svc_url", "f_status", "f_is_internal",
    "f_source", "f_category", "f_create_user", "f_create_time", "f_update_user",
    "f_update_time", "f_release_user", "f_release_time"
)
SELECT
    "f_box_id", "f_name", "f_description", "f_svc_url", "f_status", "f_is_internal",
    "f_source", "f_category", "f_create_user", "f_create_time", "f_update_user",
    "f_update_time", "f_release_user", "f_release_time"
FROM dip_data_operator_hub."t_toolbox" src
WHERE NOT EXISTS (
    SELECT 1 FROM "t_toolbox" dest
    WHERE dest."f_box_id" = src."f_box_id"
);

-- 迁移t_tool表数据
INSERT INTO "t_tool" (
    "f_tool_id", "f_box_id", "f_name", "f_description", "f_source_type", "f_source_id",
    "f_status", "f_use_count", "f_use_rule", "f_parameters", "f_create_user",
    "f_create_time", "f_update_user", "f_update_time", "f_extend_info", "f_is_deleted"
)
SELECT
    "f_tool_id", "f_box_id", "f_name", "f_description", "f_source_type", "f_source_id",
    "f_status", "f_use_count", "f_use_rule", "f_parameters", "f_create_user",
    "f_create_time", "f_update_user", "f_update_time", "f_extend_info", "f_is_deleted"
FROM dip_data_operator_hub."t_tool" src
WHERE NOT EXISTS (
    SELECT 1 FROM "t_tool" dest
    WHERE dest."f_tool_id" = src."f_tool_id"
);

-- 迁移t_mcp_server_config表数据
INSERT INTO "t_mcp_server_config" (
    "f_mcp_id", "f_name", "f_description", "f_mode", "f_url", "f_headers", "f_command",
    "f_env", "f_args", "f_status", "f_is_internal", "f_source", "f_category",
    "f_create_user", "f_create_time", "f_update_user", "f_update_time", "f_creation_type", "f_version"
)
SELECT
    "f_mcp_id", "f_name", "f_description", "f_mode", "f_url", "f_headers", "f_command",
    "f_env", "f_args", "f_status", "f_is_internal", "f_source", "f_category",
    "f_create_user", "f_create_time", "f_update_user", "f_update_time", "f_creation_type", "f_version"
FROM dip_data_operator_hub."t_mcp_server_config" src
WHERE NOT EXISTS (
    SELECT 1 FROM "t_mcp_server_config" dest
    WHERE dest."f_mcp_id" = src."f_mcp_id"
);

-- 迁移t_mcp_server_release表数据
INSERT INTO "t_mcp_server_release" (
    "f_mcp_id", "f_name", "f_description", "f_mode", "f_url", "f_headers", "f_command",
    "f_env", "f_args", "f_status", "f_is_internal", "f_source", "f_category",
    "f_create_user", "f_create_time", "f_update_user", "f_update_time", "f_version",
    "f_release_desc", "f_release_user", "f_release_time", "f_creation_type"
)
SELECT
    "f_mcp_id", "f_name", "f_description", "f_mode", "f_url", "f_headers", "f_command",
    "f_env", "f_args", "f_status", "f_is_internal", "f_source", "f_category",
    "f_create_user", "f_create_time", "f_update_user", "f_update_time", "f_version",
    "f_release_desc", "f_release_user", "f_release_time", "f_creation_type"
FROM dip_data_operator_hub."t_mcp_server_release" src
WHERE NOT EXISTS (
    SELECT 1 FROM "t_mcp_server_release" dest
    WHERE dest."f_mcp_id" = src."f_mcp_id" AND dest."f_version" = src."f_version"
);

-- 迁移t_mcp_server_release_history表数据
INSERT INTO "t_mcp_server_release_history" (
    "f_mcp_id", "f_mcp_release", "f_version", "f_release_desc", "f_create_user",
    "f_create_time", "f_update_user", "f_update_time"
)
SELECT
    "f_mcp_id", "f_mcp_release", "f_version", "f_release_desc", "f_create_user",
    "f_create_time", "f_update_user", "f_update_time"
FROM dip_data_operator_hub."t_mcp_server_release_history" src
WHERE NOT EXISTS (
    SELECT 1 FROM "t_mcp_server_release_history" dest
    WHERE dest."f_mcp_id" = src."f_mcp_id" AND dest."f_version" = src."f_version"
);

-- 迁移t_internal_component_config表数据
INSERT INTO "t_internal_component_config" (
    "f_component_type", "f_component_id", "f_config_version", "f_config_source", "f_protected_flag"
)
SELECT
    "f_component_type", "f_component_id", "f_config_version", "f_config_source", "f_protected_flag"
FROM dip_data_operator_hub."t_internal_component_config" src
WHERE NOT EXISTS (
    SELECT 1 FROM "t_internal_component_config" dest
    WHERE dest."f_component_type" = src."f_component_type" AND dest."f_component_id" = src."f_component_id"
);

-- 迁移t_operator_release表数据
INSERT INTO "t_operator_release" (
    "f_op_id", "f_name", "f_metadata_version", "f_metadata_type", "f_status", "f_operator_type",
    "f_execution_mode", "f_category", "f_source", "f_execute_control", "f_extend_info",
    "f_create_user", "f_create_time", "f_update_user", "f_update_time", "f_tag",
    "f_release_user", "f_release_time", "f_is_internal", "f_is_data_source"
)
SELECT
    "f_op_id", "f_name", "f_metadata_version", "f_metadata_type", "f_status", "f_operator_type",
    "f_execution_mode", "f_category", "f_source", "f_execute_control", "f_extend_info",
    "f_create_user", "f_create_time", "f_update_user", "f_update_time", "f_tag",
    "f_release_user", "f_release_time", "f_is_internal", "f_is_data_source"
FROM dip_data_operator_hub."t_operator_release" src
WHERE NOT EXISTS (
    SELECT 1 FROM "t_operator_release" dest
    WHERE dest."f_op_id" = src."f_op_id" AND dest."f_tag" = src."f_tag"
);

-- 迁移t_operator_release_history表数据
INSERT INTO "t_operator_release_history" (
    "f_op_id", "f_op_release", "f_metadata_version", "f_metadata_type", "f_tag",
    "f_create_user", "f_create_time", "f_update_user", "f_update_time"
)
SELECT
    "f_op_id", "f_op_release", "f_metadata_version", "f_metadata_type", "f_tag",
    "f_create_user", "f_create_time", "f_update_user", "f_update_time"
FROM dip_data_operator_hub."t_operator_release_history" src
WHERE NOT EXISTS (
    SELECT 1 FROM "t_operator_release_history" dest
    WHERE dest."f_op_id" = src."f_op_id" AND dest."f_tag" = src."f_tag"
);

-- 迁移t_category表数据
INSERT INTO "t_category" (
    "f_category_id", "f_category_name", "f_create_user", "f_create_time",
    "f_update_user", "f_update_time"
)
SELECT
    "f_category_id", "f_category_name", "f_create_user", "f_create_time",
    "f_update_user", "f_update_time"
FROM dip_data_operator_hub."t_category" src
WHERE NOT EXISTS (
    SELECT 1 FROM "t_category" dest
    WHERE dest."f_category_id" = src."f_category_id"
);

-- 迁移t_outbox_message表数据
INSERT INTO "t_outbox_message" (
    "f_event_id", "f_event_type", "f_topic", "f_payload", "f_status",
    "f_created_at", "f_updated_at", "f_next_retry_at", "f_retry_count"
)
SELECT
    "f_event_id", "f_event_type", "f_topic", "f_payload", "f_status",
    "f_created_at", "f_updated_at", "f_next_retry_at", "f_retry_count"
FROM dip_data_operator_hub."t_outbox_message" src
WHERE NOT EXISTS (
    SELECT 1 FROM "t_outbox_message" dest
    WHERE dest."f_event_id" = src."f_event_id"
);

-- 迁移t_mcp_tool表数据
INSERT INTO "t_mcp_tool" (
    "f_mcp_tool_id", "f_mcp_id", "f_mcp_version", "f_box_id", "f_box_name",
    "f_tool_id", "f_name", "f_description", "f_use_rule", "f_create_user",
    "f_create_time", "f_update_user", "f_update_time"
)
SELECT
    "f_mcp_tool_id", "f_mcp_id", "f_mcp_version", "f_box_id", "f_box_name",
    "f_tool_id", "f_name", "f_description", "f_use_rule", "f_create_user",
    "f_create_time", "f_update_user", "f_update_time"
FROM dip_data_operator_hub."t_mcp_tool" src
WHERE NOT EXISTS (
    SELECT 1 FROM "t_mcp_tool" dest
    WHERE dest."f_mcp_tool_id" = src."f_mcp_tool_id"
);