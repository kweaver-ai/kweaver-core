SET SCHEMA kweaver;
-- 从 adp 数据库复制数据到 kweaver 数据库的同名表
-- 使用 MERGE INTO 实现主键冲突时覆盖原表数据

-- t_internal_app
MERGE INTO kweaver."t_internal_app" t
USING adp."t_internal_app" s
ON t."f_app_id" = s."f_app_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_app_name" = s."f_app_name",
    t."f_app_secret" = s."f_app_secret",
    t."f_create_time" = s."f_create_time"
WHEN NOT MATCHED THEN
  INSERT ("f_app_id", "f_app_name", "f_app_secret", "f_create_time")
  VALUES (s."f_app_id", s."f_app_name", s."f_app_secret", s."f_create_time");

-- t_stream_data_pipeline
MERGE INTO kweaver."t_stream_data_pipeline" t
USING adp."t_stream_data_pipeline" s
ON t."f_pipeline_id" = s."f_pipeline_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_pipeline_name" = s."f_pipeline_name",
    t."f_tags" = s."f_tags",
    t."f_comment" = s."f_comment",
    t."f_builtin" = s."f_builtin",
    t."f_output_type" = s."f_output_type",
    t."f_index_base" = s."f_index_base",
    t."f_use_index_base_in_data" = s."f_use_index_base_in_data",
    t."f_pipeline_status" = s."f_pipeline_status",
    t."f_pipeline_status_details" = s."f_pipeline_status_details",
    t."f_deployment_config" = s."f_deployment_config",
    t."f_create_time" = s."f_create_time",
    t."f_update_time" = s."f_update_time",
    t."f_creator" = s."f_creator",
    t."f_updater" = s."f_updater",
    t."f_creator_type" = s."f_creator_type",
    t."f_updater_type" = s."f_updater_type"
WHEN NOT MATCHED THEN
  INSERT ("f_pipeline_id", "f_pipeline_name", "f_tags", "f_comment", "f_builtin", "f_output_type", "f_index_base", "f_use_index_base_in_data", "f_pipeline_status", "f_pipeline_status_details", "f_deployment_config", "f_create_time", "f_update_time", "f_creator", "f_updater", "f_creator_type", "f_updater_type")
  VALUES (s."f_pipeline_id", s."f_pipeline_name", s."f_tags", s."f_comment", s."f_builtin", s."f_output_type", s."f_index_base", s."f_use_index_base_in_data", s."f_pipeline_status", s."f_pipeline_status_details", s."f_deployment_config", s."f_create_time", s."f_update_time", s."f_creator", s."f_updater", s."f_creator_type", s."f_updater_type");
