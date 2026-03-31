import { useMemo } from 'react';
import { Avatar } from 'antd';
import { UserOutlined } from '@ant-design/icons';
import Agent from '@/assets/icons/agent';
import SystemAgentAvatarWrapper from './SystemAgentAvatarWrapper';
import BuildInAgentAvatarWrapper from './BuildInAgentAvatarWrapper';

export const AVATAR_OPTIONS = Object.keys(Agent).map(key => ({
  type: 1,
  value: key,
  img: (Agent as any)[key],
}));

// 获取当前头像显示
const AgentIcon = ({
  size = 80,
  avatar_type,
  avatar,
  name,
  onClick,
  style,
  showSystemLogo = false,
  showBuildInLogo = false,
}: {
  size?: number;
  avatar_type: number;
  avatar: string;
  name: string;
  onClick?: () => void;
  style?: Record<string, any>;
  showSystemLogo?: boolean;
  showBuildInLogo?: boolean;
}) => {
  const currentAvatar = useMemo(
    () => AVATAR_OPTIONS.find(option => option.type === avatar_type && option.value === avatar) || AVATAR_OPTIONS[0],
    [avatar_type, avatar]
  );

  const buildInAgentIconSize = size / 2 - 2;
  const systemAgentIconSize = size / 2 - 6;

  const agentAvatar = (
    <Avatar
      shape="square"
      size={size}
      style={{
        cursor: onClick ? 'pointer' : 'auto',
        ...style,
      }}
      src={<currentAvatar.img style={{ fontSize: size }} />}
      onClick={onClick}
      icon={!currentAvatar?.img && (name ? null : <UserOutlined />)}
    >
      {!currentAvatar?.img && name ? name.charAt(0).toUpperCase() : null}
    </Avatar>
  );

  // 同时显示内置 agent logo和系统 agent logo
  if (showBuildInLogo && showSystemLogo) {
    return (
      <SystemAgentAvatarWrapper systemAgentIconSize={systemAgentIconSize} size={size}>
        <BuildInAgentAvatarWrapper buildInAgentIconSize={buildInAgentIconSize} size={size}>
          {agentAvatar}
        </BuildInAgentAvatarWrapper>
      </SystemAgentAvatarWrapper>
    );
  }

  // 显示内置 agent logo
  if (showBuildInLogo) {
    return (
      <BuildInAgentAvatarWrapper buildInAgentIconSize={buildInAgentIconSize} size={size}>
        {agentAvatar}
      </BuildInAgentAvatarWrapper>
    );
  }

  // 显示系统 agent logo
  if (showSystemLogo) {
    return (
      <SystemAgentAvatarWrapper systemAgentIconSize={systemAgentIconSize} size={size}>
        {agentAvatar}
      </SystemAgentAvatarWrapper>
    );
  }

  // 只显示agent icon
  return agentAvatar;
};

export default AgentIcon;
