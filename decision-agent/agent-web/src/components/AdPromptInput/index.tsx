import { type CSSProperties, forwardRef, useImperativeHandle, useRef, useState } from 'react';
import ReactDOM from 'react-dom/client';
import './style.less';
import classNames from 'classnames';
import { Select } from 'antd';
import { getPromptVarFromString } from './assitant';
import { useLatestState, useDeepCompareEffect, useUpdateEffect } from '@/hooks';
import VarElement from './VarElement';

import type { AdPromptInputProps, AdPromptInputRef, Position } from './interface';

const AdPromptInput = forwardRef<AdPromptInputRef, AdPromptInputProps>((props, ref) => {
  const isControl = 'value' in props; // 是否受控
  const {
    className,
    style,
    value,
    onChange,
    trigger,
    onFocus,
    onBlur,
    onKeyDown,
    placeholder,
    bordered = true,
    disabled = false,
    getPopupContainer,
    footer,
  } = props;
  const [visible, setVisible, getVisible] = useLatestState<boolean>(false); // 控制Select的显示与隐藏
  const currentTriggerCharacterIndex = useRef<number>(0); // 触发字符的索引
  const currentFocusNode = useRef<Node | string>();
  const searchStrRef = useRef<string>(''); // 鼠标小手之前的字符串
  const [cursorPosition, setCursorPosition] = useState<Position>({
    x: 0,
    y: 0,
  });
  const [inputFocus, setInputFocus, getInputFocus] = useLatestState<boolean>(false);
  const inputRef = useRef<HTMLDivElement | null>(null);
  const selectRef = useRef<any>(null);
  const [selectSearchText, setSelectSearchText, getSelectSearchText] = useLatestState('');

  const [triggerOptions, setTriggerOptions] = useState<any[]>([]);

  const prefixCls = 'ad-prompt-input';

  useImperativeHandle(ref, () => ({
    setValue,
  }));

  useDeepCompareEffect(() => {
    if (isControl && !getInputFocus()) {
      setValue(value);
    }
  }, [value, trigger]);

  useUpdateEffect(() => {
    if (visible) {
      setTimeout(() => {
        selectRef.current?.focus();
      }, 0);
    }
  }, [visible]);

  const handleBlur = () => {
    setVisible(false);
    setInputFocus(false);
    onBlur?.();
  };

  const setValue = async (data?: string) => {
    const html = await textConvertHtml(data);
    inputRef.current!.innerHTML = '';
    inputRef.current!.appendChild(html);
  };

  /** 提示词文本转换为html */
  const textConvertHtml = async (data?: string) => {
    let targetHtml;
    if (data) {
      const fragment = document.createDocumentFragment();
      const inputValueArr = data.replace(/ /g, '\u00A0').split('\n');
      for (let i = 0; i < inputValueArr.length; i++) {
        const rowDiv = document.createElement('div');
        rowDiv.className = `${prefixCls}-editor-row`;
        let lineStr = inputValueArr[i];
        const vars = getPromptVarFromString(lineStr);
        let lastIndex = 0;
        for (let j = 0; j < vars.length; j++) {
          const variable = vars[j];
          const varIndex = lineStr.indexOf(variable, lastIndex);
          if (varIndex > lastIndex) {
            const textNode = document.createTextNode(lineStr.slice(lastIndex, varIndex));
            rowDiv.appendChild(textNode);
          }
          const varNode = await createVarNode(variable, undefined, { appendSpaceAfter: false });
          rowDiv.appendChild(varNode as Node);
          lastIndex = varIndex + variable.length;
          lineStr = lineStr.slice(0, varIndex) + ' '.repeat(variable.length) + lineStr.slice(lastIndex);
        }
        if (lastIndex < lineStr.length) {
          const textNode = document.createTextNode(lineStr.slice(lastIndex));
          rowDiv.appendChild(textNode);
        }
        if (!rowDiv.hasChildNodes()) {
          rowDiv.appendChild(document.createElement('br'));
        }
        fragment.appendChild(rowDiv);
      }
      targetHtml = fragment;
      // inputRef.current!.innerHTML = '';
      // inputRef.current!.appendChild(fragment);
    } else {
      const rowDiv = document.createElement('div');
      rowDiv.className = `${prefixCls}-editor-row`;
      rowDiv.appendChild(document.createElement('br'));
      targetHtml = rowDiv;
      // inputRef.current!.innerHTML = '';
      // inputRef.current!.appendChild(rowDiv);
    }
    return targetHtml;
  };

  /** 获取选择器弹框坐标 */
  const getCursorPosition = () => {
    const { x, y } = window.getSelection()?.getRangeAt(0).getBoundingClientRect() as any;
    const { x: eX, y: eY } = inputRef.current?.getBoundingClientRect() as any;
    setCursorPosition({ x: x - eX, y: y - eY });
  };

  /**
   * 监听输入框，控制下拉选择框的显示与隐藏
   */
  const handleSelectVisible = () => {
    let cursorBeforeStr = '';
    const selection: any = window.getSelection();
    if (selection?.focusNode?.data) {
      cursorBeforeStr = selection.focusNode?.data.slice(0, selection.focusOffset);
    }
    currentFocusNode.current = selection.focusNode;
    const [{ character: triggerCharacter, options } = {}] =
      trigger?.filter(item => cursorBeforeStr.endsWith(item.character)) || [];

    if (!triggerCharacter) {
      setVisible(false);
      return;
    }

    setTriggerOptions(options || []);

    const lastCharacterIndex = cursorBeforeStr?.lastIndexOf(triggerCharacter!);
    currentTriggerCharacterIndex.current = lastCharacterIndex;

    if (lastCharacterIndex !== -1) {
      getCursorPosition();
      const searchStr = cursorBeforeStr.slice(lastCharacterIndex + 1);
      // triggerCharacter之后可以输入空格以终止查询操作
      // if (!isIncludeSpacesOrLineBreak(searchStr)) {
      // console.log(searchStr, 'searchStr+++++++++');
      // todo 中文输入法存在自动补}的情况，要么补全}的情况下  禁止弹下拉选择框， 要么光标放在}后面弹框
      if (searchStr === '' || searchStr === '}') {
        searchStrRef.current = searchStr;
        setVisible(true);
        setSelectSearchText('');
      } else {
        setVisible(false);
        searchStrRef.current = '';
      }
    } else {
      setVisible(false);
    }
  };

  const editorClick = async () => {};

  const editorChange = () => {
    handleInputText();

    if (trigger?.length) {
      handleSelectVisible();
    }
  };

  /**
   * 创建变量节点
   * @param key 变量
   */
  const createVarNode = (key: string, autoFocus: boolean = false, { appendSpaceAfter = true } = {}) => {
    return new Promise(resolve => {
      let options = [];
      const isVar = key.startsWith('$');
      if (isVar) {
        options = trigger?.find(item => item.character === '$')?.options || [];
      } else {
        options = trigger?.filter(item => item.character !== '$')?.map(item => item.options) || [];
        options = options.flat();
      }
      const selectedItem = options.find(item => {
        if (key.includes('.') && item.value !== key) {
          // 处理对象属性访问，如 $history.user
          const [varKey] = key.replace(/^\$/, '').split('.');
          return item.value === (isVar ? `$${varKey}` : varKey);
        }

        return item.value === key;
      });
      const ele = document.createElement('span');
      ele.classList.add(`${prefixCls}-var`);
      if (isVar) {
        ele.contentEditable = 'false';
        if (selectedItem?.type !== 'object') {
          ele.onclick = e => {
            e.stopPropagation();
            const range = document.createRange();
            range.selectNode(ele);
            const selection = window.getSelection();
            selection?.removeAllRanges();
            selection?.addRange(range);
          };
        }
      }

      const fragment = document.createDocumentFragment();
      fragment.appendChild(ele);
      if (appendSpaceAfter) {
        // 在变量后面添加一个不间断空格
        fragment.appendChild(document.createTextNode('\u00A0'));
      }
      if (isVar) {
        ReactDOM.createRoot(ele).render(
          <VarElement
            options={options}
            type={selectedItem?.type}
            value={key}
            error={options.length > 0 && !selectedItem}
            autoFocus={autoFocus}
            editable={selectedItem?.type === 'object'}
          />
        );
      } else {
        // 非变量，当做普通文本处理
        ele.textContent = key;
      }

      setTimeout(() => {
        resolve(fragment);
      }, 100);
    });
  };

  const handleInputText = () => {
    let inputText = inputRef.current!.innerText;
    if (inputText) {
      inputText = inputText.replaceAll('\n\n', '\n');
      inputText = inputText.replace(/\u00A0/g, ' ');
      if (inputText === '\n') {
        inputText = '';
      }
    }
    onChange?.(inputText);
  };

  const onSelect = async (key: string, opt: any) => {
    const autoFocus = opt?.type === 'object';
    const selection = window.getSelection();
    const range = selection?.getRangeAt(0) as Range;
    const focusNode = currentFocusNode.current as Node;
    range.setStart(focusNode, currentTriggerCharacterIndex.current);
    range.setEnd(focusNode, currentTriggerCharacterIndex.current + 1 + searchStrRef.current.length);
    range.deleteContents();
    // 创建元素节点
    const varElement: any = await createVarNode(key, autoFocus);
    // 插入元素节点
    range.insertNode(varElement);
    if (!autoFocus) {
      // 光标移动到末尾
      range.collapse();

      inputRef.current?.focus();
    }
    // 关闭弹框
    setVisible(false);
    handleInputText();
  };

  const handlePaste = async (event: any) => {
    event.preventDefault();
    const text = event.clipboardData
      .getData('text/plain')
      .replace(/\r\n/g, '\n') // 统一换行符为\n
      .replace(/ /g, '\u00A0') // 将普通空格替换为不间断空格
      .split('\n') // 分割成行
      .join('\n'); // 重新组合

    const html = await textConvertHtml(text);
    const selection = window.getSelection();
    const range = selection?.getRangeAt(0);
    range?.deleteContents();

    // 创建临时容器
    const tempDiv = document.createElement('div');
    tempDiv.appendChild(html);

    // 使用 insertHTML 命令插入转换后的 HTML
    document.execCommand('insertHTML', false, tempDiv.innerHTML);
  };

  const placeholderStyle: CSSProperties = bordered ? { top: 8, left: 11 } : { top: 0, left: 0 };

  const { x, y } = cursorPosition;
  const selectStyle: CSSProperties = bordered ? { left: x + 11, top: y + 8 } : { left: x, top: y };
  return (
    <div
      className={classNames(prefixCls, className, {
        [`${prefixCls}-border`]: bordered,
        [`${prefixCls}-focus`]: inputFocus && bordered,
        [`${prefixCls}-disabled`]: disabled,
      })}
    >
      <div
        style={style}
        ref={inputRef}
        className={classNames(`${prefixCls}-editor`)}
        suppressContentEditableWarning
        contentEditable={!disabled}
        onInput={editorChange}
        onClick={editorClick}
        onFocus={() => {
          setInputFocus(true);
          onFocus?.();
        }}
        onBlur={() => {
          if (!getVisible()) {
            handleBlur();
          }
        }}
        onKeyUp={e => {
          if (e.keyCode === 8) {
            if (!inputRef.current!.innerText || !value) {
              inputRef.current!.innerHTML = `<div class="${prefixCls}-editor-row"><br /></div>`;
            }
          }
        }}
        onKeyDown={onKeyDown}
        onPaste={handlePaste}
      >
        <div className={`${prefixCls}-editor-row`}>
          <br />
        </div>
      </div>
      {visible && (
        <div style={{ position: 'absolute', ...selectStyle }}>
          <Select
            size="small"
            ref={selectRef}
            style={{ minWidth: 60 }}
            searchValue={selectSearchText}
            className={`${prefixCls}-select`}
            showSearch
            options={triggerOptions}
            open
            onSelect={onSelect}
            bordered={false}
            showArrow={false}
            dropdownMatchSelectWidth={false}
            onSearch={searchStr => {
              setSelectSearchText(searchStr);
            }}
            onKeyDown={e => {
              if (e.keyCode === 8) {
                if (!getSelectSearchText()) {
                  setVisible(false);
                  const selection = window.getSelection();
                  const range = selection?.getRangeAt(0) as Range;
                  const focusNode = currentFocusNode.current as Node;
                  range.setStart(focusNode, currentTriggerCharacterIndex.current);
                  range.setEnd(focusNode, currentTriggerCharacterIndex.current + 1 + searchStrRef.current.length);
                  range.collapse();
                  inputRef.current?.focus();
                }
              }
            }}
            getPopupContainer={getPopupContainer}
          />
        </div>
      )}
      {!!placeholder && (!value || value === '\n') && (
        <div style={placeholderStyle} className={`${prefixCls}-placeholder`}>
          {placeholder}
        </div>
      )}
      {footer && <div className={`${prefixCls}-footer`}>{footer}</div>}
    </div>
  );
});

export default AdPromptInput;
