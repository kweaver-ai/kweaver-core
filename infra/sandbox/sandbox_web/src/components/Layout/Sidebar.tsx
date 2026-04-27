/**
 * 侧边栏导航组件
 * 基于 Figma 设计
 */
import { useNavigate, useLocation } from 'react-router-dom';
import { Menu } from 'antd';
import {
  FileOutlined,
  CodeOutlined,
  PlayCircleOutlined,
} from '@ant-design/icons';
import type { MenuProps } from 'antd';

type MenuItem = Required<MenuProps>['items'][number];

function getItem(
  label: string,
  key: string,
  icon: React.ReactNode,
  children?: MenuItem[]
): MenuItem {
  return {
    key,
    icon,
    children,
    label,
  } as MenuItem;
}

/** 侧边栏菜单项 */
const menuItems: MenuItem[] = [
  getItem('会话管理', 'sessions', <CodeOutlined />),
  getItem('模版管理', 'templates', <FileOutlined />),
  getItem('代码执行', 'execute', <PlayCircleOutlined />),
];

/** 侧边栏组件 */
export function Sidebar() {
  const navigate = useNavigate();
  const location = useLocation();

  // 获取当前激活的菜单项
  const getSelectedKey = () => {
    const path = location.pathname;
    if (path.startsWith('/sessions')) return 'sessions';
    if (path.startsWith('/templates')) return 'templates';
    if (path.startsWith('/execute')) return 'execute';
    return 'sessions';
  };

  const handleMenuClick: MenuProps['onClick'] = (e) => {
    navigate(`/${e.key}`);
  };

  return (
    <div
      style={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      {/* Logo 区域 */}
      <div
        style={{
          padding: '24px',
          borderBottom: '1px solid #e7edf7',
        }}
      >
        <h1
          style={{
            fontSize: 16,
            fontWeight: 500,
            color: '#000000',
            margin: 0,
          }}
        >
          沙箱管理平台
        </h1>
        <p
          style={{
            fontSize: 12,
            color: '#677489',
            marginTop: 4,
            margin: '4px 0 0 0',
          }}
        >
          Sandbox Control Plane
        </p>
      </div>

      {/* 导航菜单 */}
      <div style={{ flex: 1, padding: '16px' }}>
        <Menu
          mode="inline"
          selectedKeys={[getSelectedKey()]}
          items={menuItems}
          onClick={handleMenuClick}
          inlineIndent={12}
        />
      </div>
    </div>
  );
}
