SET SCHEMA adp;

CREATE TABLE IF NOT EXISTS "t_data_agent_memory_history" (
  "f_id"          VARCHAR(40 CHAR) NOT NULL,
  "f_memory_id"   VARCHAR(40 CHAR) NOT NULL,
  "f_old_memory"  text,
  "f_new_memory"  text,
  "f_event"       VARCHAR(40 CHAR) NOT NULL,
  "f_created_at"  VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_updated_at"  VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_actor_id"    VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_role"        VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_create_by"   VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_update_by"   VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_is_deleted"  TINYINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_data_agent_memory_history_idx_memory_id ON t_data_agent_memory_history("f_memory_id");
