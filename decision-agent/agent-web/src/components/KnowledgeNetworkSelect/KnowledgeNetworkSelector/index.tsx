import React, { useState, useEffect, useMemo } from 'react';
import intl from 'react-intl-universal';
import { isEmpty, cloneDeep } from 'lodash';
import { Modal, Tree, Button, message, Spin, Checkbox } from 'antd';
import { CaretRightOutlined, CaretDownOutlined } from '@ant-design/icons';
import type { TreeDataNode } from 'antd';
import NetworkSelect from '../NetworkSelect';
import { getKnowledgeNetworkEntityList } from '../utils';
import EntiryDetail from '../EntityDetail';

// 类型定义
interface EntityProperty {
  name: string;
  value: string;
  alias: string;
  type: string;
  disabled: boolean;
  checked: boolean;
}

interface Entity {
  key: string;
  title: string;
  value: string;
  isLeaf: boolean;
  properties: EntityProperty[];
  properties_index: string[];
  color: string;
  alias: string;
}

interface SelectedData {
  spaceId: string; // 知识网络空间ID
  networkId: string; // 知识网络ID
  entities: string[]; // 实体字段名
  properties: Record<string, string[]>; // 实体属性
}

interface KnowledgeNetworkSelectorProps {
  entityCheckedKeys: string[]; // 实体选择状态
  visible: boolean;
  networkId: string;
  newtworkTreeData: TreeDataNode[];
  onCancel: (canDelete?: boolean) => void;
  onSave: (selectedData: SelectedData) => void;
  onUpdateTreeData: (treeData: TreeDataNode[]) => void;
}

