import { useReducer } from 'react';
import { Tag, Dropdown, Tooltip, Menu } from 'antd';
import { EllipsisOutlined } from '@ant-design/icons';
import ADTable from '@/components/ADTable';
import Format from '@/components/Format';
import intl from 'react-intl-universal';

const data = [
  {
    key: '1',
    name: 'John Brown',
    age: 32,
    address: 'New York No. 1 Lake Park',
    tags: ['nice', 'developer'],
  },
  {
    key: '2',
    name: 'Jim Green',
    age: 42,
    address: 'London No. 1 Lake Park',
    tags: ['loser'],
  },
  {
    key: '3',
    name: 'Joe Black',
    age: 32,
    address: 'Sidney No. 1 Lake Park',
    tags: ['cool', 'teacher'],
  },
];

const header_menu = (
  <Menu
    onClick={e => {
      console.log('头部处理函数', e);
    }}
    onContextMenu={e => {
      console.log('头部处理函数111', e);
    }}
  >
    <Menu.Item key="7" style={{ height: 40 }}>
      {intl.get('graphList.timedTask')}
    </Menu.Item>
    <Menu.Item key="8" style={{ height: 40 }} onClick={() => {}}>
      {intl.get('knowledge.export')}
    </Menu.Item>
    <Menu.Item key="9" style={{ height: 40 }} disabled={true}>
      {intl.get('knowledge.authorityManagement')}
    </Menu.Item>
  </Menu>
);

const menu = (
  <Menu
    onClick={e => {
      console.log('操作处理函数', e);
    }}
  >
    <Menu.Item key="1" style={{ height: 40 }}>
      {intl.get('graphList.timedTask')}
    </Menu.Item>
    <Menu.Item key="2" style={{ height: 40 }} onClick={() => {}}>
      {intl.get('knowledge.export')}
    </Menu.Item>
    <Menu.Item key="3" style={{ height: 40 }} disabled={true}>
      {intl.get('knowledge.authorityManagement')}
    </Menu.Item>
    <Menu.Item key="4" style={{ height: 40 }} onClick={() => {}}>
      {intl.get('knowledge.delete')}
    </Menu.Item>
  </Menu>
);

