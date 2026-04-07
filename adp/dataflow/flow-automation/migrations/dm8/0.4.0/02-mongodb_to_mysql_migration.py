#!/usr/bin/env python3
from __future__ import annotations

import hashlib
import json
import logging
import os
import sys
from dataclasses import dataclass, field
from typing import Any, Callable, Dict, Iterator, List, Sequence
import rdsdriver
from pymongo import MongoClient

logger = logging.getLogger(__name__)

DEFAULT_BATCH_SIZE = 1000
MIN_STABLE_ID = 100_000_000_000_000_000
STABLE_ID_RANGE = 900_000_000_000_000_000


def configure_logging() -> None:
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s - %(levelname)s - %(message)s",
        handlers=[logging.StreamHandler(sys.stdout)],
    )


def normalize_value(value: Any) -> Any:
    if value is None:
        return ""
    if isinstance(value, (dict, list)):
        return json.dumps(value, ensure_ascii=False, default=str)
    if isinstance(value, (tuple, set)):
        return json.dumps(list(value), ensure_ascii=False, default=str)
    if isinstance(value, (str, int, float, bool)):
        return value
    if hasattr(value, "isoformat"):
        return value.isoformat()
    return str(value)


def string_value(value: Any) -> str:
    normalized = normalize_value(value)
    if normalized == "":
        return ""
    if isinstance(normalized, bool):
        return "true" if normalized else "false"
    return str(normalized)


def int_value(value: Any, default: int = 0) -> int:
    if value is None or value == "":
        return default
    if isinstance(value, bool):
        return int(value)
    if isinstance(value, (int, float)):
        return int(value)
    text = str(value).strip()
    if not text:
        return default
    return int(text)


def bool_value(value: Any) -> bool:
    if isinstance(value, bool):
        return value
    if value in (1, "1", "true", "True", "yes", "on"):
        return True
    return False


def uint64_value(value: Any) -> int:
    result = int_value(value)
    if result < 0:
        raise ValueError(f"negative unsigned integer: {value}")
    return result


def stable_uint64(*parts: Any) -> int:
    raw = "::".join(string_value(part) for part in parts)
    digest = hashlib.sha1(raw.encode("utf-8")).digest()
    value = int.from_bytes(digest[:8], "big", signed=False)
    return MIN_STABLE_ID + (value % STABLE_ID_RANGE)


def primary_id(value: Any, *fallback_parts: Any) -> int:
    if value not in (None, ""):
        try:
            return uint64_value(value)
        except Exception:
            pass
    return stable_uint64(*fallback_parts)


def pick(document: Dict[str, Any], *keys: str, default: Any = None) -> Any:
    for key in keys:
        if key in document and document[key] is not None:
            return document[key]
    return default


def collection_name(prefix: str, suffix: str) -> str:
    clean_prefix = prefix.strip("_")
    return f"{clean_prefix}_{suffix}" if clean_prefix else suffix


def chunked(items: Sequence[Any], size: int) -> Iterator[Sequence[Any]]:
    for index in range(0, len(items), size):
        yield items[index:index + size]


def has_non_empty_command(value: Any) -> bool:
    if value is None:
        return False
    if isinstance(value, str):
        return value.strip() != ""
    if isinstance(value, (dict, list, tuple, set)):
        return bool(value)
    return True


def batch_run_id_from_vars(vars_data: Any) -> str:
    if not isinstance(vars_data, dict):
        return ""
    value = vars_data.get("batch_run_id")
    if isinstance(value, dict):
        return string_value(value.get("value"))
    return string_value(value)


