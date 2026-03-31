import { memo, useMemo, useState, useEffect } from 'react';
import intl from 'react-intl-universal';
import { Button, Table } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import Empty from '@/components/Empty';
import DipIcon from '@/components/DipIcon';
import { type ToolProcessCategoryType, type ResultProcessStrategyType } from '@/apis/agent-factory';
import BindStrategyModal from './BindStrategyModal';
import styles from './index.module.less';

interface Props {
  // 只读模式，默认false
  readonly?: boolean;
  defaultResultProcessStrategies?: ResultProcessStrategyType[];
  onUpdateResultProcessStrategies: (strategies: ResultProcessStrategyType[]) => void;
}

const getColumns = ({ onDelete, readonly }: { onDelete: (item: any) => void; readonly?: boolean }) => [
  {
    title: intl.get('dataAgent.policyCategory'),
    dataIndex: 'category',
    key: 'category',
    ellipsis: true,
    render: (category: ToolProcessCategoryType) => category.name,
  },
  {
    title: intl.get('dataAgent.policyName'),
    dataIndex: 'strategy.name',
    key: 'strategy.name',
    ellipsis: true,
    render: (_, record: ResultProcessStrategyType) => record.strategy.name,
  },
  {
    title: intl.get('dataAgent.policyDescription'),
    dataIndex: 'strategy.description',
    key: 'strategy.description',
    ellipsis: true,
    render: (_, record: ResultProcessStrategyType) => record.strategy.description || '---',
  },
  // 只读模式，不显示操作
  ...(readonly
    ? []
    : [
        {
          title: intl.get('dataAgent.config.operation'),
          dataIndex: 'action',
          key: 'action',
          width: 100,
          render: (_, record: ResultProcessStrategyType) => (
            <DipIcon type="icon-dip-trash" className="dip-pointer" onClick={() => onDelete(record)} />
          ),
        },
      ]),
];

const ResultHandlingStrategy = memo(
  ({ readonly, defaultResultProcessStrategies, onUpdateResultProcessStrategies }: Props) => {
    const [showBindModal, setShowBindModal] = useState<boolean>(false);
    const [data, setData] = useState<Array<ResultProcessStrategyType>>(defaultResultProcessStrategies || []);

    // 删除策略
    const deleteStrategy = ({ category }: ResultProcessStrategyType) => {
      // 同一分类只会有一条数据，表格中category.id是唯一的
      setData(prev => prev.filter(item => item.category.id !== category.id));
    };

    const columns = useMemo(() => getColumns({ onDelete: deleteStrategy, readonly }), []);

    useEffect(() => {
      onUpdateResultProcessStrategies(data);
    }, [data]);

    return (
      <div className={styles['container']}>
        {!readonly && (
          <Button icon={<PlusOutlined />} type="primary" className="dip-mb-16" onClick={() => setShowBindModal(true)}>
            {intl.get('dataAgent.bindPolicy')}
          </Button>
        )}
        <Table
          scroll={{ y: readonly ? 330 : 290 }}
          dataSource={data}
          columns={columns}
          pagination={false}
          locale={{
            emptyText: (
              <Empty
                style={{ height: 255 }}
                className="dip-flex-column-center"
                description={intl.get('dataAgent.noPoliciesAvailable')}
              />
            ),
          }}
        />
        {showBindModal && (
          <BindStrategyModal
            onCancel={() => setShowBindModal(false)}
            onConfirm={({ category, strategy }) => {
              setShowBindModal(false);
              setData(prev => {
                return [{ category, strategy }, ...prev.filter(item => item.category.id !== category.id)];
              });
            }}
          />
        )}
      </div>
    );
  }
);

export default ResultHandlingStrategy;
