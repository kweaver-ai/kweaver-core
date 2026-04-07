SET SEARCH_PATH TO adp;

CREATE TABLE IF NOT EXISTS `t_flow_storage` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_oss_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'OssGateway存储ID',
 `f_object_key` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '对象存储key',
 `f_name` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '原始文件名',
 `f_content_type` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'MIME类型',
 `f_size` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '文件大小',
 `f_etag` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '文件etag/hash',
 `f_status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1正常 2待删除 3已删除',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
 `f_deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT '删除时间 0表示未删除',
  PRIMARY KEY (`f_id`)
);
CREATE UNIQUE INDEX IF NOT EXISTS `uk_flow_storage_oss_id_object_key` ON `t_flow_storage` (`f_oss_id`, `f_object_key`);
CREATE INDEX IF NOT EXISTS `idx_flow_storage_status` ON `t_flow_storage` (`f_status`);
CREATE INDEX IF NOT EXISTS `idx_flow_storage_created_at` ON `t_flow_storage` (`f_created_at`);

CREATE TABLE IF NOT EXISTS `t_flow_file` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID，对应 dfs://<id>',
 `f_dag_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '流程定义ID',
 `f_dag_instance_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '流程实例ID',
 `f_storage_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '存储文件ID，未落OSS时为0',
 `f_status` TINYINT NOT NULL DEFAULT 1 COMMENT '业务状态 1待就绪 2就绪 3失效 4已过期',
 `f_name` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '文件名',
 `f_expires_at` BIGINT NOT NULL DEFAULT 0 COMMENT '过期时间 0表示不过期',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`f_id`)
);
CREATE INDEX IF NOT EXISTS `idx_flow_file_dag_id` ON `t_flow_file` (`f_dag_id`);
CREATE INDEX IF NOT EXISTS `idx_flow_file_dag_instance_id` ON `t_flow_file` (`f_dag_instance_id`);
CREATE INDEX IF NOT EXISTS `idx_flow_file_storage_id` ON `t_flow_file` (`f_storage_id`);
CREATE INDEX IF NOT EXISTS `idx_flow_file_status` ON `t_flow_file` (`f_status`);
CREATE INDEX IF NOT EXISTS `idx_flow_file_expires_at` ON `t_flow_file` (`f_expires_at`);

CREATE TABLE IF NOT EXISTS `t_flow_file_download_job` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_file_id` BIGINT UNSIGNED NOT NULL COMMENT '关联flow_file ID',
 `f_status` TINYINT NOT NULL DEFAULT 1 COMMENT '任务状态 1待执行 2执行中 3成功 4失败 5取消',
 `f_retry_count` INT NOT NULL DEFAULT 0 COMMENT '已重试次数',
 `f_max_retry` INT NOT NULL DEFAULT 3 COMMENT '最大重试次数',
 `f_next_retry_at` BIGINT NOT NULL DEFAULT 0 COMMENT '下次重试时间',
 `f_error_code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '错误码',
 `f_error_message` VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '错误信息',
 `f_download_url` VARCHAR(2048) NOT NULL DEFAULT '' COMMENT '源文件URL',
 `f_started_at` BIGINT NOT NULL DEFAULT 0 COMMENT '开始时间',
 `f_finished_at` BIGINT NOT NULL DEFAULT 0 COMMENT '结束时间',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`f_id`)
);
CREATE UNIQUE INDEX IF NOT EXISTS `uk_flow_file_download_job_file_id` ON `t_flow_file_download_job` (`f_file_id`);
CREATE INDEX IF NOT EXISTS `idx_flow_file_download_job_status_retry` ON `t_flow_file_download_job` (`f_status`, `f_next_retry_at`);

CREATE TABLE IF NOT EXISTS `t_flow_task_resume` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_task_instance_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '被阻塞的任务实例ID',
 `f_dag_instance_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '所属流程实例ID',
 `f_resource_type` VARCHAR(32) NOT NULL DEFAULT 'file' COMMENT '资源类型',
 `f_resource_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '资源ID，对文件场景即flow_file ID',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`f_id`)
);
CREATE UNIQUE INDEX IF NOT EXISTS `uk_flow_task_resume_task_instance_id` ON `t_flow_task_resume` (`f_task_instance_id`);
CREATE INDEX IF NOT EXISTS `idx_flow_task_resume_resource` ON `t_flow_task_resume` (`f_resource_type`, `f_resource_id`);