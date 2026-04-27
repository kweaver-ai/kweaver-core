from fastapi import APIRouter
from pydantic import StrictFloat, StrictStr, Field, StrictInt, conint,  conlist, constr,BaseModel

model_quota_router = APIRouter()


# 添加外部小模型信息
class AddExternalSmallModelInfo(BaseModel):
    model_id: StrictStr = Field(description="配置id", default="")
    model_name: StrictStr = Field(description="模型名称", default="")
    model_type: StrictStr = Field(description="模型类型", default="")
    model_config: dict
    adapter: bool = Field(default=False, description="是否开启适配服务")
    adapter_code: StrictStr = Field(default=None, description="适配代码")
# 大模型配额信息
class ModelQuotaInfo (BaseModel):
    conf_id: StrictStr = Field(description="配置id")
    model_id:StrictStr = Field(description="绑定的模型id")
    billing_type: StrictInt = Field(description="计费类型 0:统一计费 1:input output单独计费")
    input_tokens:StrictFloat = Field(description="输入tokens总额度")
    output_tokens:StrictFloat= Field(description="输出tokens总额度")
    currency_type:StrictInt = Field(description="计费单价货币类型,0:人名币 1:美元")
    referprice_in:StrictFloat = Field(description="输入tokens计费单价")
    referprice_out:StrictFloat = Field(description="输出tokens计费单价")
    create_time:StrictStr= Field(description="创建时间")
    update_time:StrictStr= Field(description="编辑时间")
    num_type:conlist(conint(ge=0, le=5), min_items=2, max_items=2)
    price_type: conlist(constr(regex=r'^(thousand|million)$'), min_items=2, max_items=2)

# 大模型用户配额信息
class ModelUserQuotaInfo (BaseModel):
    conf_id: StrictStr = Field(description="配置id",default="")
    model_id:StrictStr = Field(description="绑定的模型id",default="")
    user_id:StrictStr = Field(description="用户id",default="")
    input_tokens:StrictFloat = Field(description="输入tokens总额度",default=0)
    output_tokens:StrictFloat= Field(description="输出tokens总额度",default=0)
    create_time:StrictStr= Field(description="创建时间",default="")
    update_time:StrictStr= Field(description="编辑时间",default="")
    num_type:conlist(conint(ge=0, le=3), min_items=2, max_items=2)

# 大模型使用日志
class ModelUsedAuditInfo (BaseModel):
    conf_id: StrictStr = Field(description="配置id",default="")
    model_id:StrictStr = Field(description="模型id",default="")
    user_id:StrictStr = Field(description="用户id",default="")
    input_tokens:StrictInt = Field(description="使用tokens量",default=0)
    output_tokens:StrictInt= Field(description="输出tokens量",default=0)
    total_price:float = Field(description="消费金额",default=0.0)
    create_time:StrictStr= Field(description="创建时间",default="")
    currency_type:StrictInt = Field(description="计费单价货币类型,0:人名币 1:美元")
    price_type: conlist(constr(regex=r'^(thousand|million)$'), min_items=2, max_items=2)
    referprice_in: StrictFloat = Field(description="输入tokens计费单价", default=0.0)
    referprice_out: StrictFloat = Field(description="输出tokens计费单价", default=0.0)