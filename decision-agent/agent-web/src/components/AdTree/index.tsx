import React from 'react';
import classNames from 'classnames';
import { Tree, type TreeProps } from 'antd';
import { DownOutlined } from '@ant-design/icons';
import './style.less';

interface AdTreeProps extends TreeProps {
  hover?: boolean; // 节点背景是否有hover效果
}

// expandAction 控制点击目录树节点的名字，可以展开树节点 关闭需要设置false
const AdTree: React.FC<AdTreeProps> = ({ hover = true, className, expandAction = 'click', ...restProps }) => {
  return (
    <Tree
      className={classNames(className, 'ad-tree', {
        'ad-tree-noHover': !hover,
      })}
      showIcon
      showLine={{
        showLeafIcon: false,
      }}
      blockNode
      switcherIcon={<DownOutlined />}
      expandAction={expandAction}
      {...restProps}
    />
  );
};

export default AdTree;
