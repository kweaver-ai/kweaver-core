import './index.less';
import classNames from 'classnames';

interface ShinyTextProps {
  text: string;
  disabled?: boolean;
  speed?: number;
  className?: string;
  loading?: boolean;
}

const ShinyText: React.FC<ShinyTextProps> = ({ text, loading = true, speed = 2, className }) => {
  const animationDuration = `${speed}s`;
  const prefixCls = 'shiny-text';
  return (
    <div
      className={classNames(prefixCls, className, {
        [`${prefixCls}-disabled`]: !loading,
      })}
      style={{ animationDuration }}
      title={text}
    >
      {text}
    </div>
  );
};

export default ShinyText;
