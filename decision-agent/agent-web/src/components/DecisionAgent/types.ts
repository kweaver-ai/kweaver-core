export enum AgentActionEnum {
  Copy = 'copy',
  ViewAPI = 'viewAPI',
  Use = 'use',
  ConfigInfo = 'configInfo',
  Publish = 'publish',
  UpdatePublishInfo = 'updatePublishInfo',
  Unpublish = 'unpublish',
  PublishAsTemplate = 'publishAsTemplate', // 发布为模板
  // GenerateAsTemplate = 'generateAsTemplate',
  Delete = 'delete',
  // 移除Agent（自定义空间）
  RemoveAgent = 'removeAgent',
}

export enum TemplateActionEnum {
  CreateAgentFromTemplate = 'createAgentFromTemplate', // 使用此模板创建agent
  CopyTemplate = 'copyTemplate', // 复制模板
  TemplateConfig = 'templateConfig', // 配置信息
  PublishTemplate = 'publishTemplate', // 发布模板
  UnpublishTemplate = 'unpublishTemplate', // 取消模板
  DeleteTemplate = 'deleteTemplate', // 删除模板
}

export enum ModeEnum {
  DataAgent = 'dataAgent',

  // 我创建的agent列表页面
  MyAgent = 'myAgent',

  // 自定义空间的agent列表页面
  CustomSpace = 'customSpace',

  // 全部模板列表页面
  AllTemplate = 'allTemplate',

  // 我创建的模板列表页面
  MyTemplate = 'myTemplate',

  // API
  API = 'api',
}
