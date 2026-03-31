import { Radio, message, Collapse, Tooltip, Popover } from 'antd';
import intl from 'react-intl-universal';
import classNames from 'classnames';
import { importAgent as importAgentReq, ImportTypeEnum } from '@/apis/agent-factory';
import { importFile, ReadAsEnum, ImportFileErrorEnum } from '@/utils/file';
import CopyIcon from '@/assets/icons/copy.svg';
import KeyIcon from '@/assets/icons/key.svg';
import { copyToBoard } from '@/utils/handle-function';
import styles from './index.module.less';

const copy = async (text: string) => {
  const res = await copyToBoard(text);
  if (res) {
    message.success(intl.get('dataAgent.copyKeySuccess'));
  }
};

const handleErrorMessage = (modal: any, errorItemsByType: any) => {
  // 导入失败
  const errorConfigs = [
    {
      key: 'config_invalid',
      title: intl.get('dataAgent.invalidConfiguration'),
      description: intl.get('dataAgent.invalidAgentConfiguration'),
    },
    {
      key: 'no_create_system_agent_pms',
      title: intl.get('dataAgent.insufficientPermissions'),
      description: intl.get('dataAgent.systemAgentNoPermission'),
    },
    {
      key: 'agent_key_conflict',
      title: intl.get('dataAgent.agentIdConflict'),
      description: intl.get('dataAgent.existingAgentIds'),
    },
    {
      key: 'biz_domain_conflict',
      title: intl.get('dataAgent.agentIdConflict'),
      description: intl.get('dataAgent.agentIdConflictInOtherDomain'),
    },
  ];

  const errorDetails = errorConfigs
    .filter(config => errorItemsByType[config.key]?.length)
    .map(config => ({
      ...config,
      items: errorItemsByType[config.key],
    }));

  if (errorDetails.length > 0) {
    const isSingleError = errorDetails.length === 1;
    modal.info({
      title: isSingleError
        ? errorDetails[0].title + intl.get('dataAgent.itemCount', { count: errorDetails[0].items.length })
        : intl.get('dataAgent.importFailed'),
      width: 480,
      content: (
        <div
          className={classNames('dip-overflowY-auto dip-pr-24', styles['export-agent-error-modal'], {
            [styles['export-agent-error-expand-hidden']]: isSingleError,
          })}
        >
          <Collapse
            bordered={false}
            items={errorDetails.map(({ key, description, items }) => ({
              key,
              label: <span className="dip-c-black">{description}</span>,
              children: (
                <div>
                  {items.map(({ agent_key, agent_name }: { agent_key: string; agent_name: string }) => (
                    <div className={classNames(styles['panel-content-item'], 'dip-flex-align-center dip-gap-8')}>
                      <Tooltip title={agent_name}>
                        <div className="dip-ellipsis dip-text-color-85">{agent_name}</div>
                      </Tooltip>
                      <Popover
                        trigger={['hover', 'click']}
                        content={
                          <div style={{ padding: 10 }} className="dip-flex-align-center dip-gap-8">
                            <span>{`Agent Key: ${agent_key}`}</span>
                            <Tooltip title={intl.get('dataAgent.copy')}>
                              <CopyIcon
                                className={classNames(styles['export-agent-error-key-icon'], 'dip-font-14 dip-pointer')}
                                onClick={() => copy(agent_key)}
                              />
                            </Tooltip>
                          </div>
                        }
                      >
                        <KeyIcon
                          className={classNames(
                            'dip-flex-shrink-0 dip-font-12 dip-pointer',
                            styles['export-agent-error-key-icon']
                          )}
                        />
                      </Popover>
                    </div>
                  ))}
                </div>
              ),
              className: styles['error-panel'],
            }))}
            defaultActiveKey={[errorDetails[0].key]}
          />
        </div>
      ),
      centered: true,
    });
  }
};

const importAgent = async ({
  modal,
  importType,
  reload,
}: {
  modal: any;
  importType: ImportTypeEnum;
  reload: () => void;
}) => {
  // 选择本地文件
  try {
    const jsonText = await importFile({
      accept: '.json',
      readAs: ReadAsEnum.Text,
      maxFileSize: 10 * 1024 * 1024,
    });

    // 创建FormData并添加JSON内容
    const formData = new FormData();
    const blob = new Blob([jsonText], { type: 'application/json' });
    formData.append('file', blob);
    formData.append('import_type', importType);

    try {
      const { is_success, ...errorItemsByType } = await importAgentReq(formData);

      if (is_success) {
        message.success(intl.get('dataAgent.importSuccess'));
        // 导入成功后，刷新列表
        reload();
      } else {
        handleErrorMessage(modal, errorItemsByType);
      }
    } catch (ex: any) {
      if (ex?.description) {
        message.error(ex.description);
      }
    }
  } catch (ex: any) {
    switch (ex?.error) {
      case ImportFileErrorEnum.SizeLimitExceeded:
        message.info(intl.get('dataAgent.fileSizeExceedsLimit'));
        break;

      case ImportFileErrorEnum.UserCancelled:
        break;

      default:
        if (ex?.error) {
          message.error(ex.error);
        }
    }
  }
};

// 导入agent
export const handleImportAgent = (modal: any, reload: () => void) => {
  let importType = ImportTypeEnum.Upsert;
  modal.confirm({
    title: intl.get('dataAgent.selectImportMode'),
    icon: null,
    className: styles['import-modal'],
    width: 520,
    content: (
      <div>
        <div className="dip-mb-12 dip-mt-12 dip-text-color-85">{intl.get('dataAgent.importModePrompt')}</div>
        <Radio.Group
          className={styles['radio-group']}
          options={[
            {
              label: (
                <>
                  <span>{intl.get('dataAgent.updateMode')}</span>
                  <div className="dip-mb-16 dip-text-color-45">{intl.get('dataAgent.updateModeDescription')}</div>
                </>
              ),
              value: ImportTypeEnum.Upsert,
            },
            {
              label: (
                <>
                  <span>{intl.get('dataAgent.createMode')}</span>
                  <div className="dip-text-color-45 dip-mb-24">{intl.get('dataAgent.createModeDescription')}</div>
                </>
              ),
              value: ImportTypeEnum.Create,
            },
          ]}
          defaultValue={importType}
          onChange={e => {
            importType = e.target.value;
          }}
        />
      </div>
    ),
    centered: true,
    onOk() {
      importAgent({ importType, reload, modal });
    },
    onCancel() {},
    footer: (_, { OkBtn, CancelBtn }: any) => (
      <div className="dip-flex-content-end">
        <OkBtn />
        <CancelBtn />
      </div>
    ),
  });
};
