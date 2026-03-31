import { type ReactNode, type CSSProperties } from 'react';
import styles from './index.module.less';
import classNames from 'classnames';
import DipButton from '@/components/DipButton';
import { CloseCircleFilled, LoadingOutlined, RightOutlined } from '@ant-design/icons';
import intl from 'react-intl-universal';
import { Spin } from 'antd';
import DipIcon from '@/components/DipIcon';
import { useDipChatStore } from '@/components/DipChat/store';
export type SkillBarProps = {
  icon?: ReactNode;
  title?: string;
  status: string;
  readOnly?: boolean;
  loading?: boolean;
  consumeTime?: string; // 耗时
  className?: string;
  style?: CSSProperties;
  onView?: () => void;
  active?: boolean;
};
const SkillBar = (props: SkillBarProps) => {
  const { setDipChatStore } = useDipChatStore();
  const {
    icon,
    title,
    status,
    readOnly = false,
    loading = false,
    consumeTime,
    className,
    style,
    onView,
    active = false,
  } = props;

  const renderStatusIcon = () => {
    if (status === 'processing' && loading) {
      return <Spin size="small" indicator={<LoadingOutlined spin />} />;
    }
    if (status === 'completed') {
      return <DipIcon className="dip-text-color-success" type="icon-dip-chenggong" />;
    }
    if (status === 'failed') {
      return <CloseCircleFilled className="dip-text-color-error" />;
    }
  };
  return (
    <div
      style={style}
      className={classNames(styles.container, 'dip-flex-align-center dip-border-radius-8', className, {
        [styles.active]: active,
      })}
    >
      {icon}
      <div className="dip-flex-item-full-width dip-ml-8 dip-flex-align-center">
        <div style={{ maxWidth: '100%' }} className="dip-ellipsis" title={title}>
          {title}
        </div>
        <span style={{ lineHeight: 1 }} className="dip-ml-8">
          {renderStatusIcon()}
        </span>
      </div>
      {status === 'completed' && consumeTime && (
        <span className="dip-flex-align-center dip-text-color-45 dip-font-12 dip-ml-12 dip-mr-12">
          <span>耗时：{consumeTime}s</span>
        </span>
      )}
      {!readOnly && (
        <DipButton
          onClick={() => {
            // 点击正在进行中的工具，恢复工具自动展开
            if (status === 'processing') {
              setDipChatStore({
                toolAutoExpand: true,
              });
            }
            onView?.();
          }}
          type="text"
          icon={<RightOutlined />}
          iconPosition="end"
        >
          {intl.get('dipChat.view')}
        </DipButton>
      )}
    </div>
  );
};

export default SkillBar;
