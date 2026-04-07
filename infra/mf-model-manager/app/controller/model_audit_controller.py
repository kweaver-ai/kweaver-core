from app.commons.snow_id import worker
from app.interfaces import logics
from app.logs.stand_log import StandLogger
import json

from app.mydb.ConnectUtil import kafka_client


async def add_llm_model_call_log(para: logics.AddModelUsedAudit):
    """
    将token消费信息写入kafka
    :param para:
    :return:
    """
    try:
        
        # 准备消息数据，参考kafka_streams_processor.py中消费者的字段
        message_data = {
            'model_id': para.model_id,
            'user_id': para.user_id,
            'input_tokens': para.input_tokens,
            'output_tokens': para.output_tokens,
            'conf_id': str(worker.get_id()),  # 生成新的配置ID
            'total_price': 0.0,  # 这个值会在消费者端计算
            'currency_type': 0,  # 默认值，会在消费者端更新
            'price_type': ["thousand", "thousand"],    # 默认值，会在消费者端更新
            'referprice_in': 0.0,  # 默认值，会在消费者端更新
            'referprice_out': 0.0  # 默认值，会在消费者端更新
        }
        
        # 将消息数据转换为JSON格式
        message_json = json.dumps(message_data, ensure_ascii=False)
        
        # 使用异步方式发送消息到Kafka
        # 参考ConnectUtil.py中的produce_async方法
        kafka_client.produce_async(
            value=message_json.encode('utf-8'),
            key=f"{para.model_id}_{para.user_id}_{message_data['conf_id']}".encode('utf-8')  # 加入conf_id以便排查追踪
        )
        
        # StandLogger.info_log(f"成功将token消费信息写入Kafka: model_id={para.model_id}, user_id={para.user_id}, conf_id={message_data['conf_id']}")
    except Exception as e:
        StandLogger.error(f"将token消费信息写入Kafka时出错: {e}")


