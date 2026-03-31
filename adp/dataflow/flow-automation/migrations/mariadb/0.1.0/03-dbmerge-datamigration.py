#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
MariaDB 数据迁移工具 - 单连接版
"""

import rdsdriver
import os
import logging

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# ==================== 配置 ====================

# 数据库连接（同一个实例）
DB_CONFIG = {
    'host': os.environ["DB_HOST"],
    'port': os.environ["DB_PORT"],
    'user': os.environ["DB_USER"],
    'password': os.environ["DB_PASSWD"],
    'charset': 'utf8mb4'
}

# 源库和目标库
SOURCE_DATABASE = 'workflow'
TARGET_DATABASE = 'adp'

BATCH_SIZE = 2000
TABLES_TO_MIGRATE = [
    "t_model",
    "t_train_file",
    "t_automation_executor",
    "t_automation_executor_accessor",
    "t_automation_executor_action",
    "t_content_admin",
    "t_audio_segments",
    "t_automation_conf",
    "t_automation_agent",
    "t_alarm_rule",
    "t_alarm_user",
    "t_automation_dag_instance_ext_data",
    "t_task_cache_0",
    "t_task_cache_1",
    "t_task_cache_2",
    "t_task_cache_3",
    "t_task_cache_4",
    "t_task_cache_5",
    "t_task_cache_6",
    "t_task_cache_7",
    "t_task_cache_8",
    "t_task_cache_9",
    "t_task_cache_a",
    "t_task_cache_b",
    "t_task_cache_c",
    "t_task_cache_d",
    "t_task_cache_e",
    "t_task_cache_f"
]


# ==================== 数据库操作 ====================


def get_connection():
    return rdsdriver.connect(
        host=DB_CONFIG['host'],
        port=int(DB_CONFIG['port']),
        user=DB_CONFIG['user'],
        password=DB_CONFIG['password'],
        charset=DB_CONFIG['charset'],
        autocommit=True
    )


def get_primary_key(conn, database, table_name):
    cursor = conn.cursor()
    try:
        cursor.execute("""
            SELECT COLUMN_NAME FROM information_schema.KEY_COLUMN_USAGE 
            WHERE TABLE_SCHEMA = %s AND TABLE_NAME = %s AND CONSTRAINT_NAME = 'PRIMARY'
            ORDER BY ORDINAL_POSITION LIMIT 1
        """, (database, table_name))
        row = cursor.fetchone()
        return row[0] if row else None
    finally:
        cursor.close()


def get_columns(conn, database, table_name):
    cursor = conn.cursor()
    try:
        cursor.execute("""
            SELECT COLUMN_NAME FROM information_schema.COLUMNS 
            WHERE TABLE_SCHEMA = %s AND TABLE_NAME = %s ORDER BY ORDINAL_POSITION
        """, (database, table_name))
        return [row[0] for row in cursor.fetchall()]
    finally:
        cursor.close()


def table_exists(conn, database, table_name):
    cursor = conn.cursor()
    try:
        cursor.execute("""
            SELECT 1 FROM information_schema.TABLES 
            WHERE TABLE_SCHEMA = %s AND TABLE_NAME = %s
        """, (database, table_name))
        return cursor.fetchone() is not None
    finally:
        cursor.close()


def get_count(conn, database, table_name):
    cursor = conn.cursor()
    try:
        cursor.execute(f"SELECT COUNT(*) FROM `{database}`.`{table_name}`")
        return cursor.fetchone()[0]
    finally:
        cursor.close()


def get_pk_range(conn, database, table_name, pk_column):
    cursor = conn.cursor()
    try:
        cursor.execute(f"SELECT MIN(`{pk_column}`), MAX(`{pk_column}`) FROM `{database}`.`{table_name}`")
        row = cursor.fetchone()
        return row[0], row[1]
    finally:
        cursor.close()


def fetch_batch(conn, database, table_name, pk_column, columns, last_pk, batch_size):
    cursor = conn.cursor()
    try:
        cols_str = ', '.join([f'`{c}`' for c in columns])
        tbl = f'`{database}`.`{table_name}`'
        if pk_column:
            if last_pk is not None:
                sql = f"SELECT {cols_str} FROM {tbl} WHERE `{pk_column}` > %s ORDER BY `{pk_column}` LIMIT %s"
                cursor.execute(sql, (last_pk, batch_size))
            else:
                sql = f"SELECT {cols_str} FROM {tbl} ORDER BY `{pk_column}` LIMIT %s"
                cursor.execute(sql, (batch_size,))
        else:
            offset = last_pk if last_pk else 0
            sql = f"SELECT {cols_str} FROM {tbl} LIMIT %s OFFSET %s"
            cursor.execute(sql, (batch_size, offset))
        rows = cursor.fetchall()
        return [{columns[i]: row[i] for i in range(len(columns))} for row in rows]
    finally:
        cursor.close()


def get_existing_pks(conn, database, table_name, pk_column, pk_values):
    if not pk_values:
        return set()
    cursor = conn.cursor()
    try:
        tbl = f'`{database}`.`{table_name}`'
        placeholders = ','.join(['%s'] * len(pk_values))
        sql = f"SELECT `{pk_column}` FROM {tbl} WHERE `{pk_column}` IN ({placeholders})"
        cursor.execute(sql, pk_values)
        return {row[0] for row in cursor.fetchall()}
    finally:
        cursor.close()


def insert_rows(conn, database, table_name, columns, rows):
    if not rows:
        return
    cursor = conn.cursor()
    try:
        tbl = f'`{database}`.`{table_name}`'
        cols_str = ', '.join([f'`{c}`' for c in columns])
        placeholders = ', '.join(['%s'] * len(columns))
        sql = f"INSERT INTO {tbl} ({cols_str}) VALUES ({placeholders})"
        for row in rows:
            cursor.execute(sql, tuple(row[c] for c in columns))
    finally:
        cursor.close()


# ==================== 校验逻辑 ====================


def check_columns_match(src_columns, tgt_columns):
    src_set = set(c.lower() for c in src_columns)
    tgt_set = set(c.lower() for c in tgt_columns)
    if src_set != tgt_set:
        src_only = src_set - tgt_set
        tgt_only = tgt_set - src_set
        msg = []
        if src_only:
            msg.append(f"源表多: {list(src_only)}")
        if tgt_only:
            msg.append(f"目标多: {list(tgt_only)}")
        return False, '; '.join(msg)
    return True, "列名一致"


def check_need_migrate(conn, src_db, tgt_db, table_name, pk_column, columns):
    src_count = get_count(conn, src_db, table_name)
    tgt_count = get_count(conn, tgt_db, table_name)
    logger.info(f"  [行数] 源={src_count}, 目标={tgt_count}")
    
    if src_count == 0 and tgt_count == 0:
        return False, "都为空"
    if src_count != tgt_count:
        return True, f"行数差{abs(src_count - tgt_count)}"
    
    if pk_column:
        src_min, src_max = get_pk_range(conn, src_db, table_name, pk_column)
        tgt_min, tgt_max = get_pk_range(conn, tgt_db, table_name, pk_column)
        logger.info(f"  [主键] 源=[{src_min},{src_max}], 目标=[{tgt_min},{tgt_max}]")
        if src_min != tgt_min or src_max != tgt_max:
            return True, "主键范围不一致"
    
    return False, "数据一致"


# ==================== 迁移逻辑 ====================


def migrate_batch(conn, tgt_db, table_name, pk_column, columns, rows):
    if not rows:
        return 0
    
    if pk_column:
        pk_values = [row[pk_column] for row in rows]
        existing = get_existing_pks(conn, tgt_db, table_name, pk_column, pk_values)
        to_insert = [r for r in rows if r[pk_column] not in existing]
        insert_rows(conn, tgt_db, table_name, columns, to_insert)
        return len(to_insert)
    else:
        insert_rows(conn, tgt_db, table_name, columns, rows)
        return len(rows)


def migrate_table(conn, src_db, tgt_db, table_name):
    logger.info(f"\n{'='*50}")
    logger.info(f"表: {table_name}")
    
    # 1. 检查目标表存在
    if not table_exists(conn, tgt_db, table_name):
        logger.error(f"  ✗ 目标表不存在")
        return 'fail'
    
    # 2. 获取列信息
    src_columns = get_columns(conn, src_db, table_name)
    tgt_columns = get_columns(conn, tgt_db, table_name)
    
    if not src_columns or not tgt_columns:
        logger.warning(f"  跳过: 无列信息")
        return 'skip'
    
    # 3. 校验列名一致性
    is_match, msg = check_columns_match(src_columns, tgt_columns)
    if not is_match:
        logger.warning(f"  跳过: 列名不一致 - {msg}")
        return 'skip'
    
    logger.info(f"  [列] {src_columns}")
    
    # 4. 获取主键
    pk_column = get_primary_key(conn, src_db, table_name)
    logger.info(f"  [主键] {pk_column or '无'}")
    
    # 5. 数据校验
    need, reason = check_need_migrate(conn, src_db, tgt_db, table_name, pk_column, src_columns)
    if not need:
        logger.info(f"  跳过: {reason}")
        return 'skip'
    
    logger.info(f"  迁移原因: {reason}")
    
    # 6. 执行迁移
    total = get_count(conn, src_db, table_name)
    last_pk, migrated, inserted, updated, batch_num = None, 0, 0, 0, 0
    
    while True:
        batch_num += 1
        rows = fetch_batch(conn, src_db, table_name, pk_column, src_columns, last_pk, BATCH_SIZE)
        if not rows:
            break
        
        ins = migrate_batch(conn, tgt_db, table_name, pk_column, src_columns, rows)
        
        if pk_column:
            last_pk = rows[-1][pk_column]
        else:
            last_pk = migrated + len(rows)
        
        inserted += ins
        migrated += len(rows)
        
        pct = int(migrated * 100 / total) if total else 100
        logger.info(f"  批次{batch_num}: +{ins}插入 | {migrated}/{total} ({pct}%)")
    
    logger.info(f"  完成: 插入{inserted}")
    return 'done'


def run():
    logger.info("="*50)
    logger.info("MariaDB 迁移工具")
    logger.info("="*50)
    logger.info(f"连接: {DB_CONFIG['host']}:{DB_CONFIG['port']}")
    logger.info(f"源库: {SOURCE_DATABASE}")
    logger.info(f"目标: {TARGET_DATABASE}")
    
    conn = get_connection()
    
    try:
        logger.info(f"表数量: {len(TABLES_TO_MIGRATE)}")
        
        results = {'skip': [], 'done': [], 'fail': []}
        for table in TABLES_TO_MIGRATE:
            try:
                r = migrate_table(conn, SOURCE_DATABASE, TARGET_DATABASE, table)
                results[r].append(table)
            except Exception as e:
                logger.error(f"  异常: {e}")
                results['fail'].append(table)
        
        logger.info(f"\n{'='*50}\n汇总\n{'='*50}")
        logger.info(f"跳过: {len(results['skip'])} - {results['skip']}")
        logger.info(f"成功: {len(results['done'])} - {results['done']}")
        logger.info(f"失败: {len(results['fail'])} - {results['fail']}")
        
        return len(results['fail']) == 0
    finally:
        conn.close()


if __name__ == "__main__":
    exit(0 if run() else 1)