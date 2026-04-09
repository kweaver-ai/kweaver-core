'''
更新 agent_config和agent_history表中的内容
更新范围：agent_config表中的release_config,draft_config,以及agent_history表中的config
具体内容：
1. 召回块高级配置初始化
2. 原召回块reranker处理
3. 逻辑块context_length初始化为8
4. 输出name设为answer,from设为output

'''
import os
import time
# 64位ID的划分
WORKER_ID_BITS = 5
DATACENTER_ID_BITS = 5
SEQUENCE_BITS = 12

# 最大取值计算
MAX_WORKER_ID = -1 ^ (-1 << WORKER_ID_BITS)  # 2**5-1 0b11111
MAX_DATACENTER_ID = -1 ^ (-1 << DATACENTER_ID_BITS)

# 移位偏移计算
WOKER_ID_SHIFT = SEQUENCE_BITS
DATACENTER_ID_SHIFT = SEQUENCE_BITS + WORKER_ID_BITS
TIMESTAMP_LEFT_SHIFT = SEQUENCE_BITS + WORKER_ID_BITS + DATACENTER_ID_BITS

# 序号循环掩码
SEQUENCE_MASK = -1 ^ (-1 << SEQUENCE_BITS)

# Twitter元年时间戳
TWEPOCH = 1288834974657

import traceback
from datetime import datetime

import rdsdriver

class IdWorker(object):
    """
    用于生成IDs
    """

    def __init__(self, datacenter_id, worker_id, sequence=0):
        """
        初始化
        :param datacenter_id: 数据中心（机器区域）ID
        :param worker_id: 机器ID
        :param sequence: 其实序号
        """
        # sanity check
        if worker_id > MAX_WORKER_ID or worker_id < 0:
            raise ValueError('worker_id值越界')

        if datacenter_id > MAX_DATACENTER_ID or datacenter_id < 0:
            raise ValueError('datacenter_id值越界')

        self.worker_id = worker_id
        self.datacenter_id = datacenter_id
        self.sequence = sequence

        self.last_timestamp = -1  # 上次计算的时间戳

    def _gen_timestamp(self):
        """
        生成整数时间戳
        :return:int timestamp
        """
        return int(time.time() * 1000)

    def get_id(self):
        """
        获取新ID
        :return:
        """
        timestamp = self._gen_timestamp()

        # 时钟回拨
        if timestamp < self.last_timestamp:
            # logging.errors('clock is moving backwards. Rejecting requests until{}'.format(self.last_timestamp))
            raise Exception

        if timestamp == self.last_timestamp:
            self.sequence = (self.sequence + 1) & SEQUENCE_MASK
            if self.sequence == 0:
                timestamp = self._til_next_millis(self.last_timestamp)
        else:
            self.sequence = 0

        self.last_timestamp = timestamp

        new_id = ((timestamp - TWEPOCH) << TIMESTAMP_LEFT_SHIFT) | (self.datacenter_id << DATACENTER_ID_SHIFT) | \
                 (self.worker_id << WOKER_ID_SHIFT) | self.sequence
        return new_id

    def _til_next_millis(self, last_timestamp):
        """
        等到下一毫秒
        """
        timestamp = self._gen_timestamp()
        while timestamp <= last_timestamp:
            timestamp = self._gen_timestamp()
        return timestamp

def get_conn(user, password, host, port, database):
    try:
        conn = rdsdriver.connect(host=host,
                                 port=int(port),
                                 user=user,
                                 password=password,
                                 database=database,
                                 cursorclass=rdsdriver.DictCursor,
                                 autocommit=True)
    except Exception as e:
        print("connect database error: %s", str(e))
        raise e
    return conn


def fetch_llm_history(cursor, page_size=100):
    query = f"SELECT f_model_id FROM t_llm_model limit 1000"
    cursor.execute(query)
    rows = cursor.fetchall()
    return rows


def update_model(cursor):
    data = []
    now = datetime.now()
    worker = IdWorker(1, 1, 0)
    for row in fetch_llm_history(cursor=cursor):
        model_id = row["f_model_id"]
        conf_id = worker.get_id()
        value_list = [conf_id, model_id, 1, -1, -1, -1, -1, 0, now, now,
                      "[1, 1]", '["thousand", "thousand"]']
        data.append(value_list)
    
    # 批量插入
    if data:
        sql = """insert into t_model_quota_config (f_id, f_model_id, f_billing_type, f_input_tokens, f_output_tokens, f_referprice_in, 
                f_referprice_out, f_currency_type, f_create_time, f_update_time, f_num_type, f_price_type) values(%s, %s, 
                %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)"""
        cursor.executemany(sql, data)
        print(f"成功插入 {len(data)} 条记录到 t_model_quota_config 表中")

if __name__ == "__main__":
    conn = get_conn(os.environ["DB_USER"], os.environ["DB_PASSWD"],
                    os.environ["DB_HOST"], os.environ["DB_PORT"], "model_management")
    conn_cursor = conn.cursor()
    try:
        print('start update_model')
        update_model(conn_cursor)
    except Exception as e:
        print('update_model failed')
        traceback.print_exc()
        raise Exception()
    finally:
        print('update_model success')
        conn_cursor.close()
        conn.close()
