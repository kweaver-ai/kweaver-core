import React, { useState, useRef, useEffect } from 'react';
import intl from 'react-intl-universal';
import { Modal, Select, TreeSelect, Empty, Spin } from 'antd';
import { DownOutlined, LoadingOutlined } from '@ant-design/icons';
import _ from 'lodash';
import { FileTypeIcon } from '@/utils/doc';
import { useMicroWidgetProps } from '@/hooks';
import DataSource from '@/assets/icons/data-source.svg';
import { getDocsSourceListByDsId } from '@/apis/knowledge-data';
import { getDocsSourceList } from '@/apis/data-connection';
import { getInFoByPath, getDirList } from '@/apis/agent-factory';
import './style.less';

interface AnyShareSelectorProps {
  visible: boolean;
  onCancel: () => void;
  onConfirm: (selectedData: {
    dataSource: { id: string; name: string };
    selectedFiles: Array<{ name: string; path: string; source: string }>;
  }) => void;
}

const antIcon = <LoadingOutlined style={{ fontSize: 24 }} spin />;

const AnyShareSelector: React.FC<AnyShareSelectorProps> = ({ visible, onCancel, onConfirm }) => {
  const microWidgetProps = useMicroWidgetProps();

  const [loading, setLoading] = useState(false);
  const [treeLoading, setTreeLoading] = useState(false);
  const [asDataSource, setAsDataSource] = useState<any[]>([]);
  const [selectedDataSource, setSelectedDataSource] = useState<{
    id: string;
    name: string;
    binData?: any;
  } | null>(null);
  const [treeData, setTreeData] = useState<any[]>([]);
  const [treeExpandedKeys, setTreeExpandedKeys] = useState<any[]>([]);
  const [selectedFiles, setSelectedFiles] = useState<string[]>([]);
  const [isProcessing, setIsProcessing] = useState(false);

  const treeRef = useRef<any>(null);

  // 获取AS数据源列表
  const getAsSourceList = async () => {
    try {
      setLoading(true);
      const { entries } = await getDocsSourceList({
        limit: -1,
        offset: 0,
        keyword: '',
        type: '',
        sort: 'created_at',
        direction: 'desc',
      });

      // 过滤数据源的条件
      const filterConditions = ['as', 'anyshare7'];
      const authData = _.filter(entries, item => _.includes(filterConditions, item?.type));
      setAsDataSource(authData);
      setLoading(false);
    } catch {
      setLoading(false);
    }
  };

  /* 使用指定的连接信息获取目录列表 */
  async function getListDirWithConfig(
    binData: {
      connect_protocol: string;
      host: string;
      port: number;
      account: string;
      password: string;
    },
    docid: string
  ) {
    try {
      const { connect_protocol, host, port, account, password } = binData || {};
      const result = await getDirList({ protocol: connect_protocol, host, port, account, password, docid });
      result.dirs = _.map(result.dirs, item => ({ ...item, id: item.docid }));
      result.files = _.map(result.files, item => ({ ...item, id: item.docid }));
      return result;
    } catch (error) {
      console.error('获取目录列表失败:', error);
      throw error;
    }
  }

  // 获取文档库列表
  const getDocLibsList = async (binData: any) => {
    try {
      setTreeLoading(true);
      try {
        const { connect_protocol, host, port, account, password, storage_base } = binData || {};
        const postData = { protocol: connect_protocol, host, port, account, password, namepath: storage_base };
        const { docid, name, size } = await getInFoByPath(postData);
        if (!docid) return;
        const updateTreeData = handleFileTreeData({ name, id: docid, size });
        setTreeData([updateTreeData]);
      } catch (error: any) {
        console.log('error', error);
      }
      setTreeLoading(false);
    } catch {
      setTreeData([]);
      setTreeLoading(false);
    }
  };

  // 数据源选择变更
  const onDataSourceChange = async (value: string) => {
    const selectedSource = asDataSource.find(item => item.name === value);

    if (selectedSource) {
      setSelectedDataSource({
        id: selectedSource.id.toString(),
        name: selectedSource.dsname,
        binData: selectedSource.bin_data,
      });
      // 清空之前的文件选择
      setSelectedFiles([]);
      setTreeData([]);
      setTreeExpandedKeys([]);

      // 如果有bin_data信息，调用efast接口获取文档库
      if (selectedSource.bin_data) {
        getDocLibsList(selectedSource.bin_data);
      } else {
        // 否则使用原有逻辑获取文件列表
        getFileListByDataSource(selectedSource.id.toString());
      }
    }
  };

  // 获取文件列表
  const getFileListByDataSource = async (dsId: string) => {
    try {
      setTreeLoading(true);
      const param = {
        ds_id: dsId,
        data_source: 'as7',
        postfix: 'all',
      };
      const response = await getDocsSourceListByDsId(param);
      const responseData = response?.res || {};

      const updateTreeData = _.map(responseData?.output, (i: any) => handleFileTreeData(i));
      setTreeData(updateTreeData);
      setTreeLoading(false);
    } catch {
      setTreeData([]);
      setTreeLoading(false);
    }
  };

  // 处理文件树数据
  const handleFileTreeData = (data: any, parentData?: any) => {
    const size = data.size || -1;
    return {
      label: `${parentData?.label ? `${parentData?.label}/${data.name}` : data.name}`,
      title: (
        <span className="anyshare-selector-file-item">
          <FileTypeIcon name={data.name} size={size} className="dip-mr-6" />
          <span className="file-name" title={data.name}>
            {data.name}
          </span>
        </span>
      ),
      key: `${parentData?.label ? `${parentData?.label}/${data.name}` : data.name}`,
      value: `${parentData?.label ? `${parentData?.label}/${data.name}` : data.name}`,
      fileKey: data.id,
      type: data.type,
      checkable: true,
      size: data.size,
      isLeaf: size !== -1,
    };
  };

  // 加载子节点数据
  const onLoadData = (data: any, currentTreeData?: any) => {
    return new Promise(resolve => {
      if (!selectedDataSource || !selectedDataSource.binData) {
        resolve(undefined);
        return;
      }

      getListDirWithConfig(selectedDataSource.binData, data.fileKey)
        .then((response: any) => {
          if (!_.isEmpty(response?.dirs) || !_.isEmpty(response?.files)) {
            // 合并dirs和files数组
            const allItems = [...(response?.dirs || []), ...(response?.files || [])];
            const childrenTreeData = _.map(allItems, (i: any) => handleFileTreeData(i, data));
            const result = handleChildren(childrenTreeData, data, currentTreeData);
            setTreeData(result);
          }
          resolve(undefined);
        })
        .catch(() => {
          resolve(undefined);
        });
    });
  };

  // 为父节点添加children
  const handleChildren = (childrenTreeData: any, data: any, currentTreeData?: any) => {
    const result = _.cloneDeep(currentTreeData || treeData);
    const loop = (value: any) => {
      _.map(value, (i: any) => {
        if (!_.isEmpty(i?.children)) {
          loop(i?.children);
        } else if (i?.fileKey === data?.fileKey) {
          i.children = childrenTreeData;
        }
      });
    };
    loop(result);
    return result;
  };

  // 文件选择变更
  const onTreeChange = (value: any) => {
    if (isProcessing) return;
    setIsProcessing(true);

    try {
      setSelectedFiles(value || []);
    } finally {
      setIsProcessing(false);
    }
  };

  // 展开树节点
  const onTreeExpand = (keys: any) => {
    setTreeExpandedKeys(keys);
  };

  // 确认选择
  const handleConfirm = () => {
    if (!selectedDataSource || selectedFiles.length === 0) {
      return;
    }

    const selectedFileData = selectedFiles.map(path => {
      const splitPath = path.split('/');
      const name = splitPath[splitPath.length - 1];

      // 找到对应的节点以获取fileKey
      const findNodeByPath = (nodes: any[], targetPath: string): any => {
        for (const node of nodes) {
          if (node.key === targetPath) {
            return node;
          }
          if (node.children) {
            const found = findNodeByPath(node.children, targetPath);
            if (found) return found;
          }
        }
        return null;
      };

      const targetNode = findNodeByPath(treeData, path);
      const source = targetNode?.fileKey || '';

      return {
        name,
        path,
        source,
        ds_id: selectedDataSource?.id,
        type: (targetNode?.size || -1) === -1 ? 'folder' : 'file',
      };
    });

    onConfirm({
      dataSource: selectedDataSource,
      selectedFiles: selectedFileData,
    });

    // 重置状态
    setSelectedDataSource(null);
    setSelectedFiles([]);
    setTreeData([]);
    setTreeExpandedKeys([]);
  };

  // 取消操作
  const handleCancel = () => {
    setSelectedDataSource(null);
    setSelectedFiles([]);
    setTreeData([]);
    setTreeExpandedKeys([]);
    onCancel();
  };

  // 初始化时获取数据源列表
  useEffect(() => {
    if (visible) {
      getAsSourceList();
    }
  }, [visible]);

  return (
    <Modal
      title={intl.get('dataAgent.config.selectAnyShare')}
      maskClosable={false}
      centered
      open={visible}
      onCancel={handleCancel}
      onOk={handleConfirm}
      width={600}
      okButtonProps={{
        disabled: !selectedDataSource || selectedFiles.length === 0,
        className: 'dip-min-width-72',
      }}
      cancelButtonProps={{ className: 'dip-min-width-72' }}
      footer={(_, { OkBtn, CancelBtn }) => (
        <>
          <OkBtn />
          <CancelBtn />
        </>
      )}
    >
      <div className="anyshare-selector-content">
        {/* 数据源名称选择 */}
        <div className="form-item">
          <div className="form-label">{intl.get('dataAgent.config.dataSourceName')}</div>
          <Select
            className="form-control"
            placeholder={intl.get('dataAgent.config.pleaseSelect')}
            // value={selectedDataSource?.name}
            onChange={onDataSourceChange}
            allowClear
            loading={loading}
            getPopupContainer={triggerNode => triggerNode?.parentElement || document.body}
            notFoundContent={
              loading ? (
                <Spin size="small" indicator={antIcon} />
              ) : (
                <Empty
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                  description={intl.get('dataAgent.noData')}
                  // description={
                  //   <span>
                  //     暂无数据，前往「
                  //     <span
                  //       className="dip-text-blue-link"
                  //       onClick={() => {
                  //         microWidgetProps?.history.navigateToMicroWidget({
                  //           name: 'data-source',
                  //           isNewTab: true,
                  //         });
                  //       }}
                  //     >
                  //       数据连接
                  //     </span>
                  //     」添加
                  //   </span>
                  // }
                />
              )
            }
          >
            {asDataSource.map(source => (
              <Select.Option key={source.id} value={source.name}>
                <div className="data-source-option">
                  <DataSource className="source-icon" />
                  <span className="source-name" title={source.name}>
                    {source.name}
                  </span>
                </div>
              </Select.Option>
            ))}
          </Select>
        </div>

        {/* 数据选择 */}
        <div className="form-item">
          <div className="form-label">{intl.get('dataAgent.config.data')}</div>
          <TreeSelect
            ref={treeRef}
            className="form-control tree-select"
            treeLine={{
              showLeafIcon: false,
            }}
            switcherIcon={<DownOutlined />}
            treeNodeLabelProp="label"
            showCheckedStrategy="SHOW_CHILD"
            getPopupContainer={triggerNode => triggerNode?.parentElement || document.body}
            showArrow={true}
            showSearch={false}
            treeCheckable={true}
            allowClear={true}
            value={selectedFiles.length === 0 ? undefined : selectedFiles}
            treeData={treeData}
            loadData={onLoadData}
            placeholder={intl.get('dataAgent.config.selectData')}
            onDropdownVisibleChange={(open: boolean) => {
              if (!open) {
                setTreeExpandedKeys([]);
                treeRef.current?.blur();
              } else {
                setTreeData([]);
                setTreeExpandedKeys([]);
                getDocLibsList(selectedDataSource?.binData || {});
              }
            }}
            treeExpandedKeys={treeExpandedKeys}
            onTreeExpand={onTreeExpand}
            onChange={onTreeChange}
            maxTagCount={3}
            notFoundContent={
              treeLoading ? (
                <Spin size="small" indicator={antIcon} />
              ) : (
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={intl.get('dataAgent.noData')} />
              )
            }
          />
        </div>
      </div>
    </Modal>
  );
};

export default AnyShareSelector;
