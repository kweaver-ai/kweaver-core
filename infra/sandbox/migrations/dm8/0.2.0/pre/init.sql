-- ================================================================
-- Sandbox Control Plane Database Schema for DM8 (达梦数据库)
-- Version: 0.2.0
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
-- - 触发器: 实现 updated_at 自动更新
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

-- Comments for t_sandbox_template
COMMENT ON TABLE t_sandbox_template IS '沙箱模板定义表';
COMMENT ON COLUMN t_sandbox_template.f_id IS '模板唯一标识符';
COMMENT ON COLUMN t_sandbox_template.f_name IS '模板名称';
COMMENT ON COLUMN t_sandbox_template.f_description IS '模板描述';
COMMENT ON COLUMN t_sandbox_template.f_image_url IS '容器镜像URL';
COMMENT ON COLUMN t_sandbox_template.f_base_image IS '基础镜像';
COMMENT ON COLUMN t_sandbox_template.f_runtime_type IS '运行时类型(python3.11,nodejs20,java17,go1.21)';
COMMENT ON COLUMN t_sandbox_template.f_default_cpu_cores IS '默认CPU核数';
COMMENT ON COLUMN t_sandbox_template.f_default_memory_mb IS '默认内存MB';
COMMENT ON COLUMN t_sandbox_template.f_default_disk_mb IS '默认磁盘MB';
COMMENT ON COLUMN t_sandbox_template.f_default_timeout_sec IS '默认超时秒数';
COMMENT ON COLUMN t_sandbox_template.f_pre_installed_packages IS '预装包列表JSON';
COMMENT ON COLUMN t_sandbox_template.f_default_env_vars IS '默认环境变量JSON';
COMMENT ON COLUMN t_sandbox_template.f_security_context IS '安全策略JSON';
COMMENT ON COLUMN t_sandbox_template.f_is_active IS '是否激活(0:否,1:是)';
COMMENT ON COLUMN t_sandbox_template.f_created_at IS '创建时间(毫秒时间戳)';
COMMENT ON COLUMN t_sandbox_template.f_created_by IS '创建人';
COMMENT ON COLUMN t_sandbox_template.f_updated_at IS '更新时间(毫秒时间戳)';
COMMENT ON COLUMN t_sandbox_template.f_updated_by IS '更新人';
COMMENT ON COLUMN t_sandbox_template.f_deleted_at IS '删除时间(毫秒时间戳,0:未删除)';
COMMENT ON COLUMN t_sandbox_template.f_deleted_by IS '删除人';

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

-- Comments for t_sandbox_runtime_node
COMMENT ON TABLE t_sandbox_runtime_node IS '运行时节点注册表';
COMMENT ON COLUMN t_sandbox_runtime_node.f_node_id IS '节点唯一标识符';
COMMENT ON COLUMN t_sandbox_runtime_node.f_hostname IS '主机名';
COMMENT ON COLUMN t_sandbox_runtime_node.f_runtime_type IS '运行时类型(docker,kubernetes)';
COMMENT ON COLUMN t_sandbox_runtime_node.f_ip_address IS 'IP地址(IPv4/IPv6)';
COMMENT ON COLUMN t_sandbox_runtime_node.f_api_endpoint IS 'API端点URL';
COMMENT ON COLUMN t_sandbox_runtime_node.f_status IS '节点状态(online,offline,draining,maintenance)';
COMMENT ON COLUMN t_sandbox_runtime_node.f_total_cpu_cores IS '总CPU核数';
COMMENT ON COLUMN t_sandbox_runtime_node.f_total_memory_mb IS '总内存MB';
COMMENT ON COLUMN t_sandbox_runtime_node.f_allocated_cpu_cores IS '已分配CPU核数';
COMMENT ON COLUMN t_sandbox_runtime_node.f_allocated_memory_mb IS '已分配内存MB';
COMMENT ON COLUMN t_sandbox_runtime_node.f_running_containers IS '运行容器数';
COMMENT ON COLUMN t_sandbox_runtime_node.f_max_containers IS '最大容器数';
COMMENT ON COLUMN t_sandbox_runtime_node.f_cached_images IS '缓存镜像列表JSON';
COMMENT ON COLUMN t_sandbox_runtime_node.f_labels IS '节点标签JSON';
COMMENT ON COLUMN t_sandbox_runtime_node.f_last_heartbeat_at IS '最后心跳时间(毫秒时间戳)';
COMMENT ON COLUMN t_sandbox_runtime_node.f_created_at IS '创建时间(毫秒时间戳)';
COMMENT ON COLUMN t_sandbox_runtime_node.f_created_by IS '创建人';
COMMENT ON COLUMN t_sandbox_runtime_node.f_updated_at IS '更新时间(毫秒时间戳)';
COMMENT ON COLUMN t_sandbox_runtime_node.f_updated_by IS '更新人';
COMMENT ON COLUMN t_sandbox_runtime_node.f_deleted_at IS '删除时间(毫秒时间戳,0:未删除)';
COMMENT ON COLUMN t_sandbox_runtime_node.f_deleted_by IS '删除人';

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

