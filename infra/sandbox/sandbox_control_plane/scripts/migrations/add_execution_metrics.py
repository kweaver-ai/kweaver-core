"""
Alembic 迁移脚本模板

添加 execution 表的 return_value 和 metrics 字段

使用方法：
1. 确保 Alembic 已初始化：`alembic init alembic`
2. 将此文件复制到 `alembic/versions/` 目录
3. 根据实际生成的修订版本号重命名文件
4. 运行迁移：`alembic upgrade head`
"""
from alembic import op
import sqlalchemy as sa

# revision identifiers, used by Alembic
revision = '002_add_execution_metrics'
down_revision = '001_initial_schema'
branch_labels = None
depends_on = None


def upgrade():
    """添加 return_value 和 metrics 字段"""
    op.add_column(
        'executions',
        sa.Column('return_value', sa.JSON(), nullable=True)
    )
    op.add_column(
        'executions',
        sa.Column('metrics', sa.JSON(), nullable=True)
    )


def downgrade():
    """移除 return_value 和 metrics 字段"""
    op.drop_column('executions', 'metrics')
    op.drop_column('executions', 'return_value')
