import { useEffect, useMemo, useRef, useState } from 'react';
import './style.less';
import classNames from 'classnames';
import { getVarTextFromPromptVar } from '@/components/AdPromptInput/assitant';
import type { VarOptions } from '@/components/AdPromptInput/interface.ts';
export type VarElementProps = {
  className?: string;
  error?: boolean;
  value: string;
  type?: string;
  autoFocus?: boolean; // 是否自动聚焦到可编辑的变量标签上
  options: VarOptions[];
  editable?: boolean; // 是否可编辑
};
const VarElement = (props: VarElementProps) => {
  const { className, error = false, value, type, autoFocus = false, options = [], editable = false } = props;
  const editableSpanRef = useRef<any>();
  const [varError, setError] = useState(error);
  const [contentEditable, setContentEditable] = useState(editable);

  useEffect(() => {
    validateError();
  }, [error, options, type]);

  const validateError = () => {
    if (type === 'object') {
      if (error) {
        setError(true);
      } else {
        const inputValue = editableSpanRef.current!.innerText;
        let varName = '';
        if (inputValue.includes('.')) {
          varName = inputValue.split('.')[0];
        } else {
          varName = inputValue;
        }
        setError(!options.some(item => item.value === `$${varName}`));
      }
      return;
    }
    setError(error);
  };

  useEffect(() => {
    // 在父组件select选中变量生成的时候，需要主动将contentEditable置为true
    if (autoFocus && type === 'object') {
      setTimeout(() => {
        if (editableSpanRef.current) {
          editableSpanRef.current?.focus();
          // 创建一个range对象并将其设置到div的末尾
          const range = document.createRange();
          range.selectNodeContents(editableSpanRef.current);
          range.collapse(false);

          // 获取当前的selection对象并将range添加到selection中
          const selection = window.getSelection();
          selection?.removeAllRanges();
          selection?.addRange(range);
        }
      }, 100);
    }
  }, []);

  const arr = useMemo(() => {
    return getVarTextFromPromptVar(value);
  }, [value]);

  const prefixCls = 'var-element';

  const handlePaste = (event: any) => {
    // 阻止默认粘贴行为
    event.preventDefault();

    // 获取粘贴的内容
    const clipboardData = event.clipboardData || (window as any).clipboardData;
    const text = clipboardData.getData('text');

    // 插入纯文本内容
    document.execCommand('insertText', false, text);
  };

  return (
    <span
      contentEditable={false}
      className={classNames(prefixCls, className, {
        [`${prefixCls}-error`]: varError,
      })}
      onClick={() => {
        if (type === 'object') {
          setContentEditable(true);
          setTimeout(() => {
            editableSpanRef.current?.focus();
          }, 0);
        }
      }}
    >
      <span className={`${prefixCls}-block`} contentEditable={false}>
        {arr[0]}
      </span>
      <span
        ref={editableSpanRef}
        className={classNames(`${prefixCls}-text`, {
          [`${prefixCls}-text-editable`]: type === 'object',
        })}
        // contentEditable={type === 'object'}
        contentEditable={contentEditable}
        onKeyDown={event => {
          if (event.key === 'Enter') {
            event.preventDefault();
          }
        }}
        onPaste={handlePaste}
        onInput={() => {
          if (type === 'object') {
            validateError();
          }
        }}
        onBlur={() => {
          setContentEditable(false);
        }}
      >
        {arr[1]}
      </span>
      <span style={{ padding: arr[2] ? 0 : '0 4px' }} className={`${prefixCls}-block`} contentEditable={false}>
        {arr[2]}
      </span>
    </span>
  );
};

export default VarElement;
