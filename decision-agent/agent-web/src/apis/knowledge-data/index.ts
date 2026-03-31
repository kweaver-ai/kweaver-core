import { get, post } from '@/utils/http';
import qs from 'query-string';

const baseUrl = '/api/knowledge-data';
const knowledgeDataBaseUrl = '/api/kn-knowledge-data/v1';
const knExperimentDataBaseUrl = '/api/ontology-manager/v1';

/**
 * 获取数据源
 * @param {int} page 请求`页码`
 * @param {int} size 请求`每页条数`
 */
export const getDataSource = async ({
  page,
  size,
  order,
  knw_id = 0,
  ds_type = '',
  dsname = '',
  rule = 'update_time',
}: {
  page: number;
  size: number;
  order: 'descend';
  dsname?: string;
  rule?: 'update_time';
  knw_id?: number;
  ds_type?: string;
}) => {
  const params = {
    page,
    size,
    order,
    dsname,
    rule,
    ...(knw_id ? { knw_id } : {}),
    ...(ds_type ? { ds_type } : {}),
  };

  return get(`${baseUrl}/v1/ds?${qs.stringify(params)}`);
};

/* 获取知识源列表 */
export const getKnowledgeSourceList = ({
  page,
  order,
  size,
  rule,
}: {
  page: number;
  order: string;
  size: number;
  rule: string;
}) => {
  return get(`${knowledgeDataBaseUrl}/knw/get_all?page=${page}&order=${order}&size=${size}&rule=${rule}`);
};

/* 获取知识源详情 */
export const getKnowledgeSourceDetail = ({ id }: { id: string }) => {
  return get(`${knowledgeDataBaseUrl}/graph/info/onto?graph_id=${id}`);
};

export const getGraphByKnw = ({
  knw_id,
  page,
  size,
  order,
  rule,
  filter,
  name,
}: {
  knw_id: string;
  page: number;
  size: number;
  order: string;
  rule: string;
  filter: string;
  name: string;
}) => {
  return get(
    `${knowledgeDataBaseUrl}/knw/get_graph_by_knw?knw_id=${knw_id}&page=${page}&size=${size}&order=${order}&rule=${rule}&filter=${filter}&name=${name}`
  );
};

export const getDocsSourceListByDsId = ({
  ds_id,
  data_source,
  postfix,
}: {
  ds_id: string;
  data_source: string;
  postfix: string;
}) => {
  return get(`${knowledgeDataBaseUrl}/onto/gettabbydsn?ds_id=${ds_id}&data_source=${data_source}&postfix=${postfix}`);
};

export const getChildrenFile = (data: { docid: string; ds_id: string; postfix: string }) => {
  return get(`${knowledgeDataBaseUrl}/onto/dirlist`, { params: data });
};

const knDataQueryBaseUrl = '/api/kn-data-query/v1';

interface SearchConfig {
  tag: string;
  props: Array<{
    name: string;
    value: string;
    alias: string;
    type: string;
    disabled: boolean;
    checked: boolean;
  }>;
}

/**
 * 获取实体类对应的视图表
 * @param {string} kg_id 知识图谱ID
 * @param {number} page 页码
 * @param {number} size 每页条数
 * @param {Array} search_config 搜索配置
 * @param {Array} vids 视图ID列表
 */
export const getEntityViews = ({
  kg_id,
  page,
  size,
  search_config = [],
  vids = [],
}: {
  kg_id: string;
  page: number;
  size: number;
  search_config?: SearchConfig[];
  vids?: string[];
}) => {
  return post(`${knDataQueryBaseUrl}/basic-search/kgs/${kg_id}/vids`, {
    body: {
      page,
      size,
      search_config,
      kg_id,
      vids,
    },
  });
};

export const getKnExperimentList = () =>
  get(`${knExperimentDataBaseUrl}/knowledge-networks`, {
    params: { limit: -1 },
  });

export const getKnExperimentDetailsById = (kn_id: string) =>
  get(`${knExperimentDataBaseUrl}/knowledge-networks/${kn_id}?include_detail=false&include_statistics=false`);

/** 获取知识网络的对象类 */
export const getObjectTypeById = (kn_id: string) =>
  get(
    `${knExperimentDataBaseUrl}/knowledge-networks/${kn_id}/object-types?offset=0&limit=-1&direction=desc&sort=update_time&name_pattern=`
  );
