import React from 'react';
import styles from './index.module.less';
// import { CloseOutlined } from '@ant-design/icons';
import classNames from 'classnames';
import ScrollBarContainer from '@/components/ScrollBarContainer';
// import DipButton from '@/components/DipButton';
import { message } from 'antd';
// import DipIcon from '@/components/DipIcon';
import { FileTypeIcon, getFileExtension } from '@/utils/doc';

const CitesList = ({ onPreview, cites }: any) => {
  const [messageApi, contextHolder] = message.useMessage();

  return (
    <div className={classNames(styles.container, 'dip-full dip-flex-column')}>
      {/* <div className="dip-flex-space-between dip-pl-24 dip-pr-24">*/}
      {/*  <span className="dip-flex-align-center dip-font-16 dip-text-color dip-font-weight-700">*/}
      {/*    <DipIcon style={{ fontSize: 26 }} type="icon-dip-color-doc-qa" />*/}
      {/*    <span className="dip-ml-8 dip-flex-item-full-width dip-ellipsis dip-font-16 dip-text-color dip-font-weight-700">*/}
      {/*      引用文档*/}
      {/*      <span className="dip-ml-8">{cites.length}</span>*/}
      {/*    </span>*/}
      {/*  </span>*/}
      {/*  <DipButton size="small" type="text" onClick={onClose}>*/}
      {/*    <CloseOutlined className="dip-text-color-45 dip-font-16" />*/}
      {/*  </DipButton>*/}
      {/* </div>*/}
      <ScrollBarContainer className="dip-flex-item-full-height dip-pl-16 dip-pr-16 dip-pb-16">
        {cites.map((item: any) => (
          <div
            className={classNames(styles.item)}
            key={item.doc_id}
            onClick={() => {
              if (item.ds_id === '0') {
                onPreview?.({
                  fileId: item.object_id,
                  fileName: item.doc_name,
                  fileExt: item.ext_type,
                });
              } else {
                messageApi.info('外部数据源文件，暂不支持预览');
              }
            }}
          >
            <div className="dip-flex-align-center">
              <FileTypeIcon extension={getFileExtension(item.ext_type)} fontSize={16} />
              <div className="dip-flex-item-full-width dip-ellipsis dip-ml-8" title={item.doc_name}>
                {item.doc_name}
              </div>
            </div>
            <div title={item.content} className="dip-mt-12 dip-ellipsis-2 dip-font-12 dip-text-color-45">
              {item.content}
            </div>
          </div>
        ))}
      </ScrollBarContainer>
      {contextHolder}
    </div>
  );
};

export default CitesList;
