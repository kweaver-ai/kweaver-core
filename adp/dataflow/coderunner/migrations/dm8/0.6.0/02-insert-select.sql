SET SCHEMA kweaver;
-- 从 adp 数据库的 t_python_package 表复制数据到 kweaver 数据库的同名表
-- 使用 MERGE INTO 实现主键冲突时覆盖原表数据

MERGE INTO kweaver."t_python_package" t
USING adp."t_python_package" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_name" = s."f_name",
    t."f_oss_id" = s."f_oss_id",
    t."f_oss_key" = s."f_oss_key",
    t."f_creator_id" = s."f_creator_id",
    t."f_creator_name" = s."f_creator_name",
    t."f_created_at" = s."f_created_at"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_name", "f_oss_id", "f_oss_key", "f_creator_id", "f_creator_name", "f_created_at")
  VALUES (s."f_id", s."f_name", s."f_oss_id", s."f_oss_key", s."f_creator_id", s."f_creator_name", s."f_created_at");
