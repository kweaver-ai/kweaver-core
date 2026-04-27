SET SCHEMA adp;

CREATE TABLE IF NOT EXISTS "t_cron_job"
(
    "f_key_id" BIGINT NOT NULL IDENTITY(1, 1),
    "f_job_id" VARCHAR(36 CHAR) NOT NULL,
    "f_job_name" VARCHAR(64 CHAR) NOT NULL,
    "f_job_cron_time" VARCHAR(32 CHAR) NOT NULL,
    "f_job_type" TINYINT NOT NULL,
    "f_job_context" VARCHAR(10240 CHAR),
    "f_tenant_id" VARCHAR(36 CHAR),
    "f_enabled" TINYINT NOT NULL DEFAULT 1,
    "f_remarks" VARCHAR(256 CHAR),
    "f_create_time" BIGINT NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    CLUSTER PRIMARY KEY ("f_key_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_cron_job_index_job_id ON t_cron_job("f_job_id");
CREATE UNIQUE INDEX IF NOT EXISTS t_cron_job_index_job_name ON t_cron_job("f_job_name", "f_tenant_id");
CREATE INDEX IF NOT EXISTS t_cron_job_index_tenant_id ON t_cron_job("f_tenant_id");
CREATE INDEX IF NOT EXISTS t_cron_job_index_time ON t_cron_job("f_create_time", "f_update_time");



CREATE TABLE IF NOT EXISTS "t_cron_job_status"
(
    "f_key_id" BIGINT NOT NULL IDENTITY(1, 1),
    "f_execute_id" VARCHAR(36 CHAR) NOT NULL,
    "f_job_id" VARCHAR(36 CHAR) NOT NULL,
    "f_job_type" TINYINT NOT NULL,
    "f_job_name" VARCHAR(64 CHAR) NOT NULL,
    "f_job_status" TINYINT NOT NULL,
    "f_begin_time" BIGINT,
    "f_end_time" BIGINT,
    "f_executor" VARCHAR(1024 CHAR),
    "f_execute_times" INT,
    "f_ext_info" VARCHAR(1024 CHAR),
    CLUSTER PRIMARY KEY ("f_key_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_cron_job_status_index_execute_id ON t_cron_job_status("f_execute_id");
CREATE INDEX IF NOT EXISTS t_cron_job_status_index_job_id ON t_cron_job_status("f_job_id");
CREATE INDEX IF NOT EXISTS t_cron_job_status_index_job_status ON t_cron_job_status("f_job_status");
CREATE INDEX IF NOT EXISTS t_cron_job_status_index_time ON t_cron_job_status("f_begin_time","f_end_time");