import React from 'react';
import _ from 'lodash';
import classnames from 'classnames';
import { Button, ConfigProvider } from 'antd';
import './style.less';

export interface FooterModalProps {
  children?: React.ReactNode;
  className?: string;
  source: FooterModalSourceType[] | React.ReactNode;
  footerExtra?: React.ReactNode;
}

export interface FooterModalSourceType {
  label: React.ReactNode;
  type?: 'link' | 'text' | 'default' | 'ghost' | 'primary' | 'dashed' | undefined;
  style?: string;
  isDisabled?: boolean;
  btnProps?: Record<string, any>;
  onHandle: () => void;
}

const Footer = (props: FooterModalProps) => {
  const { className, source, footerExtra } = props;
  return (
    <>
      {source && (
        <div className={classnames('ad-universal-modal-footer', className)}>
          {Array.isArray(source) && (source as FooterModalSourceType[]).length > 0 ? (
            <div className="dip-flex-space-between">
              <div>{footerExtra}</div>

              <div>
                <ConfigProvider autoInsertSpaceInButton={false}>
                  {_.map(source as FooterModalSourceType[], (item, index: number) => {
                    const { label = '', type = 'default', style = '', isDisabled = false, btnProps = {} } = item;
                    const { onHandle } = item;
                    return (
                      <Button
                        key={index}
                        {...btnProps}
                        type={type}
                        className={style !== '' ? style : 'dip-ml-8'}
                        onClick={onHandle}
                        disabled={isDisabled}
                      >
                        {label}
                      </Button>
                    );
                  })}
                </ConfigProvider>
              </div>
            </div>
          ) : (
            <>{source}</>
          )}
        </div>
      )}
    </>
  );
};

export default (props: any) => {
  const { visible = true, ...other } = props;
  if (!visible) return null;
  return <Footer {...other} />;
};
