// 定义状态类型
// 定义状态类型
import type { GetStateAction } from '@/hooks/useLatestState';
import type { AiInputValue } from './components/AiInput/interface';
import type { FormInstance, TableColumnsType } from 'antd';
import type { EChartsOption } from 'echarts';
import type { GetConversationListOption } from '@/apis/super-assistant';

type ToolArgsType = {
  key: string;
  value: any;
  type: 'string';
};

export type InterruptDataType = {
  handle: any;
  data: {
    // session_id: string;
    tool_name: string;
    tool_description: string;
    tool_args: ToolArgsType[];
    interrupt_config: {
      requires_confirmation: boolean;
      confirmation_message: string;
    };
  };
};

export type DipChatItemRole = 'user' | 'common';

export type DipChatItemContentProgressType = {
  title?: string; // 工具的标题 大模型的回答没有标题
  status: 'processing' | 'completed' | 'failed'; // 生成状态
  consumeTime: string; // 耗时
  consumeTokens?: string; // tokens消耗
  cachedTokens?: string; // 提示词命中缓存token数量
  type:
    | 'llm'
    | 'docQa_tool'
    | 'common_tool'
    | 'sql_tool'
    | 'chart_tool'
    | 'code_tool'
    | 'ngql_tool'
    | 'net_search_tool'
    | 'metric_tool'; // 类型  大模型生成的  还是工具生成的
  // 大模型
  llmResult?: {
    text: string; // 大模型的最终回答, markdown文本
    thinking: string; // 大模型的思考过程
  };
  // sql工具
  sqlResult?: {
    sql: string;
    tableColumns: TableColumnsType;
    tableData: any[];
  };
  // 图表工具
  chartResult?: {
    echartsOptions: EChartsOption;
    tableColumns: TableColumnsType;
    tableData: any[];
  };
  // 代码工具（沙箱工具）
  codeResult?: {
    input: string;
    actionResult: any;
    action:
      | 'execute_code'
      | 'create_file'
      | 'read_file'
      | 'list_files'
      | 'execute_command'
      | 'upload_file'
      | 'get_status'
      | 'close_sandbox'
      | 'download_file';
    actionMessage: string;
  };
  // NGQL工具
  ngqlResult?: {
    sql: string;
    tableColumns: TableColumnsType;
    tableData: any[];
  };
  // docQa工具
  docQaToolResult?: {
    htmlText: string;
    cites: any; // 当前progress引用的数据
  };
  // 通用工具
  commonToolResult?: {
    input: string;
    output: string;
  };
  // 网络搜索工具
  netSearchResult?: {
    cites: any;
  };
  metricResult?: {
    tableColumns: TableColumnsType;
    tableData: any[];
    input: any;
  };
  skillInfo?: any; // 用到的技能信息
  originalAnswer?: string;
};

export type DipChatItemContentType = {
  cites: any[]; // 此次回答引用的数据
  progress: DipChatItemContentProgressType[];
  related_queries: string[]; // 此次回答的相关问题
  totalTokens: number; // 总消耗Token
  totalTime: string; // 总耗时
  ttftTime: string; // 首Token延迟
};

export type DipChatItem = {
  key: string;
  role: DipChatItemRole;
  content: any;
  loading?: boolean; // 流式接口是否pending
  cancel?: boolean; // 是否取消
  generating?: boolean; // 是否正在生成中
  // 用于控制聊天中断的情况
  interrupt?: InterruptDataType; // 中断数据的原始结构（和后端保持一致，中断后再次发起请求，使用这个结构，向接口传参）
  error?: string; // 储存对话过程接口报错信息
  fileList?: FileItem[]; // role 为user的时候 fileList可能会存在数据
  sourceData?: any; // 流式接口返回的原始数据
  updateTime?: number;
  agentRunId?: string;
  status?: 'processing' | 'completed' | 'failed' | 'cancelled';
};

export type ConversationItemType = {
  label: string;
  key: string;
  status: 'processing' | 'completed' | 'failed' | 'cancelled';
  children?: ConversationItemType[];
  unRead: boolean;
  timestamp: number;
};

export type PreviewFileType = {
  fileId: string;
  fileName: string;
  fileExt: string;
};

export type FileItem = {
  container_path: string;
  name: string;
  size: number;
  error: string;
  status: 'processing' | 'failed' | 'completed';
  checked: boolean;
};

