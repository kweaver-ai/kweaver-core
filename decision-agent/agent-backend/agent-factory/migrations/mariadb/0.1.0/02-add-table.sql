USE adp;

CREATE TABLE IF NOT EXISTS t_data_agent_temporary_area (
  f_temp_area_id varchar(100) NOT NULL COMMENT '临时区ID',
  f_source_id varchar(40) DEFAULT NULL COMMENT '源文件ID',
  f_conversation_id varchar(40) NOT NULL DEFAULT '' COMMENT '对话ID',
  f_id bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  f_created_at bigint(20) DEFAULT 0 COMMENT '创建时间',
  f_source_type varchar(40) DEFAULT NULL COMMENT '源文件类型',
  f_user_id varchar(40) NOT NULL COMMENT '用户ID',
  PRIMARY KEY (f_id),
  KEY idx_temp_area_id (f_temp_area_id) USING BTREE,
  KEY idx_source_id (f_source_id) USING BTREE,
  KEY idx_conversation_id (f_conversation_id) USING BTREE,
  KEY idx_created_at (f_created_at) USING BTREE
) ENGINE=InnoDB  COMMENT '临时区表';

INSERT INTO adp.t_data_agent_temporary_area (
    f_temp_area_id,
    f_source_id,
    f_conversation_id,
    f_id,
    f_created_at,
    f_source_type,
    f_user_id
)
SELECT
    f_temp_area_id,
    f_source_id,
    f_conversation_id,
    f_id,
    f_created_at,
    f_source_type,
    f_user_id
FROM dip_data_agent.t_temporary_area
WHERE NOT EXISTS (
    SELECT 1 FROM adp.t_data_agent_temporary_area t
    WHERE t.f_id = dip_data_agent.t_temporary_area.f_id
);