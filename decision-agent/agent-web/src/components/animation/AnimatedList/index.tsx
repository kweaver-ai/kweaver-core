import React, { useEffect, useState } from 'react';
import { useTrail, animated, config } from '@react-spring/web';
import styles from './index.module.less';

interface AnimatedListProps {
  children: React.ReactNode[];
  delay?: number; // 元素之间的延迟时间
  staggerDelay?: number; // 整体动画开始的延迟
  className?: string;
  style?: React.CSSProperties;
}

const AnimatedList: React.FC<AnimatedListProps> = ({
  children,
  delay = 300,
  staggerDelay = 0,
  className = '',
  style = {},
}) => {
  const [show, setShow] = useState(false);

  useEffect(() => {
    // 组件挂载后开始动画
    const timer = setTimeout(() => {
      setShow(true);
    }, staggerDelay);

    return () => clearTimeout(timer);
  }, [staggerDelay]);

  const items = React.Children.toArray(children);

  const trail = useTrail(items.length, {
    config: { ...config.gentle, tension: 280, friction: 24 },
    from: {
      opacity: 0,
      transform: 'translateX(300px)',
    },
    to: {
      opacity: show ? 1 : 0,
      transform: show ? 'translateX(0px)' : 'translateX(300px)',
    },
    trail: delay, // 这将使每个元素之间有指定的延迟
  });

  return (
    <div className={`${styles.animatedList} ${className}`} style={style}>
      {trail.map((style, index) => (
        <animated.div key={index} style={style} className={styles.animatedItem}>
          {items[index]}
        </animated.div>
      ))}
    </div>
  );
};

export default AnimatedList;
