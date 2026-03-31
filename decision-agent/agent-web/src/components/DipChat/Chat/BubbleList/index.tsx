import { Bubble } from '@ant-design/x';
import styles from './index.module.less';
import UserPanel from './UserPanel';
import ErrorPanel from './ErrorPanel';
import CommonPanel from './CommonPanel';
import { useDipChatStore } from '@/components/DipChat/store';
import type { DipChatItem } from '@/components/DipChat/interface';
type BubbleListProps = {
  readOnly?: boolean;
};
const BubbleList = ({ readOnly = false }: BubbleListProps) => {
  const {
    dipChatStore: { chatList },
  } = useDipChatStore();
  const roles: any = (chatItem: DipChatItem, chatItemIndex: number) => {
    if (chatItem.error) {
      return {
        placement: 'start',
        variant: 'borderless',
        messageRender: () => {
          return <ErrorPanel chatItemIndex={chatItemIndex} />;
        },
      };
    }
    if (chatItem.role === 'user') {
      return {
        placement: 'end',
        variant: 'borderless',
        messageRender: () => {
          return <UserPanel chatItemIndex={chatItemIndex} readOnly={readOnly} />;
        },
      };
    }
    if (chatItem.role === 'common') {
      return {
        placement: 'start',
        variant: 'borderless',
        messageRender: () => {
          return <CommonPanel chatItemIndex={chatItemIndex} readOnly={readOnly} />;
        },
      };
    }
  };
  return (
    <div className={styles.bubbleList}>
      <Bubble.List items={chatList} roles={roles} />
    </div>
  );
};

export default BubbleList;
