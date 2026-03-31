import React from 'react';
import { useSpring, animated } from '@react-spring/web';
import styles from './index.module.less';

interface GradientBorderProps {
  children: React.ReactNode;
  borderWidth?: number;
  className?: string;
  style?: React.CSSProperties;
  duration?: number; // 单位：秒
}

const GradientBorder: React.FC<GradientBorderProps> = props => {
  const { children, borderWidth = 1, className = '', style = {}, duration = 8 } = props;
  // 创建一个闪烁的动画效果
  const animation = useSpring({
    from: { backgroundPosition: '0% 0%' },
    to: { backgroundPosition: '200% 0%' },
    config: {
      duration: duration * 1000,
    },
    loop: true,
  });

  return (
    <animated.div
      className={`${styles.gradientBorderContainer} ${className}`}
      style={{
        ...style,
        padding: borderWidth,
        ...animation,
      }}
    >
      <div className={styles.content}>{children}</div>
    </animated.div>
  );
};

export default GradientBorder;
