import { useContext, createContext } from 'react';

export const MicroWidgetContext = createContext<any>(null);

const useMicroWidgetProps = () => {
  return useContext(MicroWidgetContext);
};

export default useMicroWidgetProps;
