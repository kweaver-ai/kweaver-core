# PyInstaller hook for charset_normalizer
from PyInstaller.utils.hooks import collect_all

datas, binaries, hiddenimports = collect_all("charset_normalizer")
