SET SCHEMA adp;

CREATE TABLE IF NOT EXISTS "t_resource_deploy" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_resource_id" VARCHAR(40 CHAR) NOT NULL,
    "f_type" VARCHAR(40 CHAR) NOT NULL,
    "f_version" INT NOT NULL,
    "f_name" VARCHAR(40 CHAR) NOT NULL,
    "f_description" text NOT NULL,
    "f_config" text NOT NULL,
    "f_status" VARCHAR(40 CHAR) NOT NULL,
    "f_create_user" VARCHAR(50 CHAR) NOT NULL,
    "f_create_time" BIGINT NOT NULL,
    "f_update_user" VARCHAR(50 CHAR) NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_resource_deploy_uk_resource_id ON t_resource_deploy(f_resource_id, f_type, f_version);