import { useMemo, type FC } from 'react';
import { useNavigate } from 'react-router-dom';
import classNames from 'classnames';
import { Button } from 'antd';
import intl from 'react-intl-universal';
import PlusIcon from '@/assets/icons/plus.svg';
import { getHeaderText } from './utils';
import { ModeEnum } from './types';
import styles from './index.module.less';

interface HeaderProps {
  mode: ModeEnum; // 当前的模式
  isExportMode: boolean; // 是否是导出模式
  onCreate: () => void; // 创建新的agent
}

const Header: FC<HeaderProps> = ({ mode, isExportMode, onCreate }) => {
  const navigate = useNavigate();

  // 顶部的text
  const headerText = useMemo(
    () => getHeaderText(mode, { customSpaceName: '', navigateToSpaces: () => navigate('/') }),
    [mode]
  );
  const isMine = useMemo(() => [ModeEnum.MyAgent, ModeEnum.MyTemplate].includes(mode), [mode]);

  return (
    <div
      className="dip-pt-24 dip-pr-16 dip-pl-16 dip-flex-space-between"
      // style={{ paddingBottom: isMine ? '17px' : '24px' }}
    >
      {Boolean(headerText) && (
        <div
          className={classNames(styles.sectionTitle, 'dip-font-16 dip-c-black, dip-w-100')}
          style={isMine ? { height: '32px', lineHeight: '32px' } : {}} // 固定高度，防止切换tab导致页面高度变化
        >
          {headerText}
        </div>
      )}

      {mode === ModeEnum.MyAgent && (
        <Button type="primary" disabled={isExportMode} onClick={onCreate}>
          <PlusIcon />
          <span style={{ color: 'white' }}>{intl.get('dataAgent.createNew')}</span>
        </Button>
      )}
    </div>
  );
};

export default Header;
