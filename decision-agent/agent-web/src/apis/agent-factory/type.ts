export interface Product {
  id: number;
  name: string;
  create_time: string;
  update_time: string;
  create_by: string;
  update_by: string;
}

export interface ToolBox {
  box_id: string;
  box_name: string;
  box_desc: string;
  box_svc_url: string;
  box_icon: string;
  global_headers: Record<string, string>;
  is_build_in: boolean;
  tools: string[] | ToolDetail[];
  create_user: string;
  create_time: string;
  update_user: string;
  update_time: string;
}

export interface ToolDetail {
  tool_id: string;
  tool_name: string;
  tool_desc: string;
  tool_path: string;
  tool_method: string;
  tool_input: ToolInput[];
  tool_output: ToolOutput[];
  is_build_in: boolean;
  create_user: string;
  create_time: string;
  update_user: string;
  update_time: string;
  intervention: boolean;
}

export interface ToolInput {
  input_name: string;
  input_type: string;
  input_desc: string;
  required: boolean;
  in: number;
  children?: ToolInput[] | null;
  example?: any;
  enum?: any;
  OneOfKey?: string;
  OneOfValue?: string;
  additionalProperties?: ToolInput | null;
  oneOf?: any | null;
  anyOf?: any | null;
  allOf?: any | null;
}

export interface ToolOutput {
  output_name: string;
  output_type: string;
  output_desc: string;
  children?: ToolOutput[] | null;
  example?: any;
  enum?: any;
  OneOfKey?: string;
  OneOfValue?: string;
  additionalProperties?: ToolOutput | null;
  items?: ToolOutput | null;
}

export interface Category {
  category_id: string;
  name: string;
}

export interface Agent {
  id: string;
  key: string;
  is_built_in: number;
  is_system_agent: number;
  name: string;
  category_id: string;
  categroy_name: string;
  description: string;
  version: string;
  avatar_type: number;
  avatar: string;
  product_key: number;
  product_name: string;
  publisher: string;
  publish_time: string;
  publish_user_info: {
    user_id: string;
    username: string;
  };
  update_user_info: {
    user_id: string;
    username: string;
  };
  updated_at: string;
  update_by: string;
  status: string;
  profile: string;
  published_at: number;
  updated_by: string;
  updated_by_name: string;
  published_by: string;
  published_by_name: string;
  publish_info: {
    // 是否发布到API：0-否；1-是
    is_api_agent: 0 | 1;
    // 是否发布为web SDK Agent：0-否；1-是
    is_sdk_agent: 0 | 1;
    // 是否发布为技能Agent：0-否；1-是
    is_skill_agent: 0 | 1;
  };
}

export interface LLMConfig {
  // 模型id，预留参数，目前无用
  id: string;
  // 模型名
  name: string;
  // 模型类型。llm: 非推理模型；rlm: 推理模型（可能是reasoning model）
  model_type: 'llm' | 'rlm';
  // 随机性, [ 0 .. 2 ]
  temperature: number;
  // 核采样, [ 0 .. 1 ]
  top_p: number;
  // 前k采样, >= 0
  top_k: number;
  // 频率惩罚度, [ -2 .. 2 ]
  frequency_penalty: number;
  // 话题新鲜度, [ -2 .. 2 ]
  presence_penalty: number;
  // 单次回复限制, >= 0
  max_tokens: number;
  retrieval_max_tokens: number;
}

