use adp;

-- 记忆历史表

CREATE TABLE IF NOT EXISTS `t_data_agent_memory_history` (
    `f_id` varchar(40) NOT NULL COMMENT 'ID，唯一标识',
    `f_memory_id` varchar(40) NOT NULL COMMENT '记忆ID',
    `f_old_memory` text COMMENT '旧的记忆',
    `f_new_memory` text COMMENT '新的记忆',
    `f_event` varchar(40) NOT NULL COMMENT '对记忆的操作：\n- ADD\n- UPDATE\n- DELETE',
    `f_created_at` varchar(40) NOT NULL DEFAULT '' COMMENT '创建时间',
    `f_updated_at` varchar(40) NOT NULL DEFAULT '' COMMENT '更新时间',
    `f_actor_id` varchar(40) NOT NULL DEFAULT '' COMMENT '操作者',
    `f_role` varchar(40) NOT NULL DEFAULT '' COMMENT '操作者角色',
    `f_create_by` varchar(40) NOT NULL DEFAULT '' COMMENT '创建者',
    `f_update_by` varchar(40) NOT NULL DEFAULT '' COMMENT '最后修改者',
    `f_is_deleted` tinyint NOT NULL DEFAULT 0 COMMENT '是否删除：0 - 否 1 - 是',
    PRIMARY KEY (`f_id`),
    KEY `idx_memory_id` (`f_memory_id`)
) ENGINE = InnoDB COMMENT = '记忆历史数据表';
