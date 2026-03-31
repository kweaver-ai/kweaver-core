import { getGraphByKnw, getKnowledgeSourceList, getEntityViews, getKnowledgeSourceDetail } from '@/apis/knowledge-data';

/**
 * 获取知识网络空间列表数据
 * @param params 查询参数
 * @returns 知识网络列表数据
 */
export const getKnowledgeNetworkSpace = async (): Promise<{ id: string; name: string }[]> => {
  const defaultParams = {
    size: 1000,
    page: 1,
    rule: 'update',
    order: 'desc',
  };

  try {
    const { res } = await getKnowledgeSourceList(defaultParams);
    return res?.df?.map((item: any) => ({
      id: item.id.toString(),
      name: item.knw_name,
    }));
  } catch {
    return [];
  }
};

/**
 * 根据网络空间id获取知识网络列表
 * @param knw_id 知识网络id
 * @returns 知识网络列表数据
 */
export const getKnowledgeNetworkList = async (knw_id: string): Promise<{ id: string; name: string }[]> => {
  const defaultParams = {
    knw_id,
    page: 1,
    size: 10000,
    order: 'desc',
    rule: 'create',
    filter: 'all',
    name: '',
  };

  try {
    const { res } = await getGraphByKnw(defaultParams);
    return (
      res?.df
        // ?.filter(item => item.taskstatus === 'normal')
        .map((item: any) => ({
          id: item.id.toString(),
          name: item.name,
        }))
    );
  } catch {
    return [];
  }
};

/**
 * 获取实体数据用于表格展示
 * @param searchQuery 搜索查询字符串
 * @param networkId 知识图谱ID
 * @returns 实体节点数据列表
 */
export const getEntityData = async (searchQuery: string, networkId: string) => {
  const { res } = await getEntityViews({
    kg_id: networkId,
    page: 1,
    size: 100,
    search_config: [{ tag: searchQuery }],
    vids: [],
  });

  // 返回符合新节点结构的数据格式
  return res?.nodes || [];
};

/**
 * 获取知识网络下的实体类
 * @param knw_id 知识网络id
 * @returns 实体类列表
 */
export const getKnowledgeNetworkEntityList = async (knw_id: string) => {
  const { res } = await getKnowledgeSourceDetail({ id: knw_id });
  return res?.entity?.map((item: any) => ({
    key: item.entity_id,
    title: item.name,
    value: item.entity_id,
    isLeaf: true,
    properties: item.properties,
    properties_index: item.properties_index,
    color: item.color,
    alias: item.alias,
  }));
};
