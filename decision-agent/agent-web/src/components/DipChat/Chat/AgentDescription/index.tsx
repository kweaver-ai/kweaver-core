import styles from './index.module.less';
import { useDipChatStore } from '@/components/DipChat/store';
import _ from 'lodash';
import AgentIcon from '@/components/AgentIcon';
import Markdown from '@/components/Markdown';
import classNames from 'classnames';
import { nanoid } from 'nanoid';
import type { ChatBody } from '@/components/DipChat/interface';
import dayjs from 'dayjs';

const AgentDescription = () => {
  const {
    dipChatStore: { agentDetails, chatList },
    sendChat,
  } = useDipChatStore();
  const agentConfig = _.get(agentDetails, ['config']) || {};

  const renderPreset = () => {
    const preset_questions = _.get(agentConfig, ['preset_questions']) || [];
    if (preset_questions.length > 0) {
      return (
        <div className="dip-mt-28 dip-w-100">
          <div className="dip-text-color-65 dip-mb-4">你可以问我：</div>
          {preset_questions.map((item: any, index: number) => (
            <div
              key={index}
              className={styles.question}
              onClick={() => {
                const cloneChatList = _.cloneDeep(chatList);
                cloneChatList.push({
                  key: nanoid(),
                  role: 'user',
                  content: item.question,
                  loading: false,
                  updateTime: dayjs().valueOf(),
                });
                cloneChatList.push({
                  key: nanoid(),
                  role: 'common',
                  content: '',
                  loading: true,
                });
                const body: ChatBody = { query: item.question };
                sendChat({
                  chatList: cloneChatList,
                  body,
                  activeChatItemIndex: -1,
                });
              }}
            >
              <div title={item.question} className="dip-ellipsis">
                {item.question}
              </div>
            </div>
          ))}
        </div>
      );
    }
  };

  return (
    <div className={classNames(styles.container)}>
      <AgentIcon
        avatar_type={agentDetails.avatar_type}
        avatar={agentDetails.avatar}
        size={90}
        name={agentDetails.name}
      />
      <div className="dip-mt-16 dip-flex-item-full-width">
        <div className={styles.name}>{agentDetails.name}</div>
        <div style={{ opacity: 0.65 }}>
          <Markdown
            className="dip-mb-8 dip-mt-24"
            value={agentConfig.opening_remark_config?.fixed_opening_remark}
            readOnly
          />
        </div>
        {renderPreset()}
      </div>
    </div>
  );
};

export default AgentDescription;
