import React from 'react';
import intl from 'react-intl-universal';
import { Badge, Tooltip } from 'antd';
import SystemAgentIcon from '@/assets/icons/system-agent.svg';

interface SystemAgentWrapperProps {
  systemAgentIconSize: number;
  children: React.ReactElement;
  size: number;
}

const SystemAgentAvatarWrapper = ({ children, systemAgentIconSize, size }: SystemAgentWrapperProps) => {
  return (
    <Badge
      count={
        <span>
          <Tooltip title={intl.get('dataAgent.systemAgent')} placement="right">
            <SystemAgentIcon style={{ fontSize: systemAgentIconSize, color: '#d0d0d0' }} />
          </Tooltip>
        </span>
      }
      offset={[-4, size - 4]}
    >
      {children}
    </Badge>
  );
};

export default SystemAgentAvatarWrapper;
