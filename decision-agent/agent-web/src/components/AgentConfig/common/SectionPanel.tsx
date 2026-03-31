import React from 'react';
import classNames from 'classnames';
import CollapseArrow from '@/assets/icons/collapse-arrow.svg';
import AgentIcon from '@/assets/icons/agent-icon.svg';
import styles from './SectionPanel.module.css';

interface SectionPanelProps {
  title: React.ReactNode;
  description?: string;
  children?: React.ReactNode;
  rightElement?: React.ReactNode;
  isExpanded?: boolean;
  className?: string;
  noContent?: boolean;
  onToggle?: () => void;
  icon?: React.ReactNode;
  showCollapseArrow?: boolean;
}

const SectionPanel: React.FC<SectionPanelProps> = ({
  title,
  description,
  children,
  rightElement,
  isExpanded = false,
  className = '',
  noContent = false,
  onToggle,
  icon,
  showCollapseArrow = true,
}) => {
  const handleToggle = () => {
    if (!showCollapseArrow) return; // 如果不显示折叠箭头，不触发折叠功能

    if (onToggle) {
      onToggle();
    }
  };

  return (
    <div className={`dip-p-24 ${className}`}>
      <div
        className={classNames(styles['section-panel-header'], {
          'dip-pointer': showCollapseArrow,
        })}
        onClick={handleToggle}
      >
        <div>
          <div className="dip-flex-align-center" style={{ gap: '8px' }}>
            <div className={`${styles['section-panel-arrow']} ${isExpanded ? '' : styles.collapsed}`}>
              {showCollapseArrow && <CollapseArrow />}
            </div>
            {icon ? icon : <AgentIcon className={styles['section-panel-icon']} width="20" height="20" />}
            <div className="dip-text-color-85 dip-font-16 dip-flex-align-center" style={{ gap: '8px' }}>
              {title}
            </div>
          </div>
          {description && <div className={'dip-mt-8 dip-ml-24 dip-text-color-25'}>{description}</div>}
        </div>
        {rightElement && (
          <div className={styles['section-panel-right']} onClick={e => e.stopPropagation()}>
            {rightElement}
          </div>
        )}
      </div>

      <div
        className={`${styles['section-panel-content']} ${isExpanded ? '' : styles.collapsed}`}
        style={{ display: !noContent && children ? 'block' : 'none' }}
      >
        {children}
      </div>
    </div>
  );
};

export default SectionPanel;