const Hello = () => {
  const [r, forceRender] = useReducer(s => s + 1, 0);
  const columns: any = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      sorter: true,
      fixed: 'left',
      filters: [
        {
          text: 'Joe',
          value: 'Joe',
        },
        {
          text: 'John',
          value: 'John',
        },
      ],
      // width: 300,
      render: (text: any) => <a>{text}</a>,
    },
    {
      title: '操作',
      key: 'action',
      // width: 110,
      fixed: 'left',
      // renderConfig:{},
      render: (text: any, record: any) => (
        <Dropdown
          overlay={menu}
          trigger={['click']}
          placement="bottomLeft"
          getPopupContainer={triggerNode => triggerNode?.parentElement?.parentElement || document.body}
        >
          <Tooltip title={intl.get('graphList.more')} placement="right">
            <Format.Button className="operate" type="icon">
              <EllipsisOutlined style={{ color: 'rgba(0, 0, 0, 0.85)', fontSize: '16px' }} />
            </Format.Button>
          </Tooltip>
        </Dropdown>
      ),
    },
    {
      title: 'Age',
      dataIndex: 'age',
      key: 'age',
      // width: 200
    },
    {
      title: 'Address',
      ellipsis: true,
      dataIndex: 'address',
      key: 'address',
      // width: 200
    },
    {
      title: 'Tags',
      key: 'tags',
      dataIndex: 'tags',
      // width: 200,
      render: (tags: any) => (
        <>
          {tags.map((tag: any) => {
            let color = tag.length > 5 ? 'geekblue' : 'green';
            if (tag === 'loser') {
              color = 'volcano';
            }
            return (
              <Tag color={color} key={tag} onClick={forceRender}>
                {tag.toUpperCase()}
                {r}
              </Tag>
            );
          })}
        </>
      ),
    },
  ];

  return (
    <ADTable
      // className="ad-table-1"
      // width={1000}
      title={'hello world !!!'}
      // showHeader={false}
      // showFilter={true}
      onSearchChange={(e: any) => {
        console.log('searchInput change', e);
      }}
      // searchPlaceholder="123"
      // localeImage={empty}
      // localeText="123"
      onRow={(record, index) => {
        return {
          onContextMenu: e => {
            console.log(e.pageX, e.pageY);
            e.preventDefault();
          },
        };
      }}
      onHeaderRow={(columns, index) => {
        return {
          onContextMenu: e => {
            console.log(columns, index);
            e.preventDefault();
          },
        };
      }}
      onFilterClick={(e: any, status: any) => console.log('filter', status)}
      filterToolsOptions={[
        {
          id: 1,
          // label: intl.get('cognitiveService.analysis.query'),
          // optionList: [
          //   {
          //     key: '1',
          //     value: '111',
          //     text: '1'
          //   },
          //   {
          //     key: '2',
          //     value: '2',
          //     text: '2'
          //   },
          //   {
          //     key: '3',
          //     value: '3',
          //     text: '3'
          //   }
          // ],
          // onHandle: (v: any) => {
          //   console.log('查询业务处理中', v);
          // },
          itemDom: <div>hello world</div>,
        },
        {
          id: 2,
          label: intl.get('cognitiveService.analysis.state'),
          optionList: [
            {
              key: '1',
              value: '1',
              text: '1',
            },
            {
              key: '2',
              value: '2',
              text: '2',
            },
            {
              key: '3',
              value: '3',
              text: '3',
            },
          ],
          onHandle: (v: any) => {
            console.log('状态查询业务处理中', v);
          },
        },
        {
          id: 3,
          label: intl.get('cognitiveService.analysis.associated'),
          optionList: [
            {
              key: '1',
              value: '1',
              text: '1',
            },
            {
              key: '2',
              value: '2',
              text: '2',
            },
            {
              key: '3',
              value: '3',
              text: '3',
            },
          ],
          onHandle: (v: any) => {
            console.log('相关图谱业务处理中', v);
          },
        },

        {
          id: 4,
          label: '4',
          optionList: [
            {
              key: '1',
              value: '1',
              text: '1',
            },
            {
              key: '2',
              value: '2',
              text: '2',
            },
            {
              key: '3',
              value: '3',
              text: '3',
            },
          ],
        },
        {
          id: 5,
          label: '5',
          optionList: [
            {
              key: '1',
              value: '1',
              text: '1',
            },
            {
              key: '2',
              value: '2',
              text: '2',
            },
            {
              key: '3',
              value: '3',
              text: '3',
            },
          ],
        },
        {
          id: 6,
          label: '6',
          optionList: [
            {
              key: '1',
              value: '1',
              text: '1',
            },
            {
              key: '2',
              value: '2',
              text: '2',
            },
            {
              key: '3',
              value: '3',
              text: '3',
            },
          ],
        },
        {
          id: 7,
          label: '7',
          optionList: [
            {
              key: '1',
              value: '1',
              text: '1',
            },
            {
              key: '2',
              value: '2',
              text: '2',
            },
            {
              key: '3',
              value: '3',
              text: '3',
            },
          ],
        },
      ]}
      renderButtonConfig={[
        {
          key: '1',
          // label: '排序',
          // type: 'order',
          position: 'right',
          // onHandle: () => {
          //   console.log('排序');
          // },
          // menu: <div>hello world</div>,
          // tip: true,
          itemDom: <div>hello</div>,
        },
      ]}
      onFiltersToolsClose={() => console.log('close')}
      columns={columns}
      dataSource={data}
      // scroll={{ x: 500, y: 100 }}
      rowKey="key"
      // rowSelection={{
      //   type: 'checkbox'
      // }}
      contextMenu={{ headerContextMenu: header_menu, bodyContextMenu: menu }}
    >
      {/* <div>hello world</div> */}
    </ADTable>
  );
};

export default Hello;
