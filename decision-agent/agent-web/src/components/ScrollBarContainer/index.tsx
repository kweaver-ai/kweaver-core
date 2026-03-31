import { forwardRef, type PropsWithChildren } from 'react';
import { MacScrollbar, type MacScrollbarProps } from 'mac-scrollbar';
import 'mac-scrollbar/dist/mac-scrollbar.css';

const ScrollBarContainer = forwardRef<any, PropsWithChildren<MacScrollbarProps>>(({ children, ...restProps }, ref) => {
  return (
    <MacScrollbar ref={ref} thumbStyle={() => ({ backgroundColor: 'rgba(0,0,0,.25)' })} {...restProps}>
      {children}
    </MacScrollbar>
  );
});

export default ScrollBarContainer;