const KnowledgeNetworkSelector: React.FC<KnowledgeNetworkSelectorProps> = ({
  entityCheckedKeys,
  networkId,
  newtworkTreeData,
  onCancel,
  onSave,
  onUpdateTreeData,
}) => {
  const [entities, setEntities] = useState<Entity[]>([]);
  const [selectedNetworkId, setSelectedNetworkId] = useState<string>(networkId);
  const [selectedNetworkName, setSelectedNetworkName] = useState<string>('');
  const [treeData, setTreeData] = useState<TreeDataNode[]>([]);
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);
  const [checkedKeys, setCheckedKeys] = useState<string[]>(entityCheckedKeys);
  const [networkLoading, setNetworkLoading] = useState(false);
  const [selectedEntityTitle, setSelectedEntityTitle] = useState<string>('');

  // 全部的key，用于全选
  const allKeys = useMemo(() => {
    const keys: string[] = [];
    treeData.forEach(({ key, children }) => {
      keys.push(key as string);
      children?.forEach(({ key, disabled }) => {
        if (!disabled) {
          keys.push(key as string);
        }
      });
    });

    return keys;
  }, [treeData]);

  // 处理全选和半选状态
  const [isAllChecked, isIndeterminate] = useMemo(() => {
    // 使用child key 比较是否全选
    const realCheckedKeysLength = checkedKeys.filter(item => item.includes('-')).length;
    const realAllKeysLength = allKeys.filter(item => item.includes('-')).length;
    const isChecked = checkedKeys.length > 0 && allKeys.length > 0;

    return [
      isChecked && realCheckedKeysLength === realAllKeysLength,
      isChecked && realCheckedKeysLength !== realAllKeysLength,
    ];
  }, [checkedKeys, allKeys]);

  // 获取知识网络实体类
  useEffect(() => {
    if (networkId) {
      setSelectedNetworkId(networkId);
      loadKnowledgeNetworks();
    }
  }, [networkId]);

  const loadKnowledgeNetworks = async () => {
    setNetworkLoading(true);
    try {
      await loadEntities(networkId);
    } catch {
      message.error(intl.get('dataAgent.config.failedToGetKnowledgeNetwork'));
    } finally {
      setNetworkLoading(false);
    }
  };

  // 获取实体列表
  const loadEntities = async (networkId: string) => {
    try {
      const entityList = await getKnowledgeNetworkEntityList(networkId);
      setEntities(entityList);
      buildTreeData(entityList);
    } catch {
      message.error(intl.get('dataAgent.config.failedToGetEntityList'));
    }
  };

  // 处理知识网络切换
  const handleChangeValue = (_networkId: string, networkName: string) => {
    setSelectedNetworkId(_networkId);
    if (networkId === _networkId) {
      setCheckedKeys(entityCheckedKeys);
    } else {
      setCheckedKeys([]);
    }
    setSelectedNetworkName(networkName);
    loadEntities(_networkId);
  };

  const handleChangeEntity = (entity: Entity) => {
    const props = cloneDeep(entity.properties).map((property: EntityProperty) => {
      property.disabled = !entity.properties_index.includes(property.name);
      return property;
    });
    return [{ name: '实体vid' }, ...props];
  };

  // 构建树形数据结构
  const buildTreeData = (entityList: Entity[]) => {
    const treeNodes: TreeDataNode[] = entityList.map(entity => ({
      title: (
        <span style={{ display: 'flex', alignItems: 'center' }}>
          <span
            style={{
              width: '8px',
              height: '8px',
              borderRadius: '50%',
              backgroundColor: entity.color || '#1890ff',
              marginRight: '8px',
              display: 'inline-block',
            }}
          />
          <span className="dip-ellipsis">{`${entity.alias}（${entity.title}）`}</span>
        </span>
      ),
      key: entity.title,
      isLeaf: false,
      children: handleChangeEntity(entity).map((property: any) => ({
        title: property.name,
        disabled: property.disabled,
        key: `${entity.title}-${property.name}`,
        parentKey: entity.key,
        isLeaf: true,
      })),
      switcherIcon: (props: any) => {
        if (entity.properties.length === 0) return null;
        return props.expanded ? (
          <CaretDownOutlined style={{ color: '#666', fontSize: '12px' }} />
        ) : (
          <CaretRightOutlined style={{ color: '#666', fontSize: '12px' }} />
        );
      },
    }));
    setTreeData(treeNodes);
  };

  // 处理树节点选择
  const handleTreeCheck = (checkedKeys: string[]) => {
    setCheckedKeys(checkedKeys);
  };

  // 处理树节点选择
  const handleTreeSelect = (selectedKeys: React.Key[], { node }: any) => {
    if (!node.isLeaf) {
      setSelectedEntityTitle(node.key);
    }
  };

  const handleSave = () => {
    const entityIds = checkedKeys.map(key => key.split('-')[0]);
    const entityKeys = entities.filter(entity => entityIds.includes(entity.title)).map(entity => entity.title);
    const checkedEntities = entities.filter(entity => entityIds.includes(entity.title));

    const selectedData: SelectedData = {
      networks: [{ id: selectedNetworkId, name: selectedNetworkName }],
      entities: entityKeys,
      properties: checkedEntities.reduce(
        (acc, entity) => {
          // 为每个实体收集属性名，属性的key格式为：实体key_属性名
          const entityProperties = checkedKeys
            .filter(key => key.startsWith(`${entity.title}-`))
            .map(key => key.split('-')[1]);

          acc[entity.title] = entityProperties;

          return acc;
        },
        {} as Record<string, string[]>
      ),
      checkedKeys,
    };

    onSave(selectedData);
    message.success(intl.get('dataAgent.config.knowledgeNetworkSelectionSuccess'));
    onCancel(isEmpty(selectedData) ? true : false);
  };

  return (
    <Modal
      title={<span>{intl.get('dataAgent.config.addKnowledgeNetwork')}</span>}
      open={true}
      onCancel={onCancel}
      width={1200}
      centered
      footer={[
        <Button
          key="save"
          type="primary"
          className="dip-min-width-72"
          onClick={handleSave}
          disabled={checkedKeys.length === 0}
        >
          {intl.get('dataAgent.config.save')}
        </Button>,
        <Button key="cancel" className="dip-min-width-72" onClick={onCancel}>
          {intl.get('dataAgent.cancel')}
        </Button>,
      ]}
    >
      <div style={{ display: 'flex', height: '600px' }}>
        {/* 左侧树形结构 */}
        <div
          style={{
            width: '300px',
            borderRight: '1px solid #f0f0f0',
            paddingRight: '16px',
            overflow: 'hidden',
          }}
        >
          <div style={{ marginBottom: '16px' }}>
            <NetworkSelect
              networkId={selectedNetworkId}
              networkTreeData={newtworkTreeData}
              placeholder={intl.get('dataAgent.config.pleaseSelectKnowledgeNetwork')}
              onUpdateTreeData={onUpdateTreeData}
              onChange={handleChangeValue}
            />
          </div>

          <Spin spinning={networkLoading}>
            <Checkbox
              indeterminate={isIndeterminate}
              checked={isAllChecked}
              onChange={e => {
                if (e.target.checked) {
                  setCheckedKeys(allKeys);
                } else {
                  setCheckedKeys([]);
                }
              }}
            >
              {intl.get('dataAgent.selectAll')}
            </Checkbox>
            <Tree
              checkable
              treeData={treeData}
              expandedKeys={expandedKeys}
              checkedKeys={checkedKeys}
              onExpand={setExpandedKeys}
              onCheck={handleTreeCheck}
              onSelect={handleTreeSelect}
              style={{ height: '528px', overflow: 'auto' }}
              defaultExpandAll={false}
              showLine={false}
              showIcon={false}
              switcherIcon={false}
            />
          </Spin>
        </div>

        {/* 右侧数据列表 */}
        <EntiryDetail entityTitle={selectedEntityTitle} networkId={selectedNetworkId} />
      </div>
    </Modal>
  );
};

export default KnowledgeNetworkSelector;
