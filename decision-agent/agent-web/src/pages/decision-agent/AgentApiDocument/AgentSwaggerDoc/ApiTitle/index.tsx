import { useRef, useState } from 'react';
import { message, Select } from 'antd';
import intl from 'react-intl-universal';
import { useSize, useMicroWidgetProps } from '@/hooks';
import TitleBox from '@/components/TitleBox';
import IconFont from '@/components/IconFont';
import { copyToBoard } from '@/utils/handle-function';
import { FilterEnum } from '../types';
import './style.less';
import Markdown from '@/components/Markdown';

const ApiTitle = ({ url, description, method, onUpdateFilter }: any) => {
  const [showDes, setShowDes] = useState<boolean>(false);
  const desRef = useRef<HTMLDivElement>(null);
  const containerSize = useSize(desRef);
  const microWidgetProps = useMicroWidgetProps();
  const theme = microWidgetProps?.config?.getTheme?.normal;

  const onCopy = () => {
    copyToBoard(url);
    message.success(intl.get('dataAgent.copyKeySuccess'));
  };

  return (
    <div className="rest-api-title-header">
      {/* 描述 */}
      {description && (
        <>
          <div style={{ maxHeight: showDes ? 'none' : 132, overflow: 'hidden' }}>
            <div ref={desRef} className="dip-mb-32">
              <Markdown value={description} readOnly />
            </div>
          </div>

          <div
            className="dip-pointer dip-w-100 "
            style={{ textAlign: 'center', display: containerSize?.height < 132 ? 'none' : 'block', color: theme }}
            onClick={() => setShowDes(!showDes)}
          >
            {intl.get(showDes ? 'dataAgent.apiDocument.unExpand' : 'dataAgent.apiDocument.showMore')}
            <IconFont
              type="icon-xiala"
              className="dip-ml-4"
              style={{ fontSize: 10, transform: showDes ? 'rotate(180deg)' : '' }}
            />
          </div>
        </>
      )}

      <div className="dip-mt-24">
        <TitleBox id="apiDoc" text={intl.get('dataAgent.apiDocument.apiDoc')} style={{ fontSize: 16 }} />
      </div>

      <div className="table-tbody">
        <div className="dip-flex"></div>
      </div>

      <div className="dip-flex-center api-header-box dip-mt-28">
        <Select
          options={[
            {
              label: intl.get('dataAgent.apiDocument.debugWithUserToken'),
              value: FilterEnum.User,
            },
            {
              label: intl.get('dataAgent.apiDocument.debugWithAppAccountToken'),
              value: FilterEnum.App,
            },
          ]}
          defaultValue={'user'}
          style={{
            width: 180,
          }}
          onChange={onUpdateFilter}
        />
        <div className="api-header-url dip-flex-align-center">
          <div className="api-get-post">{method}</div>
          <div className="api-url-style dip-ellipsis" title={url}>
            {url}
          </div>
          <div className="api-header-copy dip-flex-center dip-pointer" onClick={onCopy}>
            <IconFont type="icon-copy" />
          </div>
        </div>
      </div>
    </div>
  );
};

export default ApiTitle;
