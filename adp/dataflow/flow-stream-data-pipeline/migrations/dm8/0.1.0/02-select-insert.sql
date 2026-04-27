SET SCHEMA adp;

-- 迁移 "t_internal_app" 表数据
insert into "adp"."t_internal_app" (
  "f_app_id",
  "f_app_name",
  "f_app_secret",
  "f_create_time"
)
select 
  "f_app_id",
  "f_app_name",
  "f_app_secret",
  "f_create_time"
from "workflow"."t_internal_app" w
where not exists (
  select 1 
  from "adp"."t_internal_app" a 
  where a."f_app_id" = w."f_app_id"
);

-- 迁移 "t_stream_data_pipeline" 表数据
insert into "adp"."t_stream_data_pipeline" (
  "f_pipeline_id",
  "f_pipeline_name",
  "f_tags",
  "f_comment",
  "f_builtin",
  "f_output_type",
  "f_index_base",
  "f_use_index_base_in_data",
  "f_pipeline_status",
  "f_pipeline_status_details",
  "f_deployment_config",
  "f_create_time",
  "f_update_time",
  "f_creator",
  "f_updater",
  "f_creator_type",
  "f_updater_type"
)
select 
  "f_pipeline_id",
  "f_pipeline_name",
  "f_tags",
  "f_comment",
  "f_builtin",
  "f_output_type",
  "f_index_base",
  "f_use_index_base_in_data",
  "f_pipeline_status",
  "f_pipeline_status_details",
  "f_deployment_config",
  "f_create_time",
  "f_update_time",
  "f_creator",
  "f_updater",
  "f_creator_type",
  "f_updater_type"
from "workflow"."t_stream_data_pipeline" w
where not exists (
  select 1 
  from "adp"."t_stream_data_pipeline" a 
  where a."f_pipeline_id" = w."f_pipeline_id"
);