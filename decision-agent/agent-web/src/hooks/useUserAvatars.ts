import { useState, useEffect, useRef } from 'react';
import { getUserAvatarsByIds } from '../apis/user-management';

interface UserAvatarsMap {
  [userId: string]: string;
}

function arrayToObjectById(arr) {
  return arr.reduce((acc, item) => {
      acc[item.id] = item?.avatar_url || undefined;
      return acc;
  }, {});
}

/**
 * 批量获取用户头像的hook，支持懒加载
 * @param initialUserIds 初始用户ID数组
 * @param batchSize 每批加载的数量，默认20
 * @returns { userAvatars, addUserIds, loading }
 */
export function useUserAvatars(initialUserIds: string[] = [], batchSize: number = 20) {
  const [userAvatars, setUserAvatars] = useState<UserAvatarsMap>({});
  const [loading, setLoading] = useState(false);
  const [pendingUserIds, setPendingUserIds] = useState<string[]>([]);
  const processedUserIds = useRef<Set<string>>(new Set());
  const batchTimer = useRef<NodeJS.Timeout | null>(null);

  // 添加需要获取头像的用户ID
  const addUserIds = (userIds: string[]) => {
    if (!userIds || userIds.length === 0) return;

    // 过滤掉已处理的用户ID
    const newUserIds = userIds.filter(id => !processedUserIds.current.has(id));
    if (newUserIds.length === 0) return;

    // 添加到待处理队列
    setPendingUserIds(prev => [...prev, ...newUserIds]);
    // 标记为已处理
    newUserIds.forEach(id => processedUserIds.current.add(id));
  };

  // 处理批量获取头像
  useEffect(() => {
    if (pendingUserIds.length === 0) return;

    // 清除之前的定时器
    if (batchTimer.current) {
      clearTimeout(batchTimer.current);
    }

    // 设置延迟批处理，避免频繁请求
    batchTimer.current = setTimeout(() => {
      // 获取当前批次要处理的用户ID
      const batchUserIds = pendingUserIds.slice(0, batchSize);
      if (batchUserIds.length > 0) {
        // 去重，确保不请求重复的ID
        const uniqueUserIds = Array.from(new Set(batchUserIds));
        setLoading(true);
        getUserAvatarsByIds(uniqueUserIds)
          .then(res => {
            if (res?.length) {
              setUserAvatars(prev => ({
                ...prev,
                ...arrayToObjectById(res),
              }));
            }
          })
          .catch(error => {
            console.error('Failed to fetch user avatars:', error);
          })
          .finally(() => {
            setLoading(false);
            // 移除已处理的ID
            setPendingUserIds(prev => prev.slice(batchSize));
          });
      }
    }, 100);

    return () => {
      if (batchTimer.current) {
        clearTimeout(batchTimer.current);
      }
    };
  }, [pendingUserIds, batchSize]);

  // 初始化时加载提供的用户ID
  useEffect(() => {
    if (initialUserIds && initialUserIds.length > 0) {
      // 初始化时也去重
      const uniqueInitialIds = Array.from(new Set(initialUserIds));
      addUserIds(uniqueInitialIds);
    }
  }, []);

  return { userAvatars, addUserIds, loading };
}
