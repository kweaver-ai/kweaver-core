/**
 * 业务域成员下拉选项
 */
import { FC, useEffect, useState, useRef, useMemo, useCallback } from 'react';
import { Select, message } from 'antd';
import intl from 'react-intl-universal';
import { useMicroWidgetProps } from '@/hooks';
import { getBusinessDomainMembers } from '@/apis/business-system';
import { BusinessDomainMember } from '@/apis/business-system/types';

interface BusinessDomainMemberSelectProps {
  filter?: (member: BusinessDomainMember) => boolean; // 过滤条件
  value?: BusinessDomainMember[];
  className?: string;
  onChange?: (selected: BusinessDomainMember[]) => void;
}

const Limit = 200;

const BusinessDomainMemberSelect: FC<BusinessDomainMemberSelectProps> = ({ filter, value, className, onChange }) => {
  const microWidgetProps = useMicroWidgetProps();
  const offsetRef = useRef(0); // 分页偏移量
  const totalRef = useRef(0); // 总记录数
  const loadingRef = useRef(false); // 是否加载中

  const defaultValue = useMemo(
    () =>
      value?.map(item => {
        const name =
          item.id === '00000000-0000-0000-0000-000000000000' ? intl.get('businessDomain.allOrgUsers') : item.name;
        return { label: name, value: item.id, type: item.type, details: { ...item, name } };
      }),
    []
  );

  const [list, setList] = useState<{ label: string; value: string; details: BusinessDomainMember }[]>([]);

  // 使用Map来快速查找，避免每次labelRender都遍历list
  const listMap = useMemo(() => {
    const map = new Map<string, { label: string; value: string; details: BusinessDomainMember }>();

    // 首先添加defaultValue中的项（优先级更高）
    defaultValue?.forEach(item => {
      if (item.value && !map.has(item.value)) {
        map.set(item.value, item);
      }
    });

    // 然后添加list中的项（可能会覆盖defaultValue中的重复项）
    list.forEach(item => {
      if (item.value) {
        map.set(item.value, item);
      }
    });

    return map;
  }, [list, defaultValue]);

  // 使用useCallback优化labelRender函数
  const labelRender = useCallback(
    (item: any) => {
      // 使用Map快速查找对应的项以获取type信息
      // 现在Map中同时包含list和defaultValue中的项，确保能找到所有类型信息
      const originalItem = listMap.get(item.value);
      return (
        <div className="dip-flex">
          <div className="dip-ellipsis" title={item.label}>
            {item.label}
          </div>
          {originalItem?.type && <span>{`(${typeMap[originalItem.type] || originalItem.type})`}</span>}
        </div>
      );
    },
    [listMap, typeMap]
  );

  // 使用useCallback优化optionRender函数
  const optionRender = useCallback(
    (item: any) => (
      <div className="dip-flex">
        <div className="dip-ellipsis" title={item.label}>
          {item.label}
        </div>
        <span style={{ color: '#999' }} className="dip-ml-8">
          {typeMap[item.data.type] || item.data.type}
        </span>
      </div>
    ),
    [typeMap]
  );

  const typeMap = useMemo(
    () => ({
      user: intl.get('businessDomain.user'),
      department: intl.get('businessDomain.department'),
      group: intl.get('businessDomain.group'),
      role: intl.get('businessDomain.role'),
      app: intl.get('businessDomain.app'),
    }),
    []
  );

  const loadData = async () => {
    try {
      loadingRef.current = true;
      const { items, total } = await getBusinessDomainMembers({
        bdid: microWidgetProps.businessDomainID,
        offset: offsetRef.current,
        limit: Limit,
      });
      totalRef.current = total;
      offsetRef.current = offsetRef.current + Limit;
      // 过滤数据
      const filteredItems = items.filter(filter || (() => true));
      setList(prev => [
        ...prev,
        ...filteredItems.map(item => {
          const name = item.id === '00000000-0000-0000-0000-000000000000' ? '所有组织用户' : item.name;
          return { label: name, value: item.id, type: item.type, details: { ...item, name } };
        }),
      ]);
    } catch (ex: any) {
      if (ex?.description) {
        message.error(ex.description);
      }
    } finally {
      loadingRef.current = false;
    }
  };

  const handleScroll = e => {
    const { target } = e;

    // 核心滚动计算逻辑：
    // scrollTop: 滚动条距离顶部的距离
    // offsetHeight: 可视区域的高度
    // scrollHeight: 整个滚动内容的高度
    // 这里的 10 是一个缓冲值，也就是距离底部还有 10px 时就开始加载，体验更好
    if (target.scrollTop + target.offsetHeight + 10 >= target.scrollHeight) {
      // 检查是否还有更多数据
      if (list.length < totalRef.current && !loadingRef.current) {
        loadData();
      }
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  return (
    <Select
      placeholder={intl.get('businessDomain.pleaseSelect')}
      labelInValue
      defaultActiveFirstOption={false}
      defaultValue={defaultValue}
      className={className}
      mode="multiple"
      options={list}
      onPopupScroll={handleScroll}
      onChange={(_, selected: any[]) => {
        onChange?.(selected.map(item => item.details));
      }}
      labelRender={labelRender}
      optionRender={optionRender}
    />
  );
};

export default BusinessDomainMemberSelect;
