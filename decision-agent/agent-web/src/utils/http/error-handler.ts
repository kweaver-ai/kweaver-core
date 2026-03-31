import intl from 'react-intl-universal';
import { config } from './types';

export async function handleError({
  error,
  url,
  reject,
  isOffline,
}: {
  error: any;
  url: string;
  reject: (params: any) => void;
  isOffline?: boolean;
}) {
  const handleReject = (code: number | string) => {
    reject(code);
    return;
  };

  if (/\/v1\/(ping|profile|avatars|user\/get)/.test(url)) {
    handleReject(0);
    return;
  }

  if (isOffline) {
    config.toast?.warning(intl.get('error.networkError'));
    handleReject(0);
    return;
  }

  if (error.code === 'ECONNABORTED' && error.message === 'TIMEOUT') {
    config.toast?.warning(intl.get('error.timeoutError'));
    handleReject(0);
    return;
  }

  if (error.message === 'CANCEL') {
    handleReject('CANCEL');
    return;
  }

  if (!error.response) {
    config.toast?.warning(intl.get('error.serverError'));
    handleReject(0);
    return;
  }

  const { status, data } = error.response;

  if (status === 401 && config.onTokenExpired) {
    config.onTokenExpired(data?.code);
    handleReject(status);
    return;
  }

  if (status >= 500) {
    if(data?.description) {
      reject(data)
      return
    }
    const message = getServerErrorMsg(status);
    config.toast?.warning(message);
    handleReject(status);
    return;
  }

  reject(data);
}

export function getServerErrorMsg(status: number): string {
  return intl.get(`error.${status}`) || intl.get('error.serverError');
}
