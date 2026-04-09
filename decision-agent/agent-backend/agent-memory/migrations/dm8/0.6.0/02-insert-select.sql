SET SCHEMA kweaver;
-- 从 adp 数据库的 t_data_agent_memory_history 表复制数据到 kweaver 数据库的同名表
-- 注意：确保源表存在且有数据
-- 使用 WHERE NOT EXISTS 判重逻辑，避免插入重复记录（基于主键 f_id 判断）

-- 可选：如果目标表可能已有数据，先清空（根据实际情况选择是否启用）
-- DELETE FROM adp."t_data_agent_memory_history";

INSERT INTO kweaver."t_data_agent_memory_history" (
  "f_id",
  "f_memory_id",
  "f_old_memory",
  "f_new_memory",
  "f_event",
  "f_created_at",
  "f_updated_at",
  "f_actor_id",
  "f_role",
  "f_create_by",
  "f_update_by",
  "f_is_deleted"
)
SELECT
  s."f_id",
  s."f_memory_id",
  s."f_old_memory",
  s."f_new_memory",
  s."f_event",
  s."f_created_at",
  s."f_updated_at",
  s."f_actor_id",
  s."f_role",
  s."f_create_by",
  s."f_update_by",
  s."f_is_deleted"
FROM adp."t_data_agent_memory_history" s
WHERE s."f_is_deleted" = 0  -- 可选：只插入未删除的记录
  AND NOT EXISTS (
    SELECT 1
    FROM kweaver."t_data_agent_memory_history" t
    WHERE t."f_id" = s."f_id"
  );
