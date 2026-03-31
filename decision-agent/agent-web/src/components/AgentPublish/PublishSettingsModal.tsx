import React, { useEffect, useMemo, useState } from 'react';
import { Modal, Form, Select, Radio, Input, Button, Tag, Checkbox, message, Drawer } from 'antd';
import intl from 'react-intl-universal';
import { uniqBy } from 'lodash';
import {
  getAgentCategoryList,
  publishAgent,
  AgentPublishToWhereEnum,
  AgentPublishToBeEnum,
  getPublishedAgentInfo,
  updateAgentPublishInfo,
  publishTemplate,
  publishAgentAsTemplate,
  // getPublishedTemplateInfo,
  getAgentManagementPerm,
  type AgentManagementPermType,
  type PublishData,
} from '@/apis/agent-factory';
import { Category } from '@/apis/agent-factory/type';
import { apis, components } from '@aishu-tech/components/dist/dip-components.min';
import { useMicroWidgetProps } from '@/hooks';
import { getRoleByUserInfo, RoleTypeEnum } from '@/utils/role';
import { formatAccessor, formatPMSControl, transformPMSControlFromBackend } from './utils';
import './PublishSettingsModal.less';
import styles from './PublishSettingsModal.module.less';
import { VisibleRangeEnum, PublishModeEnum } from './types';
export { PublishModeEnum };

const { TextArea } = Input;

interface PublishSettingsModalProps {
  agent: any;
  mode?: PublishModeEnum;
  onCancel: () => void;
  onOk: (params: { publish_to_bes?: AgentPublishToBeEnum[]; published_at?: number }) => void;
}

/**
 * 发布设置弹窗
 */