-- Comments for t_sandbox_session
COMMENT ON TABLE t_sandbox_session IS '沙箱会话管理表';
COMMENT ON COLUMN t_sandbox_session.f_id IS '会话唯一标识符';
COMMENT ON COLUMN t_sandbox_session.f_template_id IS '模板ID引用';
COMMENT ON COLUMN t_sandbox_session.f_status IS '会话状态(creating,running,completed,failed,timeout,terminated)';
COMMENT ON COLUMN t_sandbox_session.f_runtime_type IS '运行时类型(python3.11,nodejs20,java17,go1.21)';
COMMENT ON COLUMN t_sandbox_session.f_runtime_node IS '当前运行节点';
COMMENT ON COLUMN t_sandbox_session.f_container_id IS '容器ID';
COMMENT ON COLUMN t_sandbox_session.f_pod_name IS 'Pod名称';
COMMENT ON COLUMN t_sandbox_session.f_workspace_path IS '工作区路径(S3)';
COMMENT ON COLUMN t_sandbox_session.f_resources_cpu IS 'CPU分配(如:1,2)';
COMMENT ON COLUMN t_sandbox_session.f_resources_memory IS '内存分配(如:512Mi,1Gi)';
COMMENT ON COLUMN t_sandbox_session.f_resources_disk IS '磁盘分配(如:1Gi,10Gi)';
COMMENT ON COLUMN t_sandbox_session.f_env_vars IS '环境变量JSON';
COMMENT ON COLUMN t_sandbox_session.f_timeout IS '超时时间(秒)';
COMMENT ON COLUMN t_sandbox_session.f_last_activity_at IS '最后活动时间(毫秒时间戳)';
COMMENT ON COLUMN t_sandbox_session.f_completed_at IS '完成时间(毫秒时间戳,0:未完成)';
COMMENT ON COLUMN t_sandbox_session.f_requested_dependencies IS '请求的依赖包JSON';
COMMENT ON COLUMN t_sandbox_session.f_installed_dependencies IS '已安装的依赖包JSON';
COMMENT ON COLUMN t_sandbox_session.f_dependency_install_status IS '依赖安装状态(pending,installing,completed,failed)';
COMMENT ON COLUMN t_sandbox_session.f_dependency_install_error IS '依赖安装错误信息';
COMMENT ON COLUMN t_sandbox_session.f_dependency_install_started_at IS '依赖安装开始时间(毫秒时间戳)';
COMMENT ON COLUMN t_sandbox_session.f_dependency_install_completed_at IS '依赖安装完成时间(毫秒时间戳)';
COMMENT ON COLUMN t_sandbox_session.f_created_at IS '创建时间(毫秒时间戳)';
COMMENT ON COLUMN t_sandbox_session.f_created_by IS '创建人';
COMMENT ON COLUMN t_sandbox_session.f_updated_at IS '更新时间(毫秒时间戳)';
COMMENT ON COLUMN t_sandbox_session.f_updated_by IS '更新人';
COMMENT ON COLUMN t_sandbox_session.f_deleted_at IS '删除时间(毫秒时间戳,0:未删除)';
COMMENT ON COLUMN t_sandbox_session.f_deleted_by IS '删除人';

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

