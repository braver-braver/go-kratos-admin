# 权限模型概览（API / 菜单 / 数据范围）

## 目标
- 明确 API 接口、菜单渲染、数据范围三类权限的规则与流程。
- 提供新增资源的操作指引，避免遗漏或越权。
- 让运维/开发可以快速定位“谁因为什么被拒绝”。

## 核心概念
- **认证**：JWT（`Authorization: Bearer <token>`）由认证中间件校验。
- **API 授权**：基于 Casbin 的策略，主体（用户/角色）→ 资源（路由/RPC）→ 行为（HTTP Method）。
- **菜单权限**：后端根据用户角色过滤菜单树，前端只渲染可见节点/按钮。
- **数据范围（部门/组织）**：按当前用户的数据范围对查询结果做行级过滤（本部门/含子部门/自定义/全部）。

## API 权限
- 执行位置：REST/gRPC 经过 `authn` + `authz`（Casbin）中间件。
- 白名单（无需鉴权）：`/admin/v1/login`，`/admin/v1/tenants_exists`。
- 身份来源：JWT 声明（用户 ID、租户 ID、角色列表），主体通常解析为角色或用户 ID。
- 资源定义：DB 中的路由/方法（见 `backend/sql/default-data.sql`），如 `/admin/v1/users` + `GET`。
- 新增 API 步骤：
  1) 定义路由/RPC；
  2) 在策略库中插入资源定义（路径、方法、模块、操作）；
  3) 绑定资源到角色（Casbin 策略）；
  4) 若需公开访问，则加入 REST 白名单。

## 菜单权限
- 后端返回已按角色过滤的菜单树；前端据此渲染页面/按钮，避免前端硬编码。
- 按钮/操作级权限同样以资源形式绑定角色，后端过滤不可见节点。
- 新增菜单节点：
  1) 在 DB 创建菜单节点（标题、路径、组件、父节点）；
  2) 关联所需的资源/权限标识；
  3) 分配给角色，同时确保相关 API 资源已授权。

## 数据权限（部门/组织范围）
- 目的：将列表/统计类查询限制在允许的组织范围内。
- 范围来源：用户档案 + 角色数据范围设置（仅本人、本部门、本部门含子级、自定义部门、全部）。
- 执行位置：Service/Repo 在查询前根据范围追加过滤条件（如 `dept_id IN (...)`）。
- 新增需数据范围的查询：
  1) 确定组织字段（如 `dept_id`、`tenant_id` 等）；
  2) 每次请求获取一次当前用户的数据范围；
  3) 在 list/count/export 等查询中应用范围过滤，避免绕过（直连 SQL 或缓存）。

## 时序图

### API 鉴权（JWT + Casbin）
```plantuml
@startuml
actor Client
participant Gateway as REST/gRPC
participant Authn as Authn MW
participant Authz as Casbin
participant Handler

Client -> Gateway: HTTP/gRPC + Authorization: Bearer
Gateway -> Authn: 校验 JWT
Authn -> Gateway: 返回 subject/claims (user, roles, tenant)
Gateway -> Authz: enforce(subject, object=route, action=method)
Authz --> Gateway: allow/deny
Gateway -> Handler: 允许则进入业务
Handler --> Client: 响应
Gateway --> Client: 401/403 拒绝
@enduml
```

### 菜单权限流程
```plantuml
@startuml
actor User
participant Frontend
participant Backend as Menu API
participant Authz

User -> Frontend: 打开应用
Frontend -> Backend: GET /menus (携带 JWT)
Backend -> Authz: 按用户角色过滤菜单节点
Authz --> Backend: 过滤后的菜单树
Backend --> Frontend: 返回菜单
Frontend -> User: 仅渲染可见的路由/按钮
@enduml
```

### 数据范围过滤
```plantuml
@startuml
actor User
participant Service
participant Scope as DataScope Resolver
participant Repo as Repository/DB

User -> Service: 列表请求
Service -> Scope: 解析用户数据范围（部门/租户/自定义）
Scope --> Service: 可用 ID 集合
Service -> Repo: 查询时附加范围过滤 (WHERE dept_id IN scope)
Repo --> Service: 返回结果
Service --> User: 返回过滤后的数据
@enduml
```

## 运维/自检清单
- JWT 有效且未过期，角色声明正确。
- Casbin 策略完整：所有受保护 API 均有资源与角色绑定；仅必要接口在白名单。
- 不同角色调用菜单接口，确认返回的菜单树符合预期。
- 数据范围解析与列表/导出接口一致地应用了范围过滤。
- 日志输出 `trace_id`、`span_id`、`X-Request-ID`；Jaeger 查询请使用 `Trace-Id`。
