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


CREATE TABLE IF NOT EXISTS "t_skill_index_build_task" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_task_id" VARCHAR(40 CHAR) NOT NULL,
    "f_status" VARCHAR(20 CHAR) NOT NULL,
    "f_execute_type" VARCHAR(20 CHAR) NOT NULL,
    "f_total_count" BIGINT NOT NULL DEFAULT 0,
    "f_success_count" BIGINT NOT NULL DEFAULT 0,
    "f_delete_count" BIGINT NOT NULL DEFAULT 0,
    "f_failed_count" BIGINT NOT NULL DEFAULT 0,
    "f_retry_count" BIGINT NOT NULL DEFAULT 0,
    "f_max_retry" BIGINT NOT NULL DEFAULT 0,
    "f_cursor_update_time" BIGINT NOT NULL DEFAULT 0,
    "f_cursor_skill_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    "f_error_msg" text DEFAULT NULL,
    "f_create_user" VARCHAR(50 CHAR) NOT NULL,
    "f_create_time" BIGINT NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    "f_last_finished_time" BIGINT NOT NULL DEFAULT 0,
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_skill_index_build_task_uk_task_id ON t_skill_index_build_task(f_task_id);
CREATE INDEX IF NOT EXISTS t_skill_index_build_task_idx_status_create_time ON t_skill_index_build_task(f_status, f_create_time);
CREATE INDEX IF NOT EXISTS t_skill_index_build_task_idx_exec_status_finish_time ON t_skill_index_build_task(f_execute_type, f_status, f_last_finished_time);
