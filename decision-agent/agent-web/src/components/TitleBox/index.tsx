import React from 'react';
import classNames from 'classnames';
import { Divider } from 'antd';
import './style.less';

type TypeTitleBox = {
  text: string; // 文本
  // 其他自定义参数
  [key: string]: any;
};

const TitleBox: React.FC<TypeTitleBox> = props => {
  const { text, bgColor = 'rgba(18, 110, 227, 0.06)', className, ...others } = props;
  return (
    <Divider
      className={classNames(className)}
      dashed
      orientation="left"
      orientationMargin="0"
      {...others}
      style={{ ...others?.style, flex: 1 }}
    >
      <div className={classNames('titleBoxRoot dip-flex-align-center')}>{text}</div>
    </Divider>
  );
};
export default TitleBox;
