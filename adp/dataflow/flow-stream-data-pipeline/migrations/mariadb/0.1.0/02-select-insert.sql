USE adp;

-- 从 workflow 数据库迁移数据到 adp 数据库
-- 迁移 t_internal_app 表数据
INSERT INTO adp.t_internal_app (
  f_app_id,
  f_app_name,
  f_app_secret,
  f_create_time
)
SELECT 
  f_app_id,
  f_app_name,
  f_app_secret,
  f_create_time
FROM workflow.t_internal_app
WHERE NOT EXISTS (
  SELECT 1 
  FROM adp.t_internal_app a 
  WHERE a.f_app_id = workflow.t_internal_app.f_app_id
);

-- 迁移 t_stream_data_pipeline 表数据
INSERT INTO adp.t_stream_data_pipeline (
  f_pipeline_id,
  f_pipeline_name,
  f_tags,
  f_comment,
  f_builtin,
  f_output_type,
  f_index_base,
  f_use_index_base_in_data,
  f_pipeline_status,
  f_pipeline_status_details,
  f_deployment_config,
  f_create_time,
  f_update_time,
  f_creator,
  f_creator_type,
  f_updater,
  f_updater_type
)
SELECT 
  f_pipeline_id,
  f_pipeline_name,
  f_tags,
  f_comment,
  f_builtin,
  f_output_type,
  f_index_base,
  f_use_index_base_in_data,
  f_pipeline_status,
  f_pipeline_status_details,
  f_deployment_config,
  f_create_time,
  f_update_time,
  f_creator,
  f_creator_type,
  f_updater,
  f_updater_type
FROM workflow.t_stream_data_pipeline
WHERE NOT EXISTS (
  SELECT 1 
  FROM adp.t_stream_data_pipeline s 
  WHERE s.f_pipeline_id = workflow.t_stream_data_pipeline.f_pipeline_id
);