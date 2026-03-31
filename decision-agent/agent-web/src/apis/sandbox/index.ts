import type { SandBoxQueryFileProps, SandBoxUploadFileProps } from './interface.ts';
import { post, get } from '@/utils/http';
import { message } from 'antd';
import qs from 'qs';

const sandboxBaseUrl = '/api/v1/sessions';

export const uploadFileToSandBox = async (data: SandBoxUploadFileProps) => {
  try {
    const { file, sessionId, filePath } = data;
    const uploadUrl = `${sandboxBaseUrl}/${sessionId}/files/upload?path=${filePath}`;
    const formData = new FormData();
    formData.append('file', file);
    const res = await post(uploadUrl, {
      body: formData,
      headers: { 'Content-Type': 'multipart/form-data' },
    });
    return res || true;
  } catch (error: any) {
    message.error(error.description);
    return false;
  }
};

export const getFileListFromSandBox = async (data: SandBoxQueryFileProps) => {
  try {
    const queryStr = qs.stringify({
      path: data.path,
      limit: data.limit,
    });
    const res = await get(`${sandboxBaseUrl}/${data.sessionId}/files?${queryStr}`);
    return res || true;
  } catch (error: any) {
    message.error(error.description);
    return false;
  }
};
