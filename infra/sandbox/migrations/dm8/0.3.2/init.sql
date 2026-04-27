-- ================================================================
-- Sandbox Control Plane Database Schema for DM8 (达梦数据库)
-- Version: 0.3.0
-- Database: adp
--
-- 数据表命名规范:
-- - 表名: t_{module}_{entity} (小写 + 下划线)
-- - 字段名: f_{field_name} (小写 + 下划线)
-- - 时间戳: BIGINT (毫秒级时间戳)
-- - 索引名: t_{table}_idx_{field} / t_{table}_uk_{field}
--
-- DM8 特性说明:
-- - CLUSTER PRIMARY KEY: 聚簇主键
-- - VARCHAR(N CHAR): 字符单位长度
-- - TEXT/CLOB: 大文本类型
--
-- 表说明:
-- - t_sandbox_session: 沙箱会话管理表
-- - t_sandbox_execution: 代码执行记录表
-- - t_sandbox_template: 沙箱模板定义表
-- - t_sandbox_runtime_node: 运行时节点注册表
-- ================================================================
SET SCHEMA kweaver;

-- ================================================================
-- Table: t_sandbox_template
-- ================================================================
-- 沙箱模板定义表（基础表，被 session 引用，先创建）
CREATE TABLE IF NOT EXISTS t_sandbox_template
(
    f_id                  VARCHAR(40 CHAR)  NOT NULL,
    f_name                VARCHAR(128 CHAR) NOT NULL,
    f_description         VARCHAR(500 CHAR) NOT NULL DEFAULT '',
    f_image_url           VARCHAR(512 CHAR) NOT NULL,
    f_base_image          VARCHAR(256 CHAR) NOT NULL DEFAULT '',
    f_runtime_type        VARCHAR(30 CHAR)  NOT NULL,
    f_default_cpu_cores   DECIMAL(3,1)     NOT NULL DEFAULT 0.5,
    f_default_memory_mb   INT              NOT NULL DEFAULT 512,
    f_default_disk_mb     INT              NOT NULL DEFAULT 1024,
    f_default_timeout_sec INT              NOT NULL DEFAULT 300,
    f_pre_installed_packages CLOB          NOT NULL,
    f_default_env_vars    CLOB             NOT NULL,
    f_security_context    CLOB             NOT NULL,
    f_is_active           TINYINT          NOT NULL DEFAULT 1,
    f_created_at          BIGINT           NOT NULL DEFAULT 0,
    f_created_by          VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_updated_at          BIGINT           NOT NULL DEFAULT 0,
    f_updated_by          VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_deleted_at          BIGINT           NOT NULL DEFAULT 0,
    f_deleted_by          VARCHAR(36 CHAR) NOT NULL DEFAULT '',
    CLUSTER PRIMARY KEY (f_id)
);


-- Indexes for t_sandbox_template
CREATE UNIQUE INDEX t_sandbox_template_uk_name_deleted_at ON t_sandbox_template(f_name, f_deleted_at);
CREATE INDEX t_sandbox_template_idx_runtime_type ON t_sandbox_template(f_runtime_type);
CREATE INDEX t_sandbox_template_idx_is_active ON t_sandbox_template(f_is_active);
CREATE INDEX t_sandbox_template_idx_created_at ON t_sandbox_template(f_created_at);
CREATE INDEX t_sandbox_template_idx_deleted_at ON t_sandbox_template(f_deleted_at);