def build_dag_row(document: Dict[str, Any]) -> Dict[str, Any]:
    return {
        "f_id": primary_id(document.get("_id"), "dag", document.get("_id")),
        "f_created_at": int_value(document.get("createdAt")),
        "f_updated_at": int_value(document.get("updatedAt")),
        "f_user_id": string_value(document.get("userid")),
        "f_name": string_value(document.get("name")),
        "f_desc": string_value(document.get("desc")),
        "f_trigger": string_value(document.get("trigger")),
        "f_cron": string_value(document.get("cron")),
        "f_vars": string_value(document.get("vars")),
        "f_status": string_value(document.get("status")),
        "f_tasks": string_value(document.get("tasks")),
        "f_steps": string_value(document.get("steps")),
        "f_description": string_value(document.get("description")),
        "f_shortcuts": string_value(document.get("shortcuts")),
        "f_accessors": string_value(document.get("accessors")),
        "f_type": string_value(document.get("type")),
        "f_policy_type": string_value(document.get("policy_type")),
        "f_appinfo": string_value(document.get("appinfo")),
        "f_priority": string_value(document.get("priority")),
        "f_removed": bool_value(document.get("removed")),
        "f_emails": string_value(document.get("emails")),
        "f_template": string_value(document.get("template")),
        "f_published": bool_value(pick(document, "publish", "published", default=False)),
        "f_trigger_config": string_value(document.get("trigger_config")),
        "f_sub_ids": string_value(document.get("sub_ids")),
        "f_exec_mode": string_value(document.get("exec_mode")),
        "f_category": string_value(document.get("category")),
        "f_outputs": string_value(document.get("outputs")),
        "f_instructions": string_value(document.get("instructions")),
        "f_operator_id": string_value(document.get("operator_id")),
        "f_inc_values": string_value(document.get("inc_values")),
        "f_version": string_value(document.get("version")),
        "f_version_id": string_value(document.get("versionId")),
        "f_modify_by": string_value(document.get("modify_by")),
        "f_is_debug": bool_value(document.get("is_debug")),
        "f_debug_id": string_value(document.get("debug_id")),
        "f_biz_domain_id": string_value(document.get("biz_domain_id")),
    }


def build_dag_var_rows(document: Dict[str, Any]) -> List[Dict[str, Any]]:
    dag_id = primary_id(document.get("_id"), "dag", document.get("_id"))
    vars_data = document.get("vars")
    if not isinstance(vars_data, dict):
        return []

    rows: List[Dict[str, Any]] = []
    for var_name, payload in vars_data.items():
        payload = payload if isinstance(payload, dict) else {}
        rows.append(
            {
                "f_id": stable_uint64("dag_var", dag_id, var_name),
                "f_dag_id": dag_id,
                "f_var_name": string_value(var_name),
                "f_default_value": string_value(payload.get("defaultValue")),
                "f_var_type": "string",
                "f_description": string_value(payload.get("desc")),
            }
        )
    return rows


def build_dag_step_rows(document: Dict[str, Any]) -> List[Dict[str, Any]]:
    dag_id = primary_id(document.get("_id"), "dag", document.get("_id"))
    rows: List[Dict[str, Any]] = []

    def walk(steps: Any, path_prefix: str) -> None:
        if not isinstance(steps, list):
            return

        for index, step in enumerate(steps):
            if not isinstance(step, dict):
                continue

            step_id = string_value(step.get("id")) or str(index)
            path = f"{path_prefix}.{index}.{step_id}" if path_prefix else f"{index}.{step_id}"
            operator = string_value(step.get("operator"))
            has_datasource = isinstance(step.get("dataSource"), dict)

            if operator or has_datasource:
                rows.append(
                    {
                        "f_id": stable_uint64("dag_step", dag_id, path, operator, has_datasource),
                        "f_dag_id": dag_id,
                        "f_operator": operator,
                        "f_source_id": "",
                        "f_has_datasource": has_datasource,
                    }
                )

            walk(step.get("steps"), f"{path}.steps")

            branches = step.get("branches")
            if isinstance(branches, list):
                for branch_index, branch in enumerate(branches):
                    if isinstance(branch, dict):
                        walk(branch.get("steps"), f"{path}.branches.{branch_index}")

    walk(document.get("steps"), "")
    return rows


def build_dag_accessor_rows(document: Dict[str, Any]) -> List[Dict[str, Any]]:
    dag_id = primary_id(document.get("_id"), "dag", document.get("_id"))
    accessors = document.get("accessors")
    if not isinstance(accessors, list):
        return []

    rows: List[Dict[str, Any]] = []
    for accessor in accessors:
        if not isinstance(accessor, dict):
            continue
        accessor_id = string_value(accessor.get("id"))
        if not accessor_id:
            continue
        rows.append(
            {
                "f_id": stable_uint64("dag_accessor", dag_id, accessor_id),
                "f_dag_id": dag_id,
                "f_accessor_id": accessor_id,
            }
        )
    return rows


