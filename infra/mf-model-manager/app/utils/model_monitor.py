# 定义定时任务函数
import asyncio
import json
import re
from datetime import datetime, timedelta

import aiohttp

from app.commons.snow_id import worker
from app.dao.llm_model_dao import llm_model_dao
from app.dao.model_quota_dao import model_quota_dao
from app.logs.stand_log import StandLogger


async def extract_vllm_metrics(log_text):
    # 查找generation_tokens_total和prompt_tokens_total的值
    generation_match = re.search(
        r'vllm:generation_tokens_total\{engine="[^\"]+",model_name="[^\"]+"\} ([\d.]+(?:e[-+]?\d+)?)',
        log_text)
    prompt_match = re.search(r'vllm:prompt_tokens_total\{engine="[^\"]+",model_name="[^\"]+"\} ([\d.]+(?:e[-+]?\d+)?)',
                             log_text)
    first_token_time_sum_match = re.search(
        r'vllm:time_to_first_token_seconds_sum\{engine="[^"]+",model_name="[^"]+"\} ([\d.]+(?:e[-+]?\d+)?)', log_text)
    first_token_time_count_match = re.search(
        r'vllm:time_to_first_token_seconds_count\{engine="[^"]+",model_name="[^"]+"\} ([\d.]+(?:e[-+]?\d+)?)', log_text)
    generation_tokens = int(float(generation_match.group(1))) if generation_match else 0

    prompt_tokens = int(float(prompt_match.group(1))) if prompt_match else 0
    # 计算平均首字响应时间
    if first_token_time_sum_match and first_token_time_count_match:
        first_token_time_sum = float(first_token_time_sum_match.group(1))
        first_token_time_count = float(first_token_time_count_match.group(1))
        first_token_time = first_token_time_sum / first_token_time_count if first_token_time_count != 0 else 0
    else:
        first_token_time = 0
    StandLogger.info_log({
        "generation_tokens_total": generation_tokens,
        "prompt_tokens_total": prompt_tokens,
        "average_first_token_time": first_token_time
    })
    return {
        "generation_tokens_total": generation_tokens,
        "prompt_tokens_total": prompt_tokens,
        "average_first_token_time": first_token_time
    }


async def vllm_monitor_task():
    models = llm_model_dao.get_all_tome_model_list()
    now = datetime.now()
    now_format = now.strftime("%Y-%m-%d %H:%M")
    ten_minutes_ago = now - timedelta(minutes=10)
    ten_minutes_ago_format = ten_minutes_ago.strftime("%Y-%m-%d %H:%M")
    # StandLogger.info_log(models)
    model_ids = [line["f_model_id"] for line in models]
    if not model_ids:
        StandLogger.info_log("系统中暂未配置大模型，无需探测私有化部署的推理性能")
        return
    ten_minutes_ago_monitor_data = llm_model_dao.get_ten_minutes_ago_monitor_data(model_ids, ten_minutes_ago_format)
    cache = {}
    datas = []
    for line in ten_minutes_ago_monitor_data:
        model_id = line["f_model_id"]
        generation_tokens_total = line["f_generation_tokens_total"]
        prompt_tokens_total = line["f_prompt_tokens_total"]
        cache[model_id] = [generation_tokens_total, prompt_tokens_total]
    async with aiohttp.ClientSession() as session:
        for model in models:
            data_id = worker.get_id()
            model_name = model["f_model_name"]
            model_id = model["f_model_id"]
            prev_gen_tokens, prev_prompt_tokens = cache.get(model_id, [None, None])
            model_config = json.loads(model["f_model_config"])
            metrics_url = model_config.get("api_url", "").replace("/v1/chat/completions", "/metrics")
            StandLogger.info_log(f"metrics url:{metrics_url}")
            if metrics_url:
                try:
                    async with session.get(metrics_url) as response:
                        if response.status == 200:
                            log_text = await response.text()
                            data = await extract_vllm_metrics(log_text)
                            generation_tokens_total = data['generation_tokens_total']
                            prompt_tokens_total = data['prompt_tokens_total']
                            average_first_token_time = data['average_first_token_time']
                            if prev_prompt_tokens is not None and prev_gen_tokens is not None:
                                prev_total = prev_prompt_tokens + prev_gen_tokens
                                now_total = generation_tokens_total + prompt_tokens_total
                                generation_token_speed = round((generation_tokens_total - prev_gen_tokens) / 600,
                                                               2) if generation_tokens_total > prev_gen_tokens else 0
                                total_token_speed = round((now_total - prev_total) / 600,
                                                          2) if now_total > prev_total else 0
                            else:
                                generation_token_speed = 0
                                total_token_speed = 0
                            datas.append([data_id, now_format, model_name, model_id, generation_tokens_total,
                                          prompt_tokens_total, average_first_token_time, generation_token_speed,
                                          total_token_speed])
                        else:
                            StandLogger.warn(f"call vllm server failed:{await response.text()}")
                except Exception as e:
                    StandLogger.error(f"fetching model logs failed: {str(e)}")
    StandLogger.info_log("start insert monitor data")
    llm_model_dao.add_monitor_data(datas)
    StandLogger.info_log("insert monitor data success")


async def delete_monitor_data_task():
    StandLogger.info_log("start delete 30 days ago model monitor data")
    llm_model_dao.delete_model_monitor_data()
    StandLogger.info_log("delete 30 days ago model monitor data success")


# 每个月2号凌晨删除上个月的模型消费数据
async def delete_model_quota_data_task():
    StandLogger.info_log("start delete 30 days ago model monitor data")
    model_quota_dao.delete_previous_month_model_quota()
    StandLogger.info_log("delete 30 days ago model monitor data success")


if __name__ == '__main__':
    asyncio.run(vllm_monitor_task())
