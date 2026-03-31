import Chat from './Chat';
import DipChatStore from './store';
import React from 'react';
import type { DipChatProps } from './interface';

const DipChat: React.FC<DipChatProps> = props => {
  return (
    <DipChatStore {...props}>
      <Chat />
    </DipChatStore>
  );
};

export default DipChat;
