
import os
import json
import uuid
import zlib
import hashlib
from io import BytesIO
from datetime import datetime
from jsonschema import validate


def read_file(file_input) -> bytes:
    if isinstance(file_input, str):
        current_path = os.getcwd()
        file_path = '{}/src/template/{}'.format(current_path, file_input)
        with open(file_path, 'rb') as file:
            content = file.read()
        return content
    elif isinstance(file_input, BytesIO):
        file_input.seek(0)
        return file_input.getvalue()
    else:
        raise ValueError("Unsupported input type. Please provide either a file path or BytesIO object.")

def generate_MD5_CRC32(file_input):
    content = read_file(file_input)
    md5 = hashlib.md5(content).hexdigest()
    crc32 = zlib.crc32(content)
    length = len(content)
    return md5, "%x"%(crc32 & 0xFFFFFFFF), length

def generate_timestamp():
    current_time = datetime.now()
    # 获取当前时间戳
    timestamp = int(current_time.timestamp() * 1000)  # 将秒级时间戳转换为毫秒级
    return timestamp * 1000

def validate_params(data_valid, schema_name):
    current_path = os.getcwd()
    file_path = '{}/src/schema/{}'.format(current_path, schema_name)
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