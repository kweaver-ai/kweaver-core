# -*- mode: python ; coding: utf-8 -*-


a = Analysis(
    ['src/main.py'],
    pathex=['.'],
    binaries=[],
    datas=[('mem0', 'mem0'), ('src', 'src'), ('src/config/config.yaml', '.'), ('src/locale', 'locale')],
    hiddenimports=['qdrant_client', 'openai', 'opensearchpy', 'pymysql', 'pymysql.cursors', 'dbutilsx', 'dbutilsx.pooled_db'],
    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],
    excludes=[],
    noarchive=False,
    optimize=0,
)
pyz = PYZ(a.pure)

exe = EXE(
    pyz,
    a.scripts,
    [],
    exclude_binaries=True,
    name='agent-memory',
    debug=False,
    bootloader_ignore_signals=False,
    strip=True,
    upx=True,
    console=True,
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
)
coll = COLLECT(
    exe,
    a.binaries,
    a.datas,
    strip=True,
    upx=True,
    upx_exclude=[],
    name='agent-memory',
)
