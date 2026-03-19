import React, {useState, useEffect, useCallback} from 'react';
import BlogPostItem from '@theme-original/BlogPostItem';
import {useBlogPost} from '@docusaurus/plugin-content-blog/client';
import {useColorMode} from '@docusaurus/theme-common';
import Giscus from '@giscus/react';

function LikeButton() {
  const [count, setCount] = useState<number | null>(null);
  const [discussionUrl, setDiscussionUrl] = useState<string | null>(null);

  useEffect(() => {
    function handleMessage(event: MessageEvent) {
      if (event.origin !== 'https://giscus.app') return;
      const data = event.data?.giscus;
      if (!data?.discussion) return;
      const total = data.discussion.reactions?.totalCount;
      if (typeof total === 'number') setCount(total);
      if (data.discussion.url) setDiscussionUrl(data.discussion.url);
    }
    window.addEventListener('message', handleMessage);
    return () => window.removeEventListener('message', handleMessage);
  }, []);

  const handleClick = useCallback(() => {
    if (discussionUrl) {
      window.open(discussionUrl, '_blank', 'noopener');
    }
  }, [discussionUrl]);

  return (
    <button
      onClick={handleClick}
      title={discussionUrl ? '点击前往点赞' : '加载中...'}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: '0.5rem',
        padding: '0.5rem 1.2rem',
        marginTop: '1.5rem',
        fontSize: '1rem',
        fontWeight: 500,
        color: 'var(--ifm-color-primary)',
        background: 'var(--ifm-color-emphasis-100)',
        border: '1px solid var(--ifm-color-emphasis-300)',
        borderRadius: '2rem',
        cursor: discussionUrl ? 'pointer' : 'default',
        transition: 'all 0.2s',
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.background = 'var(--ifm-color-primary)';
        e.currentTarget.style.color = '#fff';
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.background = 'var(--ifm-color-emphasis-100)';
        e.currentTarget.style.color = 'var(--ifm-color-primary)';
      }}
    >
      <span style={{fontSize: '1.2rem'}}>👍</span>
      <span>{count !== null ? `${count} 个赞` : '加载中...'}</span>
    </button>
  );
}

const GISCUS_PROPS = {
  repo: 'kweaver-ai/kweaver' as const,
  repoId: 'R_kgDOH8FKww',
  category: 'Announcements',
  categoryId: 'DIC_kwDOH8FKw84C4wpZ',
  mapping: 'pathname' as const,
  strict: '0' as const,
  reactionsEnabled: '1' as const,
  emitMetadata: '1' as const,
  inputPosition: 'bottom' as const,
  lang: 'zh-CN',
};

export default function BlogPostItemWrapper(props) {
  const {isBlogPostPage} = useBlogPost();
  const {colorMode} = useColorMode();

  return (
    <>
      <BlogPostItem {...props} />
      {isBlogPostPage && (
        <>
          <LikeButton />
          {/* Hidden Giscus: only used to fetch reaction count */}
          <div style={{position: 'absolute', width: 0, height: 0, overflow: 'hidden'}}>
            <Giscus
              {...GISCUS_PROPS}
              theme={colorMode === 'dark' ? 'dark' : 'light'}
            />
          </div>
        </>
      )}
    </>
  );
}
