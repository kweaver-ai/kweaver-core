// 卡片的最小宽度
export const minCardWidth = 380;
// 卡片的最大宽度
export const maxCardWidth = 500;
// 卡片的间距
export const gap = 16;
// 正常卡片的高度（包含间距）
export const rowHeight = 184 + gap;
// loadingMore 行的高度
export const loadingMoreRowHeight = 30;

// 计算列个数，卡片的宽度在 minCardWidth~maxCardWidth
export const computeColumnCount = (
  width: number,
  { minWidth, maxWidth }: { minWidth: number; maxWidth: number } = { minWidth: minCardWidth, maxWidth: maxCardWidth }
): number => {
  let count = 1;

  while (width / count > maxWidth) {
    count = count + 1;
  }

  if (width / count < minWidth && count > 1) {
    count = count - 1;
  }

  if (width / count > maxWidth) {
    return computeColumnCount(width, { minWidth: minWidth - 10, maxWidth: maxWidth + 10 });
  }

  return count;
};
