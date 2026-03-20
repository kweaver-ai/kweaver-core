from volcenginesdkarkruntime import Ark
import os

client = Ark(
    base_url="https://ark.cn-beijing.volces.com/api/v3",
    # 从环境变量中获取您的 API Key。此为默认方式，您可根据需要进行修改
    api_key=os.environ.get("ARK_API_KEY"),
)

print("----- embeddings request -----")
resp = client.embeddings.create(
    model="doubao-embedding-large-text-250515",
    input=["花椰菜又称菜花、花菜，是一种常见的蔬菜。"]
)
print(resp)