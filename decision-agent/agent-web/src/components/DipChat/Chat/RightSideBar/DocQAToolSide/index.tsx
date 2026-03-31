import { useDipChatStore } from '@/components/DipChat/store';
import { DipChatItemContentProgressType } from '@/components/DipChat/interface';
import Markdown from '@/components/Markdown';
import classNames from 'classnames';
import { CloseOutlined } from '@ant-design/icons';
import React, { useEffect, useRef, useState } from 'react';
import styles from './index.module.less';
import ScrollBarContainer from '@/components/ScrollBarContainer';
import CitesList from './CitesList';
import DipButton from '@/components/DipButton';
import NoData from '@/components/NoData';
import { Popover } from 'antd';
import CitePanel from './CitePanel';

const DocQAToolSide = ({ onPreview }: any) => {
  const [value, setValue] = useState('result');
  const {
    dipChatStore: { chatList, activeChatItemIndex, activeProgressIndex },
    closeSideBar,
  } = useDipChatStore();
  const chatItem = chatList[activeChatItemIndex];
  const activeProgress: DipChatItemContentProgressType = chatItem.content.progress[activeProgressIndex] || {};
  const htmlText = activeProgress.docQaToolResult?.htmlText || '';
  const cites = activeProgress?.docQaToolResult?.cites;

  const [popoverOpen, setPopoverOpen] = useState(false);
  const [popoverContent, setPopoverContent] = useState({
    citeIndex: -1,
    sliceIndex: -1,
  });
  const [popoverPos, setPopoverPos] = useState<{ left: number; top: number }>({ left: 0, top: 0 });
  const resultWrapper = useRef<HTMLDivElement>(null);
  const closeTimerRef = useRef<number | null>(null);
  const hoveringPopoverRef = useRef<boolean>(false);

  useEffect(() => {
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
      target.classList.add('dip-text-color-primary');
      target.style.cursor = 'pointer';
      const sliceIdx = target.getAttribute('slice_idx') ?? -1;
      const citeIndex = (target.textContent || '').trim() || -1;
      const rect = target.getBoundingClientRect();
      const wrapperRect = container.getBoundingClientRect();
      const left = rect.left - wrapperRect.left + rect.width / 2;
      const top = rect.top - wrapperRect.top;
      setPopoverContent({
        citeIndex: Number(citeIndex) - 1,
        sliceIndex: Number(sliceIdx),
      });
      setPopoverPos({ left, top });
      setPopoverOpen(true);
    };

    const handleLeave = (e: Event) => {
      const target = e.currentTarget as HTMLElement | null;
      if (!target) return;
      target.classList.remove('dip-text-color-primary');
      if (closeTimerRef.current) window.clearTimeout(closeTimerRef.current);
      closeTimerRef.current = window.setTimeout(() => {
        if (!hoveringPopoverRef.current) {
          setPopoverOpen(false);
        }
      }, 100);
    };

    const handleClick = () => {
      if (Array.isArray(cites) && cites.length > 0) {
        setValue('refernce');
      }
    };

    iNodes.forEach(node => {
      node.addEventListener('mouseenter', handleEnter);
      node.addEventListener('mouseleave', handleLeave);
      node.addEventListener('click', handleClick);
    });

    return () => {
      iNodes.forEach(node => {
        node.removeEventListener('mouseenter', handleEnter);
        node.removeEventListener('mouseleave', handleLeave);
        node.removeEventListener('click', handleClick);
      });
    };
  }, [htmlText, cites]);

  return (
    <div className={styles.container}>
      <div className="dip-flex-space-between dip-pt-12 dip-pb-12 dip-pl-16 dip-pr-16">
        <div className="dip-flex-item-full-width dip-flex-align-center">
          <div
            className={classNames(styles.tabItem, {
              [styles.active]: value === 'result',
            })}
            onClick={() => {
              setValue('result');
            }}
          >
            结果
          </div>
          <div
            className={classNames(styles.tabItem, {
              [styles.active]: value === 'refernce',
            })}
            onClick={() => {
              if (cites.length > 0) {
                setValue('refernce');
              }
            }}
          >
            引用 (<span className="dip-text-color-primary">{cites.length}</span>)
          </div>
        </div>
        <DipButton
          size="small"
          type="text"
          onClick={() => {
            closeSideBar();
          }}
        >
          <CloseOutlined className="dip-text-color-45 dip-font-16" />
        </DipButton>
      </div>
      <div className="dip-flex-item-full-height dip-position-r dip-overflow-hidden" ref={resultWrapper}>
        <ScrollBarContainer
          className="dip-full dip-pl-16 dip-pr-16"
          style={{ display: value === 'result' ? 'block' : 'none' }}
        >
          {htmlText ? (
            <Markdown className="dip-mb-8" value={htmlText.replaceAll('-----', '\n')} readOnly />
          ) : (
            <div className="dip-full dip-center">
              <NoData />
            </div>
          )}
        </ScrollBarContainer>
        {popoverOpen && (
          <Popover
            open={popoverOpen}
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
                <CitePanel {...popoverContent} cites={cites} onPreview={onPreview} />
              </div>
            }
            getPopupContainer={() => resultWrapper.current!}
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
        <div className="dip-full" style={{ display: value === 'refernce' ? 'block' : 'none' }}>
          <CitesList cites={cites} onPreview={onPreview} />
        </div>
      </div>
    </div>
  );
};

export default DocQAToolSide;
