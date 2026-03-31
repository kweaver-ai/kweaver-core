import React, { useEffect } from 'react';
import intl from 'react-intl-universal';
import { TreeSelect } from 'antd';
import _ from 'lodash';
import type { TreeDataNode, TreeSelectProps } from 'antd';
import { DownOutlined } from '@ant-design/icons';
import { getKnowledgeNetworkList, getKnowledgeNetworkSpace } from '../utils';

interface NetworkSelectProps {
  networkId: string;
  networkTreeData: any[];
  placeholder?: string;
  onChange: (value: string, label: any, treeData?: TreeDataNode[]) => void;
  onUpdateTreeData: (treeData: TreeDataNode[]) => void;
}

const NetworkSelect: React.FC<NetworkSelectProps> = ({
  networkId,
  networkTreeData = [],
  placeholder = intl.get('dataAgent.pleaseSelectKnowledgeNetwork'),
  onChange,
  onUpdateTreeData,
}) => {
  useEffect(() => {
    loadKnowledgeSpaces();
  }, [networkId]);

  const loadKnowledgeSpaces = async () => {
    const data = await getKnowledgeNetworkSpace();

    // 将 data 转换为标准格式
    const formattedData = data.map(item => ({
      key: item.id,
      value: item.id,
      title: item.name,
      selectable: false,
      isLeaf: false,
    }));

    // 用 networkTreeData 替换 formattedData 中同一个 key 的元素
    const mergedData = formattedData.map(newItem => {
      const existingItem = networkTreeData.find(oldItem => oldItem.key === newItem.key);
      return existingItem || newItem;
    });

    // 添加 networkTreeData 中存在但 formattedData 中不存在的元素
    const newKeys = new Set(formattedData.map(item => item.key));
    const additionalItems = networkTreeData.filter(item => !newKeys.has(item.key));

    const finalData = [...mergedData, ...additionalItems];

    onUpdateTreeData(finalData);
  };

  const loadData = async (node: any) => {
    const id = _.split(node.key, 'parent-')?.[1] || node.key;
    const data = await getKnowledgeNetworkList(id);
    const formattedData = data.map(item => ({
      key: item.id,
      value: item.id,
      isLeaf: true,
      title: `${node.title}/${item.name}`,
    }));

    const newData = networkTreeData.map(item => (item.key === id ? { ...item, children: formattedData } : item));
    onUpdateTreeData(newData);
  };

  const handleChange: TreeSelectProps['onChange'] = (newValue, label) => {
    onChange(newValue, label[0]);
  };
  const treeData = _.map(_.cloneDeep(networkTreeData), item => {
    item.key = `parent-${item.key}`;
    item.value = `parent-${item.value}`;
    return item;
  });

  return (
    <TreeSelect
      style={{ width: '100%' }}
      switcherIcon={<DownOutlined />}
      value={networkId}
      dropdownStyle={{ maxHeight: 400, overflow: 'auto' }}
      placeholder={placeholder}
      allowClear
      treeDefaultExpandAll={false}
      treeData={treeData}
      loadData={loadData}
      onChange={handleChange}
      treeIcon
      treeNodeFilterProp="title"
      treeExpandAction="click"
    />
  );
};

export default NetworkSelect;
