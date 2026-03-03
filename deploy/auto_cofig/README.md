# 使用说明

### 前置条件
1. 登录系统工作台，信息安全管理-统一身份认证-账户-用户中新建用户：test，角色与访问策略-角色管理，将数据管理员、AI管理员、应用管理员角色中添加用户test。
2. 访问 https://ip/studio 登录页，登录test用户（默认密码：123456）下方提示修改密码，修改密码为111111。

### 执行示例

1. 将脚本放到kweaver环境任意目录并赋予脚本执行权限：

```bash
chmod +x auto_config.sh
```

2. 导入供应链配置：

```bash
./auto_config.sh agent.json HD供应链业务知识网络.json dataflow.json
./auto_config.sh --step 7 contextloader工具集_020.adp
./auto_config.sh --step 7 基础结构化数据分析工具箱2.adp
```

3. 登录 Studio 进行后续配置：

- 使用 `test` 账号登录 Studio，登录地址：https://ip/studio 
- BKN引擎-业务知识网络中编辑**HD供应链业务知识网络**中对象类、关系类、行动类等，关联导入的数据源
- 决策智能体-开发-决策智能体中编辑 **HD供应链业务问答助手** 中知识来源-业务知识网络中添加**HD供应链业务知识网络**在默认模型配置中选择添加模型然后将技能中工具用到模型的选项中选择添加的模型然后发布
