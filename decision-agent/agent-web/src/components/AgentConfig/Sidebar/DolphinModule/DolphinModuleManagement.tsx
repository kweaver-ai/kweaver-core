import { memo, useState, useRef, useEffect, useMemo } from 'react';
import intl from 'react-intl-universal';
import classNames from 'classnames';
import { Modal, Tooltip, Switch, Button } from 'antd';
import { useAgentConfig } from '../../AgentConfigContext';
import AdDolphinEditor from '@/components/Editor/AdDolphinEditor';
import RevertIcon from '@/assets/icons/revert.svg';
import { type DolphinTemplateType } from '@/apis/agent-factory/type';
import styles from './DolphinModuleManagement.module.css';

interface DolphinModuleManagementProps {
  disabled: boolean;
  selectedKey: string;
  promptVarOptions: any[];
  toolOptions: any[];
  onClose: () => void;
}

// 从模板列表中找到对应key的value
const getDolphinTemplateValueByKey = (
  templates: {
    pre_dolphin: DolphinTemplateType[];
    post_dolphin: DolphinTemplateType[];
  },
  key: string
): string => {
  const item =
    templates.pre_dolphin.find(item => item.key === key) ?? templates.post_dolphin.find(item => item.key === key);
  return item?.value ?? '';
};

const DolphinModuleManagement = ({
  disabled,
  selectedKey,
  promptVarOptions,
  toolOptions,
  onClose,
}: DolphinModuleManagementProps) => {
  const { state, actions } = useAgentConfig();

  // 初始内容（用来【取消】）
  const initialValueRef = useRef<string>('');
  // 列表元素
  const listContainerRef = useRef<HTMLDivElement>(null);

  // 模块（拼接pre_dolphin和post_dolphin）
  const modules = useMemo(
    () => [...(state.config.pre_dolphin || []), ...(state.config.post_dolphin || [])],
    [state.config.pre_dolphin, state.config.post_dolphin]
  );

  // 选中的模块菜单，默认选中第一项
  const [selectedModule, setSelectedModule] = useState(modules.find(item => item.key === selectedKey) || modules[0]);
  // 是否进入编辑模式，默认非编辑模式
  const [isEditMode, setIsEditMode] = useState<boolean>(false);

  // 选中模块的模板值（用来【还原】）
  const selectedTemplateValue = useMemo(
    () => getDolphinTemplateValueByKey(state.dolphinTemplateList, selectedModule.key),
    [state.dolphinTemplateList, selectedModule.key]
  );

  // 还原按钮是否可用
  const [revertBtnDisabled, setRevertBtnDisabled] = useState<boolean>(selectedTemplateValue === selectedModule.value);

  useEffect(() => {
    // 当选中的模块变更了
    initialValueRef.current = selectedModule.value;
    setIsEditMode(false);
  }, [selectedModule.key]);

  useEffect(() => {
    setRevertBtnDisabled(selectedTemplateValue === selectedModule.value);
  }, [selectedModule.value, selectedTemplateValue]);

  // 保存
  const save = () => {
    const update = { value: selectedModule.value, edited: selectedModule.value !== selectedTemplateValue };
    actions.updatePreAndPostDolphinByKey(selectedModule.key, update);
    setSelectedModule(prev => ({
      ...prev,
      ...update,
    }));
    setIsEditMode(false);
    initialValueRef.current = selectedModule.value;
  };

  // 还原
  const revert = () => {
    // 还原成模板的值
    const update = { value: selectedTemplateValue, edited: false };
    actions.updatePreAndPostDolphinByKey(selectedModule.key, update);
    setSelectedModule(prev => ({
      ...prev,
      ...update,
    }));
    setIsEditMode(false);
  };

  // 取消
  const cancel = () => {
    // 还原成上一次保存过的值
    setSelectedModule(prev => ({
      ...prev,
      value: initialValueRef.current,
    }));
    setIsEditMode(false);
  };

  // 处理enabled变化
  const handleEnabledChange = (enabled: boolean) => {
    const updates = { enabled };
    actions.updatePreAndPostDolphinByKey(selectedModule.key, updates);
    setSelectedModule(prev => ({
      ...prev,
      ...updates,
    }));
  };

  // 处理编辑器内容变化
  const handleValueChange = (value: string | undefined) => {
    setSelectedModule(prev => ({
      ...prev,
      value: value || '',
    }));
  };

  // 处理聚焦事件
  const handleFocus = () => {
    if (disabled) return;

    // 点击输入框，自动进入编辑模式
    setIsEditMode(true);
  };

  const Content = (
    <div className={classNames('dip-flex dip-gap-16 dip-pt-12', styles['container'])}>
      {/** 导航菜单 */}
      <div className={styles['mods']} ref={listContainerRef}>
        {modules.map(module => (
          <div
            key={module.key}
            className={classNames(
              styles['mod'],
              'dip-flex-align-center dip-gap-8 dip-border-radius-8 dip-pl-8 dip-pointer',
              {
                [styles['selected']]: module.key === selectedModule.key,
              }
            )}
            onClick={() => {
              setSelectedModule(module);
            }}
          >
            <span
              className={classNames(
                styles['dot'],
                'dip-flex-shrink-0',
                module.enabled ? styles['dot-enabled'] : styles['dot-disabled']
              )}
            />
            <Tooltip title={module.name} placement="left">
              <span className="dip-ellipsis dip-flex-item-full-width">{module.name}</span>
            </Tooltip>
          </div>
        ))}
      </div>

      {/** 内容区 */}
      <div className="dip-flex-item-full-width">
        <div className="dip-flex-space-between dip-mb-4" style={{ height: '32px' }}>
          <span
            className={classNames(
              selectedModule.enabled ? styles['enabled-status'] : styles['disabled-status'],
              'dip-font-12 dip-pl-10 dip-pr-10 dip-border-radius-4'
            )}
          >
            {selectedModule.enabled ? intl.get('dataAgent.enabled') : intl.get('dataAgent.disabled')}
          </span>
          <span className="dip-flex-align-center dip-gap-10">
            <span>{intl.get('dataAgent.enableConfiguration')}</span>
            <Switch checked={selectedModule.enabled} disabled={disabled} size="small" onChange={handleEnabledChange} />
            {isEditMode && (
              <>
                <Button type="link" className="dip-p-0" onClick={save}>
                  {intl.get('dataAgent.config.save')}
                </Button>
                <Button type="link" className="dip-p-0" onClick={cancel}>
                  {intl.get('dataAgent.cancel')}
                </Button>
              </>
            )}
            {!revertBtnDisabled && !disabled && (
              <Tooltip title={intl.get('dataAgent.restore')}>
                <Button type="link" className="dip-p-0" onClick={revert}>
                  <RevertIcon />
                </Button>
              </Tooltip>
            )}
          </span>
        </div>
        <div className={classNames(styles['editor'], 'dip-border-radius-8')}>
          <AdDolphinEditor
            value={selectedModule.value}
            onChange={handleValueChange}
            placeholder={intl.get('dataAgent.pleaseEnter')}
            disabled={disabled}
            promptVarOptions={promptVarOptions}
            toolOptions={toolOptions}
            height={398}
            onFocus={handleFocus}
          />
        </div>
      </div>
    </div>
  );

  useEffect(() => {
    // 当选中的模块不是第一项时，需要让其出现在可视区域
    if (selectedModule !== modules[0] && listContainerRef.current) {
      const index = modules.findIndex(modu => modu.key === selectedModule.key);

      if (index > 0) {
        const children = listContainerRef.current.children;
        const targetElement = children[index];

        // 滚动到可视区域
        targetElement.scrollIntoView({
          behavior: 'smooth',
          block: 'center',
        });
      }
    }
  }, []);

  return (
    <Modal
      title={intl.get('dataAgent.moduleManagement')}
      open={true}
      onCancel={onClose}
      destroyOnHidden
      width={840}
      footer={[]}
      maskClosable={false}
      centered
    >
      {Content}
    </Modal>
  );
};

export default memo(DolphinModuleManagement);
