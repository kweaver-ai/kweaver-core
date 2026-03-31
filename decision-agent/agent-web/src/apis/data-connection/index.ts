import { get } from '@/utils/http';

const baseUrl = '/api/data-connection/v1';

export const getDocsSourceList = ({
  limit,
  offset,
  keyword,
  type,
  sort,
  direction,
}: {
  limit?: number;
  offset?: number;
  keyword?: string;
  type?: string;
  sort?: string;
  direction?: string;
}) => {
  return get(
    `${baseUrl}/datasource?limit=${limit}&offset=${offset}&keyword=${keyword}&type=${type}&sort=${sort}&direction=${direction}`
  );
};
