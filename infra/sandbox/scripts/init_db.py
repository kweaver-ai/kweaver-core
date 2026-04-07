#!/usr/bin/env python3
"""
数据库初始化脚本（完整功能）

提供更完整的数据库初始化功能，包括创建数据库、表结构和种子数据。
适用于开发环境设置和 CI/CD 流水线。
"""
import asyncio
import sys
from pathlib import Path
import argparse

# 添加项目根目录到路径
project_root = Path(__file__).parent.parent
sys.path.insert(0, str(project_root))

from sqlalchemy import text
from sqlalchemy.ext.asyncio import create_async_engine
from src.infrastructure.config.settings import get_settings
from src.infrastructure.persistence.database import db_manager
from src.infrastructure.persistence.seed.seeder import seed_default_data


async def create_database(database_url: str, db_name: str) -> None:
    """创建数据库（仅开发环境）"""
    # 连接到 MySQL 服务器（不指定数据库）
    url = database_url.rsplit('/', 1)[0]  # 移除数据库名
    engine = create_async_engine(url, echo=False)

    try:
        async with engine.begin() as conn:
            # 创建数据库（如果不存在）
            await conn.execute(text(f"CREATE DATABASE IF NOT EXISTS `{db_name}`"))
            print(f"✓ Database '{db_name}' created (or already exists)")
    finally:
        await engine.dispose()


async def drop_database(database_url: str, db_name: str) -> None:
    """删除数据库（仅开发环境，谨慎使用）"""
    # 连接到 MySQL 服务器（不指定数据库）
    url = database_url.rsplit('/', 1)[0]  # 移除数据库名
    engine = create_async_engine(url, echo=False)

    try:
        async with engine.begin() as conn:
            # 删除数据库（如果存在）
            await conn.execute(text(f"DROP DATABASE IF EXISTS `{db_name}`"))
            print(f"✓ Database '{db_name}' dropped")
    finally:
        await engine.dispose()


async def main():
    """主函数"""
    parser = argparse.ArgumentParser(
        description="Sandbox Control Plane - Database Initialization"
    )
    parser.add_argument(
        "--create-db",
        action="store_true",
        help="Create database (development only)"
    )
    parser.add_argument(
        "--drop-db",
        action="store_true",
        help="Drop database before creating (use with caution!)"
    )
    parser.add_argument(
        "--skip-seed",
        action="store_true",
        help="Skip seeding default data"
    )
    parser.add_argument(
        "--force-seed",
        action="store_true",
        help="Force re-seeding even if data exists"
    )

    args = parser.parse_args()
    settings = get_settings()

    print("=" * 60)
    print("Sandbox Control Plane - Database Initialization")
    print("=" * 60)
    print(f"Environment: {settings.environment}")
    print(f"Database: {settings.database_url.split('@')[-1]}")
    print("=" * 60)

    # 安全检查
    if args.drop_db and settings.environment == "production":
        print("❌ ERROR: --drop-db is not allowed in production!")
        sys.exit(1)

    # 1. 创建/删除数据库
    if args.create_db or args.drop_db:
        if settings.environment == "production":
            print("\n⚠ WARNING: Skipping database creation in production")
            print("   Use manual database setup in production environments")
        else:
            print("\n[Step 1/3] Database setup...")
            db_name = settings.database_url.split('/')[-1]

            if args.drop_db:
                await drop_database(settings.database_url, db_name)

            await create_database(settings.database_url, db_name)

    # 2. 初始化数据库管理器
    print("\n[Step 2/3] Initializing database manager...")
    db_manager.initialize()
    print("✓ Database manager initialized")

    # 3. 创建表
    print("\n[Step 3/3] Creating tables...")
    await db_manager.create_tables()
    print("✓ Database tables created")

    # 4. 初始化默认数据
    if not args.skip_seed:
        print("\n[Step 4/4] Seeding default data...")
        stats = await seed_default_data(force=args.force_seed)
        print(f"✓ Default data seeded: {stats['total']} items")
        print(f"  - Runtime nodes: {stats['runtime_nodes']}")
        print(f"  - Templates: {stats['templates']}")
    else:
        print("\n[Step 4/4] Skipping default data seeding")

    # 关闭连接
    await db_manager.close()

    print("\n" + "=" * 60)
    print("✅ Database initialization completed!")
    print("=" * 60)


if __name__ == "__main__":
    asyncio.run(main())
