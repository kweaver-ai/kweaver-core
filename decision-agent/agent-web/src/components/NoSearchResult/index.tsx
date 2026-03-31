import type { CSSProperties, ReactNode } from 'react';
import NoResult from '@/assets/icons/no-result.svg';

type NoSearchResultProps = {
  tip?: ReactNode;
  style?: CSSProperties;
};
const NoSearchResult = (props: NoSearchResultProps) => {
  const { tip = '抱歉，没有找到相关内容', style } = props;
  return (
    <div style={style} className="dip-full dip-flex-column-center">
      <NoResult />
      <div className="dip-text-color-45">{tip}</div>
    </div>
  );
};

export default NoSearchResult;
