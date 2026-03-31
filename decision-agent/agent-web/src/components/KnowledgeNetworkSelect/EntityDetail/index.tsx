import { memo, useState, useEffect } from 'react';
import classNames from 'classnames';
import intl from 'react-intl-universal';
import { Table, message, Spin } from 'antd';
import Empty from '@/components/Empty';
import LoadFailed from '@/components/LoadFailed';
import { getEntityData } from '../utils';
import styles from './index.module.less';

interface EntityProperty {
  name: string;
  value: string;
  alias: string;
  type: string;
  disabled: boolean;
  checked: boolean;
}

interface EntityProperties {
  tag: string;
  props: EntityProperty[];
}

interface EntityNode {
  id: string;
  alias: string;
  color: string;
  class_name: string;
  icon: string;
  default_property: {
    name: string;
    value: string;
    alias: string;
  };
  tags: string[];
  properties: EntityProperties[];
}

interface DataRecord {
  id: string;
  [key: string]: any;
}

enum LoadStatusEnum {
  NoEntitySelected = 'noEntitySelected',
  Loading = 'loading',
  Empty = 'empty',
  Failed = 'failed',
  Normal = 'normal',
}

// 生成表格列
const generateColumns = data => {
  if (data.length === 0) return [];

  const firstRecord = data[0];
  return Object.keys(firstRecord)
    .filter(key => key !== 'id')
    .map(key => ({
      title: key,
      dataIndex: key,
      key,
      width: 200,
      ellipsis: true,
    }));
};

const EntityDetail = ({ networkId, entityTitle }: { networkId: string; entityTitle: string }) => {
  const [loadStatus, setLoadStatus] = useState<LoadStatusEnum>(LoadStatusEnum.NoEntitySelected);
  const [entityData, setEntityData] = useState<DataRecord[]>([]);

  // 获取实体数据并转换为表格数据
  const loadEntityData = async () => {
    setLoadStatus(LoadStatusEnum.Loading);
    try {
      const data = await getEntityData(entityTitle, networkId);

      // 转换数据格式以适配表格
      const transformedData = data.map((node: EntityNode, index: number) => {
        const record: DataRecord = {
          id: node.id || `node_${index}`,
          实体vid: node.id,
        };

        // 处理properties数组中的props
        node.properties.forEach((propertyGroup: EntityProperties) => {
          propertyGroup.props.forEach((prop: EntityProperty) => {
            record[prop.alias || prop.name] = prop.value;
          });
        });

        return record;
      });

      setEntityData(transformedData);
      setLoadStatus(transformedData.length > 0 ? LoadStatusEnum.Normal : LoadStatusEnum.Empty);
    } catch (ex: any) {
      setLoadStatus(LoadStatusEnum.Failed);
      if (ex?.description) {
        message.error(ex.description);
      }
    }
  };

  useEffect(() => {
    if (entityTitle) {
      loadEntityData();
    }
  }, [entityTitle]);

  return (
    <div className="dip-pl-16 dip-overflow-hidden dip-flex-1">
      <div className="dip-h-100 dip-flex-column">
        {loadStatus === LoadStatusEnum.NoEntitySelected ? (
          <Empty
            className="dip-h-100 dip-flex-column-center"
            description={intl.get('dataAgent.config.selectAnyBasicSupplement')}
          />
        ) : loadStatus === LoadStatusEnum.Empty ? (
          <Empty className="dip-h-100 dip-flex-column-center" />
        ) : loadStatus === LoadStatusEnum.Failed ? (
          <LoadFailed className="dip-h-100 dip-flex-column-center" />
        ) : loadStatus === LoadStatusEnum.Loading && entityData.length === 0 ? (
          <div className="dip-h-100 dip-flex-column-center">
            <Spin />
          </div>
        ) : (
          <>
            <div
              style={{
                padding: '8px 12px',
                backgroundColor: '#e6f7ff',
                border: '1px solid #91d5ff',
              }}
              className="dip-mb-16 dip-border-radius-6"
            >
              <span style={{ color: '#1890ff' }} className="dip-font-weight-700">
                {intl.get('dataAgent.config.firstXSampleData', { count: entityData.length })}
              </span>
            </div>
            <div className="dip-overflow-hidden dip-flex-1">
              <Table
                className={classNames('dip-h-100', styles['table'])}
                bordered
                dataSource={entityData}
                columns={generateColumns(entityData)}
                rowKey="id"
                scroll={{ x: 600, y: 502 }}
                loading={loadStatus === LoadStatusEnum.Loading}
                size="small"
                pagination={false}
              />
            </div>
          </>
        )}
      </div>
    </div>
  );
};

export default memo(EntityDetail);
