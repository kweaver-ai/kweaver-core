/**
 * 模板管理 Hook
 */
import { useState, useCallback } from 'react';
import { message } from 'antd';
import * as templatesApi from '@apis/templates';
import type {
  TemplateResponse,
  CreateTemplateRequest,
  UpdateTemplateRequest,
} from '@apis/templates';

export function useTemplates() {
  const [templates, setTemplates] = useState<TemplateResponse[]>([]);
  const [loading, setLoading] = useState(false);

  // 获取模板列表
  const fetchTemplates = useCallback(async () => {
    setLoading(true);
    try {
      const data = await templatesApi.listTemplates();
      setTemplates(data);
    } catch (error) {
      message.error('获取模板列表失败');
      console.error(error);
    } finally {
      setLoading(false);
    }
  }, []);

  // 创建模板
  const createTemplate = useCallback(async (data: CreateTemplateRequest) => {
    setLoading(true);
    try {
      const newTemplate = await templatesApi.createTemplate(data);
      setTemplates((prev) => [...prev, newTemplate]);
      message.success('模板创建成功');
      return newTemplate;
    } catch (error) {
      message.error('模板创建失败');
      console.error(error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, []);

  // 更新模板
  const updateTemplate = useCallback(
    async (id: string, data: UpdateTemplateRequest) => {
      setLoading(true);
      try {
        const updated = await templatesApi.updateTemplate(id, data);
        setTemplates((prev) =>
          prev.map((t) => (t.id === id ? updated : t))
        );
        message.success('模板更新成功');
        return updated;
      } catch (error) {
        message.error('模板更新失败');
        console.error(error);
        throw error;
      } finally {
        setLoading(false);
      }
    },
    []
  );

  // 删除模板
  const deleteTemplate = useCallback(async (id: string) => {
    setLoading(true);
    try {
      await templatesApi.deleteTemplate(id);
      setTemplates((prev) => prev.filter((t) => t.id !== id));
      message.success('模板删除成功');
    } catch (error) {
      message.error('模板删除失败');
      console.error(error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    templates,
    loading,
    fetchTemplates,
    createTemplate,
    updateTemplate,
    deleteTemplate,
  };
}
