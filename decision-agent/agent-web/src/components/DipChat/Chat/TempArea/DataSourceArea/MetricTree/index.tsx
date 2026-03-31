import { useState } from 'react';
import AdTree from '@/components/AdTree';
import { type AdTreeDataNode, adTreeUtils } from '@/utils/handle-function';
import DipIcon from '@/components/DipIcon';
import { getMetricInfoByIds } from '@/apis/data-model';
import { flatToTreeData } from '@/utils/handle-function';
import { nanoid } from 'nanoid';
import { useDeepCompareEffect } from '@/hooks';

const MetricTree = ({ dataSource = [] }: any) => {
  const [treeProps, setTreeProps] = useState({
    treeData: [] as AdTreeDataNode[],
  });
  const ids = dataSource.map((item: any) => item.metric_model_id);
  useDeepCompareEffect(() => {
    if (ids) {
      getMetricTree();
    }
  }, [ids]);
  const getMetricTree = async () => {
    const res = await getMetricInfoByIds({ ids });
    if (Array.isArray(res)) {
      let treeData: any[] = [];
      const noGroupData = res.filter((item: any) => !item.group_id);
      const hasGroupData = res.filter((item: any) => !!item.group_id);
      if (noGroupData.length > 0) {
        const data: any = {
          id: nanoid(),
          name: '未分组',
          children: noGroupData,
        };
        treeData = [...treeData, data];
      }
      if (hasGroupData.length > 0) {
        const data = flatToTreeData(res, {
          keyField: 'id',
          titleField: 'name',
          parentKeyField: 'group_id',
          parentTitleField: 'group_name',
        });
        treeData = [...treeData, ...data];
      }
      treeData = [
        {
          id: 'metric-tree',
          name: '指标',
          children: [...treeData],
        },
      ];
      const tree = adTreeUtils.createAdTreeNodeData(treeData, {
        keyField: 'id',
        titleField: 'name',
      });
      setTreeProps(prevState => ({
        ...prevState,
        treeData: tree,
      }));
    }
  };

  return (
    <AdTree
      // @ts-expect-error
      titleRender={(nodeData: AdTreeDataNode) => {
        let Icon: any = '';
        if (nodeData.keyPath!.length === 1) {
          Icon = <DipIcon type="icon-dip-color-zhibiaometirc" />;
        } else {
          if (!nodeData.isLeaf) {
            Icon = <DipIcon type="icon-dip-color-putongwenjianjia" />;
          }
        }
        return (
          <div className="dip-flex-align-center" title={nodeData.title}>
            {Icon}
            <span style={{ whiteSpace: 'nowrap' }} className="dip-ml-8">
              {nodeData.title}
            </span>
          </div>
        );
      }}
      selectable={false}
      treeData={treeProps.treeData}
    />
  );
};

export default MetricTree;
