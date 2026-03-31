import styles from './index.module.less';
import type { DipChatItemContentProgressType } from '@/components/DipChat/interface';
import ToolIcon from '@/assets/icons/toolIcon.svg';
import AgentImg from '@/assets/icons/agent3.svg';
import MCPIcon from '@/assets/icons/mcp.svg';
import { useDipChatStore } from '@/components/DipChat/store';
import DipIcon from '@/components/DipIcon';
import SkillBar from '@/components/DipChat/components/SkillBar';

type CommonToolPanelProps = {
  progressItem: DipChatItemContentProgressType;
  chatItemIndex: number;
  progressIndex: number;
  readOnly: boolean;
};

const CommonToolPanel = ({ progressItem, chatItemIndex, progressIndex, readOnly }: CommonToolPanelProps) => {
  const {
    dipChatStore: { streamGenerating, chatList, activeProgressIndex },
    openSideBar,
    setDipChatStore,
  } = useDipChatStore();
  const skillInfo = progressItem.skillInfo;
  const loading = streamGenerating && chatItemIndex === chatList.length - 1;
  const view = () => {
    openSideBar(chatItemIndex);
    setDipChatStore({
      activeProgressIndex: progressIndex,
    });
  };
  const renderSkillIcon = () => {
    if (skillInfo.name === 'graph_qa') {
      return <DipIcon type="icon-dip-color-graph" className="dip-font-16" />;
    }
    if (skillInfo.type === 'TOOL') {
      return <ToolIcon style={{ width: '16px', height: '16px' }} />;
    }
    if (skillInfo.type === 'AGENT') {
      return <AgentImg style={{ width: '16px', height: '16px' }} />;
    }
    if (skillInfo.type === 'MCP') {
      return <MCPIcon style={{ width: '16px', height: '16px' }} />;
    }
  };
  return (
    <div className={styles.container}>
      <SkillBar
        icon={renderSkillIcon()}
        title={progressItem.title}
        status={progressItem.status}
        readOnly={readOnly}
        loading={loading}
        consumeTime={progressItem.consumeTime}
        onView={view}
        active={progressIndex === activeProgressIndex}
      />
    </div>
  );
};

export default CommonToolPanel;
