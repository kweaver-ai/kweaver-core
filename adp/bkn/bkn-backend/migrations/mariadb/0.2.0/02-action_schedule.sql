-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- Action Schedule Management
-- Supports cron-based scheduled action execution with distributed locking
USE adp;

CREATE TABLE IF NOT EXISTS t_action_schedule (
  f_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'Schedule ID',
  f_name VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'Schedule name',
  f_kn_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'Knowledge network ID',
  f_branch VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'Branch',
  f_action_type_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'Action type ID to execute',
  f_cron_expression VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'Standard 5-field cron expression (min hour dom mon dow)',
  f_unique_identities MEDIUMTEXT DEFAULT NULL COMMENT 'JSON array of target object unique identities',
  f_dynamic_params MEDIUMTEXT DEFAULT NULL COMMENT 'JSON object of dynamic parameters',
  f_status VARCHAR(20) NOT NULL DEFAULT 'inactive' COMMENT 'Schedule status: active or inactive',
  f_last_run_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT 'Last execution timestamp (ms)',
  f_next_run_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT 'Next scheduled run timestamp (ms)',
  f_lock_holder VARCHAR(64) DEFAULT NULL COMMENT 'Pod ID holding execution lock (NULL = unlocked)',
  f_lock_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT 'Lock acquisition timestamp (ms) for timeout detection',
  f_creator VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'Creator ID',
  f_creator_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT 'Creator type',
  f_create_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT 'Create timestamp (ms)',
  f_updater VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'Updater ID',
  f_updater_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT 'Updater type',
  f_update_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT 'Update timestamp (ms)',
  PRIMARY KEY (f_id),
  KEY idx_kn_branch (f_kn_id, f_branch),
  KEY idx_status_next_run (f_status, f_next_run_time),
  KEY idx_action_type (f_action_type_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = 'Action schedule for cron-based execution';
