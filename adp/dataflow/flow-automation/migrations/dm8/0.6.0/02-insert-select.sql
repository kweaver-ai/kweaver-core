SET SCHEMA kweaver;
-- 从 adp 数据库复制数据到 kweaver 数据库的同名表
-- 使用 MERGE INTO 实现主键冲突时覆盖原表数据

-- t_model
MERGE INTO kweaver."t_model" t
USING adp."t_model" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_name" = s."f_name",
    t."f_description" = s."f_description",
    t."f_train_status" = s."f_train_status",
    t."f_status" = s."f_status",
    t."f_rule" = s."f_rule",
    t."f_userid" = s."f_userid",
    t."f_type" = s."f_type",
    t."f_created_at" = s."f_created_at",
    t."f_updated_at" = s."f_updated_at",
    t."f_scope" = s."f_scope"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_name", "f_description", "f_train_status", "f_status", "f_rule", "f_userid", "f_type", "f_created_at", "f_updated_at", "f_scope")
  VALUES (s."f_id", s."f_name", s."f_description", s."f_train_status", s."f_status", s."f_rule", s."f_userid", s."f_type", s."f_created_at", s."f_updated_at", s."f_scope");

-- t_train_file
MERGE INTO kweaver."t_train_file" t
USING adp."t_train_file" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_train_id" = s."f_train_id",
    t."f_oss_id" = s."f_oss_id",
    t."f_key" = s."f_key",
    t."f_created_at" = s."f_created_at"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_train_id", "f_oss_id", "f_key", "f_created_at")
  VALUES (s."f_id", s."f_train_id", s."f_oss_id", s."f_key", s."f_created_at");

-- t_automation_executor
MERGE INTO kweaver."t_automation_executor" t
USING adp."t_automation_executor" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_name" = s."f_name",
    t."f_description" = s."f_description",
    t."f_creator_id" = s."f_creator_id",
    t."f_status" = s."f_status",
    t."f_created_at" = s."f_created_at",
    t."f_updated_at" = s."f_updated_at"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_name", "f_description", "f_creator_id", "f_status", "f_created_at", "f_updated_at")
  VALUES (s."f_id", s."f_name", s."f_description", s."f_creator_id", s."f_status", s."f_created_at", s."f_updated_at");

-- t_automation_executor_accessor
MERGE INTO kweaver."t_automation_executor_accessor" t
USING adp."t_automation_executor_accessor" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_executor_id" = s."f_executor_id",
    t."f_accessor_id" = s."f_accessor_id",
    t."f_accessor_type" = s."f_accessor_type"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_executor_id", "f_accessor_id", "f_accessor_type")
  VALUES (s."f_id", s."f_executor_id", s."f_accessor_id", s."f_accessor_type");

-- t_automation_executor_action
MERGE INTO kweaver."t_automation_executor_action" t
USING adp."t_automation_executor_action" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_executor_id" = s."f_executor_id",
    t."f_operator" = s."f_operator",
    t."f_name" = s."f_name",
    t."f_description" = s."f_description",
    t."f_group" = s."f_group",
    t."f_type" = s."f_type",
    t."f_inputs" = s."f_inputs",
    t."f_outputs" = s."f_outputs",
    t."f_config" = s."f_config",
    t."f_created_at" = s."f_created_at",
    t."f_updated_at" = s."f_updated_at"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_executor_id", "f_operator", "f_name", "f_description", "f_group", "f_type", "f_inputs", "f_outputs", "f_config", "f_created_at", "f_updated_at")
  VALUES (s."f_id", s."f_executor_id", s."f_operator", s."f_name", s."f_description", s."f_group", s."f_type", s."f_inputs", s."f_outputs", s."f_config", s."f_created_at", s."f_updated_at");

-- t_content_admin
MERGE INTO kweaver."t_content_admin" t
USING adp."t_content_admin" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_user_id" = s."f_user_id",
    t."f_user_name" = s."f_user_name"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_user_id", "f_user_name")
  VALUES (s."f_id", s."f_user_id", s."f_user_name");

-- t_audio_segments
MERGE INTO kweaver."t_audio_segments" t
USING adp."t_audio_segments" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_task_id" = s."f_task_id",
    t."f_object" = s."f_object",
    t."f_summary_type" = s."f_summary_type",
    t."f_max_segments" = s."f_max_segments",
    t."f_max_segments_type" = s."f_max_segments_type",
    t."f_need_abstract" = s."f_need_abstract",
    t."f_abstract_type" = s."f_abstract_type",
    t."f_callback" = s."f_callback",
    t."f_created_at" = s."f_created_at",
    t."f_updated_at" = s."f_updated_at"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_task_id", "f_object", "f_summary_type", "f_max_segments", "f_max_segments_type", "f_need_abstract", "f_abstract_type", "f_callback", "f_created_at", "f_updated_at")
  VALUES (s."f_id", s."f_task_id", s."f_object", s."f_summary_type", s."f_max_segments", s."f_max_segments_type", s."f_need_abstract", s."f_abstract_type", s."f_callback", s."f_created_at", s."f_updated_at");

