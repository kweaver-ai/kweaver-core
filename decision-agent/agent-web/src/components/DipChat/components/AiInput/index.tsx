import styles from './index.module.less';
import classNames from 'classnames';
import { Attachments, Sender, Suggestion } from '@ant-design/x';
import React, { forwardRef, useImperativeHandle, useRef, useState } from 'react';
import type { AiInputProps, AiInputRef, AiInputValue } from './interface';
import _ from 'lodash';
import { Col, type GetRef, Row, Tooltip } from 'antd';
import { CloseCircleFilled, LoadingOutlined } from '@ant-design/icons';
import { useLatestState } from '@/hooks';
import { FileTypeIcon, getFileExtension } from '@/utils/doc';
import FileUploadBtn, { type FileUploadBtnRef } from '../FileUploadBtn';
import ResizeObserver from '@/components/ResizeObserver';
import type { FileItem } from '@/components/DipChat/interface';

const AiInput = forwardRef<AiInputRef, AiInputProps>((props, ref) => {
  const {
    value,
    onChange,
    suggestions,
    onSubmit,
    loading,
    clearAfterSend = true,
    onCancel,
    disabled,
    agentConfig,
    tempFileList = [],
    autoSize = { minRows: 3, maxRows: 6 },
    ...restProps
  } = props;

  const fileUploadBtnRef = useRef<FileUploadBtnRef>(null);
  const senderRef = React.useRef<GetRef<typeof Sender>>(null);
  const valueRef = useRef<AiInputValue>(value);

  // 文件相关props
  const attachmentsRef = React.useRef<GetRef<typeof Attachments>>(null);

  // 建议项相关props
  const [suggestionOpen, setSuggestionOpen] = useLatestState(false);
  const currentSuggestion = useRef({
    triggerChar: '',
    triggerCharIndex: -1,
    items: [] as any,
  });

  const [colSpan, setColPan] = useState(8);

  useImperativeHandle(ref, () => ({
    reset: resetForm,
  }));

  const handleChange = (newValue: AiInputValue) => {
    valueRef.current = newValue;
    onChange?.(newValue);
  };

  const resetForm = () => {
    const newValue = _.cloneDeep(value);
    newValue.inputValue = '';
    handleChange(newValue);
    senderRef.current?.focus();
    if (agentConfig.debug) {
      fileUploadBtnRef.current?.clearFileList();
    }
  };

  const suggestionSelect = () => {
    const cursorPosition = (document.activeElement as HTMLInputElement)?.selectionStart || 0;
    const triggerCharIndex = currentSuggestion.current.triggerCharIndex;
    // 将输入值从触发字符索引到光标位置替换为空字符串
    const inputValue = value.inputValue;
    const newInputValue = inputValue.substring(0, triggerCharIndex) + inputValue.substring(cursorPosition);
    const newValue = _.cloneDeep(value);
    newValue.inputValue = newInputValue;
    handleChange(newValue);
  };

  const inputDisabled = !value.inputValue && tempFileList.length === 0;

  const handleSubmit = () => {
    if (!inputDisabled && !loading) {
      senderRef.current?.focus();
      onSubmit?.(value);
      if (clearAfterSend) {
        resetForm();
      }
    }
  };

  const renderStatusIcon = (file: FileItem) => {
    if (file.status === 'processing') {
      return <LoadingOutlined className="dip-text-color-primary" />;
    }
    if (file.status === 'failed') {
      return (
        <Tooltip title={file.error}>
          <CloseCircleFilled className="dip-text-color-error" />
        </Tooltip>
      );
    }
  };

  const senderHeader = (
    <Sender.Header title="" open={tempFileList.length > 0 && agentConfig.debug} closable={false}>
      <ResizeObserver
        onResize={({ width }) => {
          if (width < 400) {
            setColPan(12);
          } else {
            setColPan(8);
          }
        }}
      >
        <div className="dip-full">
          <Row gutter={[16, 16]} className={styles.fileWrapper}>
            {tempFileList.map(item => (
              <Col span={colSpan} key={item.container_path}>
                <div
                  onClick={() => {
                    // onPreviewFile?.(item);
                  }}
                  className={classNames(styles.fileItem, 'dip-flex-align-center')}
                >
                  <FileTypeIcon extension={getFileExtension(item.name)} fontSize={20} />
                  <div className="dip-ml-8 dip-flex-item-full-width dip-flex-align-center">
                    <div title={item.name} className="dip-ellipsis dip-flex-item-full-width">
                      {item.name}
                    </div>
                    {renderStatusIcon(item)}
                  </div>
                  {/*<div*/}
                  {/*  className={styles.delete}*/}
                  {/*  onClick={e => {*/}
                  {/*    e.stopPropagation();*/}
                  {/*  }}*/}
                  {/*>*/}
                  {/*  <CloseCircleFilled />*/}
                  {/*</div>*/}
                </div>
              </Col>
            ))}
          </Row>
        </div>
      </ResizeObserver>
    </Sender.Header>
  );

  const fileBtnDisabled = disabled || value.mode !== 'normal';

  const renderFileBtn = () => {
    let show = false;
    if (agentConfig.debug) {
      // 调试模式下，无论文件上传怎么配置，上传文件按钮都是在对话框里面
      show = true;
    }
    if (show) {
      return (
        <FileUploadBtn
          ref={fileUploadBtnRef}
          disabled={fileBtnDisabled}
          onSuccess={() => {
            senderRef.current?.focus();
          }}
        />
      );
    }
  };

  return (
    <div className={classNames(styles.container, 'ai-input')}>
      <Suggestion
        block
        items={items => items}
        onSelect={suggestionSelect}
        onOpenChange={open => {
          setSuggestionOpen(open);
        }}
      >
        {({ onTrigger, onKeyDown }) => {
          return (
            <Sender
              {...restProps}
              value={value?.inputValue}
              onChange={nextVal => {
                const newValue = _.cloneDeep(value);
                newValue.inputValue = nextVal;
                handleChange(newValue);
                if (!_.isEmpty(suggestions) && nextVal) {
                  const cursorPosition = (document.activeElement as HTMLInputElement)?.selectionStart || 0;
                  const textBeforeCursor = nextVal.slice(0, cursorPosition); // 光标之前的文本
                  const charBeforeCursor = nextVal.slice(cursorPosition - 1, cursorPosition); // 光标之前的最后一个字符;
                  // console.log(textBeforeCursor, '光标之前的文本');
                  // console.log(charBeforeCursor, '光标之前的最后一个字符');
                  const targetSuggestion = suggestions!.find(item => item.triggerChar === charBeforeCursor);
                  if (targetSuggestion) {
                    currentSuggestion.current = {
                      triggerChar: targetSuggestion.triggerChar,
                      triggerCharIndex: textBeforeCursor.lastIndexOf(targetSuggestion.triggerChar),
                      items: targetSuggestion.items,
                    };
                  }
                  const triggerCharIndex = currentSuggestion.current.triggerCharIndex;
                  if (triggerCharIndex !== -1) {
                    const searchText = textBeforeCursor.slice(triggerCharIndex + 1);
                    if (searchText) {
                      const searchItems = currentSuggestion.current.items.filter((item: any) =>
                        item.value.includes(searchText)
                      );
                      if (searchItems.length > 0) {
                        onTrigger(searchItems);
                      } else {
                        if (suggestionOpen) {
                          onTrigger(false);
                        }
                      }
                    } else {
                      onTrigger(currentSuggestion.current.items);
                    }
                  }
                  return;
                }
                if (suggestionOpen) {
                  onTrigger(false);
                }
              }}
              onKeyDown={onKeyDown}
              ref={senderRef}
              onPasteFile={(_, files) => {
                Array.from(files).forEach(file => {
                  attachmentsRef.current?.upload(file);
                });
              }}
              autoSize={autoSize}
              footer={({ components }) => {
                const { SendButton, LoadingButton } = components;
                return (
                  <div className="dip-flex-space-between">
                    <span className="dip-flex-align-center">{renderFileBtn()}</span>
                    <span>
                      {loading ? (
                        <Tooltip title="停止输出">
                          <LoadingButton onClick={onCancel} type="default" disabled={false} />
                        </Tooltip>
                      ) : (
                        <Tooltip title="请输入你的问题" open={inputDisabled ? undefined : false}>
                          <SendButton
                            onClick={() => {
                              if (!value?.inputValue) {
                                handleSubmit();
                              }
                            }}
                            shape="default"
                            type="primary"
                            disabled={inputDisabled}
                          />
                        </Tooltip>
                      )}
                    </span>
                  </div>
                );
              }}
              onSubmit={handleSubmit}
              header={senderHeader}
              disabled={disabled}
              style={{ resize: 'none' }}
              actions={false}
              onKeyPress={e => {
                // 输入框没内容但是上传了文件的时候，按Enter键允许发送
                if (e.key === 'Enter' && !e.shiftKey && !value?.inputValue) {
                  handleSubmit();
                }
              }}
            />
          );
        }}
      </Suggestion>
    </div>
  );
});

export default AiInput;
