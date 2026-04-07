/**
 * File API 实现
 */
import { API_ENDPOINTS } from '@/constants/api';
import apiClient from '@/utils/http/axios';
import type { FileUploadResponse } from './types';

/**
 * 上传文件到会话工作区
 * @param sessionId 会话ID
 * @param file 要上传的文件
 * @param path 文件在会话工作区中的路径（作为 URL 参数，默认使用文件名）
 */
export async function uploadFile(
  sessionId: string,
  file: File,
  path?: string
): Promise<FileUploadResponse> {
  const formData = new FormData();
  formData.append('file', file);

  // path 是必填参数，如果未提供则使用文件名
  const uploadPath = path || file.name;

  const response = await apiClient.post<FileUploadResponse>(
    API_ENDPOINTS.UPLOAD_FILE(sessionId),
    formData,
    {
      params: { path: uploadPath },
      headers: {
        // 覆盖默认的 application/json，让 axios 自动检测 FormData 并设置正确的 Content-Type
        'Content-Type': undefined as any,
      },
    }
  );
  return response.data;
}

/**
 * 从会话工作区下载文件
 */
export async function downloadFile(sessionId: string, filePath: string): Promise<Blob> {
  const response = await apiClient.get<Blob>(
    API_ENDPOINTS.DOWNLOAD_FILE(sessionId, filePath),
    {
      responseType: 'blob',
    }
  );
  return response.data;
}
