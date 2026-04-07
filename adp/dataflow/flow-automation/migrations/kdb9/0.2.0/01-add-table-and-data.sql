
SET SEARCH_PATH TO adp;


CREATE TABLE IF NOT EXISTS `t_cron_job`
  (
  `f_key_id` BIGSERIAL NOT NULL COMMENT '自增长ID',
  `f_job_id` VARCHAR(36) NOT NULL COMMENT '任务ID',
  `f_job_name` VARCHAR(64) NOT NULL COMMENT '任务名称',
  `f_job_cron_time` VARCHAR(32) NOT NULL COMMENT '时间计划，cron表达式',
  `f_job_type` TINYINT(4) NOT NULL COMMENT '任务类型，参考数据字典',
  `f_job_context` VARCHAR(10240) COMMENT '参考任务上下文数据结构',
  `f_tenant_id` VARCHAR(36) COMMENT '任务来源ID',
  `f_enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '启用/禁用标识',
  `f_remarks` VARCHAR(256) COMMENT '备注',
  `f_create_time` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_update_time` BIGINT(20) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`f_key_id`),
  UNIQUE KEY `idx_t_cron_job_index_job_id` (`f_job_id`),
  UNIQUE KEY `idx_t_cron_job_index_job_name` (`f_job_name`, `f_tenant_id`)
);

CREATE INDEX IF NOT EXISTS `idx_t_cron_job_index_tenant_id` ON `t_cron_job` (`f_tenant_id`);
CREATE INDEX IF NOT EXISTS `idx_t_cron_job_index_time` ON `t_cron_job` (`f_create_time`, `f_update_time`);


CREATE TABLE IF NOT EXISTS `t_cron_job_status`
  (
  `f_key_id` BIGSERIAL NOT NULL COMMENT '自增长ID',
  `f_execute_id` VARCHAR(36) NOT NULL COMMENT '执行编号，流水号',
  `f_job_id` VARCHAR(36) NOT NULL COMMENT '任务ID',
  `f_job_type` TINYINT(4) NOT NULL COMMENT '任务类型',
  `f_job_name` VARCHAR(64) NOT NULL COMMENT '任务名称',
  `f_job_status` TINYINT(4) NOT NULL COMMENT '任务状态，参考数据字典',
  `f_begin_time` BIGINT(20) COMMENT '任务本次执行开始时间',
  `f_end_time` BIGINT(20) COMMENT '任务本次执行结束时间',
  `f_executor` VARCHAR(1024) COMMENT '任务执行者',
  `f_execute_times` INT COMMENT '任务执行次数',
  `f_ext_info` VARCHAR(1024) COMMENT '扩展信息',
  PRIMARY KEY (`f_key_id`),
  UNIQUE KEY `idx_t_cron_job_status_index_execute_id` (`f_execute_id`)
);

CREATE INDEX IF NOT EXISTS `idx_t_cron_job_status_index_job_id` ON `t_cron_job_status` (`f_job_id`);
CREATE INDEX IF NOT EXISTS `idx_t_cron_job_status_index_job_status` ON `t_cron_job_status` (`f_job_status`);
CREATE INDEX IF NOT EXISTS `idx_t_cron_job_status_index_time` ON `t_cron_job_status` (`f_begin_time`,`f_end_time`);

