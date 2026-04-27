-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- Action Schedule Management
-- Supports cron-based scheduled action execution with distributed locking

SET SCHEMA adp;

CREATE TABLE IF NOT EXISTS t_action_schedule (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(100 CHAR) NOT NULL DEFAULT '',
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_action_type_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_cron_expression VARCHAR(100 CHAR) NOT NULL DEFAULT '',
  f_unique_identities TEXT DEFAULT NULL,
  f_dynamic_params TEXT DEFAULT NULL,
  f_status VARCHAR(20 CHAR) NOT NULL DEFAULT 'inactive',
  f_last_run_time BIGINT NOT NULL DEFAULT 0,
  f_next_run_time BIGINT NOT NULL DEFAULT 0,
  f_lock_holder VARCHAR(64 CHAR) DEFAULT NULL,
  f_lock_time BIGINT NOT NULL DEFAULT 0,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_id)
);

CREATE INDEX IF NOT EXISTS idx_action_schedule_kn_branch ON t_action_schedule(f_kn_id, f_branch);
CREATE INDEX IF NOT EXISTS idx_action_schedule_status_next_run ON t_action_schedule(f_status, f_next_run_time);
CREATE INDEX IF NOT EXISTS idx_action_schedule_action_type ON t_action_schedule(f_action_type_id);
