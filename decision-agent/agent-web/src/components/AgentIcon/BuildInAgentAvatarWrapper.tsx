import React, { useMemo } from 'react';
import intl from 'react-intl-universal';
import { Badge, Tooltip } from 'antd';
import BuildInAgentIcon from '@/assets/icons/build-in-agent.svg';
import BuildInAgentIconTW from '@/assets/icons/build-in-agent-tw.svg';
import BuildInAgentIconUS from '@/assets/icons/build-in-agent-us.svg';
import { useMicroWidgetProps } from '@/hooks';

interface BuildInAgentWrapperProps {
  size: number;
  buildInAgentIconSize: number;
  children: React.ReactElement;
}

const BuildInAgentAvatarWrapper = ({ children, buildInAgentIconSize, size }: BuildInAgentWrapperProps) => {
  const microWidgetProps = useMicroWidgetProps();
  const Icon = useMemo(() => {
    const lang = microWidgetProps.language.getLanguage;
    const isUS = lang === 'en-us';
    const isTW = lang === 'zh-tw';
    const isZH = !(isUS || isTW);

    return isZH ? BuildInAgentIcon : isTW ? BuildInAgentIconTW : BuildInAgentIconUS;
  }, []);

  return (
    <Badge
      count={
        <span>
          <Tooltip title={intl.get('dataAgent.builtInAgent')}>
            <Icon style={{ fontSize: buildInAgentIconSize }} />
          </Tooltip>
        </span>
      }
      offset={[-size + buildInAgentIconSize / 2, buildInAgentIconSize / 2]}
    >
      {children}
    </Badge>
  );
};

export default BuildInAgentAvatarWrapper;
