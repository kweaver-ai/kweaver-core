import React, { useEffect, useMemo } from 'react';
import intl from 'react-intl-universal';
import classNames from 'classnames';
import { Input, Select, Button, Modal, Radio, Form, Tooltip } from 'antd';
import { CloseOutlined, QuestionCircleOutlined } from '@ant-design/icons';
import { useMicroWidgetProps } from '@/hooks';
import UploadTooltipImage from '@/assets/icons/upload-sample.svg';
import UploadTooltipImageTW from '@/assets/icons/upload-sample-tw.svg';
import UploadTooltipImageUS from '@/assets/icons/upload-sample-us.svg';
import TempZoneTooltipImage from '@/assets/icons/temp-zone-sample.svg';
import TempZoneTiiltipImageTW from '@/assets/icons/temp-zone-sample-tw.svg';
import TempZoneTiiltipImageUS from '@/assets/icons/temp-zone-sample-us.svg';
import styles from './FileSettingsModal.module.less';
const { Option } = Select;

// 临时区配置类型
interface TempZoneConfig {
  name: string;
  max_file_count: number;
  single_chat_max_select_file_count: number;
  single_file_size_limit: number;
  single_file_size_limit_unit: string;
  support_data_type: string[];
  allowed_file_categories: string[];
  allowed_file_types: string[];
  tmp_file_use_type: string;
}

