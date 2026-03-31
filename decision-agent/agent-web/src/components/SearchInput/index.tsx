/**
 * 基于antd再封装的搜索框
 * 中文输入时默认劫持onChange事件, 中文输入结束才触发onChange
 *
 */

import React, { forwardRef, useEffect, useImperativeHandle, useRef, useState, useCallback } from 'react';
import type { InputProps } from 'antd';
import { Input } from 'antd';
import classNames from 'classnames';
import IconFont from '@/components/IconFont';
import './style.less';
import _ from 'lodash';

export interface SearchInputProps extends InputProps {
  onIconClick?: Function; // 点击icon回调
  onClear?: Function; // 清空搜索框回调
  iconPosition?: 'start' | 'end'; // 搜索图标的位置, 前 | 后
  debounce?: boolean; // 是否启用防抖， 默认不启用
  debounceWait?: number; // 防抖的延迟时间，默认 300 ms
}

const createDebounceChange = (wait: number) =>
  _.debounce((callback?: Function, e?: React.ChangeEvent<HTMLInputElement>) => {
    callback?.(e);
  }, wait);

const SearchInputFunc: React.ForwardRefRenderFunction<unknown, SearchInputProps> = (props, ref) => {
  const {
    className = '',
    iconPosition = 'start',
    onChange,
    onPressEnter,
    onIconClick,
    onClear,
    debounce = false,
    debounceWait = 300,
    value,
    allowClear = true,
    ...otherProps
  } = props;
  const inputRef = useRef<any>();
  const isCompos = useRef(false); // 标记键盘输入法
  const [inputValue, setInputValue] = useState(value);
  const debounceChangeRef = useRef(createDebounceChange(debounceWait));

  useEffect(() => {
    if (value) {
      inputRef.current.input.value = value;
    }
  }, []);

  useEffect(() => {
    setInputValue(value);
  }, [value]);

  // 转发ref
  useImperativeHandle(ref, () => ({
    input: inputRef.current.input,
    setValue: (value: any) => {
      setInputValue(value);
      inputRef.current.input.value = value;
      // 触发onChange事件
      const event = {
        target: { value },
        type: 'change',
      } as React.ChangeEvent<HTMLInputElement>;
      handleChange(event);
    },
  }));

  // 输入框变化
  const handleChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      // 立即更新输入框的值
      setInputValue(e.target.value);

      // 在输入法输入过程中，不触发onChange
      if (isCompos.current) return;

      // 根据是否开启防抖决定如何触发onChange
      if (debounce) {
        debounceChangeRef.current(onChange, e);
      } else {
        onChange?.(e);
      }

      // TODO antd没有暴露清除按钮的事件回调, 但清除时会触发onChange
      if (e.type === 'click' && !e.target.value) {
        setTimeout(() => {
          handleClear(e);
        }, 0);
      }
    },
    [debounce, onChange]
  );

  // 输入法开始
  const handleCompositionStart = () => {
    isCompos.current = true;
  };

  // 输入法结束
  const handleCompositionEnd = (e: any) => {
    isCompos.current = false;
    handleChange(e);
  };

  // 点击前缀搜索图标, 默认触发回车搜索
  const onPrefixClick = (e: any) => {
    onIconClick ? onIconClick(e) : onPressEnter?.(e);
  };

  // 处理清空输入框事件, 默认触发回车搜索
  const handleClear = (e: any) => {
    onClear ? onClear(e) : onPressEnter?.(e);
  };

  const iconConfig = {
    [iconPosition === 'start' ? 'prefix' : 'suffix']: (
      <IconFont type="icon-sousuo" className="s-input-icon" onClick={onPrefixClick} />
    ),
  };

  return (
    <Input
      value={inputValue}
      ref={inputRef}
      allowClear={allowClear}
      className={classNames('ad-search-input', className)}
      onChange={handleChange}
      onPressEnter={onPressEnter}
      onCompositionStart={handleCompositionStart}
      onCompositionEnd={handleCompositionEnd}
      {...iconConfig}
      {...otherProps}
    />
  );
};

const SearchInput = forwardRef(SearchInputFunc);
SearchInput.displayName = 'SearchInput';
export default SearchInput;
