/**
 * @description 使用antd封装iconfont
 */

import { createFromIconfontCN } from '@ant-design/icons';
import { type IconFontProps } from '@ant-design/icons/lib/components/IconFont';
import React from 'react';
import classNames from 'classnames';

const IconFontBase = createFromIconfontCN({
  scriptUrl: [
    /* eslint-disable */
    require('@/assets/fonts/lineIconfont.js'), // 线条风格icon
    require('@/assets/fonts/colorIconfont.js'), // 彩色风格icon
    require('@/assets/graphIcons/iconfont.js'), // 图谱icon
    require('@/assets/fonts/twotoneIconfont.js'), // 双色icon
    require('@/assets/fonts/menuIconfont.js'), // 菜单icon
    /* eslint-enable */
  ],
});

export interface AdIconProps extends IconFontProps {
  border?: boolean; // 双色图标是否显示边框 (只有传入的type是双色图标的名字才会有效)
  hideIcon?: boolean; // 当border为true的时候，通过hideIcon控制图标是否显示
  hover?: boolean; //  是否有hover效果
}

const IconFont: React.FC<AdIconProps> = ({
  border = false,
  hover = false,
  type,
  style = {},
  hideIcon = false,
  ...restProps
}) => {
  const renderContent = () => {
    // 处理双色图标加边框逻辑
    if (border) {
      const color = type.split('-').slice(-1)[0];
      if (color && color.length === 6) {
        return (
          <div
            className={classNames('dip-center ad-icon')}
            style={{
              display: 'inline-flex',
              width: hideIcon ? 16 : 32,
              height: hideIcon ? 16 : 32,
              borderRadius: 4,
              fontSize: 16,
              ...style,
            }}
          >
            {!hideIcon && <IconFontBase type={type} {...restProps} />}
          </div>
        );
      }
    }
    return <IconFontBase type={type} style={style} {...restProps} />;
  };
  return renderContent();
};

export default IconFont;
