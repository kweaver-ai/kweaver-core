import React from 'react';
import classnames from 'classnames';
import { Button as AntdButton, Tooltip } from 'antd';

import { ButtonType } from '../type';

import './style.less';
import './index.less';

const SIZE = {
  large: 'ad-format-button-large',
  middle: 'ad-format-button-middle',
  small: 'ad-format-button-small',
  smallest: 'ad-format-button-smallest',
};

const SIZE_ICON = {
  middle: 'ad-format-icon-middle',
  small: 'ad-format-icon-small',
};

const Button = (props: ButtonType) => {
  const { size = 'middle', children, className, type, tip = '', tipPosition = 'bottom', ...othersProps } = props;
  const extendSize = SIZE[size as keyof typeof SIZE] || '';
  const extendSize_Icon = SIZE_ICON[size as keyof typeof SIZE_ICON] || '';

  const prefixCls = 'ad-btn';

  const renderButton = () => {
    // 不带框纯文本---无背景色，无padding
    if (type === 'text') {
      return (
        <Tooltip title={tip} placement={tipPosition} trigger={['hover']}>
          <span
            {...othersProps}
            className={classnames(prefixCls, `${prefixCls}-text`, className, {
              [`${prefixCls}-disabled`]: othersProps.disabled,
            })}
          >
            {children}
          </span>
        </Tooltip>
      );
    }

    // 不带框纯文本---有背景色，有padding
    if (type === 'text-b') {
      return (
        <Tooltip title={tip} placement={tipPosition} trigger={['hover']}>
          <AntdButton {...othersProps} className={classnames('ad-btn-text-b', className)} type="text">
            {children}
          </AntdButton>
        </Tooltip>
      );
    }

    // 不带框文本link---无背景色，无padding，字体主题色
    if (type === 'link') {
      return (
        <Tooltip title={tip} placement={tipPosition} trigger={['hover']}>
          <AntdButton {...othersProps} className={classnames('ad-btn-link', className)} type="link">
            {children}
          </AntdButton>
        </Tooltip>
      );
    }

    // 不带框icon
    if (type === 'icon') {
      return (
        <Tooltip title={tip} placement={tipPosition} trigger={['hover']} arrow={{ pointAtCenter: true }}>
          <span
            {...othersProps}
            className={classnames(className, extendSize_Icon, 'ad-btn-icon', {
              'ad-btn-icon-disabled': othersProps.disabled,
            })}
            onClick={e => {
              if (!othersProps.disabled) {
                othersProps.onClick && othersProps.onClick(e);
              }
            }}
          >
            {children}
          </span>
        </Tooltip>
      );
    }

    // icon + 文字 + 无框 + 带背景
    if (type === 'icon-text') {
      const childrenList = React.Children.toArray(children);
      const flag_1 = childrenList.length === 2 && typeof childrenList[1] === 'string';
      const flag_2 = childrenList.length === 2 && typeof childrenList[1] === 'object';
      const flag_3 =
        childrenList.length === 3 && typeof childrenList[0] === 'object' && typeof childrenList[2] === 'object';

      return (
        <Tooltip title={tip} placement={tipPosition} trigger={['hover']}>
          <span
            {...othersProps}
            className={classnames(
              { 'ad-btn-icon-text': flag_1 },
              { 'ad-btn-text-icon': flag_2 },
              { 'ad-btn-icon-text-icon': flag_3 },
              className,
              {
                'ad-btn-icon-text-disabled': othersProps.disabled,
              }
            )}
            onClick={e => {
              if (!othersProps.disabled) {
                othersProps.onClick?.(e);
              }
            }}
          >
            {children}
          </span>
        </Tooltip>
      );
    }

    // icon + 文字 + 无框 + 不带背景
    if (type === 'icon-text-link') {
      const childrenList = React.Children.toArray(children);
      const flag_1 = childrenList.length === 2 && typeof childrenList[1] === 'string';
      const flag_2 = childrenList.length === 2 && typeof childrenList[1] === 'object';
      const flag_3 =
        childrenList.length === 3 && typeof childrenList[0] === 'object' && typeof childrenList[2] === 'object';

      return (
        <Tooltip title={tip} placement={tipPosition} trigger={['hover']}>
          <span
            {...othersProps}
            className={classnames(
              { 'ad-btn-icon-text-link': flag_1 },
              { 'ad-btn-text-icon-link': flag_2 },
              { 'ad-btn-icon-text-icon-link': flag_3 },
              className,
              {
                'ad-btn-icon-disabled': othersProps.disabled,
              }
            )}
            onClick={e => {
              if (!othersProps.disabled) {
                othersProps.onClick?.(e);
              }
            }}
          >
            {children}
          </span>
        </Tooltip>
      );
    }

    // 带框icon
    if (type === 'u-icon') {
      return (
        <Tooltip title={tip} placement={tipPosition} trigger={['hover']}>
          <AntdButton
            className={classnames(className, 'ad-btn-u-icon', {
              'ad-btn-u-icon-disabled': othersProps.disabled,
            })}
            {...othersProps}
            type={type as any}
          >
            {children}
          </AntdButton>
        </Tooltip>
      );
    }
    if (type === 'u-icon-primary') {
      return (
        <Tooltip title={tip} placement={tipPosition} trigger={['hover']}>
          <AntdButton
            className={classnames(className, 'ad-btn-u-icon', {
              'ad-btn-u-icon-disabled': othersProps.disabled,
            })}
            {...othersProps}
            type={type.slice(-7) as any}
          >
            {children}
          </AntdButton>
        </Tooltip>
      );
    }

    // icon + 文字 + 框 / 带框文本
    return (
      <Tooltip title={tip} placement={tipPosition} trigger={['hover']}>
        <AntdButton
          className={classnames(extendSize, className, 'ad-format-button')}
          {...othersProps}
          type={type as any}
        >
          {children}
        </AntdButton>
      </Tooltip>
    );
  };
  return renderButton();
};

export default Button;
