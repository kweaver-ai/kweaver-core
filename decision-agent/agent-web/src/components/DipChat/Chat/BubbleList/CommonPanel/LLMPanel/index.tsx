import Markdown from '@/components/Markdown';
import React, { useRef, useState } from 'react';
import { Popover } from 'antd';
import CitePanel from './CitePanel';
import { useDeepCompareEffect } from '@/hooks';

const LLMPanel = ({ text, cites, isLLMProcess, status, consumeTime }: any) => {
  const resultWrapper = useRef<HTMLDivElement>(null);
  const [popoverOpen, setPopoverOpen] = useState(false);
  const [popoverContent, setPopoverContent] = useState({
    citeIndex: -1,
  });
  const [popoverPos, setPopoverPos] = useState<{ left: number; top: number }>({ left: 0, top: 0 });
  const closeTimerRef = useRef<number | null>(null);
  const hoveringPopoverRef = useRef<boolean>(false);
  useDeepCompareEffect(() => {
    const container = resultWrapper.current;
    if (!container) return;
    const iNodes = Array.from(container.querySelectorAll('i')) as HTMLElement[];
    const handleEnter = (e: Event) => {
      const target = e.currentTarget as HTMLElement | null;
      if (!target) return;
      if (closeTimerRef.current) {
        window.clearTimeout(closeTimerRef.current);
        closeTimerRef.current = null;
      }
      const citeIndex = target.getAttribute('index') ?? -1;
      const rect = target.getBoundingClientRect();
      const wrapperRect = document.getElementById('adp-chat-wrapper')!.getBoundingClientRect();
      const left = rect.left - wrapperRect.left + rect.width / 2;
      const top = rect.top - wrapperRect.top;
      setPopoverContent({
        citeIndex: Number(citeIndex),
      });
      setPopoverPos({ left, top });
      setPopoverOpen(true);
    };

    const handleLeave = (e: Event) => {
      const target = e.currentTarget as HTMLElement | null;
      if (!target) return;
      if (closeTimerRef.current) window.clearTimeout(closeTimerRef.current);
      closeTimerRef.current = window.setTimeout(() => {
        if (!hoveringPopoverRef.current) {
          setPopoverOpen(false);
        }
      }, 100);
    };

    iNodes.forEach(node => {
      node.addEventListener('mouseenter', handleEnter);
      node.addEventListener('mouseleave', handleLeave);
    });

    return () => {
      iNodes.forEach(node => {
        node.removeEventListener('mouseenter', handleEnter);
        node.removeEventListener('mouseleave', handleLeave);
      });
    };
  }, [cites, text]);

  if (isLLMProcess) {
    return (
      <div className="dip-mb-8 dip-font-12 dip-text-color-45">
        {status === 'processing' && '正在调用大模型处理请求...'}
        {status === 'completed' && (
          <span>
            大模型推理完成（用时 <span>{consumeTime}</span> s）
          </span>
        )}
      </div>
    );
  }

  return (
    <>
      <div ref={resultWrapper}>
        <Markdown className="dip-mb-8" value={text} readOnly />
      </div>
      {popoverOpen && (
        <Popover
          open
          content={
            <div
              onMouseEnter={() => {
                hoveringPopoverRef.current = true;
                if (closeTimerRef.current) {
                  window.clearTimeout(closeTimerRef.current);
                  closeTimerRef.current = null;
                }
              }}
              onMouseLeave={() => {
                hoveringPopoverRef.current = false;
                setPopoverOpen(false);
              }}
            >
              <CitePanel cite={cites[popoverContent.citeIndex] ?? {}} />
            </div>
          }
          getPopupContainer={() => document.getElementById('adp-chat-wrapper')!}
          arrow={false}
        >
          <span
            style={{
              position: 'absolute',
              left: popoverPos.left,
              top: popoverPos.top,
              width: 1,
              height: 16,
              pointerEvents: 'none',
            }}
          />
        </Popover>
      )}
    </>
  );
};

export default LLMPanel;
