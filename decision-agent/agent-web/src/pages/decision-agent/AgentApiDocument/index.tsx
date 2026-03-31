import { useEffect, useMemo, useState } from 'react';
import intl from 'react-intl-universal';
import { useSearchParams } from 'react-router-dom';
import { keys, isEmpty, values } from 'lodash';
import { getHttpBaseUrl } from '@/utils/http';
import { getAPIDoc } from '@/apis/agent-app';
import AgentSwaggerDoc from './AgentSwaggerDoc';
import styles from './styles.module.css';

const AgentApiDocument = () => {
  const [searchParams] = useSearchParams();
  const agent_id = searchParams.get('id');
  const agent_name = searchParams.get('name');
  const agent_version = searchParams.get('version');
  const [apiDocData, setApiDocData] = useState<any>({});

  useEffect(() => {
    if (agent_id && agent_version) {
      fetchApiDoc({ agent_id, agent_version });
    }
  }, []);

  const fetchApiDoc = async ({ agent_id, agent_version }: { agent_id: string; agent_version: string }) => {
    const res = await getAPIDoc({
      app_key: agent_id,
      agent_id,
      agent_version,
    });

    if (res) {
      const apiPath = keys(res.paths)[0];
      res.paths[apiPath].post.parameters.forEach((paramItem: any) => {
        if (paramItem.name === 'as-user-ip') {
          paramItem.description = intl.get('dataAgent.apiDocument.asAuth');
        }
        if (paramItem.name === 'as-user-id') {
          paramItem.description = intl.get('dataAgent.apiDocument.asAuth');
        }
        if (paramItem.name === 'as-visitor-type') {
          paramItem.description = intl.get('dataAgent.apiDocument.asAuth');
        }
        if (paramItem.name === 'as-client-type') {
          paramItem.description = intl.get('dataAgent.apiDocument.asAuth');
        }
        if (paramItem.name === 'app_key') {
          paramItem.example = agent_id;
        }
      });
      setApiDocData(res);
    }
  };

  const method = useMemo(() => {
    if (!isEmpty(apiDocData)) {
      const apiPath = values(apiDocData.paths)[0];
      const method = keys(apiPath)[0];
      return method.toUpperCase();
    }
    return '';
  }, [apiDocData]);

  const url = useMemo(() => {
    if (!isEmpty(apiDocData)) {
      const apiPath = keys(apiDocData.paths)[0];
      const httpBaseUrl = getHttpBaseUrl();

      return `${httpBaseUrl}${apiPath.replace('{app_key}', agent_id!)}`;
    }
    return '';
  }, [apiDocData]);

  const apiBasePath = useMemo(() => {
    if (!isEmpty(apiDocData)) {
      return keys(apiDocData.paths)[0];
    }
    return '';
  }, [apiDocData]);

  return (
    !isEmpty(apiDocData) && (
      <div className={styles['container']}>
        <AgentSwaggerDoc
          title={agent_name as string}
          id={agent_id as string}
          description={apiDocData.info.description}
          method={method}
          url={url}
          apiDocData={apiDocData}
          apiBasePath={apiBasePath}
        />
      </div>
    )
  );
};

export default AgentApiDocument;
