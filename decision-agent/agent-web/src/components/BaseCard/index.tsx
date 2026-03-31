import { memo, useState, useEffect, useMemo, useRef } from 'react';
import intl from 'react-intl-universal';
import { isEqual } from 'lodash';
import classNames from 'classnames';
import { Card, Dropdown, Tooltip, Checkbox, Popover } from 'antd';
import { EllipsisOutlined, CloseOutlined } from '@ant-design/icons';
import UserAvatar from './UserAvatar';
import { computeColumnCount, minCardWidth, maxCardWidth, gap, rowHeight, loadingMoreRowHeight } from './utils';
import styles from './index.module.less';
import dayjs from 'dayjs';
import AgentIcon from '@/components/AgentIcon';

export { computeColumnCount, minCardWidth, maxCardWidth, gap, rowHeight, loadingMoreRowHeight };

interface BaseCardProps {
  className?: string;
  bordered?: boolean;
  closable?: boolean;
  hoverable?: boolean;
  // 是否显示复选框
  checkable?: boolean;
  // 是否勾选
  checked?: boolean;
  checkboxDisabled?: boolean;
  // 是否高亮
  isHighlighted?: boolean;
  item: any;
  name: string;
  // name后面紧跟的icon
  getNameSuffixIcon?: (id: string) => React.ReactElement | null;
  getNameSuffixStatus?: (status: string) => React.ReactElement | undefined;
  profile: string;
  // 当userName 和 userAvatar都不存在时，不显示头像
  userName?: string;
  userAvatar?: any;
  // 时间显示，比如：创建时间：2025/06/25 16:15:47
  time: string;
  // 当menuItems不存在或者长度为0时，不显示【...】
  menuItems?: any[];
  // 点击menu
  onClickMenu?: ({ key }: { key: string }, item: any) => void;
  // 点击卡片
  onClick?: (item: any) => void;
  // 关闭
  onClose?: () => void;
  // checked变化
  onCheckedChange?: (checked: boolean, item: any) => void;
}

