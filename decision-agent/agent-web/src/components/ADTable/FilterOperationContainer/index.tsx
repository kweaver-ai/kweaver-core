import React, { useCallback, useMemo, useState } from 'react';
import { Form, Select } from 'antd';
import classnames from 'classnames';
import intl from 'react-intl-universal';
import _ from 'lodash';
import Format from '@/components/Format';
import SearchInput from '@/components/SearchInput';
import IconFont from '@/components/IconFont';

import { FilterOperationContainerProps } from '../types';
import './style.less';

const { Option } = Select;

const FilterOperationContainer: React.FC<FilterOperationContainerProps> = props => {
  const {
    visible = true,
    showSearch = true,
    className,
    filterConfig,
    children,
    searchValue,
    onSearchChange,
    searchPlaceholder = '',
    onClose = () => {},
    filterToolsOptions = [],
  } = props;
  const [_isFilter, _setIsFilter] = useState(visible);
  const { isFilter, setIsFilter } = filterConfig || { isFilter: _isFilter, setIsFilter: _setIsFilter };

  const [form] = Form.useForm();
  const moreOptions = filterToolsOptions;
  const moreOptionsObj: Record<string, string | number> = {};

  const [selector, setSelector] = useState<{ cur: number; perPageTotal: number }>({ cur: 1, perPageTotal: 3 });

  const { cur, perPageTotal } = selector;

  useMemo(() => {
    moreOptions.forEach(item => {
      const { optionList = [] } = item;
      moreOptionsObj[item.id] = optionList.length > 0 ? optionList[0].value : '';
    });
  }, [moreOptions]);

  // 筛选器工具配置
  const showOptions = useCallback(() => {
    const formInputValue = onHandleFormValue(moreOptions);
    form.setFieldsValue({ ...formInputValue });
    if (moreOptions.length <= perPageTotal) {
      const res = moreOptions?.map(item => {
        const { id, label, optionList = [], onHandle = () => {}, itemDom = null, showSearch = true } = item;
        return (
          <div
            key={String(label) + cur}
            style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}
          >
            {label ? (
              <span className="dip-ellipsis" style={{ flexShrink: 0, marginRight: 10 }} title={label}>
                {label}
              </span>
            ) : null}

            <Form.Item className="dip-mr-12" name={id}>
              {itemDom === null ? (
                <Select
                  className="select-box"
                  onChange={value => {
                    onHandle(value);
                  }}
                  style={{ width: 190 }}
                  showSearch={showSearch}
                  options={_.map(optionList, (item: any) => ({ label: item.text, value: item.value }))}
                  optionFilterProp="label"
                />
              ) : (
                itemDom
              )}
            </Form.Item>
          </div>
        );
      });
      return res;
    }
    const res = moreOptions.slice((cur - 1) * perPageTotal, cur * perPageTotal)?.map(item => {
      const { id, label, optionList = [], onHandle = () => {}, itemDom = null, showSearch = true } = item;
      return (
        <div
          key={String(label) + cur}
          style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}
        >
          {label ? (
            <span className="dip-ellipsis" style={{ flexShrink: 0, marginRight: 10 }} title={label}>
              {label}
            </span>
          ) : null}

          <Form.Item className="dip-mr-12" name={id}>
            {itemDom === null ? (
              <Select
                showSearch={showSearch}
                className="select-box"
                onChange={value => {
                  onHandle(value);
                }}
                style={{ width: 190 }}
                optionFilterProp="label"
                options={optionList.map((item: any) => ({ label: item.text, value: item.value }))}
              />
            ) : (
              itemDom
            )}
          </Form.Item>
        </div>
      );
    });
    const selector = (
      <div key={cur + ''} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Format.Button
          type="icon"
          className="dip-mr-12"
          onClick={() => {
            onPageChange('pre');
          }}
          disabled={cur === 1}
        >
          <IconFont type="icon-shangfanye" />
        </Format.Button>
        <Format.Button
          type="icon"
          className="dip-mr-12"
          onClick={() => {
            onPageChange('next');
          }}
          disabled={cur === Math.ceil(moreOptions.length / perPageTotal)}
        >
          <IconFont type="icon-fanye" />
        </Format.Button>
      </div>
    );
    res.push(selector);
    return res;
  }, [moreOptions, cur, perPageTotal]);

  /**
   * 筛选框填入值
   */
  const onHandleFormValue = (data: any) => {
    let result: any = {};
    _.map(_.cloneDeep(data), (item: any) => {
      result = { ...result, [item?.id]: item?.value };
    });
    return result;
  };

  // FiltersOperations翻页控制
  const onPageChange = (label: 'pre' | 'next') => {
    label === 'pre'
      ? setSelector(oldSelector => {
          if (oldSelector.cur > 1) {
            return { cur: oldSelector.cur - 1, perPageTotal: oldSelector.perPageTotal };
          }
          return oldSelector;
        })
      : setSelector(oldSelector => {
          if (oldSelector.cur < Math.ceil(moreOptions.length / perPageTotal)) {
            return { cur: oldSelector.cur + 1, perPageTotal: oldSelector.perPageTotal };
          }
          return oldSelector;
        });
  };

  // 处理搜索
  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    onSearchChange && onSearchChange(e?.target?.value);
  };

  return (
    isFilter &&
    visible && (
      <div className={classnames('ad-table-filter', className)}>
        {children || (
          <>
            {showSearch && isFilter ? (
              <SearchInput
                defaultValue={searchValue}
                value={searchValue}
                className="search-input"
                autoComplete={'off'}
                placeholder={searchPlaceholder || intl.get('global.search')}
                onChange={e => {
                  e.persist();
                  handleSearch(e);
                }}
                debounce
              />
            ) : (
              <div />
            )}
            <Form form={form} layout="inline" initialValues={{ search: '', ...moreOptionsObj }}>
              <div className="tools">
                {showOptions()}

                <Format.Button
                  type="icon"
                  className="close-btn dip-mr-12"
                  tip={intl.get('global.clearFilter')}
                  onClick={() => {
                    setIsFilter(false);
                    onClose();
                  }}
                >
                  <IconFont type="icon-guanbiquxiao" />
                </Format.Button>
              </div>
            </Form>
          </>
        )}
      </div>
    )
  );
};

export default FilterOperationContainer;
