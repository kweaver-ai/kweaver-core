import { type ColumnOriginType } from '../useAntdResizableHeader';

export function validateColumnsFlex(cols: ColumnOriginType[] | undefined) {
  if (cols?.every(col => col.width)) {
    console.error('[use-antd-resizable-header] 请不要在所有列上添加 `width` 属性，这将导致表格无法自适应。');
  }
}
