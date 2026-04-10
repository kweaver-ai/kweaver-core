SET SCHEMA adp;

-- 迁移数据从 dip_data_agent 到 adp schema
-- 注意：使用 WHERE NOT EXISTS 避免重复插入数据

-- 1. t_data_agent_config
INSERT INTO t_data_agent_config (f_id, f_name, f_profile, f_key, f_is_built_in, f_is_system_agent, f_product_key, f_avatar_type, f_avatar, f_created_at, f_created_by, f_updated_at, f_updated_by, f_deleted_by, f_deleted_at, f_config, f_status, f_created_type, f_create_from)
SELECT f_id, f_name, f_profile, f_key, f_is_built_in, f_is_system_agent, f_product_key, f_avatar_type, f_avatar, f_created_at, f_created_by, f_updated_at, f_updated_by, f_deleted_by, f_deleted_at, f_config, f_status, f_created_type, f_create_from
FROM dip_data_agent.t_data_agent_config src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_config dest WHERE dest.f_id = src.f_id);

-- 2. t_data_agent_config_tpl
INSERT INTO t_data_agent_config_tpl (f_id, f_name, f_profile, f_key, f_is_built_in, f_product_key, f_avatar_type, f_avatar, f_created_at, f_created_by, f_updated_at, f_updated_by, f_deleted_by, f_deleted_at, f_config, f_status, f_published_at, f_published_by, f_created_type, f_create_from)
SELECT f_id, f_name, f_profile, f_key, f_is_built_in, f_product_key, f_avatar_type, f_avatar, f_created_at, f_created_by, f_updated_at, f_updated_by, f_deleted_by, f_deleted_at, f_config, f_status, f_published_at, f_published_by, f_created_type, f_create_from
FROM dip_data_agent.t_data_agent_config_tpl src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_config_tpl dest WHERE dest.f_id = src.f_id);

-- 3. t_data_agent_config_tpl_published
INSERT INTO t_data_agent_config_tpl_published (f_id, f_tpl_id, f_name, f_profile, f_key, f_is_built_in, f_product_key, f_avatar_type, f_avatar, f_config, f_published_at, f_published_by)
SELECT f_id, f_tpl_id, f_name, f_profile, f_key, f_is_built_in, f_product_key, f_avatar_type, f_avatar, f_config, f_published_at, f_published_by
FROM dip_data_agent.t_data_agent_config_tpl_published src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_config_tpl_published dest WHERE dest.f_id = src.f_id);

-- 4. t_data_agent_tpl_category_rel
INSERT INTO t_data_agent_tpl_category_rel (f_id, f_published_tpl_id, f_category_id)
SELECT f_id, f_published_tpl_id, f_category_id
FROM dip_data_agent.t_data_agent_tpl_category_rel src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_tpl_category_rel dest WHERE dest.f_id = src.f_id);

-- 5. t_product
-- 注意：init.sql中已有产品数据插入语句，这里只迁移用户自定义数据
INSERT INTO t_product (f_id, f_name, f_profile, f_key, f_created_by, f_created_at, f_updated_by, f_updated_at, f_deleted_by, f_deleted_at)
SELECT f_id, f_name, f_profile, f_key, f_created_by, f_created_at, f_updated_by, f_updated_at, f_deleted_by, f_deleted_at
FROM dip_data_agent.t_product src
WHERE NOT EXISTS (SELECT 1 FROM t_product dest WHERE dest.f_id = src.f_id)
  AND src.f_key NOT IN ('anyshare', 'dip', 'chatbi'); -- 排除已内置的产品

-- 6. t_custom_space
INSERT INTO t_custom_space (f_id, f_name, f_key, f_profile, f_created_by, f_created_at, f_updated_by, f_updated_at, f_deleted_by, f_deleted_at)
SELECT f_id, f_name, f_key, f_profile, f_created_by, f_created_at, f_updated_by, f_updated_at, f_deleted_by, f_deleted_at
FROM dip_data_agent.t_custom_space src
WHERE NOT EXISTS (SELECT 1 FROM t_custom_space dest WHERE dest.f_id = src.f_id);

-- 7. t_custom_space_member
INSERT INTO t_custom_space_member (f_id, f_space_id, f_space_key, f_obj_type, f_obj_id, f_created_at, f_created_by)
SELECT f_id, f_space_id, f_space_key, f_obj_type, f_obj_id, f_created_at, f_created_by
FROM dip_data_agent.t_custom_space_member src
WHERE NOT EXISTS (SELECT 1 FROM t_custom_space_member dest WHERE dest.f_id = src.f_id);

-- 8. t_custom_space_resource
INSERT INTO t_custom_space_resource (f_id, f_space_id, f_space_key, f_resource_type, f_resource_id, f_created_at, f_created_by)
SELECT f_id, f_space_id, f_space_key, f_resource_type, f_resource_id, f_created_at, f_created_by
FROM dip_data_agent.t_custom_space_resource src
WHERE NOT EXISTS (SELECT 1 FROM t_custom_space_resource dest WHERE dest.f_id = src.f_id);

-- 9. t_data_agent_datasource_dataset_assoc
INSERT INTO t_data_agent_datasource_dataset_assoc (f_id, f_agent_id, f_agent_version, f_dataset_id, f_created_at)
SELECT f_id, f_agent_id, f_agent_version, f_dataset_id, f_created_at
FROM dip_data_agent.t_data_agent_datasource_dataset_assoc src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_datasource_dataset_assoc dest WHERE dest.f_id = src.f_id);

