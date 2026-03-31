import { Tooltip } from 'antd';
import type { TooltipProps } from 'antd';
import { QuestionCircleFilled } from '@ant-design/icons';
import classNames from 'classnames';
import { isDef } from '@/utils/handle-function';
import './baseTipStyle.less';

export type BaseTipProps = TooltipProps & {
  iconClassName?: string;
  autoMaxWidth?: boolean;
};

/**
 * @param iconClassName 问号图标类名
 * @param autoMaxWidth 不限制最大宽度
 */
const BaseTip = (props: BaseTipProps) => {
  const { overlayClassName, iconClassName, autoMaxWidth, children, ...tipProps } = props;
  return (
    <Tooltip
      {...tipProps}
      overlayClassName={classNames('c-explain-tip', overlayClassName, { 'auto-max': autoMaxWidth })}
    >
      {isDef(children) ? (
        children
      ) : (
        <QuestionCircleFilled className={classNames('c-question-icon dip-c-watermark dip-pointer', iconClassName)} />
      )}
    </Tooltip>
  );
};

export default BaseTip;
