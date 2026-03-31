import React from 'react';
import { useSpring, animated } from '@react-spring/web';

interface FadeInFromLeftProps {
  children: React.ReactNode;
  delay?: number;
}

const FadeInFromLeft: React.FC<FadeInFromLeftProps> = ({ children, delay = 0 }) => {
  const springs = useSpring({
    from: {
      opacity: 0,
      transform: 'translateX(-50px)',
    },
    to: {
      opacity: 1,
      transform: 'translateX(0px)',
    },
    delay,
    config: {
      tension: 280,
      friction: 20,
    },
  });

  return <animated.div style={springs}>{children}</animated.div>;
};

export default FadeInFromLeft;
