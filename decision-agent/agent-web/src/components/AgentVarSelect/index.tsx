import { type CSSProperties, type ReactNode, useEffect, useMemo, useRef, useState } from 'react';
import intl from 'react-intl-universal';
import classNames from 'classnames';
import { Input, Dropdown, type MenuProps } from 'antd';
import { DownOutlined } from '@ant-design/icons';
import { useDeepCompareMemo, useLatestState, useMicroWidgetProps } from '@/hooks';
import IconFont from '@/components/IconFont';
import { selfConfigConst } from './types';
import SelfConfigVarSelector from './SelfConfigVarSelector';
import './style.less';

export type OptionItemType = {
  label: ReactNode;
  value: string;
  type: string;
  disabled?: boolean;
  editable?: boolean; // 是否可编辑
};

export type AgentVarSelectProps = {
  placeholder?: string;
  className?: string;
  style?: CSSProperties;
  options?: OptionItemType[];
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
};
const AgentVarSelect = (props: AgentVarSelectProps) => {
  const { className, style, options, value, onChange, placeholder, disabled } = props;
  const microWidgetProps = useMicroWidgetProps();
  const prefixCls = 'agent-var-select';
  const [searchProps, setSearchProps] = useState({
    value: '',
    width: 4,
  });
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [currentFocusMenuIndex, setCurrentFocusMenuIndex, getCurrentFocusMenuIndex] = useLatestState(0);
  const searchInputRef = useRef<any>();
  const varTagInputRef = useRef<any>();
  const placeholderSpanRef = useRef<any>();
  const filterMenuItemsRef = useRef<any>([]);
  const [focusTag, setFocusTag] = useState(false);
  const [showSelfConfigModal, setShowSelfConfigModal] = useState<boolean>(false);

  const menuItems: MenuProps['items'] = useDeepCompareMemo(() => {
    if (options) {
      return options.map(item => ({
        key: item.value,
        label: item.label,
        type: item.type,
        disabled: item.disabled,
      }));
    }
    return [];
  }, [options]);

  filterMenuItemsRef.current = useDeepCompareMemo(() => {
    if (searchProps.value) {
      return menuItems?.filter(item => item!.key!.toString().toLowerCase().includes(searchProps.value.toLowerCase()));
    } else {
      return menuItems;
    }
  }, [menuItems, searchProps.value]);

  useEffect(() => {
    if (dropdownOpen) {
      document.addEventListener('keydown', handleKeydown);
    }
    return () => {
      document.removeEventListener('keydown', handleKeydown);
    };
  }, [dropdownOpen]);

  const handleKeydown = (event: any) => {
    let newCurrentFocusMenuIndex = getCurrentFocusMenuIndex();
    if (event.key === 'Enter') {
      event.preventDefault();
      menuSelect(filterMenuItemsRef.current[newCurrentFocusMenuIndex]?.key);
      return;
    }
    if (event.key === 'ArrowUp') {
      event.preventDefault();
      newCurrentFocusMenuIndex =
        newCurrentFocusMenuIndex <= 0 ? filterMenuItemsRef.current.length - 1 : newCurrentFocusMenuIndex - 1;
    }
    if (event.key === 'ArrowDown') {
      event.preventDefault();
      newCurrentFocusMenuIndex =
        newCurrentFocusMenuIndex >= filterMenuItemsRef.current.length - 1 ? 0 : newCurrentFocusMenuIndex + 1;
    }
    setCurrentFocusMenuIndex(newCurrentFocusMenuIndex);
  };

  useEffect(() => {
    if (!dropdownOpen) {
      searchInputRef.current?.blur();
      clearSearchInput();
    }
  }, [dropdownOpen]);

  const mergeValue = useMemo(() => {
    return [value];
  }, [value]);

  const selectedKeys = useMemo(() => {
    if (value?.startsWith(`${selfConfigConst}.`)) {
      return [selfConfigConst];
    }

    return [value].filter(Boolean);
  }, [value]);

  const handlePaste = (event: any) => {
    // 阻止默认粘贴行为
    event.preventDefault();

    // 获取粘贴的内容
    const clipboardData = event.clipboardData || (window as any).clipboardData;
    const text = clipboardData.getData('text');

    // 插入纯文本内容
    document.execCommand('insertText', false, text);
  };

  const renderValueTag = (data: string) => {
    if (data) {
      const targetOption = options?.find(item => item.value === data);
      return (
        <div
          ref={varTagInputRef}
          onClick={event => {
            if (targetOption?.editable) {
              event.stopPropagation();
            }
          }}
          className={classNames(`${prefixCls}-item`, {
            [`${prefixCls}-item-focus`]: focusTag,
          })}
          suppressContentEditableWarning
          contentEditable={
            disabled
              ? false
              : targetOption
                ? (targetOption.editable ?? false)
                : data.startsWith(`${selfConfigConst}.`)
                  ? false
                  : true
          }
          onPaste={handlePaste}
          onBlur={event => {
            const textContent = (event.target as HTMLDivElement)?.innerText;
            onChange(textContent);
            setFocusTag(false);
          }}
          onKeyDown={event => {
            if (event.key === 'Enter') {
              event.preventDefault();
            }
          }}
          onFocus={() => {
            setFocusTag(true);
          }}
          title={data}
        >
          {data}
        </div>
      );
    }
  };

  const responsiveSearchWidth = (inputValue: string) => {
    if (placeholderSpanRef.current) {
      placeholderSpanRef.current.textContent = inputValue;
      setSearchProps(prevState => ({
        ...prevState,
        width: placeholderSpanRef.current.offsetWidth,
      }));
    }
  };

  const clearSearchInput = () => {
    responsiveSearchWidth('');
    setSearchProps(prevState => ({
      ...prevState,
      value: '',
    }));
  };

  const newMenuItems = useDeepCompareMemo(() => {
    return filterMenuItemsRef.current.map((item: any, index: number) => ({
      ...item,
      className: index === currentFocusMenuIndex ? 'agent-var-item-hover' : '',
      label: (
        <span style={{ fontWeight: 700 }}>
          <span>
            {item.label} <span>{item.label.includes('.') && intl.get('dataAgent.config.userDefined')}</span>
          </span>
          <span className="dip-font-12 dip-c-subtext dip-ml-8">
            <span>({item.type})</span>
          </span>
        </span>
      ),
    }));
  }, [currentFocusMenuIndex, filterMenuItemsRef.current]);

  const menuSelect = (key: string) => {
    if (key === value && key !== selfConfigConst) return;

    const selectValue = key;
    const targetOption = options!.find(item => item.value === key);
    if (targetOption?.editable) {
      // selectValue = `${selectValue}.`;
      setTimeout(() => {
        varTagInputRef.current?.focus();
        // 创建一个range对象并将其设置到div的末尾
        const range = document.createRange();
        range.selectNodeContents(varTagInputRef.current);
        range.collapse(false);

        // 获取当前的selection对象并将range添加到selection中
        const selection = window.getSelection();
        selection?.removeAllRanges();
        selection?.addRange(range);
      }, 0);
    }
    setDropdownOpen(false);

    if (selectValue === selfConfigConst) {
      // 弹出选择弹窗
      setShowSelfConfigModal(true);
    } else {
      onChange(selectValue);
    }
  };

  return (
    <>
      <Dropdown
        disabled={disabled}
        open={dropdownOpen}
        menu={{
          items: newMenuItems,
          selectedKeys,
          onClick: ({ key }) => {
            menuSelect(key);
          },
          selectable: true,
          className: `${prefixCls}-menu`,
        }}
        getPopupContainer={() => microWidgetProps.container}
        trigger={['click']}
        onOpenChange={open => {
          setDropdownOpen(open);
          if (open) {
            searchInputRef.current?.focus();
          }
        }}
      >
        <div
          style={style}
          className={classNames(prefixCls, className, {
            [`${prefixCls}-disabled`]: disabled,
          })}
        >
          {mergeValue.map(item => renderValueTag(item))}
          {!value && (
            <div className={`${prefixCls}-search`}>
              <Input
                ref={searchInputRef}
                value={searchProps.value}
                prefixCls={`${prefixCls}-search-input`}
                style={{ width: searchProps.width }}
                onChange={e => {
                  setCurrentFocusMenuIndex(0);
                  const inputValue = e.target.value;
                  responsiveSearchWidth(inputValue);
                  setSearchProps(prevState => ({
                    ...prevState,
                    value: e.target.value,
                  }));
                }}
              />
              <span ref={placeholderSpanRef} className={`${prefixCls}-search-mirror`} />
            </div>
          )}
          <div
            className={`${prefixCls}-clear`}
            onClick={e => {
              e.stopPropagation();
              onChange('');
            }}
          >
            <IconFont type="icon-shibai" />
          </div>
          <div className={`${prefixCls}-arrow dip-font-12`}>
            <DownOutlined />
          </div>
          {!value && placeholder && !searchProps.value && (
            <div className={`${prefixCls}-placeholder`}>{placeholder}</div>
          )}
        </div>
      </Dropdown>
      {showSelfConfigModal && (
        <SelfConfigVarSelector
          defaultValue={value === selfConfigConst || value?.startsWith(`${selfConfigConst}.`) ? value : ''}
          onCancel={() => setShowSelfConfigModal(false)}
          onConfirm={path => {
            setShowSelfConfigModal(false);
            onChange(path);
          }}
        />
      )}
    </>
  );
};

export default AgentVarSelect;