const BaseCard = ({
  className,
  bordered = true,
  closable = false,
  hoverable = false,
  checkable = false,
  checked = false,
  checkboxDisabled = false,
  isHighlighted = false,
  item: itemFromProps,
  name,
  getNameSuffixIcon,
  getNameSuffixStatus,
  profile,
  userAvatar,
  userName,
  time,
  menuItems: menuItemsFromProps,
  onClickMenu,
  onClick,
  onClose,
  onCheckedChange,
}: BaseCardProps) => {
  const prevMenuItemsRef = useRef(menuItemsFromProps);
  const itemRef = useRef(itemFromProps);

  const [dropdownOpen, setDropdownOpen] = useState<boolean>(false);

  const menuItems = useMemo(() => {
    if (isEqual(prevMenuItemsRef.current, menuItemsFromProps)) {
      return prevMenuItemsRef.current; // 内容相同，返回旧引用
    }
    return menuItemsFromProps; // 内容不同，返回新数据
  }, [menuItemsFromProps]);

  const item = useMemo(() => {
    if (isEqual(itemRef.current, itemFromProps)) {
      return itemRef.current; // 内容相同，返回旧引用
    }
    return itemFromProps; // 内容不同，返回新数据
  }, [itemFromProps]);

  const nameSuffixIcon = useMemo(() => getNameSuffixIcon?.(item?.id), [getNameSuffixIcon, item?.id]);
  const nameSuffixStatus = useMemo(() => getNameSuffixStatus?.(item?.status), [getNameSuffixStatus, item?.status]);
  itemRef.current = item;
  // 确保 ref 和当前值同步
  useEffect(() => {
    prevMenuItemsRef.current = menuItems;
  }, [menuItems]);

  const getCardIcon = (agent: any) => (
    <AgentIcon
      size={48}
      avatar_type={agent?.avatar_type}
      avatar={agent?.avatar}
      name={agent?.name}
      style={{ minWidth: '48px' }}
      showBuildInLogo={agent?.is_built_in === 1}
      showSystemLogo={agent?.is_system_agent === 1}
    />
  );

  return (
    <Card
      className={classNames(className, styles['card'], {
        [styles['highlight']]: isHighlighted,
        [styles['hoverable']]: hoverable,
      })}
      hoverable={hoverable}
      styles={{
        body: {
          padding: '24px',
        },
      }}
      variant={bordered ? 'outlined' : 'borderless'}
      onClick={() => {
        checkable && !checkboxDisabled ? onCheckedChange?.(!checked, item) : onClick?.(item);
      }}
    >
      <div className="dip-flex dip-gap-16">
        {/* 复选框 */}
        {checkable && (
          <Checkbox
            className={classNames(styles['checkbox'], 'dip-mt-2')}
            checked={checked}
            disabled={checkboxDisabled}
            onChange={e => onCheckedChange?.(e.target.checked, item)}
            onClick={e => e.stopPropagation()}
          />
        )}

        <div className="dip-flex-column dip-gap-24 dip-overflow-hidden" style={{ flex: 1 }}>
          {/* 卡片上方 */}
          <div className="dip-flex dip-gap-16">
            {/* 卡片Icon */}
            {getCardIcon(item)}

            <div className="dip-flex-item-full-width">
              <div className="dip-flex-space-between">
                <div className="dip-flex dip-overflow-hidden">
                  {/* 卡片名称 */}
                  <Tooltip title={name}>
                    <div className="dip-ellipsis dip-font-16 dip-text-color-85">{name}</div>
                  </Tooltip>
                  {nameSuffixIcon ? <span className="dip-flex-shrink-0">{nameSuffixIcon}</span> : null}
                </div>
                {nameSuffixStatus ? (
                  <Popover
                    open={item?.status === 'published_edited' ? undefined : false}
                    placement="right"
                    content={
                      <div className="dip-p-12">
                        <div>
                          <span>上一版本：</span>
                          <span>{item?.version}</span>
                        </div>
                        <div>
                          <span>{intl.get('dataAgent.publishTime')}</span>
                          <span>{dayjs(item?.published_at).format('YYYY/MM/DD HH:mm:ss')}</span>
                        </div>
                      </div>
                    }
                  >
                    <span className="dip-flex-shrink-0">{nameSuffixStatus}</span>
                  </Popover>
                ) : null}
              </div>
              {/* 卡片描述 */}
              <Tooltip title={profile}>
                <div
                  className={classNames(
                    styles['min-height-56'],
                    'dip-ellipsis-2 dip-font-13 dip-text-color-45 dip-pt-12'
                  )}
                  style={{ width: 'fit-content' }}
                >
                  {profile}
                </div>
              </Tooltip>
            </div>
          </div>

          {/* 卡片底部 */}
          <div className={classNames(styles['card-footer'], 'dip-flex-space-between')}>
            <div
              style={{
                flex: 1,
                gap: 12,
              }}
              className={'dip-flex-align-center dip-overflow-hidden'}
            >
              <div className="dip-overflow-hidden dip-flex-align-center">
                {/* 用户头像 */}
                {Boolean(userAvatar || userName) && <UserAvatar src={userAvatar} userName={userName} />}

                {/* 用户名称 */}
                {Boolean(userName) && (
                  <Tooltip title={userName}>
                    <div className={classNames(styles['footer-text-color'], 'dip-ellipsis dip-font-12')}>
                      {userName || '---'}
                    </div>
                  </Tooltip>
                )}
              </div>

              {/* 时间 */}
              <Tooltip title={time}>
                <div className={classNames(styles['footer-text-color'], 'dip-ellipsis dip-font-12')}>{time}</div>
              </Tooltip>
            </div>

            {/* 操作区 */}
            {(menuItems || []).length > 0 && (
              <div
                className={classNames(styles['actions'], 'dip-ml-12 dip-display-none')}
                style={
                  dropdownOpen
                    ? {
                        display: 'unset !important',
                      }
                    : {}
                }
                onClick={e => e.stopPropagation()}
              >
                <Dropdown
                  menu={{
                    items: menuItems,
                    onClick: info => onClickMenu?.(info, item),
                  }}
                  placement="bottomRight"
                  onOpenChange={setDropdownOpen}
                >
                  <EllipsisOutlined className={classNames(styles['more-icon'], 'dip-font-16 dip-p-8 dip-pointer')} />
                </Dropdown>
              </div>
            )}
          </div>
        </div>

        {/* 关闭 */}
        {closable && (
          <Tooltip title={intl.get('dataAgent.removeAgent')}>
            <CloseOutlined className={classNames(styles['close-icon'], 'dip-text-color-65')} onClick={onClose} />
          </Tooltip>
        )}
      </div>
    </Card>
  );
};

export default memo(BaseCard);