export interface AgentConfig {
  name: string;
  is_built_in: number;
  is_system_agent?: number;
  description?: string;
  profile: string;
  avatar_type: number;
  avatar: string;
  product_key: number;
  status: 'published' | 'unpublished';
  published_at?: number;
  is_published?: boolean;
  config: {
    input: {
      fields: Array<{
        name: string;
        type: string;
      }>;
      rewrite?: {
        enable: boolean;
        llm_config: LLMConfig;
        data_source?: Record<string, never>;
      };
      augment?: {
        enable: boolean;
        data_source?: Record<string, never>;
      };
      is_temp_zone_enabled?: number;
      temp_zone_config?: {
        name: string;
        max_file_count: number;
        single_file_size_limit: number;
        single_file_size_limit_unit: string;
        support_data_type: string[];
        allowed_file_categories: string[];
        allowed_file_types: string[];
      };
    };
    system_prompt?: string;
    dolphin?: string;
    // 在用户自定义dolphin之前执行的内置dolphin语句
    pre_dolphin?: PrePostDolphinType[];
    // 在用户自定义dolphin之后执行的内置dolphin语句
    post_dolphin?: PrePostDolphinType[];
    is_dolphin_mode?: number;
    memory?: {
      is_enabled: boolean;
    };
    // 相关问题配置（开启后，大模型会在回答用户问题时同时生成相关问题）
    related_question?: {
      // 是否启用（默认关闭）
      is_enabled: boolean;
    };
    // 任务规划模式配置（开启后，智能体将根据情况在需要时对用户的复杂请求进行智能拆解和规划，自动生成详细的任务执行列表，并按步骤逐一完成，提升处理复杂任务的效率和准确性）
    plan_mode?: {
      // 是否启用（默认关闭）
      is_enabled: boolean;
    };
    built_in_can_edit_fields?: {
      name?: boolean;
      avatar?: boolean;
      profile?: boolean;
      input_config?: boolean;
      system_prompt?: boolean;
      model?: boolean;
      skills?: boolean;
      opening_remark_config?: boolean;
      preset_questions?: boolean;
      memory?: boolean;
      'skills.tools.tool_input'?: boolean;
    };
    data_source?: {
      // 指标类型数据源
      metric?: Array<{
        // 指标模型id
        metric_model_id: string;
      }>;
      // 知识库类型数据源
      kn_entry?: Array<{
        // 知识条目id
        kn_entry_id: string;
      }>;
      knowledge_network?: Array<{
        knowledge_network_id: string;
      }>; // 储存知识网络实验版
    };
    skills?: {
      tools: Array<{
        tool_id: string;
        tool_box_id: string;
        tool_input?: Array<{
          enable: boolean;
          input_name: string;
          input_type: string;
          map_type: string;
          map_value: any;
        }>;
        intervention: boolean;
      }>;
      agents: Array<{
        agent_key: string;
        agent_version: string;
        agent_input: Array<{
          enable: boolean;
          input_name: string;
          input_type: string;
          map_type: string;
          map_value: any;
        }>;
        intervention: boolean;
        data_source_config?: SkillAgentDataSourceConfig;
        llm_config?: SkillAgentLLMConfig;
      }>;
      mcps: Array<{
        mcp_server_id: string;
      }>;
    };
    llms: Array<{
      is_default: boolean;
      llm_config: LLMConfig;
    }>;
    is_data_flow_set_enabled?: number;
    opening_remark_config?: {
      type: 'fixed' | 'dynamic';
      fixed_opening_remark?: string;
      dynamic_opening_remark_prompt?: string;
    } | null;
    preset_questions?: Array<{
      question: string;
    }>;
    output?: {
      default_format?: string;
      variables: {
        answer_var?: string;
        doc_retrieval_var?: string;
        graph_retrieval_var?: string;
        other_vars?: string;
      };
    };
  };
}

export interface AgentGenerationParams {
  params: {
    name: string;
    description: string;
    skills: string[];
    sources: string[];
  };
  from: 'system_prompt' | 'opening_remarks' | 'preset_question';
}

export interface ProcessStatuses {
  entries: Array<{
    agent_id: string;
    status: 'processing' | 'completed' | 'failed';
    progress: number;
  }>;
}

export interface Agents {
  entries: Agent[];
  total: number;
}

