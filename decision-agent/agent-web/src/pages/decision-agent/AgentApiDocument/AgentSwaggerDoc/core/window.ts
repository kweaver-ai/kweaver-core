function makeWindow() {
  let win = {
    location: {},
    history: {},
    open: () => {},
    close: () => {},
    File: () => {},
    navigator: { msSaveOrOpenBlob: (blob, download) => {} },
  };

  if (typeof window === 'undefined') {
    return win;
  }

  try {
    win = window;
    const props = ['File', 'Blob', 'FormData'];
    for (const prop of props) {
      if (prop in window) {
        win[prop] = window[prop];
      }
    }
  } catch (e) {
    console.error(e);
  }

  return win;
}

export default makeWindow();
