import React from 'react';
import BlogPostItem from '@theme-original/BlogPostItem';
import {useBlogPost} from '@docusaurus/plugin-content-blog/client';
import {useColorMode} from '@docusaurus/theme-common';
import Giscus from '@giscus/react';

const GISCUS_PROPS = {
  repo: 'kweaver-ai/kweaver' as const,
  repoId: 'R_kgDOH8FKww',
  category: 'Announcements',
  categoryId: 'DIC_kwDOH8FKw84C4wpZ',
  mapping: 'pathname' as const,
  strict: '0' as const,
  reactionsEnabled: '1' as const,
  emitMetadata: '0' as const,
  inputPosition: 'top' as const,
  lang: 'zh-CN',
};

export default function BlogPostItemWrapper(props) {
  const {isBlogPostPage} = useBlogPost();
  const {colorMode} = useColorMode();

  return (
    <>
      <BlogPostItem {...props} />
      {isBlogPostPage && (
        <div style={{marginTop: '1.5rem'}}>
          <Giscus
            {...GISCUS_PROPS}
            theme={colorMode === 'dark' ? 'dark' : 'light'}
          />
        </div>
      )}
    </>
  );
}
