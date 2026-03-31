import { Button } from 'antd';
import DipIcon from '@/components/DipIcon';
import React, { useEffect, useRef } from 'react';
import { FileTypeIcon } from '@/utils/doc';
import { apis, components } from '@aishu-tech/components/dist/dip-components.min';
import styles from './index.module.less';
import classNames from 'classnames';
type FilePreviewProps = {
  file: {
    fileId: string;
    fileName: string;
    fileExt: string;
  };
  onClose: () => void;
};

const FilePreview = ({ file: { fileId, fileName, fileExt }, onClose }: FilePreviewProps) => {
  const containerRef = useRef<HTMLDivElement | null>(null);
  useEffect(() => {
    const unmount = apis.mountComponent(
      components.Preview,
      {
        file: {
          objectId: fileId,
          name: fileName + fileExt,
        },
      },
      containerRef.current
    );

    return unmount;
  }, [fileId]);
  return (
    <div className={classNames(styles.container, 'dip-full dip-flex-column')}>
      <div className={classNames(styles.header, 'dip-flex-space-between')}>
        <div className="dip-flex-item-full-width dip-flex-align-center">
          <FileTypeIcon extension={fileExt} fontSize={20} />
          <div className="dip-ml-8 dip-flex-item-full-width dip-ellipsis">{fileName}</div>
        </div>
        <Button onClick={onClose} type="text" size="small" icon={<DipIcon type="icon-dip-close" />} />
      </div>
      <div className="dip-flex-item-full-height dip-w-100" ref={containerRef} />
    </div>
  );
};

export default FilePreview;
