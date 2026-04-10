SET SCHEMA kweaver;

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

-- t_python_package
INSERT INTO kweaver."t_python_package" ("f_id", "f_name", "f_oss_id", "f_oss_key", "f_creator_id", "f_creator_name", "f_created_at")
SELECT s."f_id", s."f_name", s."f_oss_id", s."f_oss_key", s."f_creator_id", s."f_creator_name", s."f_created_at"
FROM adp."t_python_package" s;

