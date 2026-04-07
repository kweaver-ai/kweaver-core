/**
 * 布局组件
 * 基于 Figma 设计
 */
import { Outlet } from 'react-router-dom';
import { Layout } from 'antd';
import { Sidebar } from './Sidebar';

const { Sider, Content } = Layout;

/** 布局组件 */
export default function LayoutComponent() {
  return (
    <Layout style={{ minHeight: '100vh', backgroundColor: '#fafafa' }}>
      <Sider
        width={240}
        style={{
          backgroundColor: '#ffffff',
          borderRight: '1px solid #e7edf7',
          height: '100vh',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
        }}
      >
        <Sidebar />
      </Sider>
      <Layout
        style={{
          marginLeft: 240,
          backgroundColor: '#fafafa',
        }}
      >
        <Content
          style={{
            padding: 24,
            minHeight: 'calc(100vh - 48px)',
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
