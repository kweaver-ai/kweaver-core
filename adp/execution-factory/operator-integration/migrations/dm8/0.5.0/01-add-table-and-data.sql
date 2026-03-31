SET SCHEMA adp;

CREATE TABLE IF NOT EXISTS "t_skill_repository" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_skill_id" VARCHAR(40 CHAR) NOT NULL,
    "f_name" VARCHAR(255 CHAR) NOT NULL,
    "f_description" text NOT NULL,
    "f_skill_content" text NOT NULL,
    "f_version" VARCHAR(40 CHAR) NOT NULL,
    "f_status" VARCHAR(40 CHAR) NOT NULL,
    "f_source" VARCHAR(50 CHAR) NOT NULL DEFAULT '',
    "f_extend_info" text DEFAULT NULL,
    "f_dependencies" text DEFAULT NULL,
    "f_file_manifest" text DEFAULT NULL,
    "f_create_user" VARCHAR(50 CHAR) NOT NULL,
    "f_create_time" BIGINT NOT NULL,
    "f_update_user" VARCHAR(50 CHAR) NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    "f_delete_user" VARCHAR(50 CHAR) NOT NULL DEFAULT '',
    "f_delete_time" BIGINT NOT NULL DEFAULT 0,
    "f_category" VARCHAR(50 CHAR) DEFAULT '' COMMENT '工具箱分类, 数据处理/算法模型',
    "f_is_deleted" TINYINT DEFAULT 0 COMMENT '是否删除', -- 0: 未删除, 1: 待删除
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_skill_repository_uk_skill_id ON t_skill_repository(f_skill_id);
CREATE INDEX IF NOT EXISTS t_skill_repository_idx_status_update_time ON t_skill_repository(f_status, f_update_time);
CREATE INDEX IF NOT EXISTS t_skill_repository_idx_category_update_time ON t_skill_repository(f_category, f_update_time);
CREATE INDEX IF NOT EXISTS t_skill_repository_idx_create_user_update_time ON t_skill_repository(f_create_user, f_update_time);

CREATE TABLE IF NOT EXISTS "t_skill_file_index" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_skill_id" VARCHAR(40 CHAR) NOT NULL,
    "f_skill_version" VARCHAR(40 CHAR) NOT NULL,
    "f_rel_path" VARCHAR(512 CHAR) NOT NULL,
    "f_path_hash" VARCHAR(32 CHAR) NOT NULL,
    "f_storage_id" VARCHAR(50 CHAR) NOT NULL,
    "f_storage_key" TEXT NOT NULL,
    "f_file_type" VARCHAR(40 CHAR) NOT NULL,
    "f_content_sha256" VARCHAR(64 CHAR) NOT NULL,
    "f_mime_type" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
    "f_size" BIGINT NOT NULL DEFAULT 0,
    "f_create_time" BIGINT NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_skill_file_index_uk_skill_version_rel_path ON t_skill_file_index(f_skill_id, f_skill_version, f_rel_path);
CREATE UNIQUE INDEX IF NOT EXISTS t_skill_file_index_uk_skill_version_path_hash ON t_skill_file_index(f_skill_id, f_skill_version, f_path_hash);