-- Comments for t_sandbox_execution
COMMENT ON TABLE t_sandbox_execution IS '代码执行记录表';
COMMENT ON COLUMN t_sandbox_execution.f_id IS '执行唯一标识符';
COMMENT ON COLUMN t_sandbox_execution.f_session_id IS '会话ID引用';
COMMENT ON COLUMN t_sandbox_execution.f_status IS '执行状态(pending,running,completed,failed,timeout,crashed)';
COMMENT ON COLUMN t_sandbox_execution.f_code IS '源代码';
COMMENT ON COLUMN t_sandbox_execution.f_language IS '编程语言';
COMMENT ON COLUMN t_sandbox_execution.f_entrypoint IS '入口函数';
COMMENT ON COLUMN t_sandbox_execution.f_event_data IS '事件数据JSON';
COMMENT ON COLUMN t_sandbox_execution.f_timeout_sec IS '超时时间(秒)';
COMMENT ON COLUMN t_sandbox_execution.f_return_value IS '返回值JSON';
COMMENT ON COLUMN t_sandbox_execution.f_stdout IS '标准输出';
COMMENT ON COLUMN t_sandbox_execution.f_stderr IS '标准错误';
COMMENT ON COLUMN t_sandbox_execution.f_exit_code IS '退出码';
COMMENT ON COLUMN t_sandbox_execution.f_metrics IS '性能指标JSON';
COMMENT ON COLUMN t_sandbox_execution.f_error_message IS '错误信息';
COMMENT ON COLUMN t_sandbox_execution.f_started_at IS '执行开始时间(毫秒时间戳)';
COMMENT ON COLUMN t_sandbox_execution.f_completed_at IS '执行完成时间(毫秒时间戳)';
COMMENT ON COLUMN t_sandbox_execution.f_created_at IS '创建时间(毫秒时间戳)';
COMMENT ON COLUMN t_sandbox_execution.f_created_by IS '创建人';
COMMENT ON COLUMN t_sandbox_execution.f_updated_at IS '更新时间(毫秒时间戳)';
COMMENT ON COLUMN t_sandbox_execution.f_updated_by IS '更新人';
COMMENT ON COLUMN t_sandbox_execution.f_deleted_at IS '删除时间(毫秒时间戳,0:未删除)';
COMMENT ON COLUMN t_sandbox_execution.f_deleted_by IS '删除人';

-- Indexes for t_sandbox_execution
CREATE INDEX t_sandbox_execution_idx_session_id ON t_sandbox_execution(f_session_id);
CREATE INDEX t_sandbox_execution_idx_status ON t_sandbox_execution(f_status);
CREATE INDEX t_sandbox_execution_idx_created_at ON t_sandbox_execution(f_created_at);
CREATE INDEX t_sandbox_execution_idx_deleted_at ON t_sandbox_execution(f_deleted_at);
CREATE INDEX t_sandbox_execution_idx_created_by ON t_sandbox_execution(f_created_by);

-- ================================================================
-- Triggers for ON UPDATE behavior (updated_at 自动更新)
-- ================================================================

-- Trigger for t_sandbox_template.updated_at
CREATE OR REPLACE TRIGGER trg_t_sandbox_template_updated_at
BEFORE UPDATE ON t_sandbox_template
FOR EACH ROW
BEGIN
    NEW.f_updated_at := TIMESTAMPDIFF2(SECOND, '1970-01-01 00:00:00', SYSDATE) * 1000;
END;
/

-- Trigger for t_sandbox_runtime_node.updated_at
CREATE OR REPLACE TRIGGER trg_t_sandbox_runtime_node_updated_at
BEFORE UPDATE ON t_sandbox_runtime_node
FOR EACH ROW
BEGIN
    NEW.f_updated_at := TIMESTAMPDIFF2(SECOND, '1970-01-01 00:00:00', SYSDATE) * 1000;
END;
/

-- Trigger for t_sandbox_session.updated_at
CREATE OR REPLACE TRIGGER trg_t_sandbox_session_updated_at
BEFORE UPDATE ON t_sandbox_session
FOR EACH ROW
BEGIN
    NEW.f_updated_at := TIMESTAMPDIFF2(SECOND, '1970-01-01 00:00:00', SYSDATE) * 1000;
END;
/

-- Trigger for t_sandbox_execution.updated_at
CREATE OR REPLACE TRIGGER trg_t_sandbox_execution_updated_at
BEFORE UPDATE ON t_sandbox_execution
FOR EACH ROW
BEGIN
    NEW.f_updated_at := TIMESTAMPDIFF2(SECOND, '1970-01-01 00:00:00', SYSDATE) * 1000;
END;
/

COMMIT;
