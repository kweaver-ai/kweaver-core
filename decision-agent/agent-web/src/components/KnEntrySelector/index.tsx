import { memo, useState, useRef, useMemo, useCallback, useEffect } from 'react';
import intl from 'react-intl-universal';
import { debounce } from 'lodash';
import { message, Modal, Select, Spin } from 'antd';
import { getDataDicts } from '@/apis/data-model';
import Empty from '../Empty';
import LoadFailed from '../LoadFailed';
import styles from './index.module.less';
import classNames from 'classnames';

interface KnEntrySelectorProps {
  onCancel: () => void;
  onConfirm: (selectedItems: Array<{ kn_entry_id: string; name: string }>) => void;
}

enum LoadStatus {
  Loading = 'loading',
  Empty = 'empty',
  Normal = 'normal',
  Failed = 'failed',
  LoadingMore = 'loadingMore', // 鍔犺浇鏇村锛堝姞杞戒笅涓€椤电殑鏁版嵁鏃讹級
}

const limit = 20;

// 鐭ヨ瘑鏉＄洰閫夋嫨寮圭獥
const KnEntrySelector = ({ onCancel, onConfirm }: KnEntrySelectorProps) => {
  const offsetRef = useRef<number>(0);
  const searchKeyRef = useRef<string>('');
  const requestRef = useRef<any>(undefined);
  const isLoadingMoreRef = useRef<boolean>(false);
  // 鏄惁杩樻湁鏁版嵁鏈姞杞藉畬
  const hasMoreRef = useRef<boolean>(true);

  const [data, setData] = useState<any[]>([]);
  const [loadStatus, setLoadStatus] = useState<LoadStatus>(LoadStatus.Loading);
  const [selectedItems, setSelectedItems] = useState<Array<{ kn_entry_id: string; name: string }>>([]);

  // 鑾峰彇鐭ヨ瘑鏉＄洰鍒楄〃鏁版嵁
  const fetchData = useCallback(async () => {
    if (!hasMoreRef.current) return;

    try {
      // 鍙栨秷涓婁竴娆＄殑璇锋眰
      requestRef.current?.abort?.();
      // reset request
      requestRef.current = getDataDicts({
        offset: offsetRef.current,
        limit,
        name_pattern: searchKeyRef.current,
      });
      const { entries } = await requestRef.current;
      // 璁剧疆offset
      offsetRef.current += entries.length;
      // 璁剧疆hasMore
      hasMoreRef.current = entries.length === limit;

      if (isLoadingMoreRef.current) {
        // 鍦ㄧ幇鏈夌殑鏁版嵁鍚庨潰娣诲姞
        setData(prev => [...prev, ...entries.map(({ id, name }) => ({ value: name, label: name, id, name }))]);
      } else {
        setData(entries.map(({ id, name }) => ({ value: name, label: name, id, name })));
        setLoadStatus(entries.length ? LoadStatus.Normal : LoadStatus.Empty);
      }

      isLoadingMoreRef.current = false;
    } catch (ex: any) {
      if (ex === 'CANCEL') return;

      if (ex?.description) {
        message.error(ex.description);
      }
      if (!isLoadingMoreRef.current) {
        // 鍙湁绗竴娆″姞杞界殑鏁版嵁锛屾墠璁剧疆loadStatus锛屾墠璁剧疆data涓虹┖
        setLoadStatus(LoadStatus.Failed);
        setData([]);
      }

      isLoadingMoreRef.current = false;
    }
  }, []);

  const debounceFetchData = useMemo(() => debounce(fetchData, 300), [fetchData]);

  const handleSearch = (value: string) => {
    // 璁剧疆鎼滅储鍏抽敭瀛楋紝娓呯┖offset锛岃缃甽oadStatus锛屾竻绌篸ata
    searchKeyRef.current = value;
    offsetRef.current = 0;
    hasMoreRef.current = true;
    isLoadingMoreRef.current = false;
    setLoadStatus(LoadStatus.Loading);
    setData([]);
    debounceFetchData();
  };

  const handleScroll = (e: any) => {
    if (isLoadingMoreRef.current || !hasMoreRef.current) return;

    const { scrollTop, scrollHeight, clientHeight } = e.target;
    const distanceFromBottom = scrollHeight - (scrollTop + clientHeight);

    if (distanceFromBottom < 100) {
      isLoadingMoreRef.current = true;
      debounceFetchData();
    }
  };

  useEffect(() => {
    debounceFetchData();
  }, []);

  return (
    <Modal
      open
      centered
      title={intl.get('dataAgent.selectKnowledgeEntry')}
      maskClosable={false}
      okButtonProps={{ disabled: !selectedItems.length, className: 'dip-min-width-72' }}
      cancelButtonProps={{ className: 'dip-min-width-72' }}
      onCancel={onCancel}
      onOk={() => onConfirm(selectedItems)}
      footer={(_, { OkBtn, CancelBtn }) => (
        <>
          <OkBtn />
          <CancelBtn />
        </>
      )}
    >
      <Select
        showSearch
        defaultActiveFirstOption={false}
        mode="multiple"
        placeholder={intl.get('dataAgent.pleaseSelectKnowledgeEntry')}
        className={classNames('dip-w-100 dip-mt-16 dip-mb-16', styles['select'])}
        options={data}
        notFoundContent={
          loadStatus === LoadStatus.Empty ? (
            <Empty className="dip-pt-20 dip-pb-20" />
          ) : loadStatus === LoadStatus.Failed ? (
            <LoadFailed className="dip-pt-20 dip-pb-20" />
          ) : loadStatus === LoadStatus.Loading ? (
            <div style={{ height: 80 }} className="dip-flex-center">
              <Spin />
            </div>
          ) : null
        }
        onSearch={handleSearch}
        onSelect={(_, { id, name }) => {
          // 閫変腑
          setSelectedItems(prev => [...prev, { kn_entry_id: id, name }]);
        }}
        onDeselect={(_, { id }) => {
          // 鍙栨秷閫変腑
          setSelectedItems(prev => prev.filter(selected => selected.kn_entry_id !== id));
        }}
        onFocus={() => {
          // 鑱氱劍鏃讹紝濡傛灉涓婁竴娆℃悳绱㈠€间笉涓虹┖锛屽垯瑙﹀彂鎼滅储
          if (searchKeyRef.current || loadStatus === LoadStatus.Failed) {
            handleSearch('');
          }
        }}
        onPopupScroll={handleScroll}
      />
    </Modal>
  );
};

export default memo(KnEntrySelector);


