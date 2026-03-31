import classNames from 'classnames';
import React, { type CSSProperties, type ReactNode } from 'react';
import styles from './index.module.less';
import bg from '@/assets/images/gradient-container-bg.png';
interface GradientContainerProps {
  children: ReactNode;
  className?: string;
  style?: CSSProperties;
}
const GradientContainer: React.FC<GradientContainerProps> = ({ children, className, style }) => {
  return (
    <div
      style={{
        backgroundImage: `url(${bg})`,
        ...style,
      }}
      className={classNames(styles['gradient-container'], className)}
    >
      {children}
    </div>
  );
};

export default GradientContainer;
