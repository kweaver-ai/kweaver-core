import { useRef, useState } from 'react';
import styles from './index.module.less';
import classNames from 'classnames';
import type { ConversationItemType } from '@/components/DipChat/interface';
import { Button, Input, message, Popconfirm, Spin, Tooltip, Form, Popover } from 'antd';
import intl from 'react-intl-universal';
import {
  CheckOutlined,
  ClockCircleOutlined,
  CloseOutlined,
  EditOutlined,
  LoadingOutlined,
  MessageOutlined,
} from '@ant-design/icons';
import { deleteConversationById, updateConversation } from '@/apis/super-assistant';
import DipIcon from '@/components/DipIcon';
import dayjs from 'dayjs';

type ConversationItemProps = {
  onClick?: () => void;
  refreshList?: () => void;
  onClearSelectedConversation?: () => void;
  className?: string;
  agentAppKey: string;
  id: string;
  activeKey: string;
  timestamp: number;
} & Partial<ConversationItemType>;
const ConversationItem = (props: ConversationItemProps) => {
  const [messageApi, messageContextHolder] = message.useMessage();
  const {
    className,
    onClick,
    unRead = false,
    label,
    id,
    status,
    agentAppKey,
    refreshList,
    onClearSelectedConversation,
    timestamp,
  } = props;
  const [isEdit, setIsEdit] = useState(false);
  const inputRef = useRef<any>(null);
  const [form] = Form.useForm();

  const saveTitle = async ({ name }: any) => {
    if (name === label) {
      setIsEdit(false);
    } else {
      const res = await updateConversation(agentAppKey, id!, { title: name });
      if (res) {
        messageApi.success(intl.get('dipChat.modifySuccess'));
        refreshList?.();
        setIsEdit(false);
      }
    }
  };

  const renderItem = () => {
    if (isEdit) {
      return (
        <div className="dip-flex" style={{ minHeight: 40 }}>
          <div className="dip-flex-item-full-width">
            <Form form={form} onFinish={saveTitle}>
              <Form.Item
                className="dip-mb-0"
                label=""
                name="name"
                rules={[
                  { required: true, message: intl.get('dipChat.conversationNameRequired') },
                  { max: 50, message: intl.get('global.lenErr', { len: 50 }) },
                ]}
              >
                <Input
                  autoFocus
                  ref={inputRef}
                  placeholder={intl.get('dipChat.enterConversationName')}
                  onPressEnter={() => form.submit()}
                />
              </Form.Item>
            </Form>
          </div>
          <div className="dip-ml-8">
            <Tooltip title={intl.get('dipChat.cancel')}>
              <Button
                size="small"
                onClick={() => {
                  setIsEdit(false);
                }}
                className="dip-text-color-error"
                type="text"
                icon={<CloseOutlined />}
              />
            </Tooltip>
            <Tooltip title={intl.get('dipChat.save')}>
              <Button
                size="small"
                onClick={() => form.submit()}
                className="dip-text-color-primary"
                type="text"
                icon={<CheckOutlined />}
              />
            </Tooltip>
          </div>
        </div>
      );
    }
    return (
      <div onClick={onClick} className={classNames(styles.item, 'dip-flex-align-center', className)}>
        {unRead && <div className={styles.unRead} />}
        <MessageOutlined className="dip-text-color-45" />
        <div className={classNames('dip-ml-8 dip-flex-item-full-width dip-ellipsis dip-text-color-85')} title={label}>
          {label}
        </div>
        {status === 'processing' && (
          <Tooltip title={intl.get('dipChat.taskInProgress')}>
            <Spin size="small" indicator={<LoadingOutlined style={{ fontSize: 14 }} spin />} />
          </Tooltip>
        )}
        <Tooltip title={intl.get('dipChat.rename')}>
          <Button
            onClick={e => {
              e.stopPropagation();
              setIsEdit(true);
              form.setFieldValue('name', label);
              setTimeout(() => {
                inputRef.current?.focus({
                  cursor: 'end',
                });
              }, 0);
            }}
            className={classNames(styles.btn)}
            size="small"
            type="text"
            icon={<EditOutlined />}
          />
        </Tooltip>
        <Popconfirm
          rootClassName={styles.popConfirm}
          title={intl.get('dipChat.permanentDeleteConversation')}
          description={intl.get('dipChat.deleteConversationConfirm')}
          onConfirm={async () => {
            const res = await deleteConversationById(agentAppKey, id!);
            if (res) {
              messageApi.success(intl.get('dipChat.deleteSuccess'));
              refreshList?.();
              onClearSelectedConversation?.();
            }
          }}
          okText={intl.get('dipChat.delete')}
          okButtonProps={{
            danger: true,
          }}
          onPopupClick={e => e.stopPropagation()}
        >
          <Button
            onClick={e => e.stopPropagation()}
            className={classNames(styles.btn)}
            size="small"
            type="text"
            icon={<DipIcon type="icon-dip-trash" />}
          />
        </Popconfirm>
        <Popover
          placement="right"
          content={<div className="dip-p-12">更新时间：{dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss')}</div>}
        >
          <Button size="small" type="text" className={classNames(styles.btn)} icon={<ClockCircleOutlined />} />
        </Popover>
      </div>
    );
  };

  return (
    <>
      {renderItem()}
      {messageContextHolder}
    </>
  );
};

export default ConversationItem;
