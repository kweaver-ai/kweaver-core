SET SCHEMA kweaver;

-- 迁移数据从 adp 到 kweaver schema
-- 注意：使用 WHERE NOT EXISTS 避免重复插入数据

-- 1. t_data_agent_config
INSERT INTO t_data_agent_config (f_id, f_name, f_profile, f_key, f_is_built_in, f_is_system_agent, f_product_key, f_avatar_type, f_avatar, f_created_at, f_created_by, f_updated_at, f_updated_by, f_deleted_by, f_deleted_at, f_config, f_status, f_created_type, f_create_from)
SELECT f_id, f_name, f_profile, f_key, f_is_built_in, f_is_system_agent, f_product_key, f_avatar_type, f_avatar, f_created_at, f_created_by, f_updated_at, f_updated_by, f_deleted_by, f_deleted_at, f_config, f_status, f_created_type, f_create_from
FROM adp.t_data_agent_config src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_config dest WHERE dest.f_id = src.f_id);

-- 2. t_data_agent_config_tpl
INSERT INTO t_data_agent_config_tpl (f_id, f_name, f_profile, f_key, f_is_built_in, f_product_key, f_avatar_type, f_avatar, f_created_at, f_created_by, f_updated_at, f_updated_by, f_deleted_by, f_deleted_at, f_config, f_status, f_published_at, f_published_by, f_created_type, f_create_from)
SELECT f_id, f_name, f_profile, f_key, f_is_built_in, f_product_key, f_avatar_type, f_avatar, f_created_at, f_created_by, f_updated_at, f_updated_by, f_deleted_by, f_deleted_at, f_config, f_status, f_published_at, f_published_by, f_created_type, f_create_from
FROM adp.t_data_agent_config_tpl src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_config_tpl dest WHERE dest.f_id = src.f_id);

-- 3. t_data_agent_config_tpl_published
INSERT INTO t_data_agent_config_tpl_published (f_id, f_tpl_id, f_name, f_profile, f_key, f_is_built_in, f_product_key, f_avatar_type, f_avatar, f_config, f_published_at, f_published_by)
SELECT f_id, f_tpl_id, f_name, f_profile, f_key, f_is_built_in, f_product_key, f_avatar_type, f_avatar, f_config, f_published_at, f_published_by
FROM adp.t_data_agent_config_tpl_published src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_config_tpl_published dest WHERE dest.f_id = src.f_id);

-- 4. t_data_agent_tpl_category_rel
INSERT INTO t_data_agent_tpl_category_rel (f_id, f_published_tpl_id, f_category_id)
SELECT f_id, f_published_tpl_id, f_category_id
FROM adp.t_data_agent_tpl_category_rel src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_tpl_category_rel dest WHERE dest.f_id = src.f_id);

-- 5. t_product
-- 注意：init.sql中已有产品数据插入语句，这里只迁移用户自定义数据
INSERT INTO t_product (f_id, f_name, f_profile, f_key, f_created_by, f_created_at, f_updated_by, f_updated_at, f_deleted_by, f_deleted_at)
SELECT f_id, f_name, f_profile, f_key, f_created_by, f_created_at, f_updated_by, f_updated_at, f_deleted_by, f_deleted_at
FROM adp.t_product src
WHERE NOT EXISTS (SELECT 1 FROM t_product dest WHERE dest.f_id = src.f_id)
  AND src.f_key NOT IN ('anyshare', 'dip', 'chatbi'); -- 排除已内置的产品

-- 6. t_custom_space
INSERT INTO t_custom_space (f_id, f_name, f_key, f_profile, f_created_by, f_created_at, f_updated_by, f_updated_at, f_deleted_by, f_deleted_at)
SELECT f_id, f_name, f_key, f_profile, f_created_by, f_created_at, f_updated_by, f_updated_at, f_deleted_by, f_deleted_at
FROM adp.t_custom_space src
WHERE NOT EXISTS (SELECT 1 FROM t_custom_space dest WHERE dest.f_id = src.f_id);

-- 7. t_custom_space_member
INSERT INTO t_custom_space_member (f_id, f_space_id, f_space_key, f_obj_type, f_obj_id, f_created_at, f_created_by)
SELECT f_id, f_space_id, f_space_key, f_obj_type, f_obj_id, f_created_at, f_created_by
FROM adp.t_custom_space_member src
WHERE NOT EXISTS (SELECT 1 FROM t_custom_space_member dest WHERE dest.f_id = src.f_id);

-- 8. t_custom_space_resource
INSERT INTO t_custom_space_resource (f_id, f_space_id, f_space_key, f_resource_type, f_resource_id, f_created_at, f_created_by)
SELECT f_id, f_space_id, f_space_key, f_resource_type, f_resource_id, f_created_at, f_created_by
FROM adp.t_custom_space_resource src
WHERE NOT EXISTS (SELECT 1 FROM t_custom_space_resource dest WHERE dest.f_id = src.f_id);

-- 9. t_data_agent_release
INSERT INTO t_data_agent_release (f_id, f_agent_id, f_agent_name, f_agent_config, f_agent_version, f_agent_desc, f_is_api_agent, f_is_web_sdk_agent, f_is_skill_agent, f_is_data_flow_agent, f_is_to_custom_space, f_is_to_square, f_is_pms_ctrl, f_create_time, f_update_time, f_create_by, f_update_by)
SELECT f_id, f_agent_id, f_agent_name, f_agent_config, f_agent_version, f_agent_desc, f_is_api_agent, f_is_web_sdk_agent, f_is_skill_agent, f_is_data_flow_agent, f_is_to_custom_space, f_is_to_square, f_is_pms_ctrl, f_create_time, f_update_time, f_create_by, f_update_by
FROM adp.t_data_agent_release src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_release dest WHERE dest.f_id = src.f_id);

