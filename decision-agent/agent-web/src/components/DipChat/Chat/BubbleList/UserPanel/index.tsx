import styles from './index.module.less';
import { useRef, useState } from 'react';
import { useDipChatStore } from '@/components/DipChat/store';
import _ from 'lodash';
import { FileTypeIcon, getFileExtension } from '@/utils/doc';
import classNames from 'classnames';
import { Col, Collapse, Row, Button, type GetRef } from 'antd';
import PanelFooter from '../PanelFooter';
import { nanoid } from 'nanoid';
import { Sender } from '@ant-design/x';
import intl from 'react-intl-universal';
import dayjs from 'dayjs';

const UserPanel = ({ chatItemIndex, readOnly }: any) => {
  const {
    dipChatStore: { chatList },
    sendChat,
    setDipChatStore,
    closeSideBar,
  } = useDipChatStore();
  const chatItem = chatList[chatItemIndex];
  const { content, fileList, updateTime } = chatItem;
  const [isEdit, setIsEdit] = useState(false);
  const [inputValue, setInputValue] = useState(content);
  const senderRef = useRef<GetRef<typeof Sender>>(null);
  const renderFile = () => {
    if (!_.isEmpty(fileList)) {
      return (
        <Collapse
          className="dip-w-100"
          ghost
          expandIconPosition="end"
          defaultActiveKey={['file']}
          items={[
            {
              key: 'file',
              label: intl.get('dipChat.fileList'),
              children: (
                <Row gutter={[16, 16]} className={styles.fileWrapper}>
                  {fileList!.map((item: any) => (
                    <Col span={8} key={item.id}>
                      <div
                        onClick={() => {
                          if (readOnly) {
                            return;
                          }
                          // setDipChatStore({
                          //   previewFile: {
                          //     fileId: item.id,
                          //     fileName: item.name,
                          //     fileExt: getFileExtension(item.name),
                          //   },
                          // });
                        }}
                        className={classNames(styles.fileItem, 'dip-flex-align-center')}
                      >
                        <FileTypeIcon extension={getFileExtension(item?.name)} fontSize={20} />
                        <div className="dip-ml-8 dip-flex-item-full-width">
                          <div title={item?.name} className="dip-ellipsis">
                            {item?.name}
                          </div>
                          {/* <div className="dip-text-color-45 dip-font-12">{convertFileSize(item.size)}</div>*/}
                        </div>
                      </div>
                    </Col>
                  ))}
                </Row>
              ),
            },
          ]}
        />
      );
    }
  };

  const onEdit = () => {
    setDipChatStore({
      chatListAutoScroll: false,
    });
    setInputValue(content);
    setIsEdit(!isEdit);
    setTimeout(() => {
      senderRef.current?.focus({
        cursor: 'end',
      });
    }, 0);
  };
  const renderFooter = () => {
    if (!isEdit) {
      return (
        <div className="dip-flex-align-center dip-mt-8">
          {updateTime && (
            <span style={{ marginRight: 32 }} className="dip-text-color-45 dip-font-12">
              {dayjs(updateTime).format('YYYY-MM-DD HH:mm:ss')}
            </span>
          )}
          <PanelFooter chatItemIndex={chatItemIndex} onEdit={onEdit} />
        </div>
      );
    }
  };

  const submit = () => {
    if (!inputValue) {
      return;
    }
    const newChatList = _.cloneDeep(chatList);
    const regenerate_user_message_id = chatItem.key;
    newChatList.splice(chatItemIndex);
    newChatList.push({
      key: nanoid(),
      role: 'user',
      content: inputValue,
      loading: false,
      fileList: fileList ?? [],
      updateTime: dayjs().valueOf(),
    });

    newChatList.push({
      key: nanoid(),
      role: 'common',
      content: '',
      loading: true,
    });
    const body: any = {
      query: inputValue,
      regenerate_user_message_id,
    };
    if (fileList && fileList.length > 0) {
      // 说明有文件
      body.selected_files = fileList.map(item => ({
        file_name: item.container_path,
      }));
    }
    closeSideBar();
    sendChat({
      chatList: newChatList,
      body,
    });
  };

  const renderContent = () => {
    if (isEdit) {
      return (
        <div className={styles.editWrapper}>
          <Sender
            className="dip-bg-white"
            actions={false}
            autoSize={{ minRows: 2, maxRows: 6 }}
            value={inputValue}
            onChange={nextVal => {
              setInputValue(nextVal);
            }}
            footer={() => {
              return (
                <div style={{ textAlign: 'right' }}>
                  <Button disabled={_.isEmpty(inputValue) || inputValue === content} onClick={submit} type="primary">
                    {intl.get('dipChat.confirm')}
                  </Button>
                  <Button
                    className="dip-ml-12"
                    onClick={() => {
                      setIsEdit(false);
                    }}
                  >
                    {intl.get('dipChat.cancel')}
                  </Button>
                </div>
              );
            }}
            onSubmit={submit}
            ref={senderRef}
          />
        </div>
      );
    } else {
      return <div className={styles.content}>{content}</div>;
    }
  };
  return (
    <div className={styles.container}>
      {renderFile()}
      {content.trim() && renderContent()}
      {content.trim() && !readOnly && <div className={styles.footer}>{renderFooter()}</div>}
    </div>
  );
};

export default UserPanel;
