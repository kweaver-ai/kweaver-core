import React from 'react';
import { useSpring, animated } from '@react-spring/web';

interface FadeInFromBottomProps {
  children: React.ReactNode;
  delay?: number;
}

const FadeInFromBottom: React.FC<FadeInFromBottomProps> = ({ children, delay = 0 }) => {
  const springs = useSpring({
    from: {
      opacity: 0,
      transform: 'translateY(50px)',
    },
    to: {
      opacity: 1,
      transform: 'translateY(0px)',
    },
    delay,
    config: {
      tension: 280,
      friction: 20,
    },
  });

  return <animated.div style={springs}>{children}</animated.div>;
};

export default FadeInFromBottom;
