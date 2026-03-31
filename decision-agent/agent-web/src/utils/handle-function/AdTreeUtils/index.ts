import type { TreeDataNode } from 'antd';
import _ from 'lodash';
import type { ReactNode } from 'react';
import { fuzzyMatch } from '@/utils/handle-function';

/** * 通用的处理后端数据构造antd树节点数据的utils函数。后端的原始数据会放在每一个树节点的sourceData属性上*/

export interface AdTreeDataNode extends Omit<TreeDataNode, 'children' | 'title'> {
  keyPath?: string[]; // 完整的路径
  parentKey?: string; // 父节点的key
  sourceData?: any; // 构造树节点的原始数据
  children: AdTreeDataNode[];
  title?: string;
}

export type CreateTreeDataOptions = {
  titleField?: string | ((record: any) => ReactNode); // 节点文本的字段名
  keyField?: string | ((record: any) => string); // 节点唯一值的字段名
  keyPath?: string[];
  parentKey?: string;
  isLeaf?: boolean | ((record: any) => boolean);
  selectable?: boolean;
};

/**
 * 根据后端返回的数据去生成树组件需要的树节点数据格式
 * @param dataSource 后端数据源
 */
const createAdTreeNodeData = (dataSource: any[], options: CreateTreeDataOptions = {}): AdTreeDataNode[] => {
  const newOptions = {
    titleField: 'name',
    keyField: 'key',
    keyPath: [],
    parentKey: '',
    ...options,
  };
  const loop = (data: any[], parentKey: string, keyPath: string[]): AdTreeDataNode[] => {
    return data.map((item: any) => {
      const nodeKey = typeof newOptions.keyField === 'string' ? item[newOptions.keyField] : newOptions.keyField(item);
      const isLeaf = typeof newOptions.isLeaf === 'boolean' ? newOptions.isLeaf : newOptions.isLeaf?.(item);
      const newKeyPath = [...keyPath!, nodeKey];
      return {
        title: typeof newOptions.titleField === 'string' ? item[newOptions.titleField] : newOptions.titleField(item),
        key: nodeKey, // key 是tree要求的唯一标识字段
        value: nodeKey, // value 是treeSelect要求的唯一标识字段
        keyPath: newKeyPath,
        parentKey,
        children: item.children && item.children.length > 0 ? loop(item.children, item.key, newKeyPath) : [],
        sourceData: {
          ...item,
        },
        level: newKeyPath.length,
        isLeaf: isLeaf ?? (item.children ? item.children.length === 0 : true),
        selectable: newOptions.selectable,
        checkable: item.checkable, // 支持节点级别的checkable控制
      };
    });
  };
  return loop(dataSource, newOptions.parentKey, newOptions.keyPath);
};

/**
 * 扁平树数据
 * @param treeDate
 */
const flatTreeData = (treeDate: AdTreeDataNode[]): AdTreeDataNode[] => {
  const cloneTreeData = _.cloneDeep(treeDate);
  const treeFlatDataSource: AdTreeDataNode[] = [];
  const loop = (data: AdTreeDataNode[]) => {
    data.forEach(item => {
      treeFlatDataSource.push({
        ...item,
        children: [],
      });
      if (item.children && item.children.length > 0) {
        loop(item.children);
      }
    });
  };
  loop(cloneTreeData);
  return treeFlatDataSource;
};

/**
 * 扁平数据转树结构
 */
