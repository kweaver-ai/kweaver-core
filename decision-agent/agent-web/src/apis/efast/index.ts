import { get, post } from '@/utils/http';
import { DocLibTypeEnum } from '@/utils/doc';
import { FieldEnum } from './types';
export * from './types';

const baseUrl = '/api/efast';
const urlV1 = `${baseUrl}/v1`;
const urlV2 = `${baseUrl}/v2`;

/**
 * 获取对象信息
 */
export function getFileItemsInfo({ object_id, fields }: { object_id: string; fields: string }) {
  return get(`${urlV2}/items/${object_id}/${fields}`);
}

/* 列举文档中心目录 */
export function listDir({ docid, limit, marker }: { docid: string; limit: number; marker?: string }): Promise<{
  dirs: any[];
  files: any[];
  next_marker?: string;
}> {
  return get(`${urlV1}/folders/${encodeURIComponent(docid)}/sub_objects`, {
    params: {
      sort: 'name',
      direction: 'asc',
      permission_attributes_required: false,
      limit,
      ...(marker ? { marker } : {}),
    },
  });
}

/* 通过objectId获取文档信息 */
export function getDocInfoByObjectIdBatch(objectIds: string[], fields: FieldEnum[]) {
  return post(`${urlV1}/items-batch/${fields.join(',')}`, {
    body: {
      method: 'GET',
      object_ids: objectIds,
    },
  });
}

export function getDocInfoByObjectId(objectId: string, fields: FieldEnum[]) {
  return get(`${urlV2}/items/${encodeURIComponent(objectId)}/${fields.join(',')}`);
}

export function getEntryDocLibs(type?: DocLibTypeEnum) {
  return get(`${urlV1}/entry-doc-lib`, {
    params: {
      sort: 'doc_lib_name',
      direction: 'asc',
      type,
    },
  });
}

/* 由名字获取对象信息 */
export function getDocInfoByName(namepath: string) {
  return post(`${urlV1}/file/getinfobypath`, {
    body: { namepath },
  });
}

/* 创建文件夹 */
export function createDir(data: { docid: string; name: string }) {
  return post(`${urlV1}/dir/create`, {
    body: data,
  });
}
