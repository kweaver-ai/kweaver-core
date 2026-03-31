import React, { useState, useEffect, useMemo } from 'react';
import intl from 'react-intl-universal';
import { useParams } from 'react-router-dom';
import classNames from 'classnames';
import { Button, Table, Tag, Spin, message, Tooltip, Tabs, Typography, Select, Checkbox } from 'antd';
import {
  getAgentDetail,
  getAgentDetailInUsagePage,
  getAgentVersionList,
  getTemplateDetail,
  getPublishedTemplateInfo,
  getPublishedAgentInfo,
  type PublishedTemplateInfoType,
  type PublisheAgentInfoType,
  AgentPublishToWhereEnum,
  AgentPublishToBeEnum,
  getPublishedTemplateDetail,
} from '@/apis/agent-factory';
import { AgentDetailType, AgentVersion } from '@/apis/agent-factory/type';
import { formatTimeSlash } from '@/utils/handle-function/FormatTime';
import styles from './index.module.less';
import GradientContainer from '../GradientContainer';
import { defaultTempZoneConfig } from '@/components/AgentConfig/ConfigSection/InputConfig/constants';
import FileSettingsModal from '@/components/AgentConfig/ConfigSection/FileSettingsModal';
import ModelSettingsPopover from '@/components/AgentConfig/ConfigSection/ModelSettingsPopover';
import SkillsSection from '@/components/AgentConfig/ConfigSection/sections/SkillsSection';
import { hiddenBuildInFields } from '@/components/AgentConfig/ConfigSection/utils';
import { getRemark, getPresetQuestions } from './utils';
import DolphinViewer from './DolphinViewer';
import LoadFailed from '../LoadFailed';
import DataSourceViewerTable from './DataSourceViewerTable';

const { Title, Text } = Typography;
const { Option } = Select;

const contentWidth = 1000;
const publishedStatus = 'published';
const templatePublishedVersion = 'template_published_version';
const templateCurrentVersion = 'template_current_version';

const draftVersion = 'v0';

