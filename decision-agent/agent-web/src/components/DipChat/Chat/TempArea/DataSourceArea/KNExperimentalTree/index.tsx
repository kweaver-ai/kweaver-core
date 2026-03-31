import AdTree from '@/components/AdTree';
import { useEffect, useState } from 'react';
import { type AdTreeDataNode, adTreeUtils } from '@/utils/handle-function';
import { getKnExperimentDetailsById, getObjectTypeById } from '@/apis/knowledge-data';
import DipIcon from '@/components/DipIcon';
import classNames from 'classnames';
import styles from './index.module.less';
import { nanoid } from 'nanoid';

const KNExperimentalTree = ({ dataSource }: any) => {
  const { knowledge_network_id } = dataSource || {};
  const [treeProps, setTreeProps] = useState({
    treeData: [] as AdTreeDataNode[],
  });
  useEffect(() => {
    if (knowledge_network_id) {
      getTreeData();
    }
  }, []);

  const getTreeData = async () => {
    const res: any = await getKnExperimentDetailsById(knowledge_network_id);
    console.log('res', res);
    if (res) {
      const treeDataSource = adTreeUtils.createAdTreeNodeData([res], {
        keyField: () => nanoid(),
        titleField: (record: any) => (
          <div className="dip-flex-align-center" title={record.name}>
            <div style={{ background: record.color }} className={classNames(styles.icon, 'dip-flex-center')}>
              <DipIcon style={{ fontSize: 12 }} type={record.icon} />
            </div>
            <span style={{ whiteSpace: 'nowrap' }} className="dip-ml-8">
              {record.name}
            </span>
          </div>
        ),
        isLeaf: false,
      });
      setTreeProps(prevState => ({
        ...prevState,
        treeData: treeDataSource,
      }));
    }
  };

  const onLoadData = (nodeData: AdTreeDataNode) =>
    // eslint-disable-next-line no-async-promise-executor
    new Promise<any>(async resolve => {
      const parentId = nodeData.sourceData.id as string;
      const res = await getObjectTypeById(parentId);
      if (res) {
        const objectTypeTreeData = adTreeUtils.createAdTreeNodeData(res.entries, {
          keyField: () => nanoid(),
          titleField: (record: any) => (
            <div className="dip-flex-align-center" title={record.name}>
              <div style={{ background: record.color }} className={classNames(styles.icon, 'dip-flex-center')}>
                <DipIcon style={{ fontSize: 12 }} type={record.icon} />
              </div>
              <span style={{ whiteSpace: 'nowrap' }} className="dip-ml-8">
                {record.name}
              </span>
            </div>
          ),
          isLeaf: false,
          parentKey: nodeData.key as string,
          keyPath: nodeData.keyPath,
        });
        let treeDataSource = adTreeUtils.addTreeNode(treeProps.treeData, objectTypeTreeData);

        // 组合对象类的数据属性树节点
        objectTypeTreeData.forEach(objectTypeTreeNode => {
          const dataAttributes = objectTypeTreeNode.sourceData.data_properties ?? [];
          const dataAttributeTreeData = adTreeUtils.createAdTreeNodeData(dataAttributes, {
            titleField: (record: any) => (
              <div className="dip-flex-align-center" title={`${record.display_name} (${record.name})`}>
                {/*<div className={styles.dot} style={{ background: record.color }} />*/}
                <div className="dip-flex-item-full-width dip-ellipsis">
                  {record.display_name}
                  <span className="dip-ml-8">({record.name})</span>
                </div>
              </div>
            ),
            keyField: (record: any) => `${objectTypeTreeNode.key}-${record.name}`,
            isLeaf: true,
            parentKey: objectTypeTreeNode.key as string,
            keyPath: objectTypeTreeNode.keyPath,
          });
          treeDataSource = adTreeUtils.addTreeNode(treeDataSource, dataAttributeTreeData);
        });

        setTreeProps(prevState => ({
          ...prevState,
          treeData: treeDataSource,
        }));
      }
      resolve(true);
    });

  return <AdTree loadData={onLoadData as any} selectable={false} treeData={treeProps.treeData} />;
};

export default KNExperimentalTree;