-- ================================================================
-- Table: t_sandbox_runtime_node
-- ================================================================
-- 运行时节点注册表
CREATE TABLE IF NOT EXISTS t_sandbox_runtime_node
(
    f_node_id             VARCHAR(40 CHAR)  NOT NULL,
    f_hostname            VARCHAR(128 CHAR) NOT NULL,
    f_runtime_type        VARCHAR(20 CHAR)  NOT NULL,
    f_ip_address          VARCHAR(45 CHAR)  NOT NULL,
    f_api_endpoint        VARCHAR(512 CHAR) NOT NULL DEFAULT '',
    f_status              VARCHAR(20 CHAR)  NOT NULL DEFAULT 'online',
    f_total_cpu_cores     DECIMAL(5,1)     NOT NULL,
    f_total_memory_mb     INT              NOT NULL,
    f_allocated_cpu_cores DECIMAL(5,1)     NOT NULL DEFAULT 0.0,
    f_allocated_memory_mb INT              NOT NULL DEFAULT 0,
    f_running_containers  INT              NOT NULL DEFAULT 0,
    f_max_containers      INT              NOT NULL,
    f_cached_images       CLOB             NOT NULL,
    f_labels              CLOB             NOT NULL,
    f_last_heartbeat_at   BIGINT           NOT NULL DEFAULT 0,
    f_created_at          BIGINT           NOT NULL DEFAULT 0,
    f_created_by          VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_updated_at          BIGINT           NOT NULL DEFAULT 0,
    f_updated_by          VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_deleted_at          BIGINT           NOT NULL DEFAULT 0,
    f_deleted_by          VARCHAR(36 CHAR) NOT NULL DEFAULT '',
    CLUSTER PRIMARY KEY (f_node_id)
);


-- Indexes for t_sandbox_runtime_node
CREATE UNIQUE INDEX t_sandbox_runtime_node_uk_hostname_deleted_at ON t_sandbox_runtime_node(f_hostname, f_deleted_at);
CREATE INDEX t_sandbox_runtime_node_idx_status ON t_sandbox_runtime_node(f_status);
CREATE INDEX t_sandbox_runtime_node_idx_runtime_type ON t_sandbox_runtime_node(f_runtime_type);
CREATE INDEX t_sandbox_runtime_node_idx_created_at ON t_sandbox_runtime_node(f_created_at);
CREATE INDEX t_sandbox_runtime_node_idx_deleted_at ON t_sandbox_runtime_node(f_deleted_at);

-- ================================================================
-- Table: t_sandbox_session
-- ================================================================
-- 沙箱会话管理表（含依赖安装支持）
CREATE TABLE IF NOT EXISTS t_sandbox_session
(
    f_id                          VARCHAR(255 CHAR) NOT NULL,
    f_template_id                 VARCHAR(40 CHAR)  NOT NULL,
    f_status                      VARCHAR(20 CHAR)  NOT NULL DEFAULT 'creating',
    f_runtime_type                VARCHAR(20 CHAR)  NOT NULL,
    f_runtime_node                VARCHAR(128 CHAR) NOT NULL DEFAULT '',
    f_container_id                VARCHAR(128 CHAR) NOT NULL DEFAULT '',
    f_pod_name                    VARCHAR(128 CHAR) NOT NULL DEFAULT '',
    f_workspace_path              VARCHAR(256 CHAR) NOT NULL DEFAULT '',
    f_resources_cpu               VARCHAR(16 CHAR)  NOT NULL,
    f_resources_memory            VARCHAR(16 CHAR)  NOT NULL,
    f_resources_disk              VARCHAR(16 CHAR)  NOT NULL,
    f_env_vars                    CLOB             NOT NULL,
    f_timeout                     INT              NOT NULL DEFAULT 300,
    f_last_activity_at            BIGINT           NOT NULL DEFAULT 0,
    f_completed_at                BIGINT           NOT NULL DEFAULT 0,
    f_python_package_index_url    VARCHAR(512 CHAR) NOT NULL DEFAULT 'https://pypi.org/simple/',

    -- 依赖安装字段
    f_requested_dependencies      CLOB             NOT NULL,
    f_installed_dependencies      CLOB             NOT NULL,
    f_dependency_install_status   VARCHAR(20 CHAR) NOT NULL DEFAULT 'pending',
    f_dependency_install_error    CLOB             NOT NULL,
    f_dependency_install_started_at   BIGINT       NOT NULL DEFAULT 0,
    f_dependency_install_completed_at BIGINT       NOT NULL DEFAULT 0,

    -- 审计字段
    f_created_at                  BIGINT           NOT NULL DEFAULT 0,
    f_created_by                  VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_updated_at                  BIGINT           NOT NULL DEFAULT 0,
    f_updated_by                  VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_deleted_at                  BIGINT           NOT NULL DEFAULT 0,
    f_deleted_by                  VARCHAR(36 CHAR) NOT NULL DEFAULT '',
    CLUSTER PRIMARY KEY (f_id)
);


