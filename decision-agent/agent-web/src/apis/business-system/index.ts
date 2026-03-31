import { get } from '@/utils/http';
import { GetBusinessDomainMembersRequest, GetBusinessDomainMembersResponse } from './types';

const baseUrl = '/api/business-system/v1/business-domain';

// 查询业务域成员信息
export const getBusinessDomainMembers = async ({
  bdid,
  ...body
}: GetBusinessDomainMembersRequest): Promise<GetBusinessDomainMembersResponse> =>
  get(`${baseUrl}/members/${bdid}`, { body });
