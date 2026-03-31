SET SCHEMA adp;

-- 从 workflow 数据库迁移 t_python_package 表数据到 adp 数据库
insert into "adp"."t_python_package" (
  "f_id",
  "f_name",
  "f_oss_id",
  "f_oss_key",
  "f_creator_id",
  "f_creator_name",
  "f_created_at"
)
select 
  "f_id",
  "f_name",
  "f_oss_id",
  "f_oss_key",
  "f_creator_id",
  "f_creator_name",
  "f_created_at"
from "workflow"."t_python_package" w
where not exists (
  select 1 
  from "adp"."t_python_package" a 
  where a."f_id" = w."f_id"
);