SET SCHEMA kweaver;

CREATE TABLE IF NOT EXISTS "t_skill_release" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_skill_id" VARCHAR(40 CHAR) NOT NULL,
    "f_name" VARCHAR(512 CHAR) NOT NULL,
    "f_description" text NOT NULL,
    "f_skill_content" text NOT NULL,
    "f_version" VARCHAR(40 CHAR) NOT NULL,
    "f_category" VARCHAR(50 CHAR) DEFAULT '',
    "f_source" VARCHAR(50 CHAR) DEFAULT '',
    "f_extend_info" text DEFAULT NULL,
    "f_dependencies" text DEFAULT NULL,
    "f_file_manifest" text DEFAULT NULL,
    "f_status" VARCHAR(20 CHAR) NOT NULL,
    "f_create_user" VARCHAR(50 CHAR) NOT NULL,
    "f_create_time" BIGINT NOT NULL,
    "f_update_user" VARCHAR(50 CHAR) NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    "f_release_user" VARCHAR(50 CHAR) NOT NULL,
    "f_release_time" BIGINT NOT NULL,
    "f_release_desc" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_skill_release_uk_skill_id ON t_skill_release(f_skill_id);

CREATE INDEX IF NOT EXISTS t_skill_release_idx_status_update_time ON t_skill_release(f_status, f_update_time);

CREATE INDEX IF NOT EXISTS t_skill_release_idx_release_time ON t_skill_release(f_release_time);

CREATE TABLE IF NOT EXISTS "t_skill_release_history" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_skill_id" VARCHAR(40 CHAR) NOT NULL,
    "f_version" VARCHAR(40 CHAR) NOT NULL,
    "f_skill_release" text NOT NULL,
    "f_release_desc" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
    "f_create_user" VARCHAR(50 CHAR) NOT NULL,
    "f_create_time" BIGINT NOT NULL,
    "f_update_user" VARCHAR(50 CHAR) NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_skill_release_history_uk_skill_version ON t_skill_release_history(f_skill_id, f_version);

CREATE INDEX IF NOT EXISTS t_skill_release_history_idx_skill_id_create_time ON t_skill_release_history(f_skill_id, f_create_time);
