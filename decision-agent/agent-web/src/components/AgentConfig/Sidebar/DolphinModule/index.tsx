import { memo, useMemo, useState, useRef, useEffect } from 'react';
import classNames from 'classnames';
import intl from 'react-intl-universal';
import { Button } from 'antd';
import { RightOutlined, LeftOutlined } from '@ant-design/icons';
import AutoSizer from 'react-virtualized-auto-sizer';
import lightningPng from '@/assets/images/lightning.png';
import { useAgentConfig } from '../../AgentConfigContext';
import DolphinModuleManagement from './DolphinModuleManagement';
import ModuleCard from './ModuleCard';
import styles from './index.module.less';

interface Props {
  toolOptions: any[];
  inputAndDolphinVarOptions: any[];
}

const cardSize = {
  width: 200,
  height: 58,
};

const DolphinModule = memo(({ toolOptions, inputAndDolphinVarOptions }: Props) => {
  const { state, actions } = useAgentConfig();
  const scrollContainer = useRef<HTMLDivElement | null>(null);
  const offsetX = useRef<number>(0);

  // 是否展示模块管理弹窗
  const [showDolphinModuleManagement, setShowDolphinModuleManagement] = useState<boolean>(false);
  // 选中的模块的key，用于同步到管理模块弹窗中的选中项
  const [selectedKey, setSelectedKey] = useState<string>('');
  // 是否显示箭头
  const [showArrow, setShowArrow] = useState<{ left: boolean; right: boolean }>({ left: false, right: true });
  // 模块容器宽度
  const [moduleContainerWidth, setModuleContainerWidth] = useState<number>(0);

  // dolphin模板列表个数
  const dolphinTemplateListCount = useMemo(
    () =>
      [...(state.dolphinTemplateList?.pre_dolphin || []), ...(state.dolphinTemplateList?.post_dolphin || [])].length,
    [state.dolphinTemplateList]
  );

  // 检查是否可编辑系统提示词配置
  const canEditSystemPrompt = actions.canEditField('system_prompt');

  // 已开启的dolphin模块
  const allDolphinModule = useMemo(() => {
    const enabledPre = state.config?.pre_dolphin || [];
    const enabledPost = state.config?.post_dolphin || [];

    return [...enabledPre, ...enabledPost];
  }, [state.config?.pre_dolphin, state.config?.post_dolphin]);

  // 更新箭头显示状态
  const updateArrowShowable = () => {
    if (!scrollContainer.current) return;

    const scrollWidth = scrollContainer.current.scrollWidth;
    const clientWidth = scrollContainer.current.clientWidth;

    if (offsetX.current + clientWidth > scrollWidth) {
      // 复位
      offsetX.current = 0;
      scrollContainer.current.style.transform = `translateX(-${offsetX.current}px)`;
    }

    // 左箭头是否显示：只要有x方向的偏移，则左箭头就显示；
    // 右箭头是否显示：如果右侧还有空间可移动，则右箭头就显示
    setShowArrow({ left: offsetX.current > 0, right: scrollWidth - offsetX.current - clientWidth > 0 });
  };

  const scrollToLeft = () => {
    if (!scrollContainer.current) return;

    const clientWidth = scrollContainer.current.clientWidth;

    // 移动的距离，不能小于0
    offsetX.current = Math.max(0, offsetX.current - clientWidth);
    scrollContainer.current.style.transform = `translateX(-${offsetX.current}px)`;

    updateArrowShowable();
  };

  const scrollToRight = () => {
    if (!scrollContainer.current) return;

    const scrollWidth = scrollContainer.current.scrollWidth;
    const clientWidth = scrollContainer.current.clientWidth;
    // 移动的距离，不能超过scrollWidth - clientWidth
    offsetX.current = Math.min(clientWidth + offsetX.current, scrollWidth - clientWidth);
    scrollContainer.current.style.transform = `translateX(-${offsetX.current}px)`;

    updateArrowShowable();
  };

  useEffect(() => {
    // 当模块列表的宽度发生变化，或者 dolphin模块个数发生变化时，更新箭头显示状态
    if (!scrollContainer.current || !moduleContainerWidth) {
      setShowArrow({
        left: false,
        right: false,
      });
      return;
    }

    updateArrowShowable();
  }, [moduleContainerWidth, dolphinTemplateListCount]);

  return dolphinTemplateListCount > 0 ? (
    <div>
      <div className="dip-flex-space-between">
        <span className="dip-flex-align-center">
          <img src={lightningPng} className={classNames('dip-mr-12', styles['lightning-icon'])} />
          <span className="dip-font-weight-700">{intl.get('dataAgent.moduleStatus')}</span>
          <span
            className={classNames(
              'dip-ml-8 dip-pl-12 dip-pr-12 dip-border-radius-12 dip-font-12',
              styles['module-count']
            )}
          >
            {intl.get('dataAgent.totalModulesWithCount', { count: dolphinTemplateListCount })}
          </span>
        </span>
        <span>
          <Button
            type="link"
            style={{ height: 22 }}
            className="dip-pl-0 dip-pr-0"
            onClick={e => {
              e.stopPropagation();
              setShowDolphinModuleManagement(true);
            }}
          >
            {intl.get('dataAgent.config.manageModule')}
          </Button>
        </span>
      </div>
      <div className="dip-ml-2 dip-pt-18 dip-pb-24 dip-overflow-hidden dip-position-r">
        {/** 左箭头 */}
        {showArrow.left && (
          <div
            className={classNames(
              'dip-position-a dip-flex-center',
              styles['left-arrow-wrapper'],
              styles['arrow-wrapper']
            )}
          >
            <div
              className={classNames('dip-flex-center dip-border-radius-full dip-pointer', styles['arrow-icon'])}
              onClick={scrollToLeft}
            >
              <LeftOutlined className="dip-font-16 dip-c-text-lower" />
            </div>
          </div>
        )}
        {/** 右箭头 */}
        {showArrow.right && (
          <div
            className={classNames(
              'dip-position-a dip-flex-center',
              styles['right-arrow-wrapper'],
              styles['arrow-wrapper']
            )}
          >
            <div
              className={classNames('dip-flex-center dip-border-radius-full dip-pointer', styles['arrow-icon'])}
              onClick={scrollToRight}
            >
              <RightOutlined className="dip-font-16 dip-c-text-lower" />
            </div>
          </div>
        )}
        {/** 模块卡片列表 */}
        <AutoSizer style={{ width: '100%', height: cardSize.height }}>
          {({ width }) => {
            setModuleContainerWidth(width);

            return (
              <div className="dip-flex dip-gap-12" ref={scrollContainer}>
                {allDolphinModule.map(item => (
                  <ModuleCard
                    key={item.key}
                    moduleKey={item.key}
                    name={item.name!}
                    height={cardSize.height}
                    width={cardSize.width}
                    enabled={item.enabled!}
                    onSelect={key => {
                      setShowDolphinModuleManagement(true);
                      setSelectedKey(key);
                    }}
                  />
                ))}
              </div>
            );
          }}
        </AutoSizer>
      </div>
      {showDolphinModuleManagement && (
        <DolphinModuleManagement
          disabled={!canEditSystemPrompt}
          selectedKey={selectedKey}
          promptVarOptions={inputAndDolphinVarOptions}
          toolOptions={toolOptions}
          onClose={() => {
            setShowDolphinModuleManagement(false);
            setSelectedKey('');
          }}
        />
      )}
    </div>
  ) : null;
});

export default DolphinModule;
