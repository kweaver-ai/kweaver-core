USE adp;

-- 从 workflow 数据库迁移 t_python_package 表数据到 adp 数据库
INSERT INTO adp.t_python_package (
  f_id,
  f_name,
  f_oss_id,
  f_oss_key,
  f_creator_id,
  f_creator_name,
  f_created_at
)
SELECT 
  f_id,
  f_name,
  f_oss_id,
  f_oss_key,
  f_creator_id,
  f_creator_name,
  f_created_at
FROM workflow.t_python_package
WHERE NOT EXISTS (
  SELECT 1 
  FROM adp.t_python_package pp 
  WHERE pp.f_id = workflow.t_python_package.f_id
);