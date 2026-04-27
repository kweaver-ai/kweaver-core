SET SCHEMA adp;

CREATE TABLE IF NOT EXISTS "t_python_package" (
  "f_id" VARCHAR(32 CHAR) NOT NULL,
  "f_name" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_oss_id" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_creator_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_creator_name" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  "f_created_at" BIGINT NOT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_python_package_uk_t_python_package_name ON t_python_package("f_name");

