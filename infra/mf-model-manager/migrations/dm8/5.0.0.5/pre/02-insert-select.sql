-- 迁移数据从5.3.0 model_management schema 到 5.4.0 adp schema
-- 注意：使用WHERE NOT EXISTS避免重复插入数据

SET SCHEMA adp;

-- 迁移 t_llm_model 表数据
INSERT INTO adp.t_llm_model (
    f_model_id,
    f_model_series,
    f_model_type,
    f_model_name,
    f_model,
    f_model_config,
    f_create_by,
    f_create_time,
    f_update_by,
    f_update_time,
    f_max_model_len,
    f_model_parameters,
    "f_quota",
    "f_default"
)
SELECT
    f_model_id,
    f_model_series,
    f_model_type,
    f_model_name,
    f_model,
    f_model_config,
    f_create_by,
    f_create_time,
    f_update_by,
    f_update_time,
    f_max_model_len,
    f_model_parameters,
    "f_quota",
    "f_default"
FROM model_management.t_llm_model
WHERE NOT EXISTS (
    SELECT 1 FROM adp.t_llm_model t
    WHERE t.f_model_id = model_management.t_llm_model.f_model_id
);

-- 迁移 t_small_model 表数据
INSERT INTO adp.t_small_model (
    f_model_id,
    f_model_name,
    f_model_type,
    f_model_config,
    f_create_time,
    f_update_time,
    f_create_by,
    f_update_by,
    "f_adapter",
    "f_adapter_code",
    "f_batch_size",
    "f_max_tokens",
    "f_embedding_dim"
)
SELECT
    f_model_id,
    f_model_name,
    f_model_type,
    f_model_config,
    f_create_time,
    f_update_time,
    f_create_by,
    f_update_by,
    "f_adapter",
    "f_adapter_code",
    "f_batch_size",
    "f_max_tokens",
    "f_embedding_dim"
FROM model_management.t_small_model
WHERE NOT EXISTS (
    SELECT 1 FROM adp.t_small_model t
    WHERE t.f_model_id = model_management.t_small_model.f_model_id
);

-- 迁移 t_prompt_item_list 表数据
INSERT INTO adp.t_prompt_item_list (
    f_id,
    f_prompt_item_id,
    f_prompt_item_name,
    f_prompt_item_type_id,
    f_prompt_item_type,
    f_create_by,
    f_create_time,
    f_update_by,
    f_update_time,
    f_item_is_delete,
    f_type_is_delete,
    f_built_in
)
SELECT
    f_id,
    f_prompt_item_id,
    f_prompt_item_name,
    f_prompt_item_type_id,
    f_prompt_item_type,
    f_create_by,
    f_create_time,
    f_update_by,
    f_update_time,
    f_item_is_delete,
    f_type_is_delete,
    f_built_in
FROM model_management.t_prompt_item_list
WHERE NOT EXISTS (
    SELECT 1 FROM adp.t_prompt_item_list t
    WHERE t.f_id = model_management.t_prompt_item_list.f_id
);

-- 迁移 t_prompt_list 表数据
INSERT INTO adp.t_prompt_list (
    f_prompt_id,
    f_prompt_item_id,
    f_prompt_item_type_id,
    f_prompt_service_id,
    f_prompt_type,
    f_prompt_name,
    f_prompt_desc,
    f_messages,
    f_variables,
    f_icon,
    f_model_id,
    f_model_para,
    f_opening_remarks,
    f_is_deploy,
    f_prompt_deploy_url,
    f_prompt_deploy_api,
    f_create_by,
    f_create_time,
    f_update_by,
    f_update_time,
    f_is_delete,
    f_built_in
)
SELECT
    f_prompt_id,
    f_prompt_item_id,
    f_prompt_item_type_id,
    f_prompt_service_id,
    f_prompt_type,
    f_prompt_name,
    f_prompt_desc,
    f_messages,
    f_variables,
    f_icon,
    f_model_id,
    f_model_para,
    f_opening_remarks,
    f_is_deploy,
    f_prompt_deploy_url,
    f_prompt_deploy_api,
    f_create_by,
    f_create_time,
    f_update_by,
    f_update_time,
    f_is_delete,
    f_built_in
FROM model_management.t_prompt_list
WHERE NOT EXISTS (
    SELECT 1 FROM adp.t_prompt_list t
    WHERE t.f_prompt_id = model_management.t_prompt_list.f_prompt_id
);

-- 迁移 t_prompt_template_list 表数据
INSERT INTO adp.t_prompt_template_list (
    f_prompt_id,
    f_prompt_type,
    f_prompt_name,
    f_prompt_desc,
    f_messages,
    f_variables,
    f_icon,
    f_opening_remarks,
    f_input,
    f_create_by,
    f_create_time,
    f_update_by,
    f_update_time,
    f_is_delete
)
SELECT
    f_prompt_id,
    f_prompt_type,
    f_prompt_name,
    f_prompt_desc,
    f_messages,
    f_variables,
    f_icon,
    f_opening_remarks,
    f_input,
    f_create_by,
    f_create_time,
    f_update_by,
    f_update_time,
    f_is_delete
FROM model_management.t_prompt_template_list
WHERE NOT EXISTS (
    SELECT 1 FROM adp.t_prompt_template_list t
    WHERE t.f_prompt_id = model_management.t_prompt_template_list.f_prompt_id
);

-- 迁移 t_model_quota_config 表数据
INSERT INTO adp.t_model_quota_config (
    f_id,
    f_model_id,
    f_billing_type,
    f_input_tokens,
    f_output_tokens,
    f_referprice_in,
    f_referprice_out,
    f_currency_type,
    f_create_time,
    f_update_time,
    f_num_type,
    f_price_type
)
SELECT
    f_id,
    f_model_id,
    f_billing_type,
    f_input_tokens,
    f_output_tokens,
    f_referprice_in,
    f_referprice_out,
    f_currency_type,
    f_create_time,
    f_update_time,
    f_num_type,
    f_price_type
FROM model_management.t_model_quota_config
WHERE NOT EXISTS (
    SELECT 1 FROM adp.t_model_quota_config t
    WHERE t.f_id = model_management.t_model_quota_config.f_id
);

-- 迁移 t_user_quota_config 表数据
INSERT INTO adp.t_user_quota_config (
    f_id,
    f_model_conf,
    f_user_id,
    f_input_tokens,
    f_output_tokens,
    f_create_time,
    f_update_time,
    f_num_type
)
SELECT
    f_id,
    f_model_conf,
    f_user_id,
    f_input_tokens,
    f_output_tokens,
    f_create_time,
    f_update_time,
    f_num_type
FROM model_management.t_user_quota_config
WHERE NOT EXISTS (
    SELECT 1 FROM adp.t_user_quota_config t
    WHERE t.f_id = model_management.t_user_quota_config.f_id
);