export interface AgentDetailType {
  version: string;
  description: string;
  id: string;
  name: string;
  profile: string;
  avatar_type: 1;
  avatar: '1';
  status: 'published';
  create_time: number;
  create_by: string;
  updated_at: number;
  published_at: number;
  config: {
    input: {
      fields: [{
          name: string;
          type: string;
          desc: string;
        },
      ];
      rewrite: null;
      augment: null;
      is_temp_zone_enabled: 88;
      temp_zone_config: {
        name: '临时区';
        max_file_count: 50;
        single_file_size_limit: 100;
        single_file_size_limit_unit: 'MB';
        support_data_type: ['wav', 'mp4'];
        allowed_file_categories: ['video'];
      };
    };
    system_prompt: string;
    dolphin: string;
    // 在用户自定义dolphin之前执行的内置dolphin语句
    pre_dolphin?: PrePostDolphinType[];
    // 在用户自定义dolphin之后执行的内置dolphin语句
    post_dolphin?: PrePostDolphinType[];
    is_dolphin_mode: 1;
    data_source: {
      /*
            comments: ['※vid', 'content', 'votes', 'comment_time'];
            regions: ['※vid', 'regions'];
          };
          output_fields: ['Duis irure cupidatat', 'occaecat', 'aliquip anim ullamco'];
        },
      ];
      */
      metric?: Array<{ metric_model_id: string }>;
      kn_entry?: Array<{ kn_entry_id: string }>;
      knowledge_network?: Array<{ knowledge_network_id: string }>;
    };
    skills: {
      tools: [
        {
          tool_id: '15';
          tool_box_id: '54';
          tool_input: [
            {
              enable: true;
              input_name: '泉刚';
              input_type: 'object';
              map_type: 'var';
              map_value: null;
            },
            {
              enable: false;
              input_name: '肥雅鑫';
              input_type: 'file';
              map_type: 'auto';
              map_value: null;
            },
          ];
          intervention: true;
        },
        {
          tool_id: '96';
          tool_box_id: '3';
          tool_input: [
            {
              enable: true;
              input_name: '常玉兰';
              input_type: 'object';
              map_type: 'model';
              map_value: null;
            },
            {
              enable: true;
              input_name: '澹台俊熙';
              input_type: 'file';
              map_type: 'auto';
              map_value: null;
            },
            {
              enable: false;
              input_name: '欧阳苡沫';
              input_type: 'object';
              map_type: 'fixedValue';
              map_value: null;
            },
          ];
          intervention: true;
        },
      ];
      agents: [];
      mcps: [];
    };
    llms: [
      {
        is_default: false;
        llm_config: {
          id: 'sBHd_lfv7VduG0tS-zjKS';
          name: 'deepseek-v3';
          temperature: 0.7071690415509582;
          top_p: 0.20819239685933333;
          top_k: 66;
          frequency_penalty: 'Duis';
          presence_penalty: 'fugiat nisi id sint';
          max_tokens: 'dolore ea eiusmod deserunt';
        };
      },
      {
        is_default: true;
        llm_config: {
          id: 'qmTtMVSF8V-ef01tNAxYD';
          name: 'deepseek-v3';
          temperature: 1.005584848953887;
          top_p: 0.9060435968817686;
          top_k: 96;
          frequency_penalty: 'est in sit voluptate';
          presence_penalty: 'adipisicing sunt proident';
          max_tokens: 'pariatur aute sint eu';
        };
      },
      {
        is_default: false;
        llm_config: {
          id: 'lpKugflG8peYBlFYPgmTo';
          name: 'deepseek-v3';
          temperature: 1.3941956620178484;
          top_p: 0.6668876380045565;
          top_k: 32;
          frequency_penalty: 'pariatur';
          presence_penalty: 'dolore quis in Excepteur aliqua';
          max_tokens: 'in sint sunt qui est';
        };
      },
    ];
  };
}

export interface AgentVersion {
  history_id: string;
  agent_id: string;
  agent_version: string;
  agent_desc: string;
  create_time: string;
}
export interface SpaceListInfo {
  id: string;
  name: string;
  key: string;
  profile: string;
  created_at: string;
  updated_at: string;
  created_by: string;
  created_by_name: string;
}

export enum SpaceMemberTypeEnum {
  User = 'user', // 用户
  Dept = 'dept', // 部门
  UserGroup = 'user_group', // 用户组
}

export enum SpaceResourceEnum {
  DataAgent = 'agent',
}

export interface SpaceMemberType {
  id: number;
  space_id: string;
  obj_type: SpaceMemberTypeEnum;
  obj_id: string;
  obj_name: string;
}

export interface SpaceResourceType {
  id: number;
  space_id: string;
  resource_type: SpaceResourceEnum;
  resource_id: string;
  resource_name: string;
  published_agent_info?: {
    published_at: number;
    published_by: string;
    published_by_name: string;
  };
}

export interface CreateSpaceRequestParamType {
  // 空间名称
  name: string;

  // 空间唯一标识
  key?: string;

  // 空间简介
  profile?: string;

  // 初始空间成员
  members?: {
    // 组织对象类型
    obj_type: SpaceMemberTypeEnum;

    // 组织对象ID
    obj_id: string;
  }[];

  // 初始空间资源
  resources?: {
    // 资源类型
    resource_type: SpaceResourceEnum;

    // 资源ID
    resource_id: string;
  }[];
}

