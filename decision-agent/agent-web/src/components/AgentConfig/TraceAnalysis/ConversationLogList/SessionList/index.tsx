import type { TableColumnsType } from 'antd';
import { useEffect, useState } from 'react';
import ADTable from '@/components/ADTable';
import DipButton from '@/components/DipButton';
import SessionAnalysis from './SessionAnalysis';
import { getAgentObservabilitySessionList } from '@/apis/agent-app';
import { useAgentConfig } from '../../../AgentConfigContext';
import dayjs from 'dayjs';

const SessionList = ({ conversationId, dateValue, qualityInsightAgentDetails }: any) => {
  const { state } = useAgentConfig();
  const [listData, setListData] = useState<any>([]);
  const [open, setOpen] = useState(false);
  const [activeSession, setActiveSession] = useState<any>(null);

  useEffect(() => {
    getSessionList();
  }, [conversationId, dateValue]);

  const getSessionList = async () => {
    console.log(dateValue, 'dateValue ===== SessionList');
    const res: any = await getAgentObservabilitySessionList(state.id!, conversationId, {
      page: 1,
      size: 10000,
      agent_id: state.id!,
      agent_version: 'v0',
      start_time: dateValue[0].valueOf(),
      end_time: dateValue[1].valueOf(),
    });
    if (res) {
      console.log(res, 'session列表');
      setListData(res.entries);
    }
  };

  const columns: TableColumnsType = [
    {
      title: '会话ID',
      dataIndex: 'session_id',
    },
    {
      title: '运行开始时间',
      dataIndex: 'start_time',
      render: (value: string) => {
        return dayjs(value).format('YYYY/MM/DD HH:mm:ss');
      },
    },
    {
      title: '运行结束时间',
      dataIndex: 'end_time',
      render: value => {
        return dayjs(value).format('YYYY/MM/DD HH:mm:ss');
      },
    },
    {
      title: '运行时长',
      dataIndex: 'session_duration',
      render: value => {
        return `${(value / 1000).toFixed(3)}秒`;
      },
    },
    {
      title: '操作',
      dataIndex: 'operate',
      render: (_value: any, record: any) => {
        return (
          <DipButton
            type="link"
            onClick={() => {
              setOpen(true);
              setActiveSession(record);
            }}
          >
            查看详情
          </DipButton>
        );
      },
    },
  ];
  return (
    <>
      <ADTable
        rowKey="session_id"
        showHeader={false}
        showSearch={false}
        size="small"
        columns={columns}
        dataSource={listData}
      />
      <SessionAnalysis
        open={open}
        onClose={() => {
          setOpen(false);
          setActiveSession(null);
        }}
        sessionId={activeSession?.session_id}
        activeSession={activeSession}
        conversationId={conversationId}
        dateValue={dateValue}
        qualityInsightAgentDetails={qualityInsightAgentDetails}
      />
    </>
  );
};

export default SessionList;
