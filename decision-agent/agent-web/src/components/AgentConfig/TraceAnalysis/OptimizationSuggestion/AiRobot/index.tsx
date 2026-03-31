import Lottie from 'lottie-react';
import robotJSON from './robot.json';

const AiRobot = () => {
  return <Lottie style={{ width: 48, height: 48 }} animationData={robotJSON} loop autoplay />;
};

export default AiRobot;
