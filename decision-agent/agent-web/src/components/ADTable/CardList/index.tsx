import { type CSSProperties, useMemo, useRef, useState } from 'react';
import _ from 'lodash';
import classNames from 'classnames';
import intl from 'react-intl-universal';
import { Divider, Dropdown, Menu, Pagination, Typography } from 'antd';
import { EllipsisOutlined } from '@ant-design/icons';

import { useDeepCompareEffect } from '@/hooks';

import Format from '@/components/Format';
import IconFont from '@/components/IconFont';
import LoadingMask from '@/components/LoadingMask';
import AdResizeObserver from '@/components/AdResizeObserver/AdResizeObserver';

import type { ADTableProps, CardItemMenuItemType, CardItemType, CardProps } from '../types';

import EmptyIcon from '@/assets/icons/empty.svg';
import NoResultIcon from '@/assets/icons/no-result.svg';
import './style.less';
import dayjs from 'dayjs';

const { Paragraph } = Typography;
const AdTableCardList = (props: ADTableProps & { isFilter: boolean }) => {
  const prefixCls = 'AdTableCardList';
  const {
    pagination,
    dataSource,
    card = {},
    emptyText,
    emptyImage,
    loading,
    rowKey,
    filterToolsOptions,
    searchValue,
    activeRowKey,
    isFilter,
  } = props;
  const newCard: CardProps = { rowCardNum: 4, gap: 24, cardHeight: 154, renderCardItem: () => '', ...card };
  const { rowCardNum, cardHeight, gap, renderCardItem, onClick } = newCard;
  const containerWidth = useRef<number>(0);
  const scrollIntoViewAble = useRef(false);
  const [cardWidth, setCardWidth] = useState<number>(0);
  useDeepCompareEffect(() => {
    if (containerWidth.current > 0) {
      handleStyle(containerWidth.current);
    }
  }, [dataSource, rowCardNum, gap]);

  /**
   * 处理样式
   */
  const handleStyle = (wrapperWidth: number) => {
    const cardItemDomList = document.querySelectorAll(`.${prefixCls}-item`) as any;
    if (rowCardNum && gap && cardItemDomList.length > 0) {
      // 默认值全是1024以上的情况
      let width = 0; // 卡片宽度
      let cardNum = rowCardNum; // 卡片数量
      if (wrapperWidth < 1024) {
        if (wrapperWidth > 768) {
          cardNum -= 1;
        } else {
          cardNum -= 2;
        }
      }
      const gapNum = cardNum - 1; // 间隙数量
      const gapTotalWidth = gapNum * gap; // 间隙总宽度
      width = (wrapperWidth - gapTotalWidth) / cardNum;
      cardItemDomList.forEach((domItem: HTMLDivElement) => {
        domItem.style.width = `${width}px`;
      });
      setCardWidth(width);
      handleActiveRowScrollIntoView();
    }
  };

  const handleActiveRowScrollIntoView = () => {
    const selectedCardDOM = document.querySelector(`.${prefixCls}-item-active`);
    if (selectedCardDOM && !scrollIntoViewAble.current) {
      selectedCardDOM.scrollIntoView();
      scrollIntoViewAble.current = true;
    }
  };

  const renderMenu = (record: any, menu: CardItemMenuItemType[]) => {
    const menuData = menu.filter(item => {
      if (item.visible) {
        return item.visible(record);
      }
      return true;
    });
    const getDisabled = (disabled: any) => {
      if (disabled) {
        return disabled(record);
      }
      return false;
    };
    return (
      <Dropdown
        getPopupContainer={node => node.parentElement!}
        trigger={['click']}
        overlay={
          <Menu style={{ minWidth: 120 }}>
            {menuData.map((menuItem, index) => (
              <Menu.Item
                disabled={getDisabled(menuItem.disabled)}
                key={index}
                onClick={({ domEvent }) => {
                  domEvent.stopPropagation();
                  menuItem.onClick?.(record);
                }}
              >
                {typeof menuItem.label === 'function' ? menuItem.label(record) : menuItem.label}
              </Menu.Item>
            ))}
          </Menu>
        }
      >
        <Format.Button size="small" onClick={event => event.stopPropagation()} className="ad-table-operate" type="icon">
          <EllipsisOutlined style={{ fontSize: 20 }} />
        </Format.Button>
      </Dropdown>
    );
  };
  /**
   * 获取卡片的DOM
   * @param record
   * @param index
   */
  const getCardItemDom = (record: any, index: number, cls: string) => {
    if (typeof renderCardItem === 'function') {
      return renderCardItem(record, index);
    }
    const { header, body, footer } = renderCardItem as CardItemType;
    let headerIconField;
    let headerTitleField;
    let headerMenu;
    header?.forEach(item => {
      if (item.type === 'icon') {
        headerIconField = item.field;
      }
      if (item.type === 'title') {
        headerTitleField = item.field;
      }
      if (item.type === 'menu') {
        headerMenu = item.menu;
      }
    });
    return (
      <div className={classNames('dip-w-100 dip-h-100 dip-flex-column')}>
        <div className={`${cls}-header dip-flex-align-center dip-mb-16`}>
          <span
            className="dip-flex-align-center dip-flex-item-full-width dip-pointer"
            onClick={e => {
              e.stopPropagation();
              onClick?.(record);
            }}
          >
            {headerIconField && <IconFont type={record[headerIconField]} border />}
            {headerTitleField && (
              <span
                title={record[headerTitleField]}
                style={{ color: '#000' }}
                className="dip-ml-8 dip-c-header dip-flex-item-full-width dip-ellipsis"
              >
                {record[headerTitleField]}
              </span>
            )}
          </span>
          {headerMenu && (
            <span onClick={e => e.stopPropagation()} className={`${cls}-header-operate`}>
              {renderMenu(record, headerMenu)}
            </span>
          )}
        </div>
        <div className={`${cls}-body dip-flex-item-full-height dip-mb-20`}>
          {body?.map((bodyItem, index) => {
            let dom;
            switch (bodyItem.type) {
              case 'multiText':
                let textStr: any;
                if (bodyItem.render) {
                  textStr = bodyItem.render(record);
                } else if (bodyItem.field) {
                  textStr = record[bodyItem.field];
                }
                dom = (
                  <Paragraph
                    style={{
                      color: `rgba(0,0,0,0.${textStr ? 6 : 2}5)`,
                      marginBottom: 0,
                      lineHeight: '18px',
                      minHeight: 36,
                    }}
                    ellipsis={{ rows: 2 }}
                    title={textStr}
                  >
                    {textStr || intl.get('graphList.notDes')}
                  </Paragraph>
                );
                break;
              case 'text':
                if (bodyItem.render) {
                  dom = bodyItem.render(record);
                } else if (bodyItem.field) {
                  dom = (
                    <span className="dip-ellipsis" title={record[bodyItem.field]}>
                      {record[bodyItem.field]}
                    </span>
                  );
                }
                break;
              default:
                dom = '--';
                break;
            }
            return (
              <span
                className="dip-pointer"
                key={index}
                onClick={e => {
                  e.stopPropagation();
                  onClick?.(record);
                }}
              >
                {dom}
              </span>
            );
          })}
        </div>
        {footer && (
          <div className={`${cls}-footer dip-flex-align-center dip-ellipsis`}>
            {footer.map((footerItem, index) => {
              const isLast = index === footer.length - 1;
              let dom;
              switch (footerItem.type) {
                case 'time':
                  let format = 'YYYY-MM-DD HH:mm:ss';
                  if (cardWidth < 280 && footer.length > 2) {
                    format = 'YYYY-MM-DD';
                  }
                  const timeValue = footerItem.field && dayjs(record[footerItem.field]).format(format);
                  dom = timeValue && (
                    <span title={timeValue}>
                      <IconFont type="icon-gengxinshijian" className="dip-mr-4" />
                      {timeValue}
                    </span>
                  );
                  break;
                case 'text':
                  if (footerItem.render) {
                    dom = footerItem.render(record);
                  } else if (footerItem.field) {
                    dom = (
                      <span className="dip-ellipsis" title={record[footerItem.field]}>
                        {record[footerItem.field]}
                      </span>
                    );
                  }
                  break;
                default:
                  dom = '--';
                  break;
              }
              return (
                <span key={index} className={classNames('dip-flex-align-center')} style={{ display: 'inline-flex' }}>
                  {dom}
                  {!isLast && <Divider type="vertical" />}
                </span>
              );
            })}
          </div>
        )}
      </div>
    );
  };

  const getLoading = () => {
    if (typeof loading === 'boolean') {
      return loading;
    }
    return loading?.spinning ?? false;
  };

  /** 获取缺醒图 */
  const getLocale = () => {
    // loading的时候  不显示缺省图
    if (getLoading()) {
      return undefined;
    }
    // dataSource为空，但是存在搜索或者过滤条件，显示空数据
    const filterValue = filterToolsOptions?.find(item => !!item.value && item.value !== 'all');
    let showEmptyData = false;
    if (searchValue || filterValue) {
      showEmptyData = true;
    }
    if (showEmptyData) {
      return (
        <div className={`${prefixCls}-nodata-box`}>
          <NoResultIcon />
          <div className={`${prefixCls}-nodata-text`}>{intl.get('global.noResult')}</div>
        </div>
      );
    }
    return (
      <div className={`${prefixCls}-nodata-box`}>
        {emptyImage ? <img src={emptyImage} alt="nodata" /> : <EmptyIcon />}
        <div className={`${prefixCls}-nodata-text`}>{emptyText || intl.get('global.noResult')}</div>
      </div>
    );
  };

  const paginationConfig = useMemo(() => {
    if (pagination) {
      return {
        className: `${prefixCls}-pagination ant-table-pagination ant-table-pagination-right`,
        showTotal: (total: number) => intl.get('knowledge.total', { total }),
        showSizeChanger: false,
        ...pagination,
      };
    }
    return false;
  }, [pagination]);

  const isSinglePage = useMemo(() => {
    if (pagination) {
      const total = pagination.total;
      const pageSize = pagination.pageSize;
      if (total && pageSize) {
        return total < pageSize;
      }
    }
    return true;
  }, [pagination]);

  const onResize = _.debounce(({ dom }) => {
    const width = dom.clientWidth;
    containerWidth.current = width;
    handleStyle(width);
  }, 100);

  const contentDom = () => {
    return (
      <AdResizeObserver onResize={onResize}>
        <div id={prefixCls} className={classNames(`${prefixCls} dip-flex-item-full-height`)} style={{ gap }}>
          <LoadingMask loading={getLoading()} />
          {dataSource?.length
            ? dataSource.map((record, index) => {
                const cardStyle: CSSProperties = { height: cardHeight };
                let key = record.key;
                if (rowKey) {
                  if (typeof rowKey === 'string') {
                    key = record[rowKey];
                  } else {
                    key = rowKey(record);
                  }
                }
                return (
                  <div
                    style={cardStyle}
                    key={key}
                    className={classNames(`${prefixCls}-item`, {
                      [`${prefixCls}-item-active`]:
                        activeRowKey !== undefined && key !== undefined && activeRowKey === key,
                    })}
                  >
                    {getCardItemDom(record, index, `${prefixCls}-item`)}
                  </div>
                );
              })
            : getLocale()}
        </div>
      </AdResizeObserver>
    );
  };

  return pagination && !isSinglePage && dataSource?.length ? (
    <div className="dip-flex-column dip-flex-item-full-height">
      <div className="dip-flex-column" style={{ maxHeight: '100%' }}>
        {contentDom()}
        <Pagination {...paginationConfig} />
      </div>
    </div>
  ) : (
    contentDom()
  );
};

export default AdTableCardList;