// 文件设置弹窗组件
const FileSettingsModal: React.FC<{
  isEditable: boolean;
  onCancel: () => void;
  onOk: (config: TempZoneConfig) => void;
  initialConfig: TempZoneConfig;
}> = ({ isEditable = true, onCancel, onOk, initialConfig }) => {
  const [form] = Form.useForm();
  const microWidgetProps = useMicroWidgetProps();
  const [isZH, isTW] = useMemo(() => {
    const lang = microWidgetProps.language.getLanguage;
    const isUS = lang === 'en-us';
    const isTW = lang === 'zh-tw';
    const isZH = !(isUS || isTW);
    return [isZH, isTW];
  }, []);

  // 文件类别选项
  const fileCategories = useMemo(
    () => [
      { label: intl.get('dataAgent.document'), value: 'document' },
      { label: intl.get('dataAgent.spreadsheet'), value: 'spreadsheet' },
      { label: intl.get('dataAgent.presentation'), value: 'presentation' },
      { label: 'PDF', value: 'pdf' },
      { label: intl.get('dataAgent.text'), value: 'text' },
      { label: intl.get('dataAgent.audio'), value: 'audio' },
      { label: intl.get('dataAgent.video'), value: 'video' },
      { label: intl.get('dataAgent.other'), value: 'other' },
    ],
    []
  );

  useEffect(() => {
    // 获取初始文件类型列表
    const initialCategories = initialConfig?.allowed_file_categories?.length
      ? initialConfig.allowed_file_categories
      : ['document']; // 设置文档为默认类型

    form.setFieldsValue({
      allowed_file_categories: initialCategories,
      single_file_size_limit: initialConfig?.single_file_size_limit || undefined,
      max_file_count: initialConfig?.max_file_count || undefined,
      single_chat_max_select_file_count: initialConfig?.single_chat_max_select_file_count || undefined,
      tmp_file_use_type: initialConfig?.tmp_file_use_type || 'upload',
    });
  }, [initialConfig, form]);

  const handleOk = () => {
    form.validateFields().then(values => {
      const singleFileSizeLimit = values.single_file_size_limit ? parseInt(values.single_file_size_limit, 10) : 100;

      const maxFileCount = values.max_file_count ? parseInt(values.max_file_count, 10) : 50;
      const singleChatMaxSelectFileCount = values.single_chat_max_select_file_count
        ? parseInt(values.single_chat_max_select_file_count, 10)
        : 5;

      // 移除全选选项，只保留实际的文件类型
      let fileCategories = values.allowed_file_categories || [];
      fileCategories = fileCategories.filter((cat: string) => cat !== 'all');

      onOk({
        ...initialConfig,
        ...values,
        name: '临时上传区',
        single_file_size_limit: singleFileSizeLimit,
        max_file_count: maxFileCount,
        single_chat_max_select_file_count: singleChatMaxSelectFileCount,
        single_file_size_limit_unit: 'MB',
        support_data_type: ['file'],
        allowed_file_categories: fileCategories,
        allowed_file_types: ['*'],
        tmp_file_use_type: values.tmp_file_use_type || 'upload',
      });
    });
  };

  // 全选所有文件类型
  const handleSelectAll = () => {
    const allCategories = fileCategories.map(cat => cat.value);
    form.setFieldsValue({ allowed_file_categories: allCategories });
  };

  // 处理全选项变更
  const handleCategoryChange = (values: string[]) => {
    if (values.includes('all')) {
      // 如果选择了全选，则选中所有选项（除了all自身）
      handleSelectAll();
    } else {
      // 如果取消了全选但已经全部选中，则保留当前选择
      form.setFieldsValue({ allowed_file_categories: values });
    }
  };

  return (
    <Modal
      title={intl.get('dataAgent.fileSettings')}
      open={true}
      onCancel={onCancel}
      onOk={handleOk}
      width={520}
      closeIcon={<CloseOutlined />}
      maskClosable={false}
      centered
      className={styles['file-settings-modal']}
      footer={
        isEditable
          ? [
              <Button key="submit" type="primary" className="dip-min-width-72" onClick={handleOk}>
                {intl.get('dataAgent.ok')}
              </Button>,
              <Button key="cancel" className="dip-min-width-72" onClick={onCancel}>
                {intl.get('dataAgent.cancel')}
              </Button>,
            ]
          : []
      }
    >
      <Form form={form} layout="vertical">
        <Form.Item
          label={intl.get('dataAgent.fileSelectionMethod')}
          required
          className={classNames(styles['file-selection-method'], 'dip-mb-24')}
          name="tmp_file_use_type"
        >
          <Radio.Group disabled={!isEditable}>
            <div className={classNames(styles['radio-option-item'], 'dip-mb-4')}>
              <Radio value="upload">{intl.get('dataAgent.selectFromInputBox')}</Radio>
              <Tooltip
                title={isZH ? <UploadTooltipImage /> : isTW ? <UploadTooltipImageTW /> : <UploadTooltipImageUS />}
                placement="right"
                color="white"
                overlayClassName={styles['custom-tooltip']}
              >
                <QuestionCircleOutlined className={styles['help-icon']} />
              </Tooltip>
            </div>
            <div className={styles['radio-option-item']}>
              <Radio value="select_from_temp_zone">{intl.get('dataAgent.selectFromTemporarySessionArea')}</Radio>
              <Tooltip
                title={isZH ? <TempZoneTooltipImage /> : isTW ? <TempZoneTiiltipImageTW /> : <TempZoneTiiltipImageUS />}
                placement="right"
                color="white"
                overlayClassName={styles['custom-tooltip']}
              >
                <QuestionCircleOutlined className={styles['help-icon']} style={{ width: 16, height: 16 }} />
              </Tooltip>
            </div>
          </Radio.Group>
        </Form.Item>
        <div className="dip-required dip-c-bold dip-mb-8">{intl.get('dataAgent.fileRuleConfiguration')}</div>
        <Form.Item
          name="allowed_file_categories"
          label={intl.get('dataAgent.supportedFileTypes')}
          rules={[{ required: true, message: intl.get('dataAgent.pleaseSelectSupportedFileTypes') }]}
        >
          <Select
            disabled={!isEditable}
            placeholder={intl.get('dataAgent.config.pleaseSelect')}
            mode="multiple"
            optionLabelProp="label"
            suffixIcon={<i className="anticon" />}
            onChange={handleCategoryChange}
          >
            <Option key="all" value="all" label={intl.get('dataAgent.selectAll')}>
              {intl.get('dataAgent.selectAll')}
            </Option>
            {fileCategories.map(cat => (
              <Option key={cat.value} value={cat.value} label={cat.label}>
                {cat.label}
              </Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item
          name="single_file_size_limit"
          label={intl.get('dataAgent.singleFileSizeLimit')}
          rules={[
            { required: true, message: intl.get('dataAgent.pleaseEnterFileSizeLimit') },
            {
              validator: (_, value) => {
                const num = Number(value);
                if (isNaN(num)) {
                  return Promise.reject(intl.get('dataAgent.pleaseEnterValidNumber'));
                }
                if (num < 0 || num === 0) {
                  return Promise.reject(intl.get('dataAgent.fileSizeCannotBeLessThanOrEqualToZero'));
                }
                if (num > 100) {
                  return Promise.reject(intl.get('dataAgent.fileSizeCannotExceed100MB'));
                }
                return Promise.resolve();
              },
            },
          ]}
          validateFirst={true}
        >
          <Input
            disabled={!isEditable}
            placeholder={intl.get('dataAgent.pleaseEnterMax100MB')}
            suffix="MB"
            type="number"
            min={1}
            max={100}
          />
        </Form.Item>

        <Form.Item
          noStyle
          shouldUpdate={(prevValues, currentValues) => prevValues.tmp_file_use_type !== currentValues.tmp_file_use_type}
        >
          {({ getFieldValue }) =>
            getFieldValue('tmp_file_use_type') === 'select_from_temp_zone' ? (
              <Form.Item
                name="max_file_count"
                label={intl.get('dataAgent.temporaryAreaMaxFileCount')}
                rules={[
                  { required: true, message: intl.get('dataAgent.pleaseEnterSupportedFileCount') },
                  {
                    validator: (_, value) => {
                      const num = Number(value);
                      if (isNaN(num)) {
                        return Promise.reject(intl.get('dataAgent.pleaseEnterValidNumber'));
                      }
                      if (num < 0 || num === 0) {
                        return Promise.reject(intl.get('dataAgent.maxFileCountCannotBeLessThanOrEqualToZero'));
                      }
                      if (num > 50) {
                        return Promise.reject(intl.get('dataAgent.maxFileCountCannotExceed50'));
                      }
                      return Promise.resolve();
                    },
                  },
                ]}
                validateFirst={true}
              >
                <Input
                  disabled={!isEditable}
                  placeholder={intl.get('dataAgent.pleaseEnterMax50Files')}
                  suffix={intl.get('dataAgent.fileCountUnit')}
                  type="number"
                  min={1}
                  max={50}
                />
              </Form.Item>
            ) : null
          }
        </Form.Item>

        <Form.Item
          name="single_chat_max_select_file_count"
          label={intl.get('dataAgent.maxFileCountPerConversation')}
          rules={[
            { required: true, message: intl.get('dataAgent.pleaseEnterSupportedFileCount') },
            {
              validator: (_, value) => {
                const num = Number(value);
                if (isNaN(num)) {
                  return Promise.reject(intl.get('dataAgent.pleaseEnterValidNumber'));
                }
                if (num <= 0) {
                  return Promise.reject(intl.get('dataAgent.fileCountNotLessThanOrEqualToZero'));
                }
                if (num > 5) {
                  return Promise.reject(intl.get('dataAgent.fileCountNotExceed5'));
                }
                return Promise.resolve();
              },
            },
          ]}
          validateFirst={true}
        >
          <Input
            disabled={!isEditable}
            placeholder={intl.get('dataAgent.pleaseEnterMax5Files')}
            suffix={intl.get('dataAgent.fileCountUnit')}
            type="number"
            min={1}
            max={5}
          />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default FileSettingsModal;
