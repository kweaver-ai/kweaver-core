import logging
import sqlite3
import threading
import uuid
from typing import Any, Dict, List, Optional
from src.infrastructure.db.db_pool_wrapper import connect_execute_close_db, connect_execute_commit_close_db
from datetime import datetime
import pytz

logger = logging.getLogger(__name__)


class SQLiteManager:
    def __init__(self, db_path: str = ":memory:"):
        self.db_path = db_path
        self.connection = sqlite3.connect(self.db_path, check_same_thread=False)
        self._lock = threading.Lock()
        self._migrate_history_table()
        self._create_history_table()

    def _migrate_history_table(self) -> None:
        """
        If a pre-existing history table had the old group-chat columns,
        rename it, create the new schema, copy the intersecting data, then
        drop the old table.
        """
        with self._lock:
            try:
                # Start a transaction
                self.connection.execute("BEGIN")
                cur = self.connection.cursor()

                cur.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='history'")
                if cur.fetchone() is None:
                    self.connection.execute("COMMIT")
                    return  # nothing to migrate

                cur.execute("PRAGMA table_info(history)")
                old_cols = {row[1] for row in cur.fetchall()}

                expected_cols = {
                    "id",
                    "memory_id",
                    "old_memory",
                    "new_memory",
                    "event",
                    "created_at",
                    "updated_at",
                    "is_deleted",
                    "actor_id",
                    "role",
                }

                if old_cols == expected_cols:
                    self.connection.execute("COMMIT")
                    return

                logger.info("Migrating history table to new schema (no convo columns).")

                # Clean up any existing history_old table from previous failed migration
                cur.execute("DROP TABLE IF EXISTS history_old")

                # Rename the current history table
                cur.execute("ALTER TABLE history RENAME TO history_old")

                # Create the new history table with updated schema
                cur.execute(
                    """
                    CREATE TABLE history (
                        id           TEXT PRIMARY KEY,
                        memory_id    TEXT,
                        old_memory   TEXT,
                        new_memory   TEXT,
                        event        TEXT,
                        created_at   DATETIME,
                        updated_at   DATETIME,
                        is_deleted   INTEGER,
                        actor_id     TEXT,
                        role         TEXT
                    )
                """
                )

                # Copy data from old table to new table
                intersecting = list(expected_cols & old_cols)
                if intersecting:
                    cols_csv = ", ".join(intersecting)
                    cur.execute(f"INSERT INTO history ({cols_csv}) SELECT {cols_csv} FROM history_old")

                # Drop the old table
                cur.execute("DROP TABLE history_old")

                # Commit the transaction
                self.connection.execute("COMMIT")
                logger.info("History table migration completed successfully.")

            except Exception as e:
                # Rollback the transaction on any error
                self.connection.execute("ROLLBACK")
                logger.error(f"History table migration failed: {e}")
                raise

    def _create_history_table(self) -> None:
        with self._lock:
            try:
                self.connection.execute("BEGIN")
                self.connection.execute(
                    """
                    CREATE TABLE IF NOT EXISTS history (
                        id           TEXT PRIMARY KEY,
                        memory_id    TEXT,
                        old_memory   TEXT,
                        new_memory   TEXT,
                        event        TEXT,
                        created_at   DATETIME,
                        updated_at   DATETIME,
                        is_deleted   INTEGER,
                        actor_id     TEXT,
                        role         TEXT
                    )
                """
                )
                self.connection.execute("COMMIT")
            except Exception as e:
                self.connection.execute("ROLLBACK")
                logger.error(f"Failed to create history table: {e}")
                raise

    def add_history(
        self,
        memory_id: str,
        old_memory: Optional[str],
        new_memory: Optional[str],
        event: str,
        *,
        created_at: Optional[str] = None,
        updated_at: Optional[str] = None,
        is_deleted: int = 0,
        actor_id: Optional[str] = None,
        role: Optional[str] = None,
    ) -> None:
        with self._lock:
            try:
                self.connection.execute("BEGIN")
                self.connection.execute(
                    """
                    INSERT INTO history (
                        id, memory_id, old_memory, new_memory, event,
                        created_at, updated_at, is_deleted, actor_id, role
                    )
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """,
                    (
                        str(uuid.uuid4()),
                        memory_id,
                        old_memory,
                        new_memory,
                        event,
                        created_at,
                        updated_at,
                        is_deleted,
                        actor_id,
                        role,
                    ),
                )
                self.connection.execute("COMMIT")
            except Exception as e:
                self.connection.execute("ROLLBACK")
                logger.error(f"Failed to add history record: {e}")
                raise

    def get_history(self, memory_id: str) -> List[Dict[str, Any]]:
        with self._lock:
            cur = self.connection.execute(
                """
                SELECT id, memory_id, old_memory, new_memory, event,
                       created_at, updated_at, is_deleted, actor_id, role
                FROM history
                WHERE memory_id = ?
                ORDER BY created_at ASC, DATETIME(updated_at) ASC
            """,
                (memory_id,),
            )
            rows = cur.fetchall()

        return [
            {
                "id": r[0],
                "memory_id": r[1],
                "old_memory": r[2],
                "new_memory": r[3],
                "event": r[4],
                "created_at": r[5],
                "updated_at": r[6],
                "is_deleted": bool(r[7]),
                "actor_id": r[8],
                "role": r[9],
            }
            for r in rows
        ]

    def reset(self) -> None:
        """Drop and recreate the history table."""
        with self._lock:
            try:
                self.connection.execute("BEGIN")
                self.connection.execute("DROP TABLE IF EXISTS history")
                self.connection.execute("COMMIT")
                self._create_history_table()
            except Exception as e:
                self.connection.execute("ROLLBACK")
                logger.error(f"Failed to reset history table: {e}")
                raise

    def close(self) -> None:
        if self.connection:
            self.connection.close()
            self.connection = None

    def __del__(self):
        self.close()