const flatToTreeData = (flatTreeData: AdTreeDataNode[]) => {
  const cloneFlatTreeData = _.cloneDeep(flatTreeData);
  const treeDataSource: AdTreeDataNode[] = [];
  const cacheMap: Record<string, AdTreeDataNode> = {};
  for (let i = 0; i < cloneFlatTreeData.length; i++) {
    const item = cloneFlatTreeData[i];

    const nodeKey = item.key as string;
    const parentNodeKey = item.parentKey;

    cacheMap[nodeKey] = {
      ...item,
      children: cacheMap[nodeKey]?.children ?? [],
    };
    if (!parentNodeKey) {
      // 说明是根节点
      treeDataSource.push(cacheMap[nodeKey]);
    } else {
      if (!cacheMap[parentNodeKey]) {
        // 说明还没有遍历到当前节点的父节点，给个默认值  用于后面遍历到该父节点的时候  进行合并
        cacheMap[parentNodeKey] = {
          key: parentNodeKey,
          children: [],
        };
      }
      cacheMap[parentNodeKey].children.push(cacheMap[nodeKey]);
    }
  }
  return treeDataSource;
};

/**
 * 添加树节点
 * @param treeData 源数据
 * @param children 要添加的节点
 */
const addTreeNode = (treeData: AdTreeDataNode[], children: AdTreeDataNode | AdTreeDataNode[]) => {
  const childTreeNodes = Array.isArray(children) ? children : [children];
  let newFlatTreeData = flatTreeData(treeData);
  newFlatTreeData = [...newFlatTreeData, ...childTreeNodes];
  return flatToTreeData(newFlatTreeData);
};

/**
 * 删除树节点
 * @param treeData 源数据
 * @param deleteNodeKey 要删除的节点key
 */
const deleteTreeNode = (treeData: AdTreeDataNode[], deleteNodeKey: string | string[]) => {
  const deleteKeys = Array.isArray(deleteNodeKey) ? deleteNodeKey : [deleteNodeKey];
  const newFlatTreeData = flatTreeData(treeData).filter(item => !deleteKeys.includes(item.key as string));
  return flatToTreeData(newFlatTreeData);
};

/**
 * 更新树节点
 * @param treeData 源数据
 * @param updateNode 要更新的节点
 */
const updateTreeNode = (treeData: AdTreeDataNode[], updateNode: AdTreeDataNode | AdTreeDataNode[]) => {
  const childTreeNodes = Array.isArray(updateNode) ? updateNode : [updateNode];
  const childTreeNodeMap: Record<string, AdTreeDataNode> = {};
  childTreeNodes.forEach(item => {
    childTreeNodeMap[item.key as string] = item;
  });
  const newFlatTreeData = flatTreeData(treeData);
  for (let j = 0; j < newFlatTreeData.length; j++) {
    const key = newFlatTreeData[j].key as string;
    if (childTreeNodeMap[key]) {
      newFlatTreeData[j] = childTreeNodeMap[key];
    }
  }
  return flatToTreeData(newFlatTreeData);
};

/**
 * 通过树节点的key获取树节点
 * @param treeDatasource
 * @param nodeKey
 */
export const getTreeNodeByKey = (treeDatasource: AdTreeDataNode[], nodeKey: string | string[]) => {
  const keys = typeof nodeKey === 'string' ? [nodeKey] : nodeKey;
  const flatData: AdTreeDataNode[] = flatTreeData(treeDatasource);
  return flatData.filter(item => keys.includes(item.key as string));
};

export const searchTreeNode = (treeData: AdTreeDataNode[], searchValue: string) => {
  if (searchValue) {
    const newFlatTreeData = flatTreeData(treeData);
    const matchData = newFlatTreeData.filter(p => fuzzyMatch(searchValue, p.title!));
    matchData.forEach(item => {
      if (item.parentKey) {
        const targetIndex = matchData.findIndex(i => i.key === item.parentKey);
        if (targetIndex === -1) {
          const node = newFlatTreeData.find(i => i.key === item.parentKey);
          if (node) {
            matchData.push(node);
          }
        }
      }
    });
    return flatToTreeData(matchData);
  }
  return treeData;
};

export const adTreeUtils = {
  createAdTreeNodeData,
  flatTreeData,
  flatToTreeData,
  addTreeNode,
  deleteTreeNode,
  updateTreeNode,
  getTreeNodeByKey,
  searchTreeNode,
};