def build_dag_version_row(document: Dict[str, Any]) -> Dict[str, Any]:
    return {
        "f_id": primary_id(document.get("_id"), "dag_version", document.get("_id")),
        "f_created_at": int_value(document.get("createdAt")),
        "f_updated_at": int_value(document.get("updatedAt")),
        "f_dag_id": string_value(document.get("dagId")),
        "f_user_id": string_value(document.get("userid")),
        "f_version": string_value(document.get("version")),
        "f_version_id": string_value(document.get("versionId")),
        "f_change_log": string_value(document.get("changeLog")),
        "f_config": string_value(document.get("config")),
        "f_sort_time": int_value(document.get("sortTime")),
    }


def build_dag_instance_row(document: Dict[str, Any]) -> Dict[str, Any]:
    vars_data = document.get("vars")
    command = document.get("cmd")
    return {
        "f_id": primary_id(document.get("_id"), "dag_instance", document.get("_id")),
        "f_created_at": int_value(document.get("createdAt")),
        "f_updated_at": int_value(document.get("updatedAt")),
        "f_dag_id": primary_id(document.get("dagId"), "dag_ref", document.get("dagId"), document.get("_id")),
        "f_trigger": string_value(document.get("trigger")),
        "f_worker": string_value(document.get("worker")),
        "f_source": string_value(document.get("source")),
        "f_vars": string_value(vars_data),
        "f_keywords": string_value(document.get("keywords")),
        "f_event_persistence": int_value(document.get("eventPersistence")),
        "f_event_oss_path": string_value(document.get("eventOssPath")),
        "f_share_data": string_value(document.get("shareData")),
        "f_share_data_ext": string_value(document.get("shareDataExt")),
        "f_status": string_value(document.get("status")),
        "f_reason": string_value(document.get("reason")),
        "f_cmd": string_value(command),
        "f_has_cmd": has_non_empty_command(command),
        "f_batch_run_id": batch_run_id_from_vars(vars_data),
        "f_user_id": string_value(document.get("userid")),
        "f_ended_at": int_value(document.get("endedAt")),
        "f_dag_type": string_value(document.get("dag_type")),
        "f_policy_type": string_value(document.get("policy_type")),
        "f_appinfo": string_value(document.get("appinfo")),
        "f_priority": string_value(document.get("priority")),
        "f_mode": int_value(document.get("mode")),
        "f_dump": string_value(document.get("dump")),
        "f_dump_ext": string_value(document.get("dumpExt")),
        "f_success_callback": string_value(document.get("success_callback")),
        "f_error_callback": string_value(document.get("error_callback")),
        "f_call_chain": string_value(document.get("call_chain")),
        "f_resume_data": string_value(document.get("resume_data")),
        "f_resume_status": string_value(document.get("resume_status")),
        "f_version": string_value(document.get("version")),
        "f_version_id": string_value(document.get("versionId")),
        "f_biz_domain_id": string_value(document.get("biz_domain_id")),
    }


def build_dag_instance_keyword_rows(document: Dict[str, Any]) -> List[Dict[str, Any]]:
    dag_instance_id = primary_id(document.get("_id"), "dag_instance", document.get("_id"))
    keywords = document.get("keywords")
    if not isinstance(keywords, list):
        return []

    rows: List[Dict[str, Any]] = []
    for keyword in keywords:
        keyword_text = string_value(keyword)
        if not keyword_text:
            continue
        rows.append(
            {
                "f_id": stable_uint64("dag_instance_keyword", dag_instance_id, keyword_text),
                "f_dag_ins_id": dag_instance_id,
                "f_keyword": keyword_text,
            }
        )
    return rows


def build_task_instance_row(document: Dict[str, Any]) -> Dict[str, Any]:
    updated_at = int_value(document.get("updatedAt"))
    timeout_secs = int_value(document.get("timeoutSecs"))
    return {
        "f_id": primary_id(document.get("_id"), "task_instance", document.get("_id")),
        "f_created_at": int_value(document.get("createdAt")),
        "f_updated_at": updated_at,
        "f_expired_at": updated_at + timeout_secs,
        "f_task_id": string_value(document.get("taskId")),
        "f_dag_ins_id": primary_id(
            pick(document, "dagInsId", "dagInsID", default=None),
            "task_dag_instance",
            document.get("_id"),
        ),
        "f_name": string_value(document.get("name")),
        "f_depend_on": string_value(document.get("dependOn")),
        "f_action_name": string_value(document.get("actionName")),
        "f_timeout_secs": timeout_secs,
        "f_params": string_value(document.get("params")),
        "f_traces": string_value(document.get("traces")),
        "f_status": string_value(document.get("status")),
        "f_reason": string_value(document.get("reason")),
        "f_pre_checks": string_value(pick(document, "preChecks", "preCheck")),
        "f_results": string_value(document.get("results")),
        "f_steps": string_value(document.get("steps")),
        "f_last_modified_at": int_value(document.get("lastModifiedAt")),
        "f_rendered_params": string_value(document.get("renderedParams")),
        "f_hash": string_value(document.get("hash")),
        "f_settings": string_value(document.get("settings")),
        "f_metadata": string_value(document.get("metadata")),
    }


