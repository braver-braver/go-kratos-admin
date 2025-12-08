# Admin REST 请求链路：登录 -> `/admin/v1/me`

面向排查 403/401：列出从获取令牌到携带令牌访问用户资料的完整调用链与所涉中间件。

## 1) 登录获取令牌
- 入口：`POST /admin.service.v1.AuthenticationService/Login`（`/admin/v1/login`，具体路径以前端调用为准）。
- REST 中间件顺序：`tracing.Server` -> `logging.Server` -> `applogging.Server`；登录接口在白名单内，跳过 `authn.Server`、`auth.Server`、`authz.Server`。
- 业务处理：`AuthenticationService.Login` → `doGrantTypePassword`
  - `UserCredentialRepo.VerifyCredential` 校验用户名/密码。
  - `UserRepo.GetUserByUserName` 拉取用户；`checkAuthority` 仅允许 `SYS_ADMIN`/`TENANT_ADMIN`。
  - `RoleRepo.GetRoleCodesByRoleIds` 把角色编码写入 `user.Roles`。
  - `UserTokenCacheRepo.GenerateToken`：
    - `createAccessJwtToken` 用 `authn` JWT 签发访问令牌；写入 Redis hash `access_token_key_prefix + userId`。
    - `createRefreshToken` 生成刷新令牌；写入 Redis hash `refresh_token_key_prefix + userId`。
  - 返回：`access_token` + `refresh_token`（TokenType `bearer`）。

配置要点：`configs/server.yaml` 中 `server.rest.middleware.auth.*` 控制签名方式、Redis key 前缀与过期时间；`configs/auth.yaml` 中 `authn.jwt` 提供签名密钥/算法。

## 2) 携带令牌访问 `/admin/v1/me`
- 请求头：`Authorization: Bearer <access_token>`。
- REST 中间件顺序（白名单外）：
  1. `tracing.Server`、`logging.Server`
  2. `applogging.Server`（记录操作/登录日志）
  3. `selector.Server(...).Match(newRestWhiteListMatcher())`：
     - `authn.Server(authenticator)`：校验 JWT（签名/过期），将 AuthClaims 放入 context。
     - `auth.Server()`（`pkg/middleware/auth`）：从 AuthClaims 构造 `UserTokenPayload`，写入 metadata；默认未开启 Redis 存续性校验（可通过 `WithIsExistAccessToken` 挂 Redis 检查）。
     - `authz.Server(authorizer.Engine())`：基于 `authzClaims` 做访问控制。资源 = 路径模板（如 `/admin/v1/me`），动作 = HTTP 方法（GET）。
- Handler：`UserProfileService.GetUser`
  - `auth.FromContext` 取出 `UserTokenPayload`（userId/roles 等）。
  - `UserRepo.Get` 按 `Id` 查询用户；失败抛 `authenticationV1.ErrorNotFound`.
  - `RoleRepo.GetRoleCodesByRoleIds` 补齐 `user.Roles`；返回 `userV1.User`。

## 3) 授权数据来源
- `Authorizer` 初始化：`configs/auth.yaml` 默认 `authz.type=opa`。
- `Authorizer.ResetPolicies` 生成策略：
  - 从 `sys_roles.apis` 收集角色可访问的 API id 列表。
  - 从 `sys_api_resources` 读取 path/method，组合成 OPA policy（`assets/rbac.rego`）。
  - 角色码（例如 `super`）作为 subject；若令牌中无角色，使用 `userId` 字符串。
- `Authorizer.ResetPolicies` 在启动时执行一次（`Authorizer.init`）并在 `NewRESTServer` 尾部再次执行，确保策略加载完毕。

## 4) 排查 403/401 提示
- 401 / `ErrMissingJwtToken`：确认 `Authorization` 头是否正确携带 Bearer token。
- 401 / 签名错误：核对 `auth.yaml` 的 JWT key/算法与签发时一致。
- 403 / `FORBIDDEN`：
  - 登录时 `checkAuthority` 只允许 `SYS_ADMIN`、`TENANT_ADMIN`。
  - 若角色码为空，`authz` 会使用 `userId` 作为 subject，需确保对应策略存在。
  - 确认 `sys_roles.apis` 与 `sys_api_resources` 中存在 `/admin/v1/me` 的 path/method 组合，并已在启动后成功 `ResetPolicies`。
  - 若需要校验 Redis 中 token 是否仍存在，可为 `auth.Server` 增加 `WithIsExistAccessToken` 选项，并通过 `UserTokenCacheRepo.IsExistAccessToken` 检查。