-- 10. t_data_agent_datasource_dataset
INSERT INTO t_data_agent_datasource_dataset (f_id, f_hash_sha256, f_created_at)
SELECT f_id, f_hash_sha256, f_created_at
FROM dip_data_agent.t_data_agent_datasource_dataset src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_datasource_dataset dest WHERE dest.f_id = src.f_id);

-- 11. t_data_agent_datasource_dataset_obj
INSERT INTO t_data_agent_datasource_dataset_obj (f_id, f_dataset_id, f_object_id, f_object_type, f_created_at)
SELECT f_id, f_dataset_id, f_object_id, f_object_type, f_created_at
FROM dip_data_agent.t_data_agent_datasource_dataset_obj src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_datasource_dataset_obj dest WHERE dest.f_id = src.f_id);

-- 12. t_data_agent_release
INSERT INTO t_data_agent_release (f_id, f_agent_id, f_agent_name, f_agent_config, f_agent_version, f_agent_desc, f_is_api_agent, f_is_web_sdk_agent, f_is_skill_agent, f_is_data_flow_agent, f_is_to_custom_space, f_is_to_square, f_is_pms_ctrl, f_create_time, f_update_time, f_create_by, f_update_by)
SELECT f_id, f_agent_id, f_agent_name, f_agent_config, f_agent_version, f_agent_desc, f_is_api_agent, f_is_web_sdk_agent, f_is_skill_agent, f_is_data_flow_agent, f_is_to_custom_space, f_is_to_square, f_is_pms_ctrl, f_create_time, f_update_time, f_create_by, f_update_by
FROM dip_data_agent.t_data_agent_release src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_release dest WHERE dest.f_id = src.f_id);

-- 13. t_data_agent_release_history
INSERT INTO t_data_agent_release_history (f_id, f_agent_id, f_agent_config, f_agent_version, f_agent_desc, f_create_time, f_update_time, f_create_by, f_update_by)
SELECT f_id, f_agent_id, f_agent_config, f_agent_version, f_agent_desc, f_create_time, f_update_time, f_create_by, f_update_by
FROM dip_data_agent.t_data_agent_release_history src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_release_history dest WHERE dest.f_id = src.f_id);

-- 14. t_data_agent_release_category
-- 注意：init.sql中已有分类数据插入语句，这里只迁移用户自定义数据
INSERT INTO t_data_agent_release_category (f_id, f_name, f_description, f_create_time, f_update_time, f_create_by, f_update_by)
SELECT f_id, f_name, f_description, f_create_time, f_update_time, f_create_by, f_update_by
FROM dip_data_agent.t_data_agent_release_category src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_release_category dest WHERE dest.f_id = src.f_id)
  AND src.f_id NOT IN ('01JRYRKP0M8VYHQSX4FXR5CKGX', '01JRYRKP0M8VYHQSX4FXR5CKGY', '01JRYRKP0M8VYHQSX4FXR5CKGZ', '01JRYRKP0M8VYHQSX4FXR5CKG1', '01JRYRKP0M8VYHQSX4FXR5CKG2', '01JRYRKP0M8VYHQSX4FXR5CKG3', '01JRYRKP0M8VYHQSX4FXR5CKG4', '01JRYRKP0M8VYHQSX4FXR5CKG5'); -- 排除已内置的分类

-- 15. t_data_agent_release_category_rel
INSERT INTO t_data_agent_release_category_rel (f_id, f_release_id, f_category_id)
SELECT f_id, f_release_id, f_category_id
FROM dip_data_agent.t_data_agent_release_category_rel src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_release_category_rel dest WHERE dest.f_id = src.f_id);

-- 16. t_data_agent_release_permission
INSERT INTO t_data_agent_release_permission (f_id, f_release_id, f_obj_type, f_obj_id)
SELECT f_id, f_release_id, f_obj_type, f_obj_id
FROM dip_data_agent.t_data_agent_release_permission src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_release_permission dest WHERE dest.f_id = src.f_id);

-- 17. t_data_agent_visit_history
INSERT INTO t_data_agent_visit_history (f_id, f_agent_id, f_agent_version, f_custom_space_id, f_visit_count, f_create_time, f_update_time, f_create_by, f_update_by)
SELECT f_id, f_agent_id, f_agent_version, f_custom_space_id, f_visit_count, f_create_time, f_update_time, f_create_by, f_update_by
FROM dip_data_agent.t_data_agent_visit_history src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_visit_history dest WHERE dest.f_id = src.f_id);

-- 18. t_biz_domain_agent_rel
INSERT INTO t_biz_domain_agent_rel (f_id, f_biz_domain_id, f_agent_id, f_created_at)
SELECT f_id, f_biz_domain_id, f_agent_id, f_created_at
FROM dip_data_agent.t_biz_domain_agent_rel src
WHERE NOT EXISTS (SELECT 1 FROM t_biz_domain_agent_rel dest WHERE dest.f_id = src.f_id);

-- 19. t_biz_domain_agent_tpl_rel
INSERT INTO t_biz_domain_agent_tpl_rel (f_id, f_biz_domain_id, f_agent_tpl_id, f_created_at)
SELECT f_id, f_biz_domain_id, f_agent_tpl_id, f_created_at
FROM dip_data_agent.t_biz_domain_agent_tpl_rel src
WHERE NOT EXISTS (SELECT 1 FROM t_biz_domain_agent_tpl_rel dest WHERE dest.f_id = src.f_id);

