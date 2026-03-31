import React, { useState } from 'react';
import { Button, Dropdown, Menu, Select, Tooltip } from 'antd';
import _ from 'lodash';
import classnames from 'classnames';
import intl from 'react-intl-universal';
import { CaretDownOutlined } from '@ant-design/icons';

import ContainerIsVisible from '@/components/ContainerIsVisible';
import Format from '@/components/Format';
import IconFont from '@/components/IconFont';
import SearchInput from '@/components/SearchInput';

import { ButtonConfigProps, HeaderProps, TitleProps } from '../types';
import './style.less';

const { Option } = Select;

const Title = (props: TitleProps) => {
  const { title, className, children } = props;
  return (
    <>
      <ContainerIsVisible isVisible={children || title}>
        <div data-testid={'testid-title'} className={classnames(className, 'ad-table-title')}>
          {children || title}
        </div>
      </ContainerIsVisible>
    </>
  );
};

const Header: React.FC<HeaderProps> = props => {
  const {
    title,
    className,
    children,
    visible = true,
    showFilter: isFilter_ = false,
    filterConfig,
    renderButtonConfig,
    onFilterClick = () => {},
    searchValue,
    onSearchChange = () => {},
    searchPlaceholder = '',
    filterToolsOptions = [],
    displayType = 'table',
    showSearch = true,
    onDisplayChange,
  } = props;

  const [__isFilter, __setIsFilter] = useState(isFilter_);
  const { isFilter, setIsFilter } = filterConfig || {
    isFilter: __isFilter,
    setIsFilter: __setIsFilter,
  };

  const leftButtonsList = renderButtonConfig?.filter(item => !!item && item.position === 'left') || [];
  const rightButtonsList = renderButtonConfig?.filter(item => !!item && item.position === 'right') || [];

  // 处理搜索
  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    _.debounce(query => {
      onSearchChange && onSearchChange(query);
    }, 300)(e.target.value);
  };

  const handleList = (list: ButtonConfigProps[]) => {
    return list.map((buttonItem, index) => {
      const {
        key,
        type,
        label,
        orderMenu = [],
        orderField,
        order,
        onOrderMenuClick = () => {},
        onHandle = () => {},
        itemDom = null,
        disabled,
      } = buttonItem;
      const menuRule = (
        <Menu selectedKeys={[orderField!]} className="ad-table-orderMenu" onClick={e => onOrderMenuClick?.(e)}>
          {_.map(orderMenu, (item: any) => {
            const { id, intlText } = item;
            const isSelected = id === orderField;
            const iconDirection = order === 'asc' ? '' : 'ad-table-orderMenu-icon-direction';
            return (
              <Menu.Item key={id}>
                <div className="ad-table-orderMenu-icon">
                  {isSelected && <IconFont type="icon-fanhuishangji" className={iconDirection} />}
                </div>
                <div>{intlText}</div>
              </Menu.Item>
            );
          })}
        </Menu>
      );

      if (itemDom) {
        return <div key={key ?? index}>{itemDom}</div>;
      }

      switch (type) {
        case 'add':
          return (
            <Format.Button key="add" type="primary" className="add-btn dip-mr-12" onClick={e => onHandle(e)}>
              <IconFont type="icon-Add" style={{ color: '#fff' }} />
              {label}
            </Format.Button>
          );
        case 'add-down':
          return (
            <ContainerIsVisible isVisible={true} key={key}>
              <Tooltip key="add-down" title={label} placement="bottom" visible={false}>
                <Dropdown overlay={menuRule} trigger={['click']} placement="bottomRight">
                  <Button type="primary" className="add-down-btn dip-mr-12" onClick={e => onHandle(e)}>
                    <IconFont type="icon-Add" style={{ color: '#fff' }} />
                    {label}
                    <CaretDownOutlined />
                  </Button>
                </Dropdown>
              </Tooltip>
            </ContainerIsVisible>
          );
        case 'delete':
          return (
            <ContainerIsVisible isVisible={true} key={key}>
              <Tooltip key="delete" title={label} placement="bottom" visible={false}>
                <Button disabled={disabled} className="delete-btn dip-mr-12" onClick={e => onHandle(e)}>
                  <IconFont type="icon-lajitong" />
                  {label}
                </Button>
              </Tooltip>
            </ContainerIsVisible>
          );
        case 'order':
          return (
            <Dropdown
              key="order"
              overlay={menuRule}
              trigger={['click']}
              placement="bottomRight"
              getPopupContainer={triggerNode => triggerNode.parentElement!}
            >
              <Format.Button type="icon" tip={intl.get('global.sort')} tipPosition="top">
                <IconFont type="icon-paixu11" />
              </Format.Button>
            </Dropdown>
          );
        case 'fresh':
          return (
            <Format.Button
              key="fresh"
              type="icon"
              tip={intl.get('global.refresh')}
              tipPosition="top"
              onClick={e => onHandle(e)}
            >
              <IconFont type="icon-tongyishuaxin" />
            </Format.Button>
          );
        case 'switch':
          return (
            <Format.Button
              key="switch"
              type="icon"
              tip={displayType === 'card' ? intl.get('graphList.switchList') : intl.get('graphList.switchCard')}
              tipPosition="top"
              onClick={() => {
                if (displayType === 'card') {
                  onDisplayChange?.('table');
                  onHandle?.('table');
                } else {
                  onDisplayChange?.('card');
                  onHandle?.('card');
                }
              }}
            >
              <IconFont type={`${displayType === 'card' ? 'icon-liebiao' : 'icon-wanggemoshi'}`} />
            </Format.Button>
          );
        case 'upload':
          return (
            <Format.Button key="add" type="primary" className="add-btn dip-mr-12" onClick={e => onHandle(e)}>
              <IconFont type="icon-shangchuan" style={{ color: '#fff' }} />
              {label}
            </Format.Button>
          );
        default:
          return <div key={key ?? index}>{itemDom}</div> || null;
      }
    });
  };

  return (
    <ContainerIsVisible isVisible={visible}>
      <div className={classnames('ad-table-header', className)}>
        {children || (
          <>
            {/* 标题 */}
            <Title>{title}</Title>
            {/* 左右按钮容器 */}
            {
              <div
                className="dip-mb-16"
                style={{ display: 'flex', justifyContent: 'space-between', alignContent: 'center' }}
              >
                <div className="left">{handleList(leftButtonsList)}</div>
                <div style={{ display: 'flex', alignItems: 'center' }}>
                  <ContainerIsVisible isVisible={!isFilter_}>
                    {showSearch && !isFilter_ ? (
                      <SearchInput
                        data-testid={'testid-searchInput'}
                        defaultValue={searchValue}
                        value={searchValue}
                        placeholder={searchPlaceholder || intl.get('global.search')}
                        className="dip-mr-12 search-input"
                        onChange={e => handleSearch(e)}
                        debounce
                      />
                    ) : null}

                    {filterToolsOptions.slice(0, 2).map(item => {
                      const { id, label, value, optionList = [], onHandle = () => {}, itemDom = null } = item;
                      return (
                        <div
                          key={id}
                          style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}
                          className="dip-mr-12"
                        >
                          {label ? (
                            <span className="dip-ellipsis" style={{ flexShrink: 0, marginRight: 10 }} title={label}>
                              {label}
                            </span>
                          ) : null}

                          {itemDom === null ? (
                            <Select
                              getPopupContainer={() => document.body}
                              onChange={value => {
                                onHandle(value);
                              }}
                              style={{ width: 190 }}
                              value={value}
                            >
                              {_.map(optionList, (item: any) => {
                                return (
                                  <Option key={item.key} value={item.value}>
                                    {item.text}
                                  </Option>
                                );
                              })}
                            </Select>
                          ) : (
                            itemDom
                          )}
                        </div>
                      );
                    })}
                  </ContainerIsVisible>
                  <ContainerIsVisible isVisible={isFilter_}>
                    <Format.Button
                      type="icon"
                      className="shaixuan-btn"
                      tip={intl.get('global.filter')}
                      tipPosition="top"
                      style={{ background: isFilter ? '#f5f5f5' : '' }}
                      onClick={e => {
                        setIsFilter(!isFilter);
                        onFilterClick(e, isFilter);
                      }}
                    >
                      <IconFont type="icon-shaixuan" />
                    </Format.Button>
                  </ContainerIsVisible>
                  {handleList(rightButtonsList)}
                </div>
              </div>
            }
          </>
        )}
      </div>
    </ContainerIsVisible>
  );
};

export default Header;