export type DipChatState = {
  activeChatItemIndex: number; // 激活的聊天项的索引,用于控制打开右边侧边栏，显示激活聊天项的详细信息
  chatList: DipChatItem[]; // 聊天列表
  conversationItems: ConversationItemType[]; // 会话数据源
  conversationItemsTotal: number; // 会话总数
  activeConversationKey: string; // 选中的会话key
  conversationListModalOpen: boolean; // 所有会话弹框Modal
  chatListAutoScroll: boolean; // 聊天列表是否自动滚动
  aiInputValue: AiInputValue;
  streamGenerating: boolean; // 流式输出中,
  conversationCollapsed: boolean; // 会话列表是否折叠
  previewFile?: PreviewFileType; // 预览的文件
  // 5.0.0.1新加
  agentDetails: any; // Agent的详情（接口的返回结果）
  agentAppType: AgentAppType;
  agentAppKey: string; // Agent应用的key。 默认是Agent的id
  debug: boolean; // 是否处于调试模式
  showDebuggerArea: boolean; // 是否显示调试区域
  tempFileList: FileItem[]; // 临时文件列表
  // 例如问数场景聊天项的content是一个数组，同样也适用于其他content是数组的场景
  activeProgressIndex: number; // 选中的是哪个progress
  showAgentInputParamsDrawer: boolean; // Agent 输入参数抽屉
  agentInputParamsFormValue: any; // Agent的输入参数表单的值
  agentInputParamsFormErrorFields: any[]; // Agent的输入参数表单错误
  agentInputParamForm: FormInstance<any> | null; // Agent 输入参数Form表单实例
  singleStreamResult: any[]; // 供调试用，用于开发和测试快速查看流式返回的完整结果
  toolAutoExpand: boolean; // 是否自动展开工具
  logQueryAgentDetails: Record<string, any>; // 日志查询智能体详情
  tempAreaOpen: boolean;
};

export type TempFileType = {
  file_name: string;
};

export type ChatBody = {
  query?: string;
  custom_querys?: any;
  tool?: any;
  conversation_id?: string;
  interrupted_assistant_message_id?: string;
  chat_mode?: 'deep_thinking' | 'normal';
  agent_id?: string;
  agent_version?: string;
  regenerate_user_message_id?: string;
  regenerate_assistant_message_id?: string;
  temporary_area_id?: string;
  stream?: boolean;
  inc_stream?: boolean;
  selected_files?: TempFileType[];
  // 中断的参数
  interruptAction?: 'confirm' | 'skip';
  interruptModifiedArgs?: Array<{ key: string; value: any }>;
};

export type SendChatPram = {
  chatList?: DipChatItem[]; // 要更新的聊天数据。可选
  body: ChatBody; // 流式请求体，必填
  activeChatItemIndex?: number;
  recoverConversation?: boolean; // 是否恢复对话
};

type DipChatStateKeys = keyof DipChatState;

// 定义上下文类型
export type DipChatContextType = {
  dipChatStore: DipChatState;
  setDipChatStore: (data: Partial<DipChatState>) => void;
  getDipChatStore: GetStateAction<DipChatState>;
  resetDipChatStore: (key?: DipChatStateKeys | DipChatStateKeys[]) => void;
  sendChat: (param: SendChatPram) => void; // 发送，发起流式请求
  stopChat: () => void; // 终止会话
  cancelChat: () => void; // 取消会话不终止
  openSideBar: (chatItemIndex: number) => void; // 打开侧边栏
  closeSideBar: () => void; // 关闭侧边栏
  getConversationData: (params?: GetConversationListOption) => void; // 获取会话列表
  getConversationDetailsByKey: (key: string) => Promise<
    | {
        conversationLoading: boolean;
        recoverConversation: boolean;
        chatList: DipChatItem[];
        read_message_index: number;
        message_index: number;
      }
    | false
  >; // 获取会话详情
};

// common 普通agent应用
export type AgentAppType = 'common';

export type DipChatProps = {
  agentId?: string;
  agentVersion?: string;

  // 调试页面需要用到的props
  debug?: boolean; // 调试模式下的Agent使用
  agentDetails?: any; // debugger模式下，通过直接传最新的agentDetails，优先级比agentId高
  onSaveAgent?: () => Promise<boolean | string>; // 触发保存Agent的事件

  // 外部初始控制一次组件是否发起请求的配置
  defaultChatList?: DipChatItem[]; // 默认的聊天列表
  defaultAiInputValue?: AiInputValue; // 默认的input框的初始值

  // Agent应用类型props
  agentAppType: AgentAppType;

  customSpaceId?: string; // 自定义空间Id
};
