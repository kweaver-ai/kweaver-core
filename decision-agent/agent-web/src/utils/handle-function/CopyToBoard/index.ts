/**
 * 复制内容到剪贴板
 * @param {String} text 文字内容
 */
export const copyToBoard = (text: string) => {
  // eslint-disable-next-line no-async-promise-executor
  return new Promise(async resolve => {
    if (navigator.clipboard) {
      try {
        await navigator.clipboard.writeText(text);
        resolve(true);
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
      } catch (err) {
        //
      }
    }

    if (typeof document.execCommand === 'function') {
      try {
        const input = document.createElement('textarea');
        input.style.cssText = 'position: absolute; left: -9999px; z-index: -1;';
        input.setAttribute('readonly', 'readonly');
        input.value = text;
        document.body.appendChild(input);
        input.select();
        let isSuccess = false;
        if (document.execCommand('copy')) {
          isSuccess = document.execCommand('copy');
        }
        document.body.removeChild(input);
        resolve(isSuccess);
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
      } catch (error) {
        resolve(false);
      }
    } else {
      resolve(false);
    }
  });
};

/**
 * 复制内容到剪贴板保留复制内容的格式
 * @param {String} str 文字内容
 */
export const copyToBoardArea = (str: string) => {
  try {
    const input = document.createElement('textarea');
    document.body.appendChild(input);
    input.value = str;
    input.select();
    const isSuccess = document.execCommand('copy');
    document.body.removeChild(input);

    return isSuccess;
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
  } catch (err) {
    return false;
  }
};
