#!/bin/bash

# SITE_PACKAGES="/usr/local/python3/lib/python3.8/site-packages"
# BACKUP_DIR="/usr/local/python3/lib/python3.8/site-packages.bk"
# SITE_PACKAGES="/usr/local/python3"
# BACKUP_DIR="/usr/local/python3.bk"

# # 判断 site-packages 文件夹是否为空
# if [ -z "$(ls -A $SITE_PACKAGES 2>/dev/null)" ]; then
#     cp -r $BACKUP_DIR/* $SITE_PACKAGES/
# fi


# USER_LOCAL_LIB="/usr/local/lib"
# BACKUP_LOCAL_LIB="/usr/local/lib.bk"

# # 判断 user local lib 文件夹是否为空
# if [ -z "$(ls -A $USER_LOCAL_LIB 2>/dev/null)" ]; then
#     cp -r $BACKUP_LOCAL_LIB/* $USER_LOCAL_LIB/
# fi

# USER_LOCAL_BIN="/usr/local/bin"
# BACKUP_LOCAL_BIN="/usr/local/bin.bk"

# # 判断 user local bin 文件夹是否为空
# if [ -z "$(ls -A $USER_LOCAL_BIN 2>/dev/null)" ]; then
#     cp -r $BACKUP_LOCAL_BIN/* $USER_LOCAL_BIN/
# fi

SITE_PACKAGES="/usr/local"
BACKUP_DIR="/usr/local.bk"

# 判断 local 文件夹是否为空
if [ -z "$(ls -A $SITE_PACKAGES 2>/dev/null)" ]; then
    cp -r $BACKUP_DIR/* $SITE_PACKAGES/
fi