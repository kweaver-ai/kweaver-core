import React, {useState, useRef, useEffect} from 'react';
import Info from '@theme-original/BlogPostItem/Header/Info';
import {useBlogPost} from '@docusaurus/plugin-content-blog/client';
import {QRCodeSVG} from 'qrcode.react';

function ShareButton() {
  const [open, setOpen] = useState(false);
  const [pageUrl, setPageUrl] = useState('');
  const popoverRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setPageUrl(window.location.href);
  }, []);

  useEffect(() => {
    if (!open) return;
    function handleClickOutside(e: MouseEvent) {
      if (popoverRef.current && !popoverRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [open]);

  return (
    <span style={{position: 'relative', display: 'inline-flex', marginLeft: '0.75rem'}}>
      <button
        onClick={() => setOpen(!open)}
        title="分享到微信朋友圈"
        aria-label="分享"
        style={{
          display: 'inline-flex',
          alignItems: 'center',
          gap: '0.3rem',
          padding: '0.15rem 0.6rem',
          fontSize: '0.85rem',
          color: 'var(--ifm-color-emphasis-600)',
          background: 'none',
          border: '1px solid var(--ifm-color-emphasis-300)',
          borderRadius: '1rem',
          cursor: 'pointer',
          transition: 'all 0.2s',
        }}
        onMouseEnter={(e) => {
          e.currentTarget.style.color = 'var(--ifm-color-primary)';
          e.currentTarget.style.borderColor = 'var(--ifm-color-primary)';
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.color = 'var(--ifm-color-emphasis-600)';
          e.currentTarget.style.borderColor = 'var(--ifm-color-emphasis-300)';
        }}
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <circle cx="18" cy="5" r="3"/>
          <circle cx="6" cy="12" r="3"/>
          <circle cx="18" cy="19" r="3"/>
          <line x1="8.59" y1="13.51" x2="15.42" y2="17.49"/>
          <line x1="15.41" y1="6.51" x2="8.59" y2="10.49"/>
        </svg>
        分享
      </button>
      {open && (
        <div
          ref={popoverRef}
          style={{
            position: 'absolute',
            bottom: '100%',
            right: 0,
            marginBottom: '0.5rem',
            padding: '1rem',
            background: 'var(--ifm-background-color)',
            border: '1px solid var(--ifm-color-emphasis-300)',
            borderRadius: '0.5rem',
            boxShadow: '0 4px 16px rgba(0,0,0,0.2)',
            zIndex: 1000,
            textAlign: 'center',
            whiteSpace: 'nowrap',
          }}
        >
          <QRCodeSVG value={pageUrl} size={140} />
          <div style={{
            marginTop: '0.5rem',
            fontSize: '0.8rem',
            color: 'var(--ifm-color-emphasis-600)',
          }}>
            微信扫码打开，点右上角 ··· 分享
          </div>
        </div>
      )}
    </span>
  );
}

export default function InfoWrapper(props) {
  const {isBlogPostPage} = useBlogPost();

  return (
    <div style={{display: 'flex', alignItems: 'center', flexWrap: 'wrap'}}>
      <Info {...props} />
      {isBlogPostPage && <ShareButton />}
    </div>
  );
}
