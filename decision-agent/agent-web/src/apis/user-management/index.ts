import { get } from '@/utils/http';

export function getUserAvatarsByIds(userIds: string[]) {
  return get(`/api/user-management/v1/avatars/${userIds.join(',')}`);
}