class RDSManager:
    def __init__(self, ):
        self.table_name = "t_data_agent_memory_history"
    @connect_execute_commit_close_db
    def add_history(
        self,
        memory_id: str,
        old_memory: Optional[str],
        new_memory: Optional[str] ,
        event: str,
        *,
        created_at: Optional[str] = "",
        updated_at: Optional[str] = "",
        is_deleted: int = 0,
        actor_id: Optional[str] = "",
        role: Optional[str] = "",
        connection, 
        cursor
    ) -> None:
        
        old_memory = old_memory if old_memory else ""
        created_at = created_at if created_at else datetime.now(pytz.timezone("Asia/Shanghai")).isoformat()
        updated_at = updated_at if updated_at else datetime.now(pytz.timezone("Asia/Shanghai")).isoformat()
        actor_id = actor_id if actor_id else ""
        role = role if role else ""

        sql = f"""
            INSERT INTO {self.table_name} (
                f_id, f_memory_id, f_old_memory, f_new_memory, f_event,
                f_created_at, f_updated_at, f_is_deleted, f_actor_id, f_role
            )
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
        """
        
        bind_values = [
            str(uuid.uuid4()),
            memory_id,
            old_memory,
            new_memory,
            event,
            created_at,
            updated_at,
            is_deleted,
            actor_id,
            role,
        ]
            
        cursor.execute(sql, bind_values)     
        
    @connect_execute_close_db
    def get_history(self, memory_id: str, connection, cursor) -> List[Dict[str, Any]]:
        
        sql = f"""
            SELECT f_id, f_memory_id, f_old_memory, f_new_memory, f_event,
                f_created_at, f_updated_at, f_is_deleted, f_actor_id, f_role
            FROM {self.table_name}
            WHERE f_memory_id = %s
            ORDER BY f_updated_at DESC
        """
        cursor.execute(sql, memory_id)
        rows = cursor.fetchall()

        return [
            {
                "id": r["f_id"],
                "memory_id": r["f_memory_id"],
                "old_memory": r["f_old_memory"],
                "new_memory": r["f_new_memory"],
                "event": r["f_event"],
                "created_at": r["f_created_at"],
                "updated_at": r["f_updated_at"],
                "is_deleted": bool(r["f_is_deleted"]),
                "actor_id": r["f_actor_id"],
                "role": r["f_role"],
            }
            for r in rows
        ]

    def reset(self) -> None:
        """Drop and recreate the history table."""
        

    def close(self) -> None:
        pass