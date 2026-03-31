import { useEffect, useState, useRef } from 'react';
import SwaggerUI from 'swagger-ui-react';
import { getCommonHttpHeaders } from '@/utils/http';

import _ from 'lodash';
import { getComponentEventEmitter, ComponentIDEnum } from '@/utils/event-bus';
import Parameters from './Parameters';
import OperationContainer from './core/containers/OperationContainer';
import Operations from './core/components/operations';
import Responses from './core/components/responses';
import ApiTitle from './ApiTitle';
import { FilterEnum } from './types';
import './swagger-ui.css';
import './style.less';

const plugins = [
  () => ({
    components: {
      Fallback: () => null, // 去除错误
      parameters: Parameters,
      Operations,
      OperationContainer,
      responses: Responses,
    },
  }),
];

export type AdSwaggerDocPros = {
  // 数据的意思是  根据那条数据生成的restful API 文档
  title: string; // 数据的名称
  id: string; // 数据的id
  description: string; // api的描述
  method: string; // 请求方式
  url: string; // restful URL
  apiDocData?: any; // 后端接口返回的api文档数据
  apiBasePath: string;
};

const AdSwaggerDoc = (props: AdSwaggerDocPros) => {
  const { title, id, description, method, url, apiDocData, apiBasePath } = props;
  const filterRef = useRef<FilterEnum>(FilterEnum.User);

  const [apiData, setApiData] = useState<any>({});

  useEffect(() => {
    const newData = _.cloneDeep(apiDocData);
    const newApiBasePath = apiBasePath.replace('{app_key}', id);

    const value = newData.paths[apiBasePath];
    delete newData.paths[apiBasePath];
    newData.paths[newApiBasePath] = value;
    setApiData(newData);
  }, [apiDocData]);

  return (
    <div className="l-swagger-box">
      <div className="api-title-box">
        <div className="api-title dip-ellipsis dip-mb-4" title={title}>
          {title || '--'}
        </div>
        <span className="dip-c-subtext" style={{ fontSize: 12 }}>
          ID：{id || '--'}
        </span>
      </div>
      <div className="apiDocContent">
        <div className="id-name-des-box">
          <ApiTitle
            description={description}
            method={method}
            url={url}
            title={title}
            onUpdateFilter={(filter: FilterEnum) => {
              const emitter = getComponentEventEmitter(ComponentIDEnum.APIDoc);
              emitter.emit('changeFilter', { filter });
              filterRef.current = filter;
            }}
          />

          {!_.isEmpty(apiData) && (
            <SwaggerUI
              plugins={plugins}
              spec={apiData}
              requestInterceptor={req => {
                const headers = getCommonHttpHeaders();
                req.headers['Accept-Language'] = headers['Accept-Language'];
                req.headers['x-Language'] = headers['Accept-Language'];
                req.headers['x-business-domain'] = headers['x-business-domain'];
                // 根据选中的filter是 以用户token调试还是以应用账号token调试，决定Authorization的值
                req.headers['Authorization'] =
                  filterRef.current === FilterEnum.User
                    ? headers['Authorization']
                    : document.querySelector('input[placeholder="Authorization"]')?.value;
                return req;
              }}
            />
          )}
        </div>
      </div>
    </div>
  );
};

export default AdSwaggerDoc;
