import Lottie from 'lottie-react';
import loadingJSON from './loading.json';

const Loading = () => {
  return <Lottie style={{ width: 120, height: 120 }} animationData={loadingJSON} loop autoplay />;
};

export default Loading;