-- Indexes for t_sandbox_session
CREATE INDEX t_sandbox_session_idx_template_id ON t_sandbox_session(f_template_id);
CREATE INDEX t_sandbox_session_idx_status ON t_sandbox_session(f_status);
CREATE INDEX t_sandbox_session_idx_runtime_node ON t_sandbox_session(f_runtime_node);
CREATE INDEX t_sandbox_session_idx_last_activity_at ON t_sandbox_session(f_last_activity_at);
CREATE INDEX t_sandbox_session_idx_dependency_install_status ON t_sandbox_session(f_dependency_install_status);
CREATE INDEX t_sandbox_session_idx_created_at ON t_sandbox_session(f_created_at);
CREATE INDEX t_sandbox_session_idx_deleted_at ON t_sandbox_session(f_deleted_at);
CREATE INDEX t_sandbox_session_idx_created_by ON t_sandbox_session(f_created_by);

-- ================================================================
-- Table: t_sandbox_execution
-- ================================================================
-- 代码执行记录表
CREATE TABLE IF NOT EXISTS t_sandbox_execution
(
    f_id              VARCHAR(40 CHAR)  NOT NULL,
    f_session_id      VARCHAR(255 CHAR) NOT NULL,
    f_status          VARCHAR(20 CHAR)  NOT NULL DEFAULT 'pending',
    f_code            CLOB              NOT NULL,
    f_language        VARCHAR(32 CHAR)  NOT NULL,
    f_entrypoint      VARCHAR(255 CHAR) NOT NULL DEFAULT '',
    f_event_data      CLOB              NOT NULL,
    f_timeout_sec     INT               NOT NULL,
    f_return_value    CLOB              NOT NULL,
    f_stdout          CLOB              NOT NULL,
    f_stderr          CLOB              NOT NULL,
    f_exit_code       INT               NOT NULL DEFAULT 0,
    f_metrics         CLOB              NOT NULL,
    f_error_message   CLOB              NOT NULL,
    f_started_at      BIGINT            NOT NULL DEFAULT 0,
    f_completed_at    BIGINT            NOT NULL DEFAULT 0,

    -- 审计字段
    f_created_at      BIGINT            NOT NULL DEFAULT 0,
    f_created_by      VARCHAR(40 CHAR)  NOT NULL DEFAULT '',
    f_updated_at      BIGINT            NOT NULL DEFAULT 0,
    f_updated_by      VARCHAR(40 CHAR)  NOT NULL DEFAULT '',
    f_deleted_at      BIGINT            NOT NULL DEFAULT 0,
    f_deleted_by      VARCHAR(36 CHAR)  NOT NULL DEFAULT '',
    CLUSTER PRIMARY KEY (f_id)
);


-- Indexes for t_sandbox_execution
CREATE INDEX t_sandbox_execution_idx_session_id ON t_sandbox_execution(f_session_id);
CREATE INDEX t_sandbox_execution_idx_status ON t_sandbox_execution(f_status);
CREATE INDEX t_sandbox_execution_idx_created_at ON t_sandbox_execution(f_created_at);
CREATE INDEX t_sandbox_execution_idx_deleted_at ON t_sandbox_execution(f_deleted_at);
CREATE INDEX t_sandbox_execution_idx_created_by ON t_sandbox_execution(f_created_by);