-- 10. t_data_agent_release_history
INSERT INTO t_data_agent_release_history (f_id, f_agent_id, f_agent_config, f_agent_version, f_agent_desc, f_create_time, f_update_time, f_create_by, f_update_by)
SELECT f_id, f_agent_id, f_agent_config, f_agent_version, f_agent_desc, f_create_time, f_update_time, f_create_by, f_update_by
FROM adp.t_data_agent_release_history src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_release_history dest WHERE dest.f_id = src.f_id);

-- 11. t_data_agent_release_category
-- 注意：init.sql中已有分类数据插入语句，这里只迁移用户自定义数据
INSERT INTO t_data_agent_release_category (f_id, f_name, f_description, f_create_time, f_update_time, f_create_by, f_update_by)
SELECT f_id, f_name, f_description, f_create_time, f_update_time, f_create_by, f_update_by
FROM adp.t_data_agent_release_category src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_release_category dest WHERE dest.f_id = src.f_id)
  AND src.f_id NOT IN ('01JRYRKP0M8VYHQSX4FXR5CKGX', '01JRYRKP0M8VYHQSX4FXR5CKGY', '01JRYRKP0M8VYHQSX4FXR5CKGZ', '01JRYRKP0M8VYHQSX4FXR5CKG1', '01JRYRKP0M8VYHQSX4FXR5CKG2', '01JRYRKP0M8VYHQSX4FXR5CKG3', '01JRYRKP0M8VYHQSX4FXR5CKG4', '01JRYRKP0M8VYHQSX4FXR5CKG5'); -- 排除已内置的分类

-- 12. t_data_agent_release_category_rel
INSERT INTO t_data_agent_release_category_rel (f_id, f_release_id, f_category_id)
SELECT f_id, f_release_id, f_category_id
FROM adp.t_data_agent_release_category_rel src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_release_category_rel dest WHERE dest.f_id = src.f_id);

-- 13. t_data_agent_release_permission
INSERT INTO t_data_agent_release_permission (f_id, f_release_id, f_obj_type, f_obj_id)
SELECT f_id, f_release_id, f_obj_type, f_obj_id
FROM adp.t_data_agent_release_permission src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_release_permission dest WHERE dest.f_id = src.f_id);

-- 14. t_data_agent_visit_history
INSERT INTO t_data_agent_visit_history (f_id, f_agent_id, f_agent_version, f_custom_space_id, f_visit_count, f_create_time, f_update_time, f_create_by, f_update_by)
SELECT f_id, f_agent_id, f_agent_version, f_custom_space_id, f_visit_count, f_create_time, f_update_time, f_create_by, f_update_by
FROM adp.t_data_agent_visit_history src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_visit_history dest WHERE dest.f_id = src.f_id);

-- 15. t_biz_domain_agent_rel
INSERT INTO t_biz_domain_agent_rel (f_id, f_biz_domain_id, f_agent_id, f_created_at)
SELECT f_id, f_biz_domain_id, f_agent_id, f_created_at
FROM adp.t_biz_domain_agent_rel src
WHERE NOT EXISTS (SELECT 1 FROM t_biz_domain_agent_rel dest WHERE dest.f_id = src.f_id);

-- 16. t_biz_domain_agent_tpl_rel
INSERT INTO t_biz_domain_agent_tpl_rel (f_id, f_biz_domain_id, f_agent_tpl_id, f_created_at)
SELECT f_id, f_biz_domain_id, f_agent_tpl_id, f_created_at
FROM adp.t_biz_domain_agent_tpl_rel src
WHERE NOT EXISTS (SELECT 1 FROM t_biz_domain_agent_tpl_rel dest WHERE dest.f_id = src.f_id);

-- 17. t_data_agent_conversation
INSERT INTO t_data_agent_conversation (f_id, f_agent_app_key, f_title, f_origin, f_message_index, f_read_message_index, f_ext, f_create_time, f_update_time, f_create_by, f_update_by, f_is_deleted)
SELECT f_id, f_agent_app_key, f_title, f_origin, f_message_index, f_read_message_index, f_ext, f_create_time, f_update_time, f_create_by, f_update_by, f_is_deleted
FROM adp.t_data_agent_conversation src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_conversation dest WHERE dest.f_id = src.f_id);

-- 18. t_data_agent_conversation_message
INSERT INTO t_data_agent_conversation_message (f_id, f_agent_app_key, f_conversation_id, f_agent_id, f_agent_version, f_reply_id, f_index, f_role, f_content, f_content_type, f_status, f_ext, f_create_time, f_update_time, f_create_by, f_update_by, f_is_deleted)
SELECT f_id, f_agent_app_key, f_conversation_id, f_agent_id, f_agent_version, f_reply_id, f_index, f_role, f_content, f_content_type, f_status, f_ext, f_create_time, f_update_time, f_create_by, f_update_by, f_is_deleted
FROM adp.t_data_agent_conversation_message src
WHERE NOT EXISTS (SELECT 1 FROM t_data_agent_conversation_message dest WHERE dest.f_id = src.f_id);
