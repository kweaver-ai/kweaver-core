import { memo } from 'react';
import intl from 'react-intl-universal';
import classNames from 'classnames';
import { Tooltip } from 'antd';
import styles from './ModuleCard.module.less';

interface Props {
  width: number;
  height: number;
  moduleKey: string;
  name: string;
  enabled: boolean;
  onSelect: (key: string) => void;
}

const ModuleCard = memo(({ width, height, moduleKey, name, enabled, onSelect }: Props) => {
  return (
    <div
      className={classNames(
        'dip-pl-8 dip-border-radius-4 dip-flex-align-center dip-gap-12 dip-flex-shrink-0 dip-pointer',
        styles['container']
      )}
      style={{ width, height }}
      onClick={() => onSelect(moduleKey)}
    >
      <span
        className={classNames(
          'dip-border-radius-full dip-position-r dip-flex-shrink-0',
          styles['outer-circle'],
          enabled ? styles['enabled-outer-circle'] : styles['disabled-outer-circle']
        )}
      >
        <span className={classNames('dip-border-radius-full dip-position-center', styles['inner-circle'])} />
      </span>

      <span className="dip-flex-column dip-gap-5 dip-pt-5 dip-overflow-hidden">
        <Tooltip title={name}>
          <span className="dip-line-height-20 dip-c-black dip-ellipsis">{name}</span>
        </Tooltip>
        <span
          className={classNames(
            'dip-line-height-20 dip-font-12',
            enabled ? styles['enabled-color'] : styles['disabled-color']
          )}
        >
          {enabled ? intl.get('dataAgent.enabled') : intl.get('dataAgent.disabled')}
        </span>
      </span>
    </div>
  );
});

export default ModuleCard;
