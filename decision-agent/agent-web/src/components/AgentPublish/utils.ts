import { AccessorTypeEnum, VisibleRangeEnum } from './types';

export const formatAccessor = (accessor: any) => {
  let id = '';
  let name = accessor.name;

  switch (accessor.type) {
    case AccessorTypeEnum.User:
      id = accessor.userid || accessor.id;
      if (!name) {
        name = accessor.user?.displayName;
      }
      break;

    case AccessorTypeEnum.Department:
      id = accessor.depid || accessor.id;
      break;

    default:
      id = accessor.id;
  }

  return {
    id,
    type: accessor.type,
    name,
  };
};

// 将前端数据，转换成后端需要的数据
export const formatPMSControl = (accessors: any[]) => {
  if (!accessors.length) return null;

  return accessors.reduce(
    (prev, { id, type }) => {
      switch (type) {
        case AccessorTypeEnum.User:
          return {
            ...prev,
            user_ids: [...prev.user_ids, id],
          };

        case AccessorTypeEnum.Department:
          return {
            ...prev,
            department_ids: [...prev.department_ids, id],
          };

        case AccessorTypeEnum.Group:
          return {
            ...prev,
            user_group_ids: [...prev.user_group_ids, id],
          };

        case AccessorTypeEnum.App:
          return {
            ...prev,
            app_account_ids: [...prev.app_account_ids, id],
          };

        case AccessorTypeEnum.Role:
          return {
            ...prev,
            role_ids: [...prev.role_ids, id],
          };

        default:
          return prev;
      }
    },
    {
      role_ids: [],
      user_ids: [],
      user_group_ids: [],
      department_ids: [],
      app_account_ids: [],
    }
  );
};

// 将后端数据，转换成前端展示用的数据
export const transformPMSControlFromBackend = (pms_control: any) => {
  if (!pms_control) {
    return {
      visibleRange: VisibleRangeEnum.AllUser,
      accessors: [],
    };
  }

  let users: any[] = [];
  Object.keys(pms_control).map(key => {
    const value = pms_control[key];

    switch (key) {
      case 'user':
        users = [
          ...users,
          ...value.map(({ user_id, username }) => ({
            id: user_id,
            name: username,
            type: AccessorTypeEnum.User,
          })),
        ];
        break;

      case 'user_group':
        users = [
          ...users,
          ...value.map(({ user_group_id, user_group_name }) => ({
            id: user_group_id,
            name: user_group_name,
            type: AccessorTypeEnum.Group,
          })),
        ];
        break;

      case 'department':
        users = [
          ...users,
          ...value.map(({ department_id, department_name }) => ({
            id: department_id,
            name: department_name,
            type: AccessorTypeEnum.Department,
          })),
        ];
        break;

      case 'app_account':
        users = [
          ...users,
          ...value.map(({ app_account_id, app_account_name }) => ({
            id: app_account_id,
            name: app_account_name,
            type: AccessorTypeEnum.App,
          })),
        ];
        break;

      case 'roles':
        users = [
          ...users,
          ...value.map(({ role_id, role_name }) => ({
            id: role_id,
            name: role_name,
            type: AccessorTypeEnum.Role,
          })),
        ];
        break;

      default:
        break;
    }
  });

  return {
    visibleRange: VisibleRangeEnum.SpecifiedRange,
    accessors: users,
  };
};