const AgentDetail: React.FC<{
  isTemplate?: boolean;
  onlyShowPublishedVersion?: boolean;
}> = ({ isTemplate = false, onlyShowPublishedVersion = false }) => {
  const { id: agentId } = useParams<{ id: string }>();

  const [loading, setLoading] = useState(true);
  const [versionsLoading, setVersionsLoading] = useState(true);
  const [agentConfig, setAgentConfig] = useState<AgentDetailType | null>(null);
  const [activeTab, setActiveTab] = useState('basicInfo');
  const [fileSettingsVisible, setFileSettingsVisible] = useState(false);
  const [currentVersion, setCurrentVersion] = useState('');
  const [availableVersions, setAvailableVersions] = useState<AgentVersion[]>([]);
  const [publishedInfo, setPublishedInfo] = useState<PublishedTemplateInfoType | PublisheAgentInfoType | undefined>(
    undefined
  );

  // 输入字段类型
  const inputTypes = useMemo(
    () => [
      { value: 'string', label: intl.get('dataAgent.config.string') },
      { value: 'file', label: intl.get('dataAgent.config.file') },
      { value: 'object', label: intl.get('dataAgent.config.object') },
    ],
    []
  );

  const remark = useMemo(() => getRemark(agentConfig), [agentConfig]);
  const presetQuestions = useMemo(() => getPresetQuestions(agentConfig), [agentConfig]);
  const publishedPmsControlNames = useMemo(() => {
    if (isTemplate || !(publishedInfo as PublisheAgentInfoType)?.pms_control) return [];

    let names: string[] = [];
    const pmsControl = (publishedInfo as PublisheAgentInfoType).pms_control;

    Object.keys(pmsControl).forEach(key => {
      const value = pmsControl[key];

      if (value?.length === 0) return;

      switch (key) {
        case 'user':
          names = [...names, ...value.map(({ username }) => username)];
          break;

        case 'user_group':
          names = [...names, ...value.map(({ user_group_name }) => user_group_name)];
          break;

        case 'department':
          names = [...names, ...value.map(({ department_name }) => department_name)];
          break;

        case 'roles':
          names = [...names, ...value.map(({ role_name }) => role_name)];
          break;

        case 'app_account':
          names = [...names, ...value.map(({ app_account_name }) => app_account_name)];
          break;
      }
    });

    return names;
  }, [publishedInfo, isTemplate]);

  useEffect(() => {
    if (agentId && !isTemplate) {
      fetchAgentVersions();
    }
  }, [agentId]);

  useEffect(() => {
    if (agentId && currentVersion && !isTemplate) {
      fetchAgentDetail();
    }
  }, [agentId, currentVersion]);

  // 获取模板详情
  useEffect(() => {
    if (agentId && isTemplate) {
      fetchTemplateDetail('');
    }
  }, [agentId, isTemplate]);

  // 重试
  const retry = () => {
    if (agentId && currentVersion && !isTemplate) {
      fetchAgentDetail();
    } else if (agentId && isTemplate) {
      fetchTemplateDetail(currentVersion || '');
    }
  };

  const fetchAgentVersions = async () => {
    if (!agentId) return;

    try {
      setVersionsLoading(true);
      const [{ status }, response] = await Promise.all([getAgentDetail(agentId), await getAgentVersionList(agentId)]);
      const versions =
        status === 'unpublished' ? [{ agent_version: draftVersion }, ...response.entries] : response.entries || [];
      setAvailableVersions(versions);

      // 设置默认版本为第一个版本
      if (versions.length > 0) {
        setCurrentVersion(versions[0].agent_version);
      }
    } catch (error: any) {
      if (error?.description) {
        message.error(error.description);
      }
      console.error('Error loading agent versions:', error);
    } finally {
      setVersionsLoading(false);
    }
  };

  // 查看agent的已发布信息
  const fetchAgentPublishedInfo = async (id: string) => {
    try {
      const info = await getPublishedAgentInfo(id);
      setPublishedInfo(info);
    } catch {}
  };

  // 获取agent配置信息
  const fetchAgentDetail = async () => {
    if (!agentId || !currentVersion) return;

    try {
      setLoading(true);
      const config = await getAgentDetailInUsagePage({ id: agentId, version: currentVersion });
      setAgentConfig(config);

      // 只有最新的非v0版本，才能查看发布信息
      if (
        (availableVersions?.[0]?.agent_version !== draftVersion &&
          availableVersions?.[0]?.agent_version === currentVersion) ||
        (availableVersions?.[0]?.agent_version === draftVersion &&
          availableVersions?.[1]?.agent_version === currentVersion)
      ) {
        fetchAgentPublishedInfo(agentId);
      } else {
        setPublishedInfo(undefined);
      }
    } catch (ex: any) {
      if (ex?.description) {
        message.error(ex.description);
      }
      setAgentConfig(null);
    } finally {
      setLoading(false);
    }
  };

  // 查看模板的已发布信息
  const fetchTemplatePublishedInfo = async (id: string) => {
    try {
      const info = await getPublishedTemplateInfo(id);
      setPublishedInfo(info);
    } catch {}
  };

  // 获取模板的version
  const getTemplateVersions = async (config: any) => {
    let version: any[] = [];

    if (config.status !== publishedStatus && !onlyShowPublishedVersion) {
      // 最新状态不是已发布，则添加版本：当前配置
      version = [
        ...version,
        {
          agent_version: templateCurrentVersion,
        },
      ];
    }

    if (config.published_at) {
      // 有发布时间，则添加版本：已发布配置
      version = [...version, { agent_version: templatePublishedVersion }];
    }

    setAvailableVersions(version);
    setCurrentVersion(version[0]?.agent_version);
    setVersionsLoading(false);
  };

  // 查看模板配置信息
  const fetchTemplateDetail = async (version?: string) => {
    try {
      setLoading(true);

      if ((!version || version === templateCurrentVersion) && !onlyShowPublishedVersion) {
        // 如果版本无，或者是当前配置，则获取最新模板详情
        const config = await getTemplateDetail(agentId!);
        setAgentConfig(config);

        if (config.status === publishedStatus) {
          // 如果最新版本是已发布的状态，才需要获取发布到的信息
          fetchTemplatePublishedInfo(agentId!);
        } else {
          setPublishedInfo(undefined);
        }

        if (!availableVersions?.length) {
          getTemplateVersions(config);
        }
      } else {
        // 获取已发布模板详情
        const config = await getPublishedTemplateDetail(agentId!);
        setAgentConfig(config);
        fetchTemplatePublishedInfo(agentId!);

        if (!availableVersions?.length) {
          getTemplateVersions(config);
        }
      }
    } catch (ex: any) {
      if (ex?.description) {
        message.error(ex.description);
      }
      setAgentConfig(null);
    } finally {
      setLoading(false);
    }
  };

  const handleVersionChange = (version: string) => {
    setCurrentVersion(version);

    if (isTemplate) {
      fetchTemplateDetail(version);
    }
  };

  // 渲染输入配置表格
  const renderInputConfigTable = () => {
    const input = agentConfig?.config?.input?.fields?.filter?.(({ name }) => !hiddenBuildInFields.includes(name)) || [];
    if (input.length === 0) {
      return (
        <div style={{ textAlign: 'center', padding: '40px 0', color: '#999' }}>
          {intl.get('dataAgent.noInputConfigurationAvailable')}
        </div>
      );
    }

    const columns = [
      {
        title: intl.get('dataAgent.config.variableName'),
        dataIndex: 'name',
        key: 'name',
      },
      {
        title: intl.get('dataAgent.config.type'),
        dataIndex: 'type',
        key: 'type',
        render: (type: string) => inputTypes.find(item => item.value === type)?.label || type,
      },
      {
        title: intl.get('dataAgent.config.operation'),
        key: 'action',
        width: 100,
        render: (_: any, record: any) => {
          if (record.type === 'file') {
            return (
              <Button
                type="link"
                size="small"
                onClick={() => {
                  setFileSettingsVisible(true);
                }}
                style={{ padding: 0 }}
              >
                {intl.get('dataAgent.config.viewConfiguration')}
              </Button>
            );
          }
          return null;
        },
      },
    ];

    return <Table columns={columns} dataSource={input} pagination={false} size="small" rowKey="name" />;
  };

  // 渲染模型配置
  const renderModelConfig = () => {
    if (!agentConfig?.config?.llms || agentConfig.config.llms.length === 0) {
      return (
        <div style={{ textAlign: 'center', padding: '40px 0', color: '#999' }}>
          {intl.get('dataAgent.noModelConfigurationAvailable')}
        </div>
      );
    }

    const columns = [
      {
        title: intl.get('dataAgent.modelName'),
        dataIndex: 'llm_config',
        key: 'llm_config',
        render: (llmConfig: any, record: any) => (
          <div>
            <div className={styles.modelIcon}>
              <div className={styles.modelIconInner}></div>
            </div>
            <div>
              <div style={{ fontWeight: 500 }}>
                {llmConfig.name}
                {record.is_default && (
                  <Tag color="blue" style={{ marginLeft: 8 }}>
                    {intl.get('dataAgent.config.default')}
                  </Tag>
                )}
              </div>
            </div>
          </div>
        ),
      },
      {
        title: intl.get('dataAgent.config.operation'),
        key: 'action',
        width: 100,
        render: (_: any, record: any) => (
          <ModelSettingsPopover initialSettings={record.llm_config} onSettingsChange={() => {}} isEditable={false}>
            <Button type="link" size="small" className="dip-pl-0 dip-pr-0">
              {intl.get('dataAgent.config.viewConfiguration')}
            </Button>
          </ModelSettingsPopover>
        ),
      },
    ];

    return <Table columns={columns} dataSource={agentConfig.config.llms} pagination={false} size="small" />;
  };

  // 渲染基本信息内容
  const renderBasicInfo = () => (
    <div className={styles.basicInfoContent}>
      {/* 名称 */}
      <div className={classNames(styles.sectionTitle, 'dip-font-20', styles['bottom-divider'])}>
        {intl.get('dataAgent.basicInfo')}
      </div>
      <div className={styles.infoSection}>
        <div className={classNames(styles.label, 'dip-font-weight-700')}>{intl.get('dataAgent.name')}</div>
        <div className={styles.value}>{agentConfig?.name}</div>
      </div>

      {/* 简介 */}
      <div className={styles.infoSection}>
        <div className={classNames(styles.label, 'dip-font-weight-700')}>{intl.get('dataAgent.briefIntroduction')}</div>
        <div className={styles.value}>{agentConfig?.profile}</div>
      </div>

      {/* 人设和任务描述 */}
      <div className={classNames(styles.sectionTitle, 'dip-font-20', styles['bottom-divider'])}>
        {intl.get('dataAgent.personaAndTaskDescription')}
      </div>
      <div className={styles.infoSection}>
        <div className={classNames(styles.label, 'dip-font-weight-700')}>{intl.get('dataAgent.description')}</div>
        {agentConfig?.config?.is_dolphin_mode ? (
          <DolphinViewer config={agentConfig} />
        ) : (
          <div className={styles.value}>
            {agentConfig?.config.system_prompt || intl.get('dataAgent.noDescriptionAvailable')}
          </div>
        )}
      </div>

      {/* 输入配置 */}
      <div className={classNames(styles.sectionTitle, 'dip-font-20', styles['bottom-divider'])}>
        {intl.get('dataAgent.config.inputConfig')}
      </div>
      <div className={styles.tableSection}>{renderInputConfigTable()}</div>

      {/* 知识来源 */}
      <div className={classNames(styles.sectionTitle, 'dip-font-20', styles['bottom-divider'])}>
        {intl.get('dataAgent.config.knowledgeSource')}
      </div>
      <div className={styles.tableSection}>
        <DataSourceViewerTable config={agentConfig} />
      </div>

      {/* 技能 */}
      <div className={classNames(styles.sectionTitle, 'dip-font-20', styles['bottom-divider'])}>
        {intl.get('dataAgent.config.skill')}
      </div>
      <div className={styles.skillsSection}>
        <SkillsSection viewSkills={agentConfig?.config?.skills as any} readonly={true} />
      </div>

      {/* 模型配置 */}
      <div className={classNames(styles.sectionTitle, 'dip-font-20', styles['bottom-divider'])}>
        {intl.get('dataAgent.modelConfiguration')}
      </div>
      <div className={styles.modelsSection}>{renderModelConfig()}</div>

      {/* 其它体验 */}
      <div className={classNames(styles.sectionTitle, 'dip-font-20', styles['bottom-divider'])}>
        {intl.get('dataAgent.otherExperiences')}
      </div>
      <div className={styles.infoSection}>
        <div className={classNames(styles.label, 'dip-font-weight-700')}>
          {intl.get('dataAgent.config.defaultGreeting')}
        </div>
        <div className={styles.value}>{remark || intl.get('dataAgent.noDefaultOpeningLineSet')}</div>
      </div>
      <div className={styles.infoSection}>
        <div className={classNames(styles.label, 'dip-font-weight-700')}>{intl.get('dataAgent.presetQuestion')}</div>
        <ul
          className={classNames(styles.value, { 'dip-pl-20': presetQuestions?.length > 0 })}
          style={{ listStyle: 'disc' }}
        >
          {presetQuestions?.length
            ? presetQuestions.map(({ question }, index) => <li key={index}>{question}</li>)
            : intl.get('dataAgent.noPresetQuestionsSet')}
        </ul>
      </div>

      {/* 发布设置 */}
      {Boolean(publishedInfo) && (
        <>
          <div className={classNames(styles.sectionTitle, 'dip-font-20', styles['bottom-divider'])}>
            {intl.get('dataAgent.publishSettings')}
          </div>
          <div className={styles.infoSection}>
            <div className={classNames(styles.label, 'dip-font-weight-700')}>{intl.get('dataAgent.category')}</div>
            <div className={styles.value}>{publishedInfo?.categories?.[0]?.name}</div>
          </div>
          {!isTemplate && (
            <>
              <div className={styles.infoSection}>
                <div className={classNames(styles.label, 'dip-font-weight-700')}>
                  {intl.get('dataAgent.publishDescription')}
                </div>
                <div className={styles.value}>
                  {(publishedInfo as PublisheAgentInfoType).description ||
                    intl.get('dataAgent.noPublishDescriptionAvailable')}
                </div>
              </div>
              <div className={styles.infoSection}>
                <div className={classNames(styles.label, 'dip-font-weight-700')}>
                  {intl.get('dataAgent.publishPlatform')}
                </div>
                {[
                  {
                    label: intl.get('dataAgent.publishToDataAgent'),
                    value: AgentPublishToWhereEnum.Square,
                  },
                ].map(({ label, value }) => (
                  <div>
                    <Checkbox
                      disabled
                      checked={(publishedInfo as PublisheAgentInfoType).publish_to_where?.includes(value)}
                    >
                      <span className={classNames('dip-text-color-85', styles['value'])}>{label}</span>
                    </Checkbox>
                  </div>
                ))}
              </div>
              <div className={styles.infoSection}>
                <div className={classNames(styles.label, 'dip-font-weight-700')}>
                  {intl.get('dataAgent.visibilityRange')}
                </div>
                {(publishedInfo as PublisheAgentInfoType).pms_control === null ? (
                  intl.get('dataAgent.allUsers')
                ) : (
                  <div className={styles['pms-control-user-tags']}>
                    {publishedPmsControlNames?.map((name, index) => (
                      <Tooltip title={name} key={index}>
                        <Tag className={classNames('dip-ellipsis', styles['user-tag'])}>{name}</Tag>
                      </Tooltip>
                    ))}
                  </div>
                )}
              </div>
              <div className={styles.infoSection}>
                <div className={classNames(styles.label, 'dip-font-weight-700')}>
                  {intl.get('dataAgent.publishAsOtherTypesSimultaneously')}
                </div>
                {[
                  {
                    label: 'API',
                    value: AgentPublishToBeEnum.ApiAgent,
                  },
                  {
                    label: intl.get('dataAgent.config.skill'),
                    value: AgentPublishToBeEnum.SkillAgent,
                  },
                ].map(({ label, value }) => (
                  <div>
                    <Checkbox
                      disabled
                      checked={(publishedInfo as PublisheAgentInfoType).publish_to_bes?.includes(value)}
                    >
                      <span className={classNames('dip-text-color-85', styles['value'])}>{label}</span>
                    </Checkbox>
                  </div>
                ))}
              </div>
            </>
          )}
        </>
      )}
    </div>
  );

  if (loading) {
    return (
      <div className={classNames(styles.loadingContainer, 'dip-w-100 dip-h-100 dip-flex-center')}>
        <Spin size="large" />
      </div>
    );
  }

  if (!agentConfig) {
    return <LoadFailed className="dip-position-center dip-text-align-center" onRetry={retry} />;
  }

  // 标签页配置(模板无)
  const tabItems = isTemplate
    ? []
    : [
        {
          key: 'basicInfo',
          label: intl.get('dataAgent.versionInfo'),
          // children: renderBasicInfo(),
        },
        // {
        //   key: 'benchmark',
        //   label: 'benchmark',
        //   children: <div style={{ padding: '40px 0', textAlign: 'center', color: '#999' }}>benchmark内容</div>,
        // },
        // {
        //   key: 'permissions',
        //   label: '权限设置',
        //   children: <div style={{ padding: '40px 0', textAlign: 'center', color: '#999' }}>权限设置内容</div>,
        // },
        // {
        //   key: 'visualization',
        //   label: '可视化设置',
        //   children: <div style={{ padding: '40px 0', textAlign: 'center', color: '#999' }}>可视化设置内容</div>,
        // },
      ];

  return (
    <div className={styles.agentDetailContainer}>
      <GradientContainer className={classNames(styles['bottom-divider'], styles['content'])}>
        <div style={{ width: contentWidth, margin: '0 auto' }} className={classNames({ 'dip-mb-16': isTemplate })}>
          {/* 页面标题和版本信息 */}
          <div className={styles.titleSection}>
            <div className={styles.titleContent}>
              <Tooltip title={agentConfig.name}>
                <Title
                  level={2}
                  className={classNames(styles.pageTitle, 'dip-font-24 dip-ellipsis')}
                  style={{ whiteSpace: 'pre' }}
                >
                  {agentConfig.name}
                </Title>
              </Tooltip>

              {/* 模板页面，当只有一个版本时，不显示版本；其它情况，都显示版本 */}
              {isTemplate && availableVersions?.length === 1 ? null : (
                <div className="dip-ml-20">
                  <Select
                    value={currentVersion}
                    onChange={handleVersionChange}
                    style={{ minWidth: 160 }}
                    size="large"
                    loading={versionsLoading}
                    disabled={versionsLoading || availableVersions.length === 0}
                  >
                    {availableVersions.map(version => (
                      <Option key={version.agent_version} value={version.agent_version}>
                        {!isTemplate &&
                          (version?.agent_version === draftVersion
                            ? intl.get('dataAgent.currentConfiguration')
                            : intl.get('dataAgent.versionWithPlaceholder', { version: version.agent_version }))}

                        {isTemplate &&
                          (version?.agent_version === templatePublishedVersion
                            ? intl.get('dataAgent.publishedConfiguration')
                            : intl.get('dataAgent.currentConfiguration'))}
                      </Option>
                    ))}
                  </Select>
                </div>
              )}
            </div>
          </div>
          {/* agent描述信息profile */}
          <div className={classNames(styles.profileValue, 'dip-text-color-65')}>
            <Typography.Paragraph
              ellipsis={{
                rows: 1,
                expandable: 'collapsible',
              }}
              style={{ whiteSpace: 'pre-wrap' }}
            >
              {agentConfig?.profile}
            </Typography.Paragraph>
          </div>

          {/* 更新时间 */}
          <div className={styles.updateTimeSection}>
            <Text type="secondary" className={classNames(styles.updateTime, 'dip-text-color-85')}>
              {/* 模板：如果状态是已发布 或者 版本是已发布配置，则显示发布时间；agent，如果版本不是v0，则显示发布时间；其它，都显示更新时间 */}
              {(isTemplate && agentConfig?.status === publishedStatus) ||
              currentVersion === templatePublishedVersion ||
              (!isTemplate && currentVersion !== draftVersion)
                ? intl.get('dataAgent.publishTime') + formatTimeSlash(agentConfig.published_at)
                : intl.get('dataAgent.updateTime') + formatTimeSlash(agentConfig.updated_at)}
            </Text>
          </div>

          {tabItems?.length > 0 && (
            <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} className={styles.detailTabs} />
          )}
        </div>
      </GradientContainer>

      <div style={{ width: contentWidth, margin: '0 auto' }}>
        <div className={styles.tabsSection}>{renderBasicInfo()}</div>
        {fileSettingsVisible && (
          <FileSettingsModal
            onCancel={() => setFileSettingsVisible(false)}
            onOk={() => {}}
            initialConfig={(agentConfig?.config?.input?.temp_zone_config as any) || defaultTempZoneConfig}
            isEditable={false}
          />
        )}
      </div>
    </div>
  );
};

export default AgentDetail;
