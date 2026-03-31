import { memo, CSSProperties } from 'react';
import { Empty as EmptyAntd } from 'antd';
import intl from 'react-intl-universal';
import empty from '@/assets/images/empty.png';

const Empty = ({
  description,
  className,
  style,
}: {
  description?: string;
  className?: string;
  style?: CSSProperties;
}) => {
  return (
    <EmptyAntd
      className={className}
      style={style}
      image={<img src={empty} />}
      description={description || intl.get('dataAgent.noData')}
    />
  );
};

export default memo(Empty);
