import { useCallback, useEffect, useState } from 'react';
import { useMicroWidgetProps } from './index';
import { cacheableGet } from '@/utils/http';

interface BusinessDomain {
  id: string; //  业务域ID
  name: string; // 业务域名
  description: string; // 业务域描述
  creator: string; // 创建者
  products: string[]; // 关联产品
  create_time: string; // 创建时间
}

const useBusinessDomain = () => {
  const microWidgetProps = useMicroWidgetProps();
  const [businessDomainState, setBusinessDomainState] = useState<{
    publicBusinessDomain: BusinessDomain | undefined;
    currentBusinessDomain: BusinessDomain | undefined;
    allBusinessDomain: BusinessDomain[] | undefined;
    publicAndCurrentDomainIds: string[] | undefined; // 公共业务域和当前业务域ID组成的去重数组
    businessDomainMap: Record<string, BusinessDomain & { isCurrent?: boolean }> | undefined; // 业务域ID到业务域对象的映射
  }>({
    publicBusinessDomain: undefined,
    currentBusinessDomain: undefined,
    allBusinessDomain: undefined,
    publicAndCurrentDomainIds: undefined,
    businessDomainMap: undefined,
  });

  // 获取业务域
  const getBusinessDomain = useCallback(async () => {
    try {
      const allBusinessDomain: BusinessDomain[] = await cacheableGet('/api/business-system/v1/business-domain', {
        expires: 1000 * 60 * 5,
      });
      const publicBusinessDomain = allBusinessDomain.find(item => item.id === 'bd_public');
      const currentBusinessDomain = allBusinessDomain.find(item => item.id === microWidgetProps?.businessDomainID);

      // 构建公共业务域和当前业务域ID的去重数组
      const publicAndCurrentDomainIds = [
        ...new Set([
          ...(publicBusinessDomain?.id ? [publicBusinessDomain.id] : []),
          ...(currentBusinessDomain?.id ? [currentBusinessDomain.id] : []),
        ]),
      ];

      // 构建业务域ID到业务域对象的映射
      const businessDomainMap = allBusinessDomain.reduce(
        (acc, domain) => {
          if (domain.id) {
            if (domain.id === currentBusinessDomain?.id) {
              // @ts-ignore
              domain.isCurrent = true;
            }
            acc[domain.id] = domain;
          }
          return acc;
        },
        {} as Record<string, BusinessDomain & { isCurrent?: boolean }>
      );

      setBusinessDomainState({
        publicBusinessDomain,
        currentBusinessDomain,
        allBusinessDomain,
        publicAndCurrentDomainIds,
        businessDomainMap,
      });
    } catch {}
  }, [microWidgetProps?.businessDomainID]);

  useEffect(() => {
    getBusinessDomain();
  }, [getBusinessDomain]);

  return businessDomainState;
};

export default useBusinessDomain;