export interface EditSpaceRequestParamType {
  // 空间名称
  name?: string;

  // 空间简介
  profile?: string;

  // 初始空间成员
  members?: {
    // 组织对象类型
    obj_type: SpaceMemberTypeEnum;

    // 组织对象ID
    obj_id: string;
  }[];

  // 初始空间资源
  resources?: {
    // 资源类型
    resource_type: SpaceResourceEnum;

    // 资源ID
    resource_id: string;
  }[];
}

// 发布为枚举
export enum AgentPublishToBeEnum {
  SkillAgent = 'skill_agent',
  ApiAgent = 'api_agent',
  WebSDKAgent = 'web_sdk_agent',
  AgentTPL = 'agent_tpl',
  All = '', // 无值时代表全部
}

// 发布到的目标枚举
export enum AgentPublishToWhereEnum {
  // 发布到自定义空间
  CustomSpace = 'custom_space',
  // 发布到Decision Agent
  Square = 'square',
}

export interface PublishData {
  business_domain_id: string; // 业务域id
  category_ids: string[];
  description: string;
  publish_to_where: AgentPublishToWhereEnum[];
  pms_control: {
    // 角色列表
    role_ids?: string[];
    // 用户列表
    user_ids: string[];
    // 用户组列表
    user_group_ids: string[];
    // 部门列表
    department_ids: string[];
  };
  publish_to_bes: AgentPublishToBeEnum[];
}

// 发布状态枚举
export enum PublishStatusEnum {
  Draft = 'unpublished',
  Published = 'published',
  PublishedEdited = 'published_edited',
  // 无值代表全部
  All = '',
}

export interface AgentManagementPermType {
  // 自定义空间的权限
  custom_space: {
    // 是否有自定义空间的创建权限
    create: boolean;
  };

  // agnet的权限
  agent: {
    // 是否有 个人空间-发布 权限
    publish: boolean;
    // 是否有 个人空间-取消发布 权限
    unpublish: boolean;
    // 是否有 已发布空间-取消发布别人的agent 权限
    unpublish_other_user_agent: boolean;

    // 是否有 发布为其他类型-技能Agent 权限
    publish_to_be_skill_agent: boolean;

    // 是否有 发布为其他类型-Web SDK Agent 权限
    publish_to_be_web_sdk_agent: boolean;

    // 是否有 发布为其他类型-API Agent 权限
    publish_to_be_api_agent: boolean;

    // 是否有发布为数据流agent的权限
    publish_to_be_data_flow_agent: boolean;

    // 是否有 创建系统Agent 权限
    create_system_agent: boolean;

    // 是否有 查看轨迹分析权限
    see_trajectory_analysis: boolean;
  };

  // 模板的权限
  agent_tpl: {
    // 是否有 个人空间-发布 权限
    publish: boolean;
    // 是否有 个人空间-取消发布 权限
    unpublish: boolean;
    // 是否有 已发布空间-取消发布别人的模板 权限
    unpublish_other_user_agent_tpl: boolean;
  };
}

// 获取已发布信息接口返回
export interface PublisheAgentInfoType {
  // 分类列表
  categories: {
    id: string;
    name: string;
  }[];

  // 发布描述
  description: string;

  // 发布到的目标
  publish_to_where: AgentPublishToWhereEnum[];

  pms_control: {
    roles: {
      role_id: string;
      role_name: string;
    }[];

    user: {
      user_id: string;
      username: string;
    }[];

    user_group: {
      user_group_id: string;
      user_group_name: string;
    }[];

    department: {
      department_id: string;
      department_name: string;
    }[];

    app_account: {
      app_account_id: string;
      app_account_name: string;
    };
  };

  publish_to_bes: AgentPublishToBeEnum;
}

export enum DatasourceConfigTypeEnum {
  // 使用自身配置
  SelfConfigured = 'self_configured',
  // 继承主 Agent 数据源
  InheritMain = 'inherit_main',
}

export enum DatasourceConfigSpecificInheritEnum {
  // 仅继承文档数据源
  DocsOnly = 'docs_only',
  // 仅继承图谱数据源
  GraphOnly = 'graph_only',
  // 继承所有类型数据源
  All = 'all',
}

export enum LLMConfigTypeEnum {
  // 使用自身配置
  SelfConfigured = 'self_configured',
  // 继承主 Agent 大模型
  InheritMain = 'inherit_main',
}

