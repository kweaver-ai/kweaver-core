# BizDomain Test Handler

业务域HTTP接口测试handler模块，按功能拆分成多个独立文件。

## 📁 文件结构

```
bizdomain/
├── README.md                        # 本文件
├── common.go                        # 公共请求结构体
├── handler.go                       # Handler定义和路由注册
├── associate_resource.go            # 资源关联测试handler
├── query_resource_associations.go   # 关联关系查询测试handler
├── disassociate_resource.go         # 资源取消关联测试handler
├── has_resource_association.go      # 单个资源关联检查handler
└── full_flow.go                     # 完整流程测试handler
```

## 📝 文件说明

### common.go
- 定义公共的请求结构体 `TestBizDomainHttpRequest`
- 包含 `agent_id` 字段

### handler.go
- 定义 `BizDomainTestHandler` 结构体
- 实现 `NewBizDomainTestHandler()` 构造函数
- 实现 `RegisterRoutes()` 方法注册所有路由

### associate_resource.go
- 实现 `AssociateResourceTestHandler()` 方法
- 处理资源关联测试请求
- 路由: `POST /test/bizdomain/associate-resource`

### query_resource_associations.go
- 实现 `QueryResourceAssociationsTestHandler()` 方法
- 处理关联关系查询测试请求
- 路由: `POST /test/bizdomain/query-resource-associations`

### disassociate_resource.go
- 实现 `DisassociateResourceTestHandler()` 方法
- 处理资源取消关联测试请求
- 路由: `POST /test/bizdomain/disassociate-resource`

### has_resource_association.go
- 实现 `HasResourceAssociationTestHandler()` 方法
- 处理单个资源关联检查测试请求
- 路由: `POST /test/bizdomain/has-resource-association`

### full_flow.go
- 实现 `TestBizDomainHttp()` 方法
- 处理完整流程测试请求（关联→查询→取消关联）
- 路由: `POST /test/bizdomain-http`

## 🔌 使用方式

在 `testhandler/define.go` 中：

```go
import (
    "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-factory/src/driveradapter/api/public/v3/testhandler/bizdomain"
)

type testHTTPHandler struct {
    bizDomainHandler *bizdomain.BizDomainTestHandler
}

func (t *testHTTPHandler) RegPubRouter(router *gin.RouterGroup) {
    // 委托给bizdomain handler注册路由
    t.bizDomainHandler.RegisterRoutes(router)
}

func NewTestHTTPHandler() ihandlerportdriver.IHTTPRouter {
    handlerOnce.Do(func() {
        _handler = &testHTTPHandler{
            bizDomainHandler: bizdomain.NewBizDomainTestHandler(),
        }
    })
    return _handler
}
```

## 🎯 设计原则

1. **单一职责**: 每个文件只负责一个具体的handler功能
2. **模块化**: 便于维护和扩展
3. **清晰结构**: 文件名直接反映功能
4. **统一风格**: 所有handler遵循相同的代码结构

## 🔄 添加新的测试端点

1. 在 `bizdomainsvc` 中添加新的测试方法
2. 在 `bizdomain/` 目录下创建新的handler文件
3. 在新文件中实现handler方法
4. 在 `handler.go` 的 `RegisterRoutes()` 中注册新路由

## 📚 相关文件

- Service层: `/src/domain/service/v3/bizdomainsvc/test_tmp.go`
- 测试脚本: `/.local/5003/Makefile`
- 测试文档: `/.local/5003/README.md`