def build_token_row(document: Dict[str, Any]) -> Dict[str, Any]:
    return {
        "f_id": primary_id(document.get("_id"), "token", document.get("_id")),
        "f_created_at": int_value(document.get("createdAt")),
        "f_updated_at": int_value(document.get("updatedAt")),
        "f_user_id": string_value(document.get("userid")),
        "f_user_name": string_value(document.get("username")),
        "f_refresh_token": string_value(document.get("refresh_token")),
        "f_token": string_value(document.get("token")),
        "f_expires_in": int_value(document.get("expires_in")),
        "f_login_ip": string_value(document.get("login_ip")),
        "f_is_app": bool_value(document.get("isapp")),
    }


def build_inbox_row(document: Dict[str, Any]) -> Dict[str, Any]:
    return {
        "f_id": primary_id(document.get("_id"), "inbox", document.get("_id")),
        "f_created_at": int_value(document.get("createdAt")),
        "f_updated_at": int_value(document.get("updatedAt")),
        "f_msg": string_value(document.get("msg")),
        "f_topic": string_value(document.get("topic")),
        "f_docid": string_value(document.get("docid")),
        "f_dag": string_value(pick(document, "dag", "dags")),
    }


def build_client_row(document: Dict[str, Any]) -> Dict[str, Any]:
    return {
        "f_id": primary_id(
            document.get("_id"),
            "client",
            document.get("_id"),
            document.get("client_name"),
            document.get("client_id"),
        ),
        "f_created_at": int_value(document.get("createdAt")),
        "f_updated_at": int_value(document.get("updatedAt")),
        "f_client_name": string_value(document.get("client_name")),
        "f_client_id": string_value(document.get("client_id")),
        "f_client_secret": string_value(document.get("client_secret")),
    }


def build_switch_row(document: Dict[str, Any]) -> Dict[str, Any]:
    return {
        "f_id": primary_id(document.get("_id"), "switch", document.get("_id")),
        "f_created_at": int_value(document.get("createdAt")),
        "f_updated_at": int_value(document.get("updatedAt")),
        "f_name": string_value(document.get("name")),
        "f_status": bool_value(document.get("status")),
    }


def build_log_row(document: Dict[str, Any]) -> Dict[str, Any]:
    return {
        "f_id": primary_id(document.get("_id"), "log", document.get("_id")),
        "f_created_at": int_value(document.get("createdAt")),
        "f_updated_at": int_value(document.get("updatedAt")),
        "f_ossid": string_value(document.get("ossid")),
        "f_key": string_value(document.get("key")),
        "f_filename": string_value(document.get("filename")),
    }


def build_outbox_row(document: Dict[str, Any]) -> Dict[str, Any]:
    return {
        "f_id": primary_id(document.get("_id"), "outbox", document.get("_id")),
        "f_created_at": int_value(document.get("createdAt")),
        "f_updated_at": int_value(document.get("updatedAt")),
        "f_msg": string_value(document.get("msg")),
        "f_topic": string_value(document.get("topic")),
    }


@dataclass(frozen=True)
class ChildMapping:
    table_name: str
    pk_column: str
    build_rows: Callable[[Dict[str, Any]], List[Dict[str, Any]]]


@dataclass(frozen=True)
class Mapping:
    name: str
    collection_suffix: str
    table_name: str
    pk_column: str
    build_row: Callable[[Dict[str, Any]], Dict[str, Any]]
    child_mappings: Sequence[ChildMapping] = field(default_factory=tuple)