const PublishSettingsModal: React.FC<PublishSettingsModalProps> = ({
  agent,
  mode = PublishModeEnum.PublishAgent,
  onCancel,
  onOk,
}) => {
  const microWidgetProps = useMicroWidgetProps();
  const [form] = Form.useForm();

  const [categories, setCategories] = useState<Category[]>([]);

  // 可见范围
  const [visibleRange, setVisibleRange] = useState<VisibleRangeEnum>(VisibleRangeEnum.AllUser); // 可见范围的选中项
  const [isVisibleRangeHovered, setIsVisibleRangeHovered] = useState(false); // 指定范围元素是否正在hover
  const [accessors, setAccessors] = useState<any[]>([]); // 指定范围选择的用户
  const [visibleRangeError, setVisibleRangeError] = useState<string>(''); // 可见范围的校验错误信息

  // 发布到
  const [publishToWhere, setPublishToWhere] = useState<AgentPublishToWhereEnum[]>([AgentPublishToWhereEnum.Square]); // 发布到的选中项
  const [publishToWhereError, setPublishToWhereError] = useState<string>(''); // 发布到的校验错误信息

  const [loading, setLoading] = useState<boolean>(false);
  // 权限
  const [perms, setPerms] = useState<AgentManagementPermType>({
    custom_space: {
      create: false,
    },
    agent: {
      publish: false,
      unpublish: false,
      unpublish_other_user_agent: false,
      publish_to_be_api_agent: false,
      publish_to_be_data_flow_agent: false,
      publish_to_be_skill_agent: false,
      publish_to_be_web_sdk_agent: false,
      create_system_agent: false,
    },
    agent_tpl: {
      publish: false,
      unpublish: false,
      unpublish_other_user_agent_tpl: false,
    },
  });

  const theme = microWidgetProps?.config?.getTheme?.normal;
  const isTemplate = useMemo(
    () => [PublishModeEnum.PublishTemplate, PublishModeEnum.PublishAgentAsTemplate].includes(mode),
    [mode]
  );
  const title = useMemo(
    () =>
      mode === PublishModeEnum.UpdatePublishAgent
        ? intl.get('dataAgent.updatePublishSettings')
        : mode === PublishModeEnum.PublishAgent
          ? intl.get('dataAgent.publishSettings')
          : intl.get('dataAgent.publishTemplate'),
    [mode]
  );

  // 添加用户
  const addUser = () => {
    let container: any = document.createElement('div');
    const { isAdmin, roleType } = getRoleByUserInfo(microWidgetProps?.config?.userInfo);
    // 只有超级管理员、系统管理员能查看应用账号
    const showApp = [RoleTypeEnum.SuperAdmin, RoleTypeEnum.SysAdmin].includes(roleType);
    const unmount = apis.mountComponent(
      components.AccessorPicker,
      {
        isAdmin,
        tabs: showApp ? ['organization', 'group', 'app'] : ['organization', 'group'],
        range: showApp ? ['user', 'department', 'group', 'app'] : ['user', 'department', 'group'],
        role: roleType,
        title: intl.get('dataAgent.visibilityRange'),
        onSelect: async (items: any[]) => {
          const added = items.map(formatAccessor);
          setAccessors(original => uniqBy([...added, ...original], 'id'));
          unmount();
          container = null;
        },
        onCancel: () => {
          unmount();
          container = null;
        },
      },
      container
    );
  };

  // 获取分类列表
  const fetchCategoryList = async () => {
    try {
      const categories = await getAgentCategoryList();
      setCategories(categories);
    } catch (error) {
      console.error('获取分类数据时发生错误', error);
    }
  };

  // 自定义处理取消
  const handleCancel = () => {
    onCancel();
  };

  // 校验发布到的值
  const validatePublishPlatformValue = (
    publishToWhere?: AgentPublishToWhereEnum[],
    publishToBe?: AgentPublishToBeEnum[]
  ) => {
    if (!publishToWhere?.length && !publishToBe?.length) {
      message.info(intl.get('dataAgent.pleaseSelectPublishWay'));
      return false;
    }

    return true;
  };

  // 校验可见范围的值
  const validateVisibleRangeValue = () => {
    if (visibleRange === VisibleRangeEnum.SpecifiedRange && accessors?.length === 0) {
      setVisibleRangeError(intl.get('dataAgent.pleaseSelectSpecifiedRange'));
      return false;
    }

    return true;
  };

  // 提交表单
  const handleSubmit = async () => {
    const values = await form.validateFields();
    const isPublishPlarformValid = isTemplate ? true : validatePublishPlatformValue(publishToWhere, values.publishToBe);
    const isVisibleRangeValid = isTemplate ? true : validateVisibleRangeValue();

    if (!isPublishPlarformValid || !isVisibleRangeValid) return;

    const publishData = {
      business_domain_id: microWidgetProps.businessDomainID,
      category_ids: [values.category_id],
      description: values.description || '',
      publish_to_where: publishToWhere,
      pms_control: formatPMSControl(accessors),
      publish_to_bes: values.publishToBe,
    };

    publish(publishData);
  };

  const publish = async (publishData: PublishData) => {
    try {
      setLoading(true);

      let result;

      switch (mode) {
        case PublishModeEnum.PublishAgent:
          // 发布agent
          result = await publishAgent(agent.id, publishData);
          break;

        case PublishModeEnum.UpdatePublishAgent:
          // 更新发布agent
          result = await updateAgentPublishInfo(agent.id, publishData);
          break;

        case PublishModeEnum.PublishTemplate:
          // 发布模板
          result = await publishTemplate({
            id: agent.tpl_id ?? agent.id,
            category_ids: publishData.category_ids,
            business_domain_id: publishData.business_domain_id,
          });
          break;

        case PublishModeEnum.PublishAgentAsTemplate:
          // 将agent发布为模板
          result = await publishAgentAsTemplate({
            agent_id: agent.id,
            category_ids: publishData.category_ids,
            business_domain_id: publishData.business_domain_id,
          });
          break;

        default:
          break;
      }

      message.success(intl.get('dataAgent.operationSuccess'));
      onOk({
        published_at: result?.published_at,
        ...([PublishModeEnum.PublishTemplate, PublishModeEnum.PublishAgentAsTemplate].includes(mode)
          ? {}
          : { publish_to_bes: publishData.publish_to_bes }),
      });
    } catch (ex: any) {
      if (ex?.description) {
        message.error(ex.description);
      }
    } finally {
      setLoading(false);
    }
  };

  // 获取已发布agent信息
  const loadPublishedAgentInfo = async () => {
    try {
      const { categories, description, publish_to_where, pms_control, publish_to_bes } = await getPublishedAgentInfo(
        agent.id
      );

      form.setFieldsValue({
        category_id: categories?.[0]?.id,
        description,
        publishToBe: publish_to_bes,
      });
      setPublishToWhere(publish_to_where);
      const { visibleRange, accessors } = transformPMSControlFromBackend(pms_control);
      setVisibleRange(visibleRange);
      setAccessors(accessors);
    } catch (ex: any) {
      // 发布agent时，如果获取已发布信息报错，无需处理错误
      if (mode === PublishModeEnum.PublishAgent) return;

      if (ex?.description) {
        message.error(ex.description);
      }
    }
  };

  // 获取已发布模板信息
  // const loadPublishedTemplateInfo = async () => {
  //   try {
  //     const { categories } = await getPublishedTemplateInfo(agent.id);

  //     form.setFieldsValue({
  //       category_id: categories?.[0]?.id,
  //     });
  //   } catch (ex: any) {
  //     // 发布模板时，如果获取已发布信息报错，无需处理错误
  //     if (mode === PublishModeEnum.PublishTemplate) return;

  //     if (ex?.description) {
  //       message.error(ex.description);
  //     }
  //   }
  // };

  useEffect(() => {
    setPublishToWhereError('');
  }, [publishToWhere]);

  useEffect(() => {
    if (visibleRange === VisibleRangeEnum.AllUser) {
      setAccessors([]);
    }
    setVisibleRangeError('');
  }, [visibleRange]);

  useEffect(() => {
    setVisibleRangeError('');
  }, [accessors]);

  // 加载分类列表
  useEffect(() => {
    fetchCategoryList();
  }, []);

  useEffect(() => {
    // 获取已发布信息，用于回显在弹窗上
    if ([PublishModeEnum.PublishAgent, PublishModeEnum.UpdatePublishAgent].includes(mode)) {
      loadPublishedAgentInfo();
    }
    // 模板不用回显
    // else {
    //   loadPublishedTemplateInfo();
    // }
  }, []);

  useEffect(() => {
    const fetchPerms = async () => {
      try {
        // 获取权限
        const perms = await getAgentManagementPerm();
        setPerms(perms);
      } catch (ex: any) {
        if (ex?.description) {
          message.error(ex.description);
        }
      }
    };

    fetchPerms();
  }, []);

  const renderContent = () => (
    <Form
      form={form}
      layout="vertical"
      initialValues={{
        category_id: undefined,
        description: '',
        permission_range: [],
        publishToWhere: [AgentPublishToWhereEnum.Square],
        visibleRange,
        publishCategory: [],
      }}
    >
      <Form.Item
        label={intl.get('dataAgent.category')}
        name="category_id"
        rules={[{ required: true, message: intl.get('dataAgent.pleaseSelectCategory') }]}
        required
      >
        <Select placeholder={intl.get('dataAgent.pleaseSelectCategory')}>
          {categories.map(category => (
            <Select.Option key={category.category_id} value={category.category_id}>
              {category.name}
            </Select.Option>
          ))}
        </Select>
      </Form.Item>

      {!isTemplate && (
        <>
          <Form.Item label={intl.get('dataAgent.publishDescription')} name="description">
            <TextArea
              placeholder={intl.get('dataAgent.pleaseEnter')}
              autoSize={{ minRows: 4, maxRows: 6 }}
              showCount
              maxLength={200}
            />
          </Form.Item>

          <Form.Item label={intl.get('dataAgent.publishTo')} name="publishToWhere">
            <div style={{ marginBottom: '-16px' }}>
              <Checkbox.Group
                options={[
                  {
                    label: (
                      <>
                        {intl.get('dataAgent.publishToDataAgent')}
                        <div className={styles['tip-text']}>{intl.get('dataAgent.publishAgentToMarketDesc')}</div>
                      </>
                    ),
                    value: AgentPublishToWhereEnum.Square,
                  },
                ]}
                className={styles['publish-platform']}
                value={publishToWhere}
                onChange={setPublishToWhere}
              />
              <div className={styles['error']}>
                {Boolean(publishToWhereError) && <div className={styles['error-text']}>{publishToWhereError}</div>}
              </div>
            </div>
          </Form.Item>

          <Form.Item
            label={intl.get('dataAgent.visibilityRange')}
            name="visibleRange"
            rules={[{ required: true, message: intl.get('dataAgent.pleaseSelectVisibilityRange') }]}
          >
            <div style={{ marginBottom: '-16px' }}>
              <Radio.Group
                options={[
                  {
                    label: (
                      <>
                        {intl.get('dataAgent.allUsers')}
                        <div className={styles['tip-text']}>{intl.get('dataAgent.visibleAndUsableByAll')}</div>
                      </>
                    ),
                    value: VisibleRangeEnum.AllUser,
                  },
                  {
                    label: (
                      <>
                        {intl.get('dataAgent.specifiedRange')}
                        <div className={styles['tip-text']}>
                          {intl.get('dataAgent.visibleAndUsableBySpecifiedRangeUsers')}
                        </div>
                      </>
                    ),
                    value: VisibleRangeEnum.SpecifiedRange,
                  },
                ]}
                className={styles['visible-range']}
                value={visibleRange}
                onChange={e => {
                  setVisibleRange(e.target.value);
                }}
              />
              {visibleRange === VisibleRangeEnum.SpecifiedRange && (
                <div
                  className={styles['accessor-list']}
                  style={isVisibleRangeHovered ? { borderColor: theme } : {}}
                  onClick={addUser}
                  onMouseEnter={() => setIsVisibleRangeHovered(true)}
                  onMouseLeave={() => setIsVisibleRangeHovered(false)}
                >
                  {accessors.map(user => (
                    <Tag
                      className={styles['tag']}
                      closable
                      onClose={() => {
                        setAccessors(origin => origin.filter(({ id }) => id !== user.id));
                      }}
                    >
                      <span className={styles['user-name']} key={user.id} title={user.name}>
                        {user.name}
                      </span>
                    </Tag>
                  ))}
                  {accessors.length === 0 && (
                    <span className={styles['placeholder']}>{intl.get('dataAgent.config.pleaseSelect')}</span>
                  )}
                </div>
              )}
              <div className={styles['error']}>
                {Boolean(visibleRangeError) && <div className={styles['error-text']}>{visibleRangeError}</div>}
              </div>
            </div>
          </Form.Item>

          <Form.Item label={intl.get('dataAgent.publishAsOtherTypesSimultaneously')} name="publishToBe">
            <Checkbox.Group
              options={[
                {
                  label: 'API',
                  value: AgentPublishToBeEnum.ApiAgent,
                  disabled: !perms?.agent?.publish_to_be_api_agent,
                },
                {
                  label: intl.get('dataAgent.config.skill'),
                  value: AgentPublishToBeEnum.SkillAgent,
                  disabled: !perms?.agent?.publish_to_be_skill_agent,
                },
              ]}
              className={styles['publish-category']}
            />
          </Form.Item>
        </>
      )}
    </Form>
  );

  return isTemplate ? (
    <Modal
      title={title}
      open={true}
      onCancel={handleCancel}
      destroyOnHidden
      width={520}
      styles={{
        body: {
          maxHeight: '600px',
          overflowY: 'auto',
          marginRight: '-24px',
          paddingRight: '24px',
        },
      }}
      footer={[
        <Button key="submit" type="primary" className="dip-min-width-72" loading={loading} onClick={handleSubmit}>
          {intl.get('dataAgent.ok')}
        </Button>,
        <Button key="cancel" className="dip-min-width-72" onClick={handleCancel}>
          {intl.get('dataAgent.cancel')}
        </Button>,
      ]}
      maskClosable={false}
      centered
    >
      {renderContent()}
    </Modal>
  ) : (
    <Drawer
      title={title}
      open={true}
      onClose={handleCancel}
      destroyOnHidden
      width={'40%'}
      rootClassName={styles['drawer']}
    >
      <div className="dip-flex-column dip-gap-16 dip-h-100">
        <div style={{ flex: 1, marginRight: '-24px' }} className="dip-pr-24 dip-overflow-auto">
          {renderContent()}
        </div>
        <div className="dip-flex-content-end dip-gap-8">
          <Button key="submit" type="primary" className="dip-min-width-72" loading={loading} onClick={handleSubmit}>
            {intl.get('dataAgent.ok')}
          </Button>
          <Button key="cancel" className="dip-min-width-72" onClick={handleCancel}>
            {intl.get('dataAgent.cancel')}
          </Button>
        </div>
      </div>
    </Drawer>
  );
};

export default PublishSettingsModal;
