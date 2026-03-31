import TraceAnalysisTitle from '../Title';
import SearchInput from '@/components/SearchInput';
import { useEffect, useState } from 'react';
import { Collapse, Divider, Pagination } from 'antd';
import { CaretRightOutlined, MessageOutlined } from '@ant-design/icons';
import styles from './index.module.less';
import intl from 'react-intl-universal';
import SessionList from './SessionList';
import { getAgentObservabilityConversationList } from '@/apis/agent-app';
import { useAgentConfig } from '../../AgentConfigContext';
import NoResultIcon from '@/assets/images/noResult.svg';
import EmptyIcon from '@/assets/images/empty.svg';
import dayjs from 'dayjs';

const ConversationLogList = ({ dateValue, qualityInsightAgentDetails, refreshTraceAnalysis }: any) => {
  const { state } = useAgentConfig();
  const [listProps, setListProps] = useState({
    current: 1,
    pageSize: 5,
    total: 0,
    searchValue: '',
    data: [],
  });

  useEffect(() => {
    getData();
  }, [listProps.current, listProps.pageSize, listProps.searchValue, dateValue]);

  const getData = async () => {
    const res = await getAgentObservabilityConversationList(state.key!, {
      page: listProps.current,
      size: listProps.pageSize,
      agent_id: state.id!,
      title: listProps.searchValue,
      agent_version: 'v0',
      start_time: dateValue[0].valueOf(),
      end_time: dateValue[1].valueOf(),
    });
    if (res) {
      setListProps(prev => ({
        ...prev,
        total: res.total_count,
        data: res.entries ?? [],
      }));
    }
  };

  const items = listProps.data.map((item: any) => ({
    key: item.conversation.id,
    label: (
      <div>
        <div className="dip-flex-align-center">
          <MessageOutlined className="dip-text-color-45" />
          <span className="dip-ml-8 dip-text-color">{item.conversation.title}</span>
        </div>
        <div className="dip-flex-align-center dip-text-color-45 dip-font-12 dip-mt-4">
          <span>更新时间：{dayjs(item.conversation.update_time).format('YYYY-MM-DD HH:mm:ss')}</span>
          <Divider type="vertical" />
          <span>会话数量：{item.session_count}</span>
        </div>
      </div>
    ),
    children: (
      <SessionList
        conversationId={item.conversation.id}
        dateValue={dateValue}
        qualityInsightAgentDetails={qualityInsightAgentDetails}
        refreshTraceAnalysis={refreshTraceAnalysis}
      />
    ),
  }));

  const renderList = () => {
    if (items.length === 0) {
      if (listProps.searchValue) {
        return (
          <div className="dip-text-color-65 dip-font-12 dip-flex-column-center dip-p-24">
            <NoResultIcon />
            <div>{intl.get('global.noResult')}</div>
          </div>
        );
      }
      return (
        <div className="dip-text-color-65 dip-font-12 dip-flex-column-center dip-p-24">
          <EmptyIcon />
          <div>{intl.get('global.noData')}</div>
        </div>
      );
    }
    return (
      <>
        <div className={styles.collapseWrapper}>
          <Collapse
            expandIcon={({ isActive }) => (
              <CaretRightOutlined className="dip-text-color-25" style={{ fontSize: 12 }} rotate={isActive ? 90 : 0} />
            )}
            items={items}
            expandIconPosition="end"
            bordered={false}
          />
        </div>
        <Pagination
          className="dip-mt-16"
          align="end"
          current={listProps.current}
          onChange={(page, pageSize) => {
            setListProps(prev => ({
              ...prev,
              current: page,
              pageSize: pageSize!,
            }));
          }}
          showTotal={(total: number) => intl.get('knowledge.total', { total })}
          showSizeChanger={false}
          hideOnSinglePage={true}
          total={listProps.total}
          pageSize={listProps.pageSize}
        />
      </>
    );
  };

  return (
    <div className="dip-mt-24">
      <TraceAnalysisTitle
        title="对话日志列表"
        description="所有对话的执行记录"
        extra={
          <SearchInput
            placeholder="搜索对话名称"
            value={listProps.searchValue}
            onChange={e => {
              setListProps(prev => ({
                ...prev,
                current: 1,
                searchValue: e.target.value,
              }));
            }}
            style={{ width: 260 }}
          />
        }
      />
      {renderList()}
    </div>
  );
};
export default ConversationLogList;