-- t_automation_conf
MERGE INTO kweaver."t_automation_conf" t
USING adp."t_automation_conf" s
ON t."f_key" = s."f_key"
WHEN MATCHED THEN
  UPDATE SET
    t."f_value" = s."f_value"
WHEN NOT MATCHED THEN
  INSERT ("f_key", "f_value")
  VALUES (s."f_key", s."f_value");

-- t_automation_agent
MERGE INTO kweaver."t_automation_agent" t
USING adp."t_automation_agent" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_name" = s."f_name",
    t."f_agent_id" = s."f_agent_id",
    t."f_version" = s."f_version"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_name", "f_agent_id", "f_version")
  VALUES (s."f_id", s."f_name", s."f_agent_id", s."f_version");

-- t_alarm_rule
MERGE INTO kweaver."t_alarm_rule" t
USING adp."t_alarm_rule" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_rule_id" = s."f_rule_id",
    t."f_dag_id" = s."f_dag_id",
    t."f_frequency" = s."f_frequency",
    t."f_threshold" = s."f_threshold",
    t."f_created_at" = s."f_created_at"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_rule_id", "f_dag_id", "f_frequency", "f_threshold", "f_created_at")
  VALUES (s."f_id", s."f_rule_id", s."f_dag_id", s."f_frequency", s."f_threshold", s."f_created_at");

-- t_alarm_user
MERGE INTO kweaver."t_alarm_user" t
USING adp."t_alarm_user" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_rule_id" = s."f_rule_id",
    t."f_user_id" = s."f_user_id",
    t."f_user_name" = s."f_user_name",
    t."f_user_type" = s."f_user_type"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_rule_id", "f_user_id", "f_user_name", "f_user_type")
  VALUES (s."f_id", s."f_rule_id", s."f_user_id", s."f_user_name", s."f_user_type");

-- t_automation_dag_instance_ext_data
MERGE INTO kweaver."t_automation_dag_instance_ext_data" t
USING adp."t_automation_dag_instance_ext_data" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_created_at" = s."f_created_at",
    t."f_updated_at" = s."f_updated_at",
    t."f_dag_id" = s."f_dag_id",
    t."f_dag_ins_id" = s."f_dag_ins_id",
    t."f_field" = s."f_field",
    t."f_oss_id" = s."f_oss_id",
    t."f_oss_key" = s."f_oss_key",
    t."f_size" = s."f_size",
    t."f_removed" = s."f_removed"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_created_at", "f_updated_at", "f_dag_id", "f_dag_ins_id", "f_field", "f_oss_id", "f_oss_key", "f_size", "f_removed")
  VALUES (s."f_id", s."f_created_at", s."f_updated_at", s."f_dag_id", s."f_dag_ins_id", s."f_field", s."f_oss_id", s."f_oss_key", s."f_size", s."f_removed");

