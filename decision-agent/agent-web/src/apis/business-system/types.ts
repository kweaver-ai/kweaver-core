// 业务域成员类型
export enum BusinessDomainMemberType {
  User = 'user', // 用户
  Department = 'department', // 部门
  Group = 'group', // 组
  Role = 'role', // 角色
  App = 'app', // 应用
}

export interface BusinessDomainMember {
  id: string; // 成员ID
  name: string; // 成员名
  role: 'administrator' | 'developer' | 'viewer'; // 成员角色
  type: BusinessDomainMemberType; // 成员类型
  parent_names: string[]; // 父部门名称数组
}

export interface GetBusinessDomainMembersRequest {
  bdid: string; // 业务域id
  offset?: number; // 分页偏移：默认0
  limit?: number; // 分页限制：默认20
}

export interface GetBusinessDomainMembersResponse {
  limit: number;
  offset: number;
  total: number;
  items: BusinessDomainMember[];
}
