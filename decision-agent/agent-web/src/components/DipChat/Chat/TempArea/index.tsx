import styles from './index.module.less';
import { Splitter } from 'antd';
import DataSourceArea from './DataSourceArea';
import FileArea from './FileArea';
import classNames from 'classnames';
import { useDipChatStore } from '@/components/DipChat/store';

const TempArea = () => {
  const {
    dipChatStore: { agentDetails, previewFile },
  } = useDipChatStore();
  const { data_source } = agentDetails?.config || {};
  const knExperimentalDataSource = data_source?.knowledge_network ?? [];
  const metricTreeDataSource = data_source?.metric ?? [];

  const renderContent = () => {
    // 只配置了配置了临时区
    if (knExperimentalDataSource.length === 0 && metricTreeDataSource.length === 0) {
      return <FileArea />;
    }

    return (
      <Splitter layout="vertical">
        <Splitter.Panel>
          <FileArea />
        </Splitter.Panel>
        <Splitter.Panel>
          <DataSourceArea />
        </Splitter.Panel>
      </Splitter>
    );
  };

  return (
    <div className={classNames(styles.container)} style={{ display: previewFile ? 'none' : 'block' }}>
      {renderContent()}
    </div>
  );
};

export default TempArea;