MAPPINGS: Sequence[Mapping] = (
    Mapping(
        name="dag",
        collection_suffix="dag",
        table_name="t_flow_dag",
        pk_column="f_id",
        build_row=build_dag_row,
        child_mappings=(
            ChildMapping("t_flow_dag_var", "f_id", build_dag_var_rows),
            ChildMapping("t_flow_dag_step", "f_id", build_dag_step_rows),
            ChildMapping("t_flow_dag_accessor", "f_id", build_dag_accessor_rows),
        ),
    ),
    Mapping("dag_version", "dag_version", "t_flow_dag_version", "f_id", build_dag_version_row),
    Mapping(
        name="dag_instance",
        collection_suffix="dag_instance",
        table_name="t_flow_dag_instance",
        pk_column="f_id",
        build_row=build_dag_instance_row,
        child_mappings=(ChildMapping("t_flow_dag_instance_keyword", "f_id", build_dag_instance_keyword_rows),),
    ),
    Mapping("task_instance", "task_instance", "t_flow_task_instance", "f_id", build_task_instance_row),
    Mapping("token", "token", "t_flow_token", "f_id", build_token_row),
    Mapping("inbox", "inbox", "t_flow_inbox", "f_id", build_inbox_row),
    Mapping("client", "client", "t_flow_client", "f_id", build_client_row),
    Mapping("switch", "switch", "t_flow_switch", "f_id", build_switch_row),
    Mapping("log", "log", "t_flow_log", "f_id", build_log_row),
    Mapping("outbox", "outbox", "t_flow_outbox", "f_id", build_outbox_row),
)


class DatabaseManager:
    def __init__(self, mongo_database: str, mongo_prefix: str, mysql_database: str) -> None:
        self.mongo_database = mongo_database
        self.mongo_prefix = mongo_prefix
        self.mysql_database = mysql_database
        self.mongo_client = None
        self.mongo_db = None
        self.mysql_conn = None

    def connect_mongodb(self):
        mongodb_host = os.environ["MONGODB_HOST"]
        mongodb_port = os.environ["MONGODB_PORT"]
        mongodb_user = os.environ["MONGODB_USER"]
        mongodb_pwd = os.environ["MONGODB_PASSWORD"]
        mongodb_auth_source = os.environ["MONGODB_AUTH_SOURCE"]

        dns = f"mongodb://{mongodb_user}:{mongodb_pwd}@{mongodb_host}:{mongodb_port}?authSource={mongodb_auth_source}"

        client = MongoClient(dns, serverSelectionTimeoutMS=5000)
        client.admin.command("ping")
        self.mongo_client = client
        self.mongo_db = client[self.mongo_database]
        logger.info("MongoDB connected: %s", self.mongo_database)
        return self.mongo_db

    def connect_mysql(self):
        params = {
            'host': os.environ["DB_HOST"],
            'port': int(os.environ["DB_PORT"]),
            'user': os.environ["DB_USER"],
            'password': os.environ["DB_PASSWD"],
            'autocommit': True,
            'charset': 'utf8mb4'
        }
        try:
            self.mysql_conn = rdsdriver.connect(**params)
        except TypeError as e:
            logger.error("MySQL connection error: %s", str(e))
            raise
        logger.info("MySQL connected: %s:%s", params["host"], params["port"])
        return self.mysql_conn

    def close(self) -> None:
        if self.mysql_conn:
            self.mysql_conn.close()
        if self.mongo_client:
            self.mongo_client.close()


def load_existing_ids(conn, database: str, table_name: str, pk_column: str, ids: Sequence[Any], batch_size: int) -> set:
    existing_ids = set()
    if not ids:
        return existing_ids

    table = f"{database}.{table_name}"
    for batch in chunked(list(ids), batch_size):
        placeholders = ", ".join(["%s"] * len(batch))
        sql = f"SELECT {pk_column} FROM {table} WHERE {pk_column} IN ({placeholders})"
        cursor = conn.cursor()
        try:
            cursor.execute(sql, list(batch))
            existing_ids.update(row[0] for row in cursor.fetchall())
        finally:
            cursor.close()
    return existing_ids


def filter_new_rows(rows: List[Dict[str, Any]], existing_ids: set, pk_column: str) -> List[Dict[str, Any]]:
    filtered: List[Dict[str, Any]] = []
    seen = set(existing_ids)
    for row in rows:
        pk_value = row[pk_column]
        if pk_value in seen:
            continue
        seen.add(pk_value)
        filtered.append(row)
    return filtered


def insert_rows(conn, database: str, table_name: str, rows: List[Dict[str, Any]]) -> None:
    if not rows:
        return

    columns = list(rows[0].keys())
    column_sql = ", ".join(f"{column}" for column in columns)
    placeholder_sql = ", ".join(["%s"] * len(columns))
    sql = f"INSERT INTO {database}.{table_name} ({column_sql}) VALUES ({placeholder_sql})"
    values = [tuple(row[column] for column in columns) for row in rows]

    cursor = conn.cursor()
    try:
        cursor.executemany(sql, values)
    finally:
        cursor.close()


