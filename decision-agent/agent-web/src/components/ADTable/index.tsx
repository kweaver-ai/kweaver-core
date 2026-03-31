import React, { memo, useState } from 'react';

import AdTableCardList from '@/components/ADTable/CardList';
import Header from './Header';
import FilterOperationContainer from './FilterOperationContainer';
import ITable from './Table';

import { ADTableProps, DisplayTypeProps, FilterOperationContainerProps, HeaderProps, ITableProps } from './types';

/**
 * 表单模板, 包含标题，搜索框，过滤工具
 * @title --- null，不渲染标题
 * @showHeader --- false，不显示表头
 * @contextMenu --- 表头和表体的上下文菜单
 * @showFilter --- 是否显示筛选器, 默认不显示筛选器
 * @renderButtonConfig --- 渲染新建，刷新，排序等按钮，不配置时，默认显示搜索框或者筛选器
 * @FilterToolsOptions --- 渲染筛选器中的工具-下拉选择框，默认为空；不显示筛选器时，配置后只显示第一项
 * @onFilterClick --- 筛选按钮回调函数
 * @onSearchChange --- 搜索框回调函数
 * @searchText --- 搜索框placeholder
 * @persistenceID --- 持久化列宽，使用不同列宽时改变persistenceID值，否则使用同一列宽
 */
const ADtable: React.FC<ADTableProps> = props => {
  const {
    className = '',
    title = null,
    width = 'auto',
    showHeader = true,
    showTableHeader = true,
    // columns = [],
    dataSource: data,
    contextMenu = {
      headerContextMenu: <></>,
      bodyContextMenu: <></>,
    },
    showFilter = false,
    showSearch = true,
    defaultDisplayType = 'table',
    renderButtonConfig,
    filterToolsOptions,
    onFilterClick = () => {},
    searchValue,
    onSearchChange = () => {},
    searchPlaceholder = '',
    onFiltersToolsClose = () => {},
    persistenceID,
    children,
    extraContent = null,
    customBody = null,
    onDisplayTypeChange,
    onScrollCapture,
    showSorterTooltip = {},
    ...resetProps
  } = props;

  const [isFilter, setIsFilter] = useState(showFilter);
  const [displayType, setDisplayType] = useState<DisplayTypeProps>(defaultDisplayType);

  const onDisplayChange = (type: DisplayTypeProps) => {
    setDisplayType(type);
    onDisplayTypeChange?.(type);
  };

  return (
    <div className="dip-w-100 dip-h-100 dip-flex-column" data-testid={'testid-ADTable'}>
      {showHeader &&
        (children || (
          <>
            <Header
              title={title}
              showFilter={showFilter}
              showSearch={showSearch}
              filterConfig={{ isFilter, setIsFilter }}
              renderButtonConfig={renderButtonConfig}
              onFilterClick={onFilterClick}
              searchValue={searchValue}
              onSearchChange={onSearchChange}
              searchPlaceholder={searchPlaceholder}
              filterToolsOptions={filterToolsOptions}
              onDisplayChange={onDisplayChange}
              displayType={displayType}
            />
            <FilterOperationContainer
              showSearch={showSearch}
              visible={isFilter}
              filterConfig={{ isFilter, setIsFilter }}
              searchValue={searchValue}
              onSearchChange={onSearchChange}
              searchPlaceholder={searchPlaceholder}
              onClose={onFiltersToolsClose}
              filterToolsOptions={filterToolsOptions}
            />
          </>
        ))}
      {extraContent}
      {customBody ||
        (displayType === 'table' ? (
          <ITable
            persistenceID={persistenceID}
            width={width}
            title={() => {
              if (!showHeader) return null;
            }}
            dataSource={data}
            contextMenu={contextMenu}
            showHeader={showTableHeader}
            filterToolsOptions={filterToolsOptions}
            searchValue={searchValue}
            className={className}
            onScrollCapture={onScrollCapture}
            showSorterTooltip={showSorterTooltip}
            {...resetProps}
          />
        ) : (
          <AdTableCardList {...props} isFilter={isFilter} />
        ))}
    </div>
  );
};

export type ADtableCombineProps = React.FC<ADTableProps> & {
  Header: React.FC<HeaderProps>;
  FilterOperationContainer: React.FC<FilterOperationContainerProps>;
  ITable: React.FC<ITableProps>;
};

const ADTable = ADtable as ADtableCombineProps;

ADTable.Header = Header;
ADTable.FilterOperationContainer = FilterOperationContainer;
ADTable.ITable = ITable;

ADTable.displayName = 'ADTable';
export { Header, FilterOperationContainer, ITable };
export default memo(ADTable);
