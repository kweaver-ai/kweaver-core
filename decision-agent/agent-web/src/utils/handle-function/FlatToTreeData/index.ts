import _ from 'lodash';

type FlatTreeDataOption = {
  keyField: string;
  titleField: string;
  parentKeyField: string; // 父节点的key字段
  parentTitleField: string; // 父节点的title字段
};

/**
 * 扁平数据转树结构
 */
export const flatToTreeData = (flatTreeData: any[], option: FlatTreeDataOption) => {
  if (!_.isEmpty(option)) {
    const { keyField, titleField, parentKeyField, parentTitleField } = option;
    const cloneFlatTreeData = _.cloneDeep(flatTreeData);
    const treeDataSource: any[] = [];
    const cacheMap: Record<string, any> = {};
    for (let i = 0; i < cloneFlatTreeData.length; i++) {
      const item = cloneFlatTreeData[i];

      const nodeKey = item[keyField] as string;
      const parentNodeKey = item[parentKeyField];
      const parentNodeTitle = item[parentTitleField];

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
            [keyField]: parentNodeKey,
            [titleField]: parentNodeTitle,
            children: [],
          };
        }
        cacheMap[parentNodeKey].children.push(cacheMap[nodeKey]);
        delete cacheMap[nodeKey];
      }
    }
    if (treeDataSource.length === 0) {
      // 说明没有根节点
      return Object.values(cacheMap);
    }
    return treeDataSource;
  }
  return [];
};
