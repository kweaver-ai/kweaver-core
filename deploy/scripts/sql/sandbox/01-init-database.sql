-- ================================================================
-- Sandbox Control Plane Database Schema for MariaDB
-- Version: 0.2.0
-- Database: adp
--
-- 数据表命名规范:
-- - 表名: t_{module}_{entity} (小写 + 下划线)
-- - 字段名: f_{field_name} (小写 + 下划线)
-- - 时间戳: BIGINT (毫秒级时间戳)
-- - 索引名: t_{table}_idx_{field} / t_{table}_uk_{field}
--
-- 表说明:
-- - t_sandbox_session: 沙箱会话管理表
-- - t_sandbox_execution: 代码执行记录表
-- - t_sandbox_template: 沙箱模板定义表
-- - t_sandbox_runtime_node: 运行时节点注册表
-- ================================================================

USE adp;

-- ================================================================
-- Table: t_sandbox_template
-- ================================================================
-- 沙箱模板定义表（基础表，被 session 引用，先创建）
CREATE TABLE IF NOT EXISTS `t_sandbox_template` (
  `f_id` varchar(40) NOT NULL,
  `f_name` varchar(128) NOT NULL,
  `f_description` varchar(500) NOT NULL,
  `f_image_url` varchar(512) NOT NULL,
  `f_base_image` varchar(256) NOT NULL,
  `f_runtime_type` varchar(30) NOT NULL,
  `f_default_cpu_cores` decimal(3,1) NOT NULL,
  `f_default_memory_mb` int(11) NOT NULL,
  `f_default_disk_mb` int(11) NOT NULL,
  `f_default_timeout_sec` int(11) NOT NULL,
  `f_pre_installed_packages` text NOT NULL,
  `f_default_env_vars` text NOT NULL,
  `f_security_context` text NOT NULL,
  `f_is_active` int(11) NOT NULL,
  `f_created_at` bigint(20) NOT NULL,
  `f_created_by` varchar(40) NOT NULL,
  `f_updated_at` bigint(20) NOT NULL,
  `f_updated_by` varchar(40) NOT NULL,
  `f_deleted_at` bigint(20) NOT NULL,
  `f_deleted_by` varchar(36) NOT NULL,
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `t_sandbox_template_uk_name_deleted_at` (`f_name`,`f_deleted_at`),
  KEY `t_sandbox_template_idx_runtime_type` (`f_runtime_type`),
  KEY `t_sandbox_template_idx_is_active` (`f_is_active`),
  KEY `t_sandbox_template_idx_created_at` (`f_created_at`),
  KEY `t_sandbox_template_idx_deleted_at` (`f_deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ================================================================
-- Table: t_sandbox_runtime_node
-- ================================================================
-- 运行时节点注册表
CREATE TABLE IF NOT EXISTS `t_sandbox_runtime_node` (
  `f_node_id` varchar(40) NOT NULL,
  `f_hostname` varchar(128) NOT NULL,
  `f_runtime_type` varchar(20) NOT NULL,
  `f_ip_address` varchar(45) NOT NULL,
  `f_api_endpoint` varchar(512) NOT NULL,
  `f_status` varchar(20) NOT NULL,
  `f_total_cpu_cores` decimal(5,1) NOT NULL,
  `f_total_memory_mb` int(11) NOT NULL,
  `f_allocated_cpu_cores` decimal(5,1) NOT NULL,
  `f_allocated_memory_mb` int(11) NOT NULL,
  `f_running_containers` int(11) NOT NULL,
  `f_max_containers` int(11) NOT NULL,
  `f_cached_images` text NOT NULL,
  `f_labels` text NOT NULL,
  `f_last_heartbeat_at` bigint(20) NOT NULL,
  `f_created_at` bigint(20) NOT NULL,
  `f_created_by` varchar(40) NOT NULL,
  `f_updated_at` bigint(20) NOT NULL,
  `f_updated_by` varchar(40) NOT NULL,
  `f_deleted_at` bigint(20) NOT NULL,
  `f_deleted_by` varchar(36) NOT NULL,
  PRIMARY KEY (`f_node_id`),
  UNIQUE KEY `f_hostname` (`f_hostname`),
  UNIQUE KEY `t_sandbox_runtime_node_uk_hostname_deleted_at` (`f_hostname`,`f_deleted_at`),
  KEY `t_sandbox_runtime_node_idx_status` (`f_status`),
  KEY `t_sandbox_runtime_node_idx_runtime_type` (`f_runtime_type`),
  KEY `t_sandbox_runtime_node_idx_created_at` (`f_created_at`),
  KEY `t_sandbox_runtime_node_idx_deleted_at` (`f_deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ================================================================
-- Table: t_sandbox_session
-- ================================================================
-- 沙箱会话管理表（含依赖安装支持）
CREATE TABLE IF NOT EXISTS `t_sandbox_session` (
  `f_id` varchar(255) NOT NULL,
  `f_template_id` varchar(40) NOT NULL,
  `f_status` varchar(20) NOT NULL,
  `f_runtime_type` varchar(20) NOT NULL,
  `f_runtime_node` varchar(128) NOT NULL,
  `f_container_id` varchar(128) NOT NULL,
  `f_pod_name` varchar(128) NOT NULL,
  `f_workspace_path` varchar(256) NOT NULL,
  `f_resources_cpu` varchar(16) NOT NULL,
  `f_resources_memory` varchar(16) NOT NULL,
  `f_resources_disk` varchar(16) NOT NULL,
  `f_env_vars` text NOT NULL,
  `f_timeout` int(11) NOT NULL,
  `f_last_activity_at` bigint(20) NOT NULL,
  `f_completed_at` bigint(20) NOT NULL,
  `f_requested_dependencies` text NOT NULL,
  `f_installed_dependencies` text NOT NULL,
  `f_dependency_install_status` varchar(20) NOT NULL,
  `f_dependency_install_error` text NOT NULL,
  `f_dependency_install_started_at` bigint(20) NOT NULL,
  `f_dependency_install_completed_at` bigint(20) NOT NULL,
  `f_created_at` bigint(20) NOT NULL,
  `f_created_by` varchar(40) NOT NULL,
  `f_updated_at` bigint(20) NOT NULL,
  `f_updated_by` varchar(40) NOT NULL,
  `f_deleted_at` bigint(20) NOT NULL,
  `f_deleted_by` varchar(36) NOT NULL,
  PRIMARY KEY (`f_id`),
  KEY `t_sandbox_session_idx_last_activity_at` (`f_last_activity_at`),
  KEY `t_sandbox_session_idx_dependency_install_status` (`f_dependency_install_status`),
  KEY `t_sandbox_session_idx_created_by` (`f_created_by`),
  KEY `t_sandbox_session_idx_created_at` (`f_created_at`),
  KEY `t_sandbox_session_idx_template_id` (`f_template_id`),
  KEY `t_sandbox_session_idx_deleted_at` (`f_deleted_at`),
  KEY `t_sandbox_session_idx_status` (`f_status`),
  KEY `t_sandbox_session_idx_runtime_node` (`f_runtime_node`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ================================================================
-- Table: t_sandbox_execution
-- ================================================================
-- 代码执行记录表
CREATE TABLE IF NOT EXISTS `t_sandbox_execution` (
  `f_id` varchar(40) NOT NULL,
  `f_session_id` varchar(255) NOT NULL,
  `f_status` varchar(20) NOT NULL,
  `f_code` text NOT NULL,
  `f_language` varchar(32) NOT NULL,
  `f_entrypoint` varchar(255) NOT NULL,
  `f_event_data` text NOT NULL,
  `f_timeout_sec` int(11) NOT NULL,
  `f_return_value` text NOT NULL,
  `f_stdout` text NOT NULL,
  `f_stderr` text NOT NULL,
  `f_exit_code` int(11) NOT NULL,
  `f_metrics` text NOT NULL,
  `f_error_message` text NOT NULL,
  `f_started_at` bigint(20) NOT NULL,
  `f_completed_at` bigint(20) NOT NULL,
  `f_created_at` bigint(20) NOT NULL,
  `f_created_by` varchar(40) NOT NULL,
  `f_updated_at` bigint(20) NOT NULL,
  `f_updated_by` varchar(40) NOT NULL,
  `f_deleted_at` bigint(20) NOT NULL,
  `f_deleted_by` varchar(36) NOT NULL,
  PRIMARY KEY (`f_id`),
  KEY `t_sandbox_execution_idx_deleted_at` (`f_deleted_at`),
  KEY `t_sandbox_execution_idx_created_by` (`f_created_by`),
  KEY `t_sandbox_execution_idx_session_id` (`f_session_id`),
  KEY `t_sandbox_execution_idx_status` (`f_status`),
  KEY `t_sandbox_execution_idx_created_at` (`f_created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
