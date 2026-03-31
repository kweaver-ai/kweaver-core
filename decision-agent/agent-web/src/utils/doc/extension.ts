import { last } from 'lodash';

/**
 * 获取文件的扩展名，.txt之类的
 */
export function getFileExtension(docname: string, hasDot: boolean = true) {
  return (hasDot ? '.' : '') + last(docname?.split('.'));
}
