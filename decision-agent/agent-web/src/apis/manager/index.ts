import { get } from '@/utils/http';

const baseUrl = '/api/manager/v1';

export const getSwaggerDoc = () => get(`${baseUrl}/swaggerDoc`);
