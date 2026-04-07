
import os
import json
import sys
from urllib.parse import urlparse
import uuid
import zlib
import hashlib
from io import BytesIO
from datetime import datetime
from jsonschema import validate
import tornado


def generate_MD5_CRC32(file_input, chunk_size=8192):
    md5_hash = hashlib.md5()
    crc32_value = 0
    data_length = 0
    
    # 统一处理为文件流
    if isinstance(file_input, str):
        file_obj = open(file_input, 'rb')
        should_close = True
    else:
        file_obj = file_input
        file_obj.seek(0)
        should_close = False
    
    try:
        while True:
            chunk = file_obj.read(chunk_size)
            if not chunk:
                break
            md5_hash.update(chunk)
            crc32_value = zlib.crc32(chunk, crc32_value)
            data_length += len(chunk)
        
        return md5_hash.hexdigest(), "%x"%(crc32_value & 0xFFFFFFFF), data_length
    
    finally:
        if should_close:
            file_obj.close()
        elif hasattr(file_obj, 'seek'):
            file_obj.seek(0)  # 重置流位置

def generate_timestamp():
    current_time = datetime.now()
    # 获取当前时间戳
    timestamp = int(current_time.timestamp() * 1000)  # 将秒级时间戳转换为毫秒级
    return timestamp * 1000

def validate_params(data_valid, schema_name):
    base_path = os.getcwd()
    possible_paths = [
        f"{base_path}/src/schema/{schema_name}",
        f"{base_path}/dataflowtools/src/schema/{schema_name}",
        f"{base_path}/_internal/src/schema/{schema_name}"
    ]

    file_path = ''
    for path in possible_paths:
        if os.path.exists(path):
            file_path = path
            break
    else:
        raise FileNotFoundError(f"Schema file not found: {schema_name}")

    with open(file_path, 'rb') as file:
        content = file.read()
    schema = json.loads(content)
    validate(instance=data_valid, schema=schema)

def split_file_type(file_path):
    _, file_extension = os.path.splitext(file_path)
    return file_extension.lower()

def generate_random_filename():
    # 使用 uuid4 方法生成一个随机的UUID
    random_uuid = uuid.uuid4()
    # 将 UUID 转换为字符串，并去掉连接符
    random_filename = str(random_uuid).replace("-", "")
    return random_filename

def generate_random_id():
    # 使用 uuid4 方法生成一个随机的UUID
    random_uuid = uuid.uuid4()
    # 将 UUID 转换为字符串，并去掉连接符
    random_name = str(random_uuid).replace("-", "")
    return random_name

# 判断字节是否丢失
def is_byte_loss(len_buff, start_byte, end_byte, spacing_byte, file_size):
    if end_byte == start_byte:
        if len_buff != 1:
            return True
    else:
        if file_size <= spacing_byte:
            if len_buff != file_size:
                return True
        else:
            if start_byte == 0:
                if len_buff != (end_byte - start_byte) + 1:
                    return True
            elif end_byte != file_size:
                if len_buff != (end_byte - start_byte) + 1:
                    return True
            else:
                # 末字节为空
                if len_buff != (end_byte - start_byte):
                    return True
    return False

async def run_in_thread(self, func, *args):
    """在线程池中运行阻塞函数"""
    loop = tornado.ioloop.IOLoop.current()
    return await loop.run_in_executor(self.executor, func, *args)

def is_valid_url(url: str) -> bool:
    """
    使用urllib.parse判断URL是否合法
    """
    try:
        result = urlparse(url)
        # 检查是否有scheme和netloc
        return all([result.scheme in ("http", "https"), result.netloc])
    except Exception:
        return False