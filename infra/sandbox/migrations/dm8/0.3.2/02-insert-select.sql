SET SCHEMA kweaver;

-- 迁移数据从 adp 到 kweaver schema
-- 注意：使用 WHERE NOT EXISTS 避免重复插入数据

-- 1. t_sandbox_template
INSERT INTO t_sandbox_template (
    f_id, f_name, f_description, f_image_url, f_base_image, f_runtime_type,
    f_default_cpu_cores, f_default_memory_mb, f_default_disk_mb, f_default_timeout_sec,
    f_pre_installed_packages, f_default_env_vars, f_security_context, f_is_active,
    f_created_at, f_created_by, f_updated_at, f_updated_by, f_deleted_at, f_deleted_by
)
SELECT
    f_id, f_name, f_description, f_image_url, f_base_image, f_runtime_type,
    f_default_cpu_cores, f_default_memory_mb, f_default_disk_mb, f_default_timeout_sec,
    f_pre_installed_packages, f_default_env_vars, f_security_context, f_is_active,
    f_created_at, f_created_by, f_updated_at, f_updated_by, f_deleted_at, f_deleted_by
FROM adp.t_sandbox_template src
WHERE NOT EXISTS (SELECT 1 FROM t_sandbox_template dest WHERE dest.f_id = src.f_id);

-- 2. t_sandbox_runtime_node
INSERT INTO t_sandbox_runtime_node (
    f_node_id, f_hostname, f_runtime_type, f_ip_address, f_api_endpoint, f_status,
    f_total_cpu_cores, f_total_memory_mb, f_allocated_cpu_cores, f_allocated_memory_mb,
    f_running_containers, f_max_containers, f_cached_images, f_labels, f_last_heartbeat_at,
    f_created_at, f_created_by, f_updated_at, f_updated_by, f_deleted_at, f_deleted_by
)
SELECT
    f_node_id, f_hostname, f_runtime_type, f_ip_address, f_api_endpoint, f_status,
    f_total_cpu_cores, f_total_memory_mb, f_allocated_cpu_cores, f_allocated_memory_mb,
    f_running_containers, f_max_containers, f_cached_images, f_labels, f_last_heartbeat_at,
    f_created_at, f_created_by, f_updated_at, f_updated_by, f_deleted_at, f_deleted_by
FROM adp.t_sandbox_runtime_node src
WHERE NOT EXISTS (SELECT 1 FROM t_sandbox_runtime_node dest WHERE dest.f_node_id = src.f_node_id);

-- 3. t_sandbox_session
INSERT INTO t_sandbox_session (
    f_id, f_template_id, f_status, f_runtime_type, f_runtime_node, f_container_id,
    f_pod_name, f_workspace_path, f_resources_cpu, f_resources_memory, f_resources_disk,
    f_env_vars, f_timeout, f_last_activity_at, f_completed_at, f_python_package_index_url,
    f_requested_dependencies, f_installed_dependencies, f_dependency_install_status,
    f_dependency_install_error, f_dependency_install_started_at, f_dependency_install_completed_at,
    f_created_at, f_created_by, f_updated_at, f_updated_by, f_deleted_at, f_deleted_by
)
SELECT
    f_id, f_template_id, f_status, f_runtime_type, f_runtime_node, f_container_id,
    f_pod_name, f_workspace_path, f_resources_cpu, f_resources_memory, f_resources_disk,
    f_env_vars, f_timeout, f_last_activity_at, f_completed_at, f_python_package_index_url,
    f_requested_dependencies, f_installed_dependencies, f_dependency_install_status,
    f_dependency_install_error, f_dependency_install_started_at, f_dependency_install_completed_at,
    f_created_at, f_created_by, f_updated_at, f_updated_by, f_deleted_at, f_deleted_by
FROM adp.t_sandbox_session src
WHERE NOT EXISTS (SELECT 1 FROM t_sandbox_session dest WHERE dest.f_id = src.f_id);

-- 4. t_sandbox_execution
INSERT INTO t_sandbox_execution (
    f_id, f_session_id, f_status, f_code, f_language, f_entrypoint, f_event_data,
    f_timeout_sec, f_return_value, f_stdout, f_stderr, f_exit_code, f_metrics, f_error_message,
    f_started_at, f_completed_at, f_created_at, f_created_by, f_updated_at, f_updated_by,
    f_deleted_at, f_deleted_by
)
SELECT
    f_id, f_session_id, f_status, f_code, f_language, f_entrypoint, f_event_data,
    f_timeout_sec, f_return_value, f_stdout, f_stderr, f_exit_code, f_metrics, f_error_message,
    f_started_at, f_completed_at, f_created_at, f_created_by, f_updated_at, f_updated_by,
    f_deleted_at, f_deleted_by
FROM adp.t_sandbox_execution src
WHERE NOT EXISTS (SELECT 1 FROM t_sandbox_execution dest WHERE dest.f_id = src.f_id);
