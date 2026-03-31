import styles from './index.module.less';
import classNames from 'classnames';
import _ from 'lodash';
import { FileTypeIcon, getFileExtension } from '@/utils/doc';
import React from 'react';
import ScrollBarContainer from '@/components/ScrollBarContainer';
import { message } from 'antd';
const CitePanel = ({ citeIndex, sliceIndex, cites, onPreview }: any) => {
  const [messageApi, contextHolder] = message.useMessage();
  const renderContent = () => {
    if (citeIndex > -1) {
      const cite = cites[citeIndex];
      const pages = _.get(cite, ['slices', sliceIndex, 'pages']) ?? [];
      return (
        <div className={classNames(styles.container, 'dip-flex-column')}>
          <div
            onClick={() => {
              if (cite.ds_id === '0') {
                onPreview?.({
                  fileId: cite.object_id,
                  fileName: cite.doc_name,
                  fileExt: cite.ext_type,
                });
              } else {
                messageApi.info('外部数据源文件，暂不支持预览');
              }
            }}
            className="dip-flex-align-center dip-pl-10 dip-pr-10 dip-pb-8 dip-border-b dip-text-link dip-text-color-85"
          >
            <FileTypeIcon extension={getFileExtension(cite.ext_type)} fontSize={16} />
            <div className="dip-flex-item-full-width dip-ellipsis dip-ml-8" title={cite.doc_name}>
              {cite.doc_name}
            </div>
          </div>
          <ScrollBarContainer className="dip-flex-item-full-height dip-pl-10 dip-pr-10 dip-pt-8">
            <span className="dip-text-color-primary dip-mr-8">[第{pages.join('、')}页]</span>
            {cite.content}
          </ScrollBarContainer>
        </div>
      );
    }
  };
  return (
    <>
      {renderContent()}
      {contextHolder}
    </>
  );
};

export default CitePanel;
