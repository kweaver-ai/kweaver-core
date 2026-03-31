import { last } from 'lodash';
import { DocLibTypeEnum } from './types';
export { default as FileTypeIcon } from './FileTypeIcon';
export * from './file-icon';
export * from './extension';

export { DocLibTypeEnum };

export const IdPreFix = 'gns://';
export const DocIdSeparator = '/';
function formatDocsStringToArray(docid: string) {
  return docid?.replace(IdPreFix, '').split(DocIdSeparator);
}

/**
 * 获取最后一个docid
 * @param docid
 * @returns
 */
export function getLastDocId(docid: string) {
  return docid === '' ? '' : last(formatDocsStringToArray(docid));
}
