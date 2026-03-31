import styles from './index.module.less';
import classNames from 'classnames';
import React from 'react';
const CitePanel = ({ cite }: any) => {
  const renderContent = () => {
    return (
      <div
        className={classNames(styles.container)}
        onClick={() => {
          window.open(cite.link);
        }}
      >
        <div className="dip-flex-align-center">
          {cite.icon && <img className={classNames(styles.icon, 'dip-mr-8')} src={cite.icon} alt="" />}
          <span className={classNames(styles.link, 'dip-flex-item-full-width dip-ellipsis')} title={cite.title}>
            {cite.title}
          </span>
        </div>
        <div className="dip-mt-8 dip-text-color-45 dip-ellipsis-2">{cite.content}</div>
      </div>
    );
  };
  return <>{renderContent()}</>;
};

export default CitePanel;