// 数据源配置
export interface SkillAgentDataSourceConfig {
  type: DatasourceConfigTypeEnum;
  specific_inherit?: DatasourceConfigSpecificInheritEnum;
}

// 大模型配置
export interface SkillAgentLLMConfig {
  type: LLMConfigTypeEnum;
}

// 模板的发布信息
export interface PublishedTemplateInfoType {
  categories: {
    id: string;
    name: string;
  }[];
}

// 已发布智能体信息
export interface PublishedAgentInfoType extends AgentDetailType {
  // agent 标识
  key: string;
  // 是否是内置智能体。0: 否；1: 是
  is_built_in: 0 | 1;
  // 是否是系统智能体。0: 否；1: 是
  is_system_agent: 0 | 1;
  // 发布者uid
  published_by: string;
  // 发布者名称
  published_by_name: string;
  publish_info: {
    // 是否发布为API Agent
    is_api_agent: 0 | 1;
    // 是否发布为web SDK Agent
    is_sdk_agent: 0 | 1;
    // 是否发布为技能Agent
    is_skill_agent: 0 | 1;
  };
}

export interface GetAgentsByPostRequestBody {
  // 根据名称模糊查询
  name?: string;

  // agent ID数组
  ids?: string[];

  // agent key数组
  agent_keys?: string[];

  // 需要排除的agent keys
  exclude_agent_keys?: string[];

  // 分类ID
  category_id?: string;

  // 发布为标识("api_agent", "web_sdk_agent", "skill_agent", "data_flow_agent")
  publish_to_be?: AgentPublishToBeEnum;

  // 自定义空间ID（如果不是自定义空间，传空。如广场）
  custom_space_id?: string;

  // 每页返回条数
  size?: number;

  // 分页marker（用于获取下一页数据）
  pagination_marker_str?: string;

  // 获取发布到自定义空间的智能体
  is_to_custom_space?: 0 | 1;

  // 获取发布到广场的智能体
  is_to_square?: 0 | 1;

  // 业务域ID数组。如果不传，会使用headers中的x-business-domain。
  business_domain_ids?: string[];
}

// 内置dolphin语句类型
export interface DolphinTemplateType {
  // dolphin块的key
  key: string;
  // dolphin块的名称
  name?: string;
  // dolphin块的dolphin语句
  value: string;
}

export interface PrePostDolphinType extends DolphinTemplateType {
  // 是否启用
  enabled?: boolean;
  // 是否编辑过
  edited?: boolean;
}

// 导入类型
export enum ImportTypeEnum {
  // 如果agent_key存在则更新，否则创建
  Upsert = 'upsert',

  // 只能创建，如果agent_key存在则失败
  Create = 'create',
}

interface ImportAgentResultErrorAgentType {
  // agent标识
  agent_key: string;
  // agent名称
  agent_name: string;
}

// 导入agent的结果
export interface ImportAgentResult {
  // 是否导入成功
  is_success: boolean;

  // 配置无效的agent列表
  config_invalid: ImportAgentResultErrorAgentType[];

  // 需要“创建系统agent权限”的agent列表
  no_create_system_agent_pms: ImportAgentResultErrorAgentType[];

  // agent标识冲突的agent列表(当前域)
  agent_key_conflict: ImportAgentResultErrorAgentType[];

  // agent标识冲突的agent列表(其它域)
  biz_domain_conflict: ImportAgentResultErrorAgentType[];
}

export interface ToolProcessCategoryType {
  // 分类id
  id: string;

  // 分类名称
  name: string;

  // 分类描述
  description: string;
}

export interface ToolProcessStrategyType {
  // 策略id
  id: string;

  // 策略名称
  name: string;

  // 策略描述
  description: string;
}

export interface GetToolProcessCategoriesResponseType {
  // 总数
  total: number;

  entries: Array<ToolProcessCategoryType>;
}

export interface GetToolProcessStrategiesByCategoryResponseType {
  // 总数
  total: number;

  entries: Array<ToolProcessStrategyType>;
}

export interface ResultProcessStrategyType {
  category: ToolProcessCategoryType;
  strategy: ToolProcessStrategyType;
}

export interface SelfConfigSchemaType {
  name: string;
  type: string;
  description: string;
  children: Array<SelfConfigSchemaType>;
}
