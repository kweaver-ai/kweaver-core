// 管理员的roleId枚举
export enum RoleIdEnum {
  // 超级管理员
  SuperAdmin = '7dcfcc9c-ad02-11e8-aa06-000c29358ad6',

  // 系统管理员
  SysAdmin = 'd2bd2082-ad03-11e8-aa06-000c29358ad6',

  // 安全管理员
  SecAdmin = 'd8998f72-ad03-11e8-aa06-000c29358ad6',

  // 审计管理员
  AuditAdmin = 'def246f2-ad03-11e8-aa06-000c29358ad6',

  // 组织管理员
  OrgManager = 'e63e1c88-ad03-11e8-aa06-000c29358ad6',

  // 组织审计员
  OrgAudit = 'f06ac18e-ad03-11e8-aa06-000c29358ad6',
}

// AccessPicker组件需要的role枚举
export enum RoleTypeEnum {
  // 超级管理员
  SuperAdmin = 'super_admin',

  // 系统管理员
  SysAdmin = 'sys_admin',

  // 安全管理员
  SecAdmin = 'sec_admin',

  // 组织审计员
  OrgAudit = 'org_audit',

  // 组织管理员
  OrgManager = 'org_manager',

  // 审计管理员
  AuditAdmin = 'audit_admin',

  // 普通用户
  NormalUser = 'normal_user',
}