-- t_task_cache_0
MERGE INTO kweaver."t_task_cache_0" t
USING adp."t_task_cache_0" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_hash" = s."f_hash",
    t."f_type" = s."f_type",
    t."f_status" = s."f_status",
    t."f_oss_id" = s."f_oss_id",
    t."f_oss_key" = s."f_oss_key",
    t."f_ext" = s."f_ext",
    t."f_size" = s."f_size",
    t."f_err_msg" = s."f_err_msg",
    t."f_create_time" = s."f_create_time",
    t."f_modify_time" = s."f_modify_time",
    t."f_expire_time" = s."f_expire_time"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
  VALUES (s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time");

-- t_task_cache_1
MERGE INTO kweaver."t_task_cache_1" t
USING adp."t_task_cache_1" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_hash" = s."f_hash",
    t."f_type" = s."f_type",
    t."f_status" = s."f_status",
    t."f_oss_id" = s."f_oss_id",
    t."f_oss_key" = s."f_oss_key",
    t."f_ext" = s."f_ext",
    t."f_size" = s."f_size",
    t."f_err_msg" = s."f_err_msg",
    t."f_create_time" = s."f_create_time",
    t."f_modify_time" = s."f_modify_time",
    t."f_expire_time" = s."f_expire_time"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
  VALUES (s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time");

-- t_task_cache_2
MERGE INTO kweaver."t_task_cache_2" t
USING adp."t_task_cache_2" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_hash" = s."f_hash",
    t."f_type" = s."f_type",
    t."f_status" = s."f_status",
    t."f_oss_id" = s."f_oss_id",
    t."f_oss_key" = s."f_oss_key",
    t."f_ext" = s."f_ext",
    t."f_size" = s."f_size",
    t."f_err_msg" = s."f_err_msg",
    t."f_create_time" = s."f_create_time",
    t."f_modify_time" = s."f_modify_time",
    t."f_expire_time" = s."f_expire_time"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
  VALUES (s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time");

-- t_task_cache_3
MERGE INTO kweaver."t_task_cache_3" t
USING adp."t_task_cache_3" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_hash" = s."f_hash",
    t."f_type" = s."f_type",
    t."f_status" = s."f_status",
    t."f_oss_id" = s."f_oss_id",
    t."f_oss_key" = s."f_oss_key",
    t."f_ext" = s."f_ext",
    t."f_size" = s."f_size",
    t."f_err_msg" = s."f_err_msg",
    t."f_create_time" = s."f_create_time",
    t."f_modify_time" = s."f_modify_time",
    t."f_expire_time" = s."f_expire_time"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
  VALUES (s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time");

-- t_task_cache_4
MERGE INTO kweaver."t_task_cache_4" t
USING adp."t_task_cache_4" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_hash" = s."f_hash",
    t."f_type" = s."f_type",
    t."f_status" = s."f_status",
    t."f_oss_id" = s."f_oss_id",
    t."f_oss_key" = s."f_oss_key",
    t."f_ext" = s."f_ext",
    t."f_size" = s."f_size",
    t."f_err_msg" = s."f_err_msg",
    t."f_create_time" = s."f_create_time",
    t."f_modify_time" = s."f_modify_time",
    t."f_expire_time" = s."f_expire_time"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
  VALUES (s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time");

-- t_task_cache_5
MERGE INTO kweaver."t_task_cache_5" t
USING adp."t_task_cache_5" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_hash" = s."f_hash",
    t."f_type" = s."f_type",
    t."f_status" = s."f_status",
    t."f_oss_id" = s."f_oss_id",
    t."f_oss_key" = s."f_oss_key",
    t."f_ext" = s."f_ext",
    t."f_size" = s."f_size",
    t."f_err_msg" = s."f_err_msg",
    t."f_create_time" = s."f_create_time",
    t."f_modify_time" = s."f_modify_time",
    t."f_expire_time" = s."f_expire_time"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
  VALUES (s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time");

-- t_task_cache_6
MERGE INTO kweaver."t_task_cache_6" t
USING adp."t_task_cache_6" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_hash" = s."f_hash",
    t."f_type" = s."f_type",
    t."f_status" = s."f_status",
    t."f_oss_id" = s."f_oss_id",
    t."f_oss_key" = s."f_oss_key",
    t."f_ext" = s."f_ext",
    t."f_size" = s."f_size",
    t."f_err_msg" = s."f_err_msg",
    t."f_create_time" = s."f_create_time",
    t."f_modify_time" = s."f_modify_time",
    t."f_expire_time" = s."f_expire_time"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
  VALUES (s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time");

-- t_task_cache_7
MERGE INTO kweaver."t_task_cache_7" t
USING adp."t_task_cache_7" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_hash" = s."f_hash",
    t."f_type" = s."f_type",
    t."f_status" = s."f_status",
    t."f_oss_id" = s."f_oss_id",
    t."f_oss_key" = s."f_oss_key",
    t."f_ext" = s."f_ext",
    t."f_size" = s."f_size",
    t."f_err_msg" = s."f_err_msg",
    t."f_create_time" = s."f_create_time",
    t."f_modify_time" = s."f_modify_time",
    t."f_expire_time" = s."f_expire_time"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
  VALUES (s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time");

-- t_task_cache_8
MERGE INTO kweaver."t_task_cache_8" t
USING adp."t_task_cache_8" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_hash" = s."f_hash",
    t."f_type" = s."f_type",
    t."f_status" = s."f_status",
    t."f_oss_id" = s."f_oss_id",
    t."f_oss_key" = s."f_oss_key",
    t."f_ext" = s."f_ext",
    t."f_size" = s."f_size",
    t."f_err_msg" = s."f_err_msg",
    t."f_create_time" = s."f_create_time",
    t."f_modify_time" = s."f_modify_time",
    t."f_expire_time" = s."f_expire_time"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
  VALUES (s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time");

-- t_task_cache_9
MERGE INTO kweaver."t_task_cache_9" t
USING adp."t_task_cache_9" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_hash" = s."f_hash",
    t."f_type" = s."f_type",
    t."f_status" = s."f_status",
    t."f_oss_id" = s."f_oss_id",
    t."f_oss_key" = s."f_oss_key",
    t."f_ext" = s."f_ext",
    t."f_size" = s."f_size",
    t."f_err_msg" = s."f_err_msg",
    t."f_create_time" = s."f_create_time",
    t."f_modify_time" = s."f_modify_time",
    t."f_expire_time" = s."f_expire_time"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
  VALUES (s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time");

-- t_task_cache_a
MERGE INTO kweaver."t_task_cache_a" t
USING adp."t_task_cache_a" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_hash" = s."f_hash",
    t."f_type" = s."f_type",
    t."f_status" = s."f_status",
    t."f_oss_id" = s."f_oss_id",
    t."f_oss_key" = s."f_oss_key",
    t."f_ext" = s."f_ext",
    t."f_size" = s."f_size",
    t."f_err_msg" = s."f_err_msg",
    t."f_create_time" = s."f_create_time",
    t."f_modify_time" = s."f_modify_time",
    t."f_expire_time" = s."f_expire_time"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
  VALUES (s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time");

-- t_task_cache_b
MERGE INTO kweaver."t_task_cache_b" t
USING adp."t_task_cache_b" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_hash" = s."f_hash",
    t."f_type" = s."f_type",
    t."f_status" = s."f_status",
    t."f_oss_id" = s."f_oss_id",
    t."f_oss_key" = s."f_oss_key",
    t."f_ext" = s."f_ext",
    t."f_size" = s."f_size",
    t."f_err_msg" = s."f_err_msg",
    t."f_create_time" = s."f_create_time",
    t."f_modify_time" = s."f_modify_time",
    t."f_expire_time" = s."f_expire_time"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
  VALUES (s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time");

-- t_task_cache_c
MERGE INTO kweaver."t_task_cache_c" t
USING adp."t_task_cache_c" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_hash" = s."f_hash",
    t."f_type" = s."f_type",
    t."f_status" = s."f_status",
    t."f_oss_id" = s."f_oss_id",
    t."f_oss_key" = s."f_oss_key",
    t."f_ext" = s."f_ext",
    t."f_size" = s."f_size",
    t."f_err_msg" = s."f_err_msg",
    t."f_create_time" = s."f_create_time",
    t."f_modify_time" = s."f_modify_time",
    t."f_expire_time" = s."f_expire_time"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
  VALUES (s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time");

-- t_task_cache_d
MERGE INTO kweaver."t_task_cache_d" t
USING adp."t_task_cache_d" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_hash" = s."f_hash",
    t."f_type" = s."f_type",
    t."f_status" = s."f_status",
    t."f_oss_id" = s."f_oss_id",
    t."f_oss_key" = s."f_oss_key",
    t."f_ext" = s."f_ext",
    t."f_size" = s."f_size",
    t."f_err_msg" = s."f_err_msg",
    t."f_create_time" = s."f_create_time",
    t."f_modify_time" = s."f_modify_time",
    t."f_expire_time" = s."f_expire_time"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
  VALUES (s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time");

-- t_task_cache_e
MERGE INTO kweaver."t_task_cache_e" t
USING adp."t_task_cache_e" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_hash" = s."f_hash",
    t."f_type" = s."f_type",
    t."f_status" = s."f_status",
    t."f_oss_id" = s."f_oss_id",
    t."f_oss_key" = s."f_oss_key",
    t."f_ext" = s."f_ext",
    t."f_size" = s."f_size",
    t."f_err_msg" = s."f_err_msg",
    t."f_create_time" = s."f_create_time",
    t."f_modify_time" = s."f_modify_time",
    t."f_expire_time" = s."f_expire_time"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
  VALUES (s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time");

-- t_task_cache_f
MERGE INTO kweaver."t_task_cache_f" t
USING adp."t_task_cache_f" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_hash" = s."f_hash",
    t."f_type" = s."f_type",
    t."f_status" = s."f_status",
    t."f_oss_id" = s."f_oss_id",
    t."f_oss_key" = s."f_oss_key",
    t."f_ext" = s."f_ext",
    t."f_size" = s."f_size",
    t."f_err_msg" = s."f_err_msg",
    t."f_create_time" = s."f_create_time",
    t."f_modify_time" = s."f_modify_time",
    t."f_expire_time" = s."f_expire_time"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
  VALUES (s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time");

-- t_dag_instance_event
MERGE INTO kweaver."t_dag_instance_event" t
USING adp."t_dag_instance_event" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_type" = s."f_type",
    t."f_instance_id" = s."f_instance_id",
    t."f_operator" = s."f_operator",
    t."f_task_id" = s."f_task_id",
    t."f_status" = s."f_status",
    t."f_name" = s."f_name",
    t."f_data" = s."f_data",
    t."f_size" = s."f_size",
    t."f_inline" = s."f_inline",
    t."f_visibility" = s."f_visibility",
    t."f_timestamp" = s."f_timestamp"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_type", "f_instance_id", "f_operator", "f_task_id", "f_status", "f_name", "f_data", "f_size", "f_inline", "f_visibility", "f_timestamp")
  VALUES (s."f_id", s."f_type", s."f_instance_id", s."f_operator", s."f_task_id", s."f_status", s."f_name", s."f_data", s."f_size", s."f_inline", s."f_visibility", s."f_timestamp");

-- t_cron_job
MERGE INTO kweaver."t_cron_job" t
USING adp."t_cron_job" s
ON t."f_key_id" = s."f_key_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_job_id" = s."f_job_id",
    t."f_job_name" = s."f_job_name",
    t."f_job_cron_time" = s."f_job_cron_time",
    t."f_job_type" = s."f_job_type",
    t."f_job_context" = s."f_job_context",
    t."f_tenant_id" = s."f_tenant_id",
    t."f_enabled" = s."f_enabled",
    t."f_remarks" = s."f_remarks",
    t."f_create_time" = s."f_create_time",
    t."f_update_time" = s."f_update_time"
WHEN NOT MATCHED THEN
  INSERT ("f_key_id", "f_job_id", "f_job_name", "f_job_cron_time", "f_job_type", "f_job_context", "f_tenant_id", "f_enabled", "f_remarks", "f_create_time", "f_update_time")
  VALUES (s."f_key_id", s."f_job_id", s."f_job_name", s."f_job_cron_time", s."f_job_type", s."f_job_context", s."f_tenant_id", s."f_enabled", s."f_remarks", s."f_create_time", s."f_update_time");

-- t_cron_job_status
MERGE INTO kweaver."t_cron_job_status" t
USING adp."t_cron_job_status" s
ON t."f_key_id" = s."f_key_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_execute_id" = s."f_execute_id",
    t."f_job_id" = s."f_job_id",
    t."f_job_type" = s."f_job_type",
    t."f_job_name" = s."f_job_name",
    t."f_job_status" = s."f_job_status",
    t."f_begin_time" = s."f_begin_time",
    t."f_end_time" = s."f_end_time",
    t."f_executor" = s."f_executor",
    t."f_execute_times" = s."f_execute_times",
    t."f_ext_info" = s."f_ext_info"
WHEN NOT MATCHED THEN
  INSERT ("f_key_id", "f_execute_id", "f_job_id", "f_job_type", "f_job_name", "f_job_status", "f_begin_time", "f_end_time", "f_executor", "f_execute_times", "f_ext_info")
  VALUES (s."f_key_id", s."f_execute_id", s."f_job_id", s."f_job_type", s."f_job_name", s."f_job_status", s."f_begin_time", s."f_end_time", s."f_executor", s."f_execute_times", s."f_ext_info");

-- t_flow_dag
MERGE INTO kweaver."t_flow_dag" t
USING adp."t_flow_dag" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_created_at" = s."f_created_at",
    t."f_updated_at" = s."f_updated_at",
    t."f_user_id" = s."f_user_id",
    t."f_name" = s."f_name",
    t."f_desc" = s."f_desc",
    t."f_trigger" = s."f_trigger",
    t."f_cron" = s."f_cron",
    t."f_vars" = s."f_vars",
    t."f_status" = s."f_status",
    t."f_tasks" = s."f_tasks",
    t."f_steps" = s."f_steps",
    t."f_description" = s."f_description",
    t."f_shortcuts" = s."f_shortcuts",
    t."f_accessors" = s."f_accessors",
    t."f_type" = s."f_type",
    t."f_policy_type" = s."f_policy_type",
    t."f_appinfo" = s."f_appinfo",
    t."f_priority" = s."f_priority",
    t."f_removed" = s."f_removed",
    t."f_emails" = s."f_emails",
    t."f_template" = s."f_template",
    t."f_published" = s."f_published",
    t."f_trigger_config" = s."f_trigger_config",
    t."f_sub_ids" = s."f_sub_ids",
    t."f_exec_mode" = s."f_exec_mode",
    t."f_category" = s."f_category",
    t."f_outputs" = s."f_outputs",
    t."f_instructions" = s."f_instructions",
    t."f_operator_id" = s."f_operator_id",
    t."f_inc_values" = s."f_inc_values",
    t."f_version" = s."f_version",
    t."f_version_id" = s."f_version_id",
    t."f_modify_by" = s."f_modify_by",
    t."f_is_debug" = s."f_is_debug",
    t."f_debug_id" = s."f_debug_id",
    t."f_biz_domain_id" = s."f_biz_domain_id"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_created_at", "f_updated_at", "f_user_id", "f_name", "f_desc", "f_trigger", "f_cron", "f_vars", "f_status", "f_tasks", "f_steps", "f_description", "f_shortcuts", "f_accessors", "f_type", "f_policy_type", "f_appinfo", "f_priority", "f_removed", "f_emails", "f_template", "f_published", "f_trigger_config", "f_sub_ids", "f_exec_mode", "f_category", "f_outputs", "f_instructions", "f_operator_id", "f_inc_values", "f_version", "f_version_id", "f_modify_by", "f_is_debug", "f_debug_id", "f_biz_domain_id")
  VALUES (s."f_id", s."f_created_at", s."f_updated_at", s."f_user_id", s."f_name", s."f_desc", s."f_trigger", s."f_cron", s."f_vars", s."f_status", s."f_tasks", s."f_steps", s."f_description", s."f_shortcuts", s."f_accessors", s."f_type", s."f_policy_type", s."f_appinfo", s."f_priority", s."f_removed", s."f_emails", s."f_template", s."f_published", s."f_trigger_config", s."f_sub_ids", s."f_exec_mode", s."f_category", s."f_outputs", s."f_instructions", s."f_operator_id", s."f_inc_values", s."f_version", s."f_version_id", s."f_modify_by", s."f_is_debug", s."f_debug_id", s."f_biz_domain_id");

-- t_flow_dag_var
MERGE INTO kweaver."t_flow_dag_var" t
USING adp."t_flow_dag_var" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_dag_id" = s."f_dag_id",
    t."f_var_name" = s."f_var_name",
    t."f_default_value" = s."f_default_value",
    t."f_var_type" = s."f_var_type",
    t."f_description" = s."f_description"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_dag_id", "f_var_name", "f_default_value", "f_var_type", "f_description")
  VALUES (s."f_id", s."f_dag_id", s."f_var_name", s."f_default_value", s."f_var_type", s."f_description");

-- t_flow_dag_instance_keyword
MERGE INTO kweaver."t_flow_dag_instance_keyword" t
USING adp."t_flow_dag_instance_keyword" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_dag_ins_id" = s."f_dag_ins_id",
    t."f_keyword" = s."f_keyword"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_dag_ins_id", "f_keyword")
  VALUES (s."f_id", s."f_dag_ins_id", s."f_keyword");

-- t_flow_dag_step
MERGE INTO kweaver."t_flow_dag_step" t
USING adp."t_flow_dag_step" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_dag_id" = s."f_dag_id",
    t."f_operator" = s."f_operator",
    t."f_source_id" = s."f_source_id",
    t."f_has_datasource" = s."f_has_datasource"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_dag_id", "f_operator", "f_source_id", "f_has_datasource")
  VALUES (s."f_id", s."f_dag_id", s."f_operator", s."f_source_id", s."f_has_datasource");

-- t_flow_dag_accessor
MERGE INTO kweaver."t_flow_dag_accessor" t
USING adp."t_flow_dag_accessor" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_dag_id" = s."f_dag_id",
    t."f_accessor_id" = s."f_accessor_id"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_dag_id", "f_accessor_id")
  VALUES (s."f_id", s."f_dag_id", s."f_accessor_id");

-- t_flow_dag_version
MERGE INTO kweaver."t_flow_dag_version" t
USING adp."t_flow_dag_version" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_created_at" = s."f_created_at",
    t."f_updated_at" = s."f_updated_at",
    t."f_dag_id" = s."f_dag_id",
    t."f_user_id" = s."f_user_id",
    t."f_version" = s."f_version",
    t."f_version_id" = s."f_version_id",
    t."f_change_log" = s."f_change_log",
    t."f_config" = s."f_config",
    t."f_sort_time" = s."f_sort_time"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_created_at", "f_updated_at", "f_dag_id", "f_user_id", "f_version", "f_version_id", "f_change_log", "f_config", "f_sort_time")
  VALUES (s."f_id", s."f_created_at", s."f_updated_at", s."f_dag_id", s."f_user_id", s."f_version", s."f_version_id", s."f_change_log", s."f_config", s."f_sort_time");

-- t_flow_dag_instance
MERGE INTO kweaver."t_flow_dag_instance" t
USING adp."t_flow_dag_instance" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_created_at" = s."f_created_at",
    t."f_updated_at" = s."f_updated_at",
    t."f_dag_id" = s."f_dag_id",
    t."f_trigger" = s."f_trigger",
    t."f_worker" = s."f_worker",
    t."f_source" = s."f_source",
    t."f_vars" = s."f_vars",
    t."f_keywords" = s."f_keywords",
    t."f_event_persistence" = s."f_event_persistence",
    t."f_event_oss_path" = s."f_event_oss_path",
    t."f_share_data" = s."f_share_data",
    t."f_share_data_ext" = s."f_share_data_ext",
    t."f_status" = s."f_status",
    t."f_reason" = s."f_reason",
    t."f_cmd" = s."f_cmd",
    t."f_has_cmd" = s."f_has_cmd",
    t."f_batch_run_id" = s."f_batch_run_id",
    t."f_user_id" = s."f_user_id",
    t."f_ended_at" = s."f_ended_at",
    t."f_dag_type" = s."f_dag_type",
    t."f_policy_type" = s."f_policy_type",
    t."f_appinfo" = s."f_appinfo",
    t."f_priority" = s."f_priority",
    t."f_mode" = s."f_mode",
    t."f_dump" = s."f_dump",
    t."f_dump_ext" = s."f_dump_ext",
    t."f_success_callback" = s."f_success_callback",
    t."f_error_callback" = s."f_error_callback",
    t."f_call_chain" = s."f_call_chain",
    t."f_resume_data" = s."f_resume_data",
    t."f_resume_status" = s."f_resume_status",
    t."f_version" = s."f_version",
    t."f_version_id" = s."f_version_id",
    t."f_biz_domain_id" = s."f_biz_domain_id"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_created_at", "f_updated_at", "f_dag_id", "f_trigger", "f_worker", "f_source", "f_vars", "f_keywords", "f_event_persistence", "f_event_oss_path", "f_share_data", "f_share_data_ext", "f_status", "f_reason", "f_cmd", "f_has_cmd", "f_batch_run_id", "f_user_id", "f_ended_at", "f_dag_type", "f_policy_type", "f_appinfo", "f_priority", "f_mode", "f_dump", "f_dump_ext", "f_success_callback", "f_error_callback", "f_call_chain", "f_resume_data", "f_resume_status", "f_version", "f_version_id", "f_biz_domain_id")
  VALUES (s."f_id", s."f_created_at", s."f_updated_at", s."f_dag_id", s."f_trigger", s."f_worker", s."f_source", s."f_vars", s."f_keywords", s."f_event_persistence", s."f_event_oss_path", s."f_share_data", s."f_share_data_ext", s."f_status", s."f_reason", s."f_cmd", s."f_has_cmd", s."f_batch_run_id", s."f_user_id", s."f_ended_at", s."f_dag_type", s."f_policy_type", s."f_appinfo", s."f_priority", s."f_mode", s."f_dump", s."f_dump_ext", s."f_success_callback", s."f_error_callback", s."f_call_chain", s."f_resume_data", s."f_resume_status", s."f_version", s."f_version_id", s."f_biz_domain_id");

-- t_flow_inbox
MERGE INTO kweaver."t_flow_inbox" t
USING adp."t_flow_inbox" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_created_at" = s."f_created_at",
    t."f_updated_at" = s."f_updated_at",
    t."f_msg" = s."f_msg",
    t."f_topic" = s."f_topic",
    t."f_docid" = s."f_docid",
    t."f_dag" = s."f_dag"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_created_at", "f_updated_at", "f_msg", "f_topic", "f_docid", "f_dag")
  VALUES (s."f_id", s."f_created_at", s."f_updated_at", s."f_msg", s."f_topic", s."f_docid", s."f_dag");

-- t_flow_outbox
MERGE INTO kweaver."t_flow_outbox" t
USING adp."t_flow_outbox" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_created_at" = s."f_created_at",
    t."f_updated_at" = s."f_updated_at",
    t."f_msg" = s."f_msg",
    t."f_topic" = s."f_topic"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_created_at", "f_updated_at", "f_msg", "f_topic")
  VALUES (s."f_id", s."f_created_at", s."f_updated_at", s."f_msg", s."f_topic");

-- t_flow_task_instance
MERGE INTO kweaver."t_flow_task_instance" t
USING adp."t_flow_task_instance" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_created_at" = s."f_created_at",
    t."f_updated_at" = s."f_updated_at",
    t."f_expired_at" = s."f_expired_at",
    t."f_task_id" = s."f_task_id",
    t."f_dag_ins_id" = s."f_dag_ins_id",
    t."f_name" = s."f_name",
    t."f_depend_on" = s."f_depend_on",
    t."f_action_name" = s."f_action_name",
    t."f_timeout_secs" = s."f_timeout_secs",
    t."f_params" = s."f_params",
    t."f_traces" = s."f_traces",
    t."f_status" = s."f_status",
    t."f_reason" = s."f_reason",
    t."f_pre_checks" = s."f_pre_checks",
    t."f_results" = s."f_results",
    t."f_steps" = s."f_steps",
    t."f_last_modified_at" = s."f_last_modified_at",
    t."f_rendered_params" = s."f_rendered_params",
    t."f_hash" = s."f_hash",
    t."f_settings" = s."f_settings",
    t."f_metadata" = s."f_metadata"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_created_at", "f_updated_at", "f_expired_at", "f_task_id", "f_dag_ins_id", "f_name", "f_depend_on", "f_action_name", "f_timeout_secs", "f_params", "f_traces", "f_status", "f_reason", "f_pre_checks", "f_results", "f_steps", "f_last_modified_at", "f_rendered_params", "f_hash", "f_settings", "f_metadata")
  VALUES (s."f_id", s."f_created_at", s."f_updated_at", s."f_expired_at", s."f_task_id", s."f_dag_ins_id", s."f_name", s."f_depend_on", s."f_action_name", s."f_timeout_secs", s."f_params", s."f_traces", s."f_status", s."f_reason", s."f_pre_checks", s."f_results", s."f_steps", s."f_last_modified_at", s."f_rendered_params", s."f_hash", s."f_settings", s."f_metadata");

-- t_flow_token
MERGE INTO kweaver."t_flow_token" t
USING adp."t_flow_token" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_created_at" = s."f_created_at",
    t."f_updated_at" = s."f_updated_at",
    t."f_user_id" = s."f_user_id",
    t."f_user_name" = s."f_user_name",
    t."f_refresh_token" = s."f_refresh_token",
    t."f_token" = s."f_token",
    t."f_expires_in" = s."f_expires_in",
    t."f_login_ip" = s."f_login_ip",
    t."f_is_app" = s."f_is_app"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_created_at", "f_updated_at", "f_user_id", "f_user_name", "f_refresh_token", "f_token", "f_expires_in", "f_login_ip", "f_is_app")
  VALUES (s."f_id", s."f_created_at", s."f_updated_at", s."f_user_id", s."f_user_name", s."f_refresh_token", s."f_token", s."f_expires_in", s."f_login_ip", s."f_is_app");

-- t_flow_client
MERGE INTO kweaver."t_flow_client" t
USING adp."t_flow_client" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_created_at" = s."f_created_at",
    t."f_updated_at" = s."f_updated_at",
    t."f_client_name" = s."f_client_name",
    t."f_client_id" = s."f_client_id",
    t."f_client_secret" = s."f_client_secret"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_created_at", "f_updated_at", "f_client_name", "f_client_id", "f_client_secret")
  VALUES (s."f_id", s."f_created_at", s."f_updated_at", s."f_client_name", s."f_client_id", s."f_client_secret");

-- t_flow_switch
MERGE INTO kweaver."t_flow_switch" t
USING adp."t_flow_switch" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_created_at" = s."f_created_at",
    t."f_updated_at" = s."f_updated_at",
    t."f_name" = s."f_name",
    t."f_status" = s."f_status"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_created_at", "f_updated_at", "f_name", "f_status")
  VALUES (s."f_id", s."f_created_at", s."f_updated_at", s."f_name", s."f_status");

-- t_flow_log
MERGE INTO kweaver."t_flow_log" t
USING adp."t_flow_log" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_created_at" = s."f_created_at",
    t."f_updated_at" = s."f_updated_at",
    t."f_ossid" = s."f_ossid",
    t."f_key" = s."f_key",
    t."f_filename" = s."f_filename"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_created_at", "f_updated_at", "f_ossid", "f_key", "f_filename")
  VALUES (s."f_id", s."f_created_at", s."f_updated_at", s."f_ossid", s."f_key", s."f_filename");

-- t_flow_storage
MERGE INTO kweaver."t_flow_storage" t
USING adp."t_flow_storage" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_oss_id" = s."f_oss_id",
    t."f_object_key" = s."f_object_key",
    t."f_name" = s."f_name",
    t."f_content_type" = s."f_content_type",
    t."f_size" = s."f_size",
    t."f_etag" = s."f_etag",
    t."f_status" = s."f_status",
    t."f_created_at" = s."f_created_at",
    t."f_updated_at" = s."f_updated_at",
    t."f_deleted_at" = s."f_deleted_at"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_oss_id", "f_object_key", "f_name", "f_content_type", "f_size", "f_etag", "f_status", "f_created_at", "f_updated_at", "f_deleted_at")
  VALUES (s."f_id", s."f_oss_id", s."f_object_key", s."f_name", s."f_content_type", s."f_size", s."f_etag", s."f_status", s."f_created_at", s."f_updated_at", s."f_deleted_at");

-- t_flow_file
MERGE INTO kweaver."t_flow_file" t
USING adp."t_flow_file" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_dag_id" = s."f_dag_id",
    t."f_dag_instance_id" = s."f_dag_instance_id",
    t."f_storage_id" = s."f_storage_id",
    t."f_status" = s."f_status",
    t."f_name" = s."f_name",
    t."f_expires_at" = s."f_expires_at",
    t."f_created_at" = s."f_created_at",
    t."f_updated_at" = s."f_updated_at"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_dag_id", "f_dag_instance_id", "f_storage_id", "f_status", "f_name", "f_expires_at", "f_created_at", "f_updated_at")
  VALUES (s."f_id", s."f_dag_id", s."f_dag_instance_id", s."f_storage_id", s."f_status", s."f_name", s."f_expires_at", s."f_created_at", s."f_updated_at");

-- t_flow_file_download_job
MERGE INTO kweaver."t_flow_file_download_job" t
USING adp."t_flow_file_download_job" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_file_id" = s."f_file_id",
    t."f_status" = s."f_status",
    t."f_retry_count" = s."f_retry_count",
    t."f_max_retry" = s."f_max_retry",
    t."f_next_retry_at" = s."f_next_retry_at",
    t."f_error_code" = s."f_error_code",
    t."f_error_message" = s."f_error_message",
    t."f_download_url" = s."f_download_url",
    t."f_started_at" = s."f_started_at",
    t."f_finished_at" = s."f_finished_at",
    t."f_created_at" = s."f_created_at",
    t."f_updated_at" = s."f_updated_at"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_file_id", "f_status", "f_retry_count", "f_max_retry", "f_next_retry_at", "f_error_code", "f_error_message", "f_download_url", "f_started_at", "f_finished_at", "f_created_at", "f_updated_at")
  VALUES (s."f_id", s."f_file_id", s."f_status", s."f_retry_count", s."f_max_retry", s."f_next_retry_at", s."f_error_code", s."f_error_message", s."f_download_url", s."f_started_at", s."f_finished_at", s."f_created_at", s."f_updated_at");

-- t_flow_task_resume
MERGE INTO kweaver."t_flow_task_resume" t
USING adp."t_flow_task_resume" s
ON t."f_id" = s."f_id"
WHEN MATCHED THEN
  UPDATE SET
    t."f_task_instance_id" = s."f_task_instance_id",
    t."f_dag_instance_id" = s."f_dag_instance_id",
    t."f_resource_type" = s."f_resource_type",
    t."f_resource_id" = s."f_resource_id",
    t."f_created_at" = s."f_created_at",
    t."f_updated_at" = s."f_updated_at"
WHEN NOT MATCHED THEN
  INSERT ("f_id", "f_task_instance_id", "f_dag_instance_id", "f_resource_type", "f_resource_id", "f_created_at", "f_updated_at")
  VALUES (s."f_id", s."f_task_instance_id", s."f_dag_instance_id", s."f_resource_type", s."f_resource_id", s."f_created_at", s."f_updated_at");
