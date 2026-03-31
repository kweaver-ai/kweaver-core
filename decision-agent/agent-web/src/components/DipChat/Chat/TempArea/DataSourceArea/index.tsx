import styles from './index.module.less';
import classNames from 'classnames';
import ScrollBarContainer from '@/components/ScrollBarContainer';
import { useDipChatStore } from '@/components/DipChat/store';
import MetricTree from './MetricTree';
import _ from 'lodash';
import KNExperimentalTree from './KNExperimentalTree';
import { useEffect } from 'react';
const DataSourceArea = () => {
  const {
    dipChatStore: { agentDetails },
    setDipChatStore,
  } = useDipChatStore();
  const { data_source } = agentDetails?.config || {};
  const knExperimentalDataSource = data_source?.knowledge_network ?? [];
  const metricTreeDataSource = data_source?.metric ?? [];

  useEffect(() => {
    setDipChatStore({ tempAreaOpen: true });
  }, []);

  const renderKNExperimentalTree = () => {
    return knExperimentalDataSource.map((item: any) => (
      <KNExperimentalTree key={item.knowledge_network_id} dataSource={item} />
    ));
  };

  const renderContent = () => {
    return (
      <div>
        {renderKNExperimentalTree()}
        {!_.isEmpty(metricTreeDataSource) && <MetricTree dataSource={metricTreeDataSource} />}
      </div>
    );
  };
  return (
    <div className={classNames(styles.container, 'dip-flex-column')}>
      <div className="dip-pl-12 dip-pr-12">
        <span className="dip-font-weight-700">知识来源</span>
      </div>
      <ScrollBarContainer className="dip-flex-item-full-height dip-pl-12 dip-pr-12 dip-pt-8">
        {renderContent()}
      </ScrollBarContainer>
    </div>
  );
};

export default DataSourceArea;
