import type { CSSProperties, ReactNode } from 'react';
import emptyIcon from '@/assets/images/empty.png';

type NoDataProps = {
  tip?: ReactNode;
  style?: CSSProperties;
};
const NoData = (props: NoDataProps) => {
  const { tip = '抱歉，没有找到相关内容', style } = props;
  return (
    <div style={style} className="dip-full dip-flex-column-center">
      <img style={{ width: 64, height: 64 }} src={emptyIcon} alt="" />
      <div className="dip-text-color-45">{tip}</div>
    </div>
  );
};

export default NoData;