def iter_documents(collection, batch_size: int) -> Iterator[List[Dict[str, Any]]]:
    cursor = collection.find({}, no_cursor_timeout=True).sort("_id", 1).batch_size(batch_size)
    batch: List[Dict[str, Any]] = []
    try:
        for document in cursor:
            batch.append(document)
            if len(batch) >= batch_size:
                yield batch
                batch = []
        if batch:
            yield batch
    finally:
        cursor.close()


def write_table_rows(
    conn,
    database: str,
    table_name: str,
    pk_column: str,
    rows: List[Dict[str, Any]],
    batch_size: int,
) -> tuple[int, int]:
    if not rows:
        return 0, 0

    pk_values = [row[pk_column] for row in rows]
    existing_ids = load_existing_ids(conn, database, table_name, pk_column, pk_values, batch_size)
    new_rows = filter_new_rows(rows, existing_ids, pk_column)
    skipped = len(rows) - len(new_rows)

    if new_rows:
        insert_rows(conn, database, table_name, new_rows)
    return len(new_rows), skipped


def migrate_mapping(db_manager: DatabaseManager, mapping: Mapping, batch_size: int) -> None:
    source_collection = collection_name(db_manager.mongo_prefix, mapping.collection_suffix)
    collection = db_manager.mongo_db[source_collection]

    scanned = 0
    inserted = 0
    skipped = 0
    failed = 0

    logger.info("migrating `%s` from `%s`", mapping.table_name, source_collection)

    for documents in iter_documents(collection, batch_size):
        main_rows: List[Dict[str, Any]] = []
        child_rows: Dict[str, List[Dict[str, Any]]] = {
            child.table_name: [] for child in mapping.child_mappings
        }

        for document in documents:
            scanned += 1
            try:
                main_rows.append(mapping.build_row(document))
                for child in mapping.child_mappings:
                    child_rows[child.table_name].extend(child.build_rows(document))
            except Exception as exc:
                failed += 1
                logger.exception(
                    "failed to build rows for `%s` document `%s`: %s",
                    mapping.name,
                    document.get("_id"),
                    exc,
                )

        main_inserted, main_skipped = write_table_rows(
            db_manager.mysql_conn,
            db_manager.mysql_database,
            mapping.table_name,
            mapping.pk_column,
            main_rows,
            batch_size,
        )
        inserted += main_inserted
        skipped += main_skipped

        for child in mapping.child_mappings:
            child_inserted, child_skipped = write_table_rows(
                db_manager.mysql_conn,
                db_manager.mysql_database,
                child.table_name,
                child.pk_column,
                child_rows[child.table_name],
                batch_size,
            )
            inserted += child_inserted
            skipped += child_skipped

    logger.info(
        "finished `%s`: scanned=%s inserted=%s skipped=%s failed=%s",
        mapping.table_name,
        scanned,
        inserted,
        skipped,
        failed,
    )


def migrate(mongo_database: str, mongo_prefix: str, mysql_database: str, batch_size: int) -> int:
    db_manager = DatabaseManager(
        mongo_database=mongo_database,
        mongo_prefix=mongo_prefix,
        mysql_database=mysql_database,
    )

    try:
        db_manager.connect_mongodb()
        db_manager.connect_mysql()
        for mapping in MAPPINGS:
            migrate_mapping(db_manager, mapping, batch_size)
        return 0
    except Exception as exc:
        logger.exception("migration failed: %s", exc)
        return 1
    finally:
        db_manager.close()


def main() -> int:
    configure_logging()

    mongo_database = os.getenv("MONGODB_DATABASE", os.getenv("MONGO_DATABASE", "automation"))
    mongo_prefix = os.getenv("MONGODB_PREFIX", os.getenv("MONGO_PREFIX", os.getenv("STORE_PREFIX", "flow")))
    mysql_database = os.getenv("DB_NAME", os.getenv("DB_DATABASE", os.getenv("MYSQL_DATABASE", "adp")))
    batch_size = DEFAULT_BATCH_SIZE

    return migrate(
        mongo_database=mongo_database,
        mongo_prefix=mongo_prefix,
        mysql_database=mysql_database,
        batch_size=batch_size,
    )


if __name__ == "__main__":
    sys.exit(main())
