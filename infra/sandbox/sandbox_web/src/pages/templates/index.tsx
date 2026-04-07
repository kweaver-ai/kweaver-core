/**
 * 模板管理页面
 */
import { useState, useEffect } from 'react';
import { Button, Input, Table, Modal, Form, InputNumber, Select, Tag, Space, Popconfirm, Descriptions } from 'antd';
import { PlusOutlined, SearchOutlined, EditOutlined, DeleteOutlined, EyeOutlined } from '@ant-design/icons';
import { useTemplates } from '@hooks/useTemplates';
import { RUNTIME_TYPE_LABELS, RESOURCE_OPTIONS } from '@constants/runtime';
import type { TemplateResponse, CreateTemplateRequest, UpdateTemplateRequest } from '@apis/templates';
import * as templatesApi from '@apis/templates';

export default function TemplatesPage() {
  const { templates, loading, fetchTemplates, createTemplate, deleteTemplate } =
    useTemplates();
  const [searchQuery, setSearchQuery] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showDetailModal, setShowDetailModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState<TemplateResponse | null>(null);
  const [form] = Form.useForm<CreateTemplateRequest>();
  const [editForm] = Form.useForm<UpdateTemplateRequest>();

  useEffect(() => {
    fetchTemplates();
  }, [fetchTemplates]);

  // 过滤模板
  const filteredTemplates = templates.filter(
    (t) =>
      t.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      t.id.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // 创建模板
  const handleCreate = async () => {
    try {
      const values = await form.validateFields();
      await createTemplate(values);
      setShowCreateModal(false);
      form.resetFields();
    } catch (error) {
      // 表单验证失败或创建失败
    }
  };

  // 查看详情
  const handleViewDetail = async (template: TemplateResponse) => {
    try {
      const detail = await templatesApi.getTemplate(template.id);
      setSelectedTemplate(detail);
      setShowDetailModal(true);
    } catch (error) {
      // 获取详情失败
    }
  };

  // 打开编辑模态框
  const handleOpenEdit = async (template: TemplateResponse) => {
    try {
      const detail = await templatesApi.getTemplate(template.id);
      setSelectedTemplate(detail);
      editForm.setFieldsValue({
        name: detail.name,
        image_url: detail.image_url,
        default_cpu_cores: detail.default_cpu_cores,
        default_memory_mb: detail.default_memory_mb,
        default_disk_mb: detail.default_disk_mb,
        default_timeout: detail.default_timeout_sec,
        default_env_vars: detail.default_env_vars,
      });
      setShowEditModal(true);
    } catch (error) {
      // 获取详情失败
    }
  };

  // 保存编辑
  const handleEditSave = async () => {
    if (!selectedTemplate) return;
    try {
      const values = await editForm.validateFields();
      await templatesApi.updateTemplate(selectedTemplate.id, values);
      setShowEditModal(false);
      editForm.resetFields();
      setSelectedTemplate(null);
      fetchTemplates();
    } catch (error) {
      // 保存失败
    }
  };

  // 删除模板
  const handleDelete = async (id: string) => {
    try {
      await deleteTemplate(id);
    } catch (error) {
      // 删除失败
    }
  };

  // 表格列定义
  const columns = [
    {
      title: '模版ID',
      dataIndex: 'id',
      key: 'id',
      width: 200,
    },
    {
      title: '模版名称',
      dataIndex: 'name',
      key: 'name',
      width: 200,
    },
    {
      title: '运行时',
      dataIndex: 'runtime_type',
      key: 'runtime_type',
      width: 120,
      render: (type: string) => RUNTIME_TYPE_LABELS[type as keyof typeof RUNTIME_TYPE_LABELS] || type,
    },
    {
      title: '资源配置',
      key: 'resources',
      width: 200,
      render: (_: unknown, record: TemplateResponse) =>
        `${record.default_cpu_cores}核 / ${record.default_memory_mb}MB / ${record.default_disk_mb}MB`,
    },
    {
      title: '状态',
      dataIndex: 'is_active',
      key: 'is_active',
      width: 80,
      render: (active: boolean) =>
        active ? (
          <Tag color="success">活跃</Tag>
        ) : (
          <Tag color="error">停用</Tag>
        ),
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
      render: (_: unknown, record: TemplateResponse) => (
        <Space size="small">
          <Button
            type="text"
            icon={<EyeOutlined />}
            size="small"
            title="查看详情"
            onClick={() => handleViewDetail(record)}
          />
          <Button
            type="text"
            icon={<EditOutlined />}
            size="small"
            title="编辑"
            onClick={() => handleOpenEdit(record)}
          />
          <Popconfirm
            title="确认删除"
            description="确定要删除该模版吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="text" danger icon={<DeleteOutlined />} size="small" />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      {/* 页面标题 */}
      <div style={{ marginBottom: 24 }}>
        <div style={{ display: 'flex', alignItems: 'center', marginBottom: 8 }}>
          <div
            style={{
              width: 2,
              height: 18,
              backgroundColor: '#126ee3',
              borderRadius: 4,
              marginRight: 8,
            }}
          />
          <h2
            style={{
              fontSize: 15,
              fontWeight: 500,
              margin: 0,
              color: '#000000',
            }}
          >
            模版管理
          </h2>
        </div>
        <p style={{ fontSize: 13, color: '#677489', marginLeft: 12, marginTop: 0, marginBottom: 0 }}>
          创建和管理代码执行环境模版
        </p>
      </div>

      {/* 内容区域 */}
      <div
        style={{
          backgroundColor: '#ffffff',
          borderRadius: 12,
          border: '1px solid #e7edf7',
          padding: 24,
        }}
      >
        {/* 搜索和操作栏 */}
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 24 }}>
          <Input
            placeholder="搜索模版名称或ID"
            prefix={<SearchOutlined />}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={{ width: 320 }}
            allowClear
          />
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setShowCreateModal(true)}>
            创建模版
          </Button>
        </div>

        {/* 模板表格 */}
        <Table
          columns={columns}
          dataSource={filteredTemplates}
          rowKey="id"
          loading={loading}
          pagination={{ pageSize: 10 }}
          style={{ overflow: 'auto' }}
        />
      </div>

      {/* 创建模版对话框 */}
      <Modal
        title="创建新模版"
        open={showCreateModal}
        onOk={handleCreate}
        onCancel={() => {
          setShowCreateModal(false);
          form.resetFields();
        }}
        width={600}
        okText="创建"
        cancelText="取消"
      >
        <Form form={form} layout="vertical" style={{ marginTop: 24 }}>
          <Form.Item
            name="id"
            label="模版ID"
            rules={[{ required: true, message: '请输入模版ID' }]}
          >
            <Input placeholder="例如: python-3.11-advanced" />
          </Form.Item>

          <Form.Item
            name="name"
            label="模版名称"
            rules={[{ required: true, message: '请输入模版名称' }]}
          >
            <Input placeholder="例如: Python 3.11 高级环境" />
          </Form.Item>

          <Form.Item
            name="image_url"
            label="镜像 URL"
            rules={[{ required: true, message: '请输入镜像 URL' }]}
          >
            <Input placeholder="例如: python:3.11-slim" />
          </Form.Item>

          <Form.Item
            name="runtime_type"
            label="运行时类型"
            rules={[{ required: true, message: '请选择运行时类型' }]}
            initialValue="python3.11"
          >
            <Select>
              {Object.entries(RUNTIME_TYPE_LABELS).map(([value, label]) => (
                <Select.Option key={value} value={value}>
                  {label}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
            <Form.Item
              name="default_cpu_cores"
              label="CPU核心数"
              rules={[{ required: true, message: '请输入CPU核心数' }]}
              initialValue={1}
            >
              <InputNumber min={0.1} max={4} step={0.1} style={{ width: '100%' }} />
            </Form.Item>

            <Form.Item
              name="default_memory_mb"
              label="内存(MB)"
              rules={[{ required: true, message: '请输入内存' }]}
              initialValue={512}
            >
              <InputNumber min={128} max={8192} style={{ width: '100%' }} />
            </Form.Item>

            <Form.Item
              name="default_disk_mb"
              label="磁盘(MB)"
              rules={[{ required: true, message: '请输入磁盘' }]}
              initialValue={1024}
            >
              <InputNumber min={256} max={51200} style={{ width: '100%' }} />
            </Form.Item>

            <Form.Item
              name="default_timeout_sec"
              label="超时(秒)"
              rules={[{ required: true, message: '请输入超时时间' }]}
              initialValue={300}
            >
              <InputNumber min={60} max={3600} style={{ width: '100%' }} />
            </Form.Item>
          </div>
        </Form>
      </Modal>

      {/* 详情模态框 */}
      <Modal
        title="模板详情"
        open={showDetailModal}
        onCancel={() => {
          setShowDetailModal(false);
          setSelectedTemplate(null);
        }}
        footer={[
          <Button key="close" onClick={() => setShowDetailModal(false)}>
            关闭
          </Button>,
        ]}
        width={700}
      >
        {selectedTemplate && (
          <Descriptions bordered column={2} size="small">
            <Descriptions.Item label="模板ID" span={2}>
              <code>{selectedTemplate.id}</code>
            </Descriptions.Item>
            <Descriptions.Item label="模板名称" span={2}>
              {selectedTemplate.name}
            </Descriptions.Item>
            <Descriptions.Item label="镜像 URL" span={2}>
              <code>{selectedTemplate.image_url}</code>
            </Descriptions.Item>
            <Descriptions.Item label="运行时类型">
              {RUNTIME_TYPE_LABELS[selectedTemplate.runtime_type] || selectedTemplate.runtime_type}
            </Descriptions.Item>
            <Descriptions.Item label="状态">
              {selectedTemplate.is_active ? <Tag color="success">活跃</Tag> : <Tag color="error">停用</Tag>}
            </Descriptions.Item>
            <Descriptions.Item label="CPU核心数">
              {selectedTemplate.default_cpu_cores}
            </Descriptions.Item>
            <Descriptions.Item label="内存(MB)">
              {selectedTemplate.default_memory_mb}
            </Descriptions.Item>
            <Descriptions.Item label="磁盘(MB)">
              {selectedTemplate.default_disk_mb}
            </Descriptions.Item>
            <Descriptions.Item label="超时(秒)">
              {selectedTemplate.default_timeout_sec}
            </Descriptions.Item>
            {selectedTemplate.created_at && (
              <Descriptions.Item label="创建时间" span={2}>
                {selectedTemplate.created_at}
              </Descriptions.Item>
            )}
            {selectedTemplate.updated_at && (
              <Descriptions.Item label="更新时间" span={2}>
                {selectedTemplate.updated_at}
              </Descriptions.Item>
            )}
          </Descriptions>
        )}
      </Modal>

      {/* 编辑模态框 */}
      <Modal
        title="编辑模板"
        open={showEditModal}
        onOk={handleEditSave}
        onCancel={() => {
          setShowEditModal(false);
          editForm.resetFields();
          setSelectedTemplate(null);
        }}
        width={600}
        okText="保存"
        cancelText="取消"
      >
        <Form form={editForm} layout="vertical" style={{ marginTop: 24 }}>
          <Form.Item
            name="name"
            label="模版名称"
            rules={[{ required: true, message: '请输入模版名称' }]}
          >
            <Input placeholder="例如: Python 3.11 高级环境" />
          </Form.Item>

          <Form.Item
            name="image_url"
            label="镜像 URL"
            rules={[{ required: true, message: '请输入镜像 URL' }]}
          >
            <Input placeholder="例如: python:3.11-slim" />
          </Form.Item>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
            <Form.Item
              name="default_cpu_cores"
              label="CPU核心数"
              rules={[{ required: true, message: '请输入CPU核心数' }]}
            >
              <InputNumber min={0.1} max={4} step={0.1} style={{ width: '100%' }} />
            </Form.Item>

            <Form.Item
              name="default_memory_mb"
              label="内存(MB)"
              rules={[{ required: true, message: '请输入内存' }]}
            >
              <InputNumber min={128} max={8192} style={{ width: '100%' }} />
            </Form.Item>

            <Form.Item
              name="default_disk_mb"
              label="磁盘(MB)"
              rules={[{ required: true, message: '请输入磁盘' }]}
            >
              <InputNumber min={256} max={51200} style={{ width: '100%' }} />
            </Form.Item>

            <Form.Item
              name="default_timeout"
              label="超时(秒)"
              rules={[{ required: true, message: '请输入超时时间' }]}
            >
              <InputNumber min={60} max={3600} style={{ width: '100%' }} />
            </Form.Item>
          </div>
        </Form>
      </Modal>
    </>
  );
}
