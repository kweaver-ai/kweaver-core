import { RoleIdEnum, RoleTypeEnum } from './types';
export { RoleTypeEnum };

export const getRoleByUserInfo = (userInfo: {
  user: {
    roles: {
      id: string;
    }[];
  };
}) => {
  const roles = userInfo?.user?.roles;
  let isAdmin: boolean = false;
  let role: RoleTypeEnum = RoleTypeEnum.NormalUser;

  const roleIds = roles.map(({ id }) => id);
  const adminRoleMap: any = [
    [RoleIdEnum.SuperAdmin, RoleTypeEnum.SuperAdmin],
    [RoleIdEnum.SysAdmin, RoleTypeEnum.SysAdmin],
    [RoleIdEnum.SecAdmin, RoleTypeEnum.SecAdmin],
    [RoleIdEnum.AuditAdmin, RoleTypeEnum.AuditAdmin],
    [RoleIdEnum.OrgManager, RoleTypeEnum.OrgManager],
    [RoleIdEnum.OrgAudit, RoleTypeEnum.OrgAudit],
  ];

  for (const [roleId, roleType] of adminRoleMap) {
    if (roleIds.includes(roleId)) {
      isAdmin = true;
      role = roleType;
      break; // 只需要匹配第一个找到的角色
    }
  }

  return {
    isAdmin,
    roleType: role,
  };
};
