import styles from './index.module.less';
import classNames from 'classnames';
import ScrollBarContainer from '@/components/ScrollBarContainer';
import FileUploadBtn, { type FileUploadBtnRef } from '../../../components/FileUploadBtn';
import NoData from '@/components/NoData';
import { Button, Checkbox, message, Spin } from 'antd';
import DipIcon from '@/components/DipIcon';
import { FileTypeIcon, getFileExtension } from '@/utils/doc';
import { LoadingOutlined } from '@ant-design/icons';
import { useDipChatStore } from '@/components/DipChat/store';
import intl from 'react-intl-universal';
import { useEffect, useRef } from 'react';
const FileArea = () => {
  const fileUploadBtnRef = useRef<FileUploadBtnRef>(null);
  const {
    dipChatStore: { tempFileList, activeConversationKey },
    setDipChatStore,
  } = useDipChatStore();
  const [messageApi, contextHolder] = message.useMessage();

  useEffect(() => {
    if (activeConversationKey) {
      fileUploadBtnRef.current?.getFileList();
    }
    return () => {
      if (activeConversationKey) {
        setDipChatStore({ tempFileList: [] });
      }
    };
  }, [activeConversationKey]);

  // const deleteFile = async (fileIds: string[]) => {};

  const renderContent = () => {
    if (tempFileList.length === 0) {
      return (
        <div className="dip-full dip-center">
          <div className="dip-flex-column-center">
            <NoData tip={intl.get('dipChat.noTempData')} />
            <div className="dip-mt-12">
              <FileUploadBtn
                customBtn={
                  <Button size="small" icon={<DipIcon className="dip-font-12" type="icon-dip-upload" />}>
                    {intl.get('dipChat.uploadData')}
                  </Button>
                }
              />
            </div>
          </div>
        </div>
      );
    }
    return (
      <div>
        {tempFileList.map(file => {
          return (
            <div className={styles.fileItem} key={file.container_path}>
              <div className="dip-flex-align-center dip-flex-item-full-width">
                <Checkbox
                  disabled={file.status !== 'completed'}
                  checked={file.checked}
                  onChange={e => {
                    if (e.target.checked) {
                      const checkedFiles = tempFileList.filter(item => item.checked);
                      if (checkedFiles.length >= 1) {
                        messageApi.warning(intl.get('dipChat.singleChatMaxFiles', { count: 1 }));
                        return;
                      }
                    }
                    setDipChatStore({
                      tempFileList: tempFileList.map(item => {
                        if (item.container_path === file.container_path) {
                          return {
                            ...item,
                            checked: e.target.checked,
                          };
                        }
                        return { ...item };
                      }),
                    });
                  }}
                />
                <div className="dip-flex-align-center dip-flex-item-full-width dip-ml-8 dip-pointer">
                  <FileTypeIcon extension={getFileExtension(file.name)} fontSize={16} />
                  <span
                    title={file.name}
                    className={classNames(styles.fileName, 'dip-flex-item-full-width dip-ellipsis')}
                  >
                    {file.name}
                  </span>
                  {file.status === 'processing' && (
                    <span className="dip-ml-4 dip-mr-4">
                      <Spin indicator={<LoadingOutlined spin />} size="small" />
                    </span>
                  )}
                </div>
              </div>
              {/*<span className={styles.btn}>*/}
              {/*  <Tooltip title={intl.get('dipChat.remove')}>*/}
              {/*    <DipButton*/}
              {/*      onClick={e => {*/}
              {/*        e.stopPropagation();*/}
              {/*        deleteFile([file.container_path]);*/}
              {/*      }}*/}
              {/*      size="small"*/}
              {/*      type="text"*/}
              {/*      icon={<DipIcon type="icon-dip-trash" />}*/}
              {/*    />*/}
              {/*  </Tooltip>*/}
              {/*</span>*/}
            </div>
          );
        })}
      </div>
    );
  };

  return (
    <div className={classNames(styles.container, 'dip-flex-column')}>
      {contextHolder}
      <div className="dip-flex-space-between dip-pl-8 dip-pr-8">
        <span className="dip-font-weight-700">临时文件</span>
        <FileUploadBtn
          customBtn={<Button type="text" size="small" icon={<DipIcon type="icon-dip-upload" />} />}
          ref={fileUploadBtnRef}
        />
      </div>
      <ScrollBarContainer className="dip-flex-item-full-height dip-pl-8 dip-pr-8">{renderContent()}</ScrollBarContainer>
    </div>
  );
};

export default FileArea;
