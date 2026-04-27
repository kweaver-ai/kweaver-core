# agent-executor.spec
# -*- mode: python ; coding: utf-8 -*-
# PyInstaller specification file for Agent-Executor

from PyInstaller.utils.hooks import copy_metadata, collect_all

project_root = os.path.abspath('.')

# Data files to include
datas = [
    ("data_migrations", "data_migrations"),
]
datas += copy_metadata('setuptools')

# Collect all charset_normalizer files (including mypyc compiled .so modules)
cn_datas, cn_binaries, cn_hiddenimports = collect_all('charset_normalizer')
datas += cn_datas

# Collect mypyc runtime shared library from top-level site-packages
# charset_normalizer 3.x uses mypyc to compile modules, the runtime .so file
# (e.g. 81d243bd2c585b0f4821__mypyc.*.so) is placed at site-packages root,
# NOT inside charset_normalizer/, so collect_all won't find it.
import glob
_cn_pkg_dir = os.path.dirname(__import__('charset_normalizer').__file__)
_site_packages = os.path.dirname(_cn_pkg_dir)
for _mypyc_file in glob.glob(os.path.join(_site_packages, '*__mypyc*')):
    cn_binaries.append((_mypyc_file, '.'))

# Analysis configuration
a = Analysis(
    ['main.py'],
    pathex=[project_root],
    binaries=cn_binaries,
    datas=datas,
    hookspath=['hooks'],
    hiddenimports=[
        'pkg_resources',
        'setuptools',
        'setuptools._distutils',
        'opentelemetry.instrumentation.dependencies',
        'charset_normalizer',
        'charset_normalizer.md',
        'chardet',
        'requests.packages.chardet',
    ] + cn_hiddenimports,
    excludes=[
        'tkinter',
        'matplotlib',
        'IPython',
        'jupyter',
        'notebook',
        'pytest',
        'pip',
        'pylint',
        'coverage',
    ],
)

pyz = PYZ(a.pure, a.zipped_data)

exe = EXE(
    pyz,
    a.scripts,
    exclude_binaries=True,
    name='agent-executor',
    console=True,
)

coll = COLLECT(
    exe,
    a.binaries,
    a.zipfiles,
    a.datas,
    name='agent-executor',
)
