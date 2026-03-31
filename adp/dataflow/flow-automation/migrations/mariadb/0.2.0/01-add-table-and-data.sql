use adp;
CREATE TABLE IF NOT EXISTS `t_cron_job`
(
    `f_key_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增长ID',
    `f_job_id` varchar(36) NOT NULL COMMENT '任务ID',
    `f_job_name` varchar(64) NOT NULL COMMENT '任务名称',
    `f_job_cron_time` varchar(32) NOT NULL COMMENT '时间计划，cron表达式',
    `f_job_type` tinyint(4) NOT NULL COMMENT '任务类型，参考数据字典',
    `f_job_context` varchar(10240) COMMENT '参考任务上下文数据结构',
    `f_tenant_id` varchar(36) COMMENT '任务来源ID',
    `f_enabled` tinyint(2) NOT NULL DEFAULT 1 COMMENT '启用/禁用标识',
    `f_remarks` varchar(256) COMMENT '备注',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_update_time` bigint(20) NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`f_key_id`),
    UNIQUE KEY `index_job_id`(`f_job_id`) USING BTREE,
    UNIQUE KEY `index_job_name`(`f_job_name`, `f_tenant_id`) USING BTREE,
    KEY `index_tenant_id`(`f_tenant_id`) USING BTREE,
    KEY `index_time`(`f_create_time`, `f_update_time`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '定时任务信息表';

CREATE TABLE IF NOT EXISTS `t_cron_job_status`
(
    `f_key_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增长ID',
    `f_execute_id` varchar(36) NOT NULL COMMENT '执行编号，流水号',
    `f_job_id` varchar(36) NOT NULL COMMENT '任务ID',
    `f_job_type` tinyint(4) NOT NULL COMMENT '任务类型',
    `f_job_name` varchar(64) NOT NULL COMMENT '任务名称',
    `f_job_status` tinyint(4) NOT NULL COMMENT '任务状态，参考数据字典',
    `f_begin_time` bigint(20) COMMENT '任务本次执行开始时间',
    `f_end_time` bigint(20) COMMENT '任务本次执行结束时间',
    `f_executor` varchar(1024) COMMENT '任务执行者',
    `f_execute_times` int COMMENT '任务执行次数',
    `f_ext_info` varchar(1024) COMMENT '扩展信息',
    PRIMARY KEY (`f_key_id`),
    UNIQUE KEY `index_execute_id`(`f_execute_id`) USING BTREE,
    KEY `index_job_id`(`f_job_id`) USING BTREE,
    KEY `index_job_status`(`f_job_status`) USING BTREE,
    KEY `index_time`(`f_begin_time`,`f_end_time`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '定时任务状态表';
