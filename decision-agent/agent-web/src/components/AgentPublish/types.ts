export enum AccessorTypeEnum {
  User = 'user',
  Department = 'department',
  Group = 'group',
  App = 'app',
  Role = 'role',
}

// 可见范围枚举
export enum VisibleRangeEnum {
  // 全部用户
  AllUser = 'allUser',
  // 指定范围
  SpecifiedRange = 'specifiedRange',
}

export enum PublishModeEnum {
  PublishAgent = 'publishAgent', // 发布agent
  UpdatePublishAgent = 'updatePublishAgent', // 更新发布agent
  PublishTemplate = 'publishTemplate', // 发布模板
  PublishAgentAsTemplate = 'publishAgentAsTemplate', // 复制agent为模板并发布
}
