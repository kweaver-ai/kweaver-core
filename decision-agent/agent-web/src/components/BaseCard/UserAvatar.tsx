import { memo } from 'react';
import classNames from 'classnames';
import { Avatar } from 'antd';
import styles from './index.module.less';

interface UserAvatarProps {
  src?: string;
  userName?: string;
}
const UserAvatar = ({ src, userName }: UserAvatarProps) => {
  return (
    <Avatar size={24} className={classNames(styles['user-avatar'], 'dip-mr-4')} src={src}>
      {userName?.charAt(0)}
    </Avatar>
  );
};

export default memo(UserAvatar);
