# Sequence Diagrams / 时序图

## Login & Navigation / 登录与跳转
```mermaid
sequenceDiagram
  participant User as User
  participant LoginView as Login View
  participant AuthStore as useAuthStore
  participant AuthService as AuthenticationService
  participant AccessStore as useAccessStore
  participant Router as Vue Router

  User->>LoginView: Submit username/password
  LoginView->>AuthStore: authLogin(params)
  AuthStore->>AuthStore: Encrypt password (AES key from env)
  AuthStore->>AuthService: POST /login (username, encrypted password, grant_type=password)
  AuthService-->>AuthStore: LoginResponse (access_token, refresh_token)
  AuthStore->>AccessStore: setAccessToken(token)
  AuthStore->>AuthStore: fetchUserInfo() & fetchAccessCodes()
  AuthStore->>AccessStore: setAccessCodes(codes)
  AuthStore->>Router: push(homePath or DEFAULT_HOME_PATH)
  Router-->>User: Render authorized layout
```

## Protected Request with Refresh / 受保护请求与刷新
```mermaid
sequenceDiagram
  participant View as Any View
  participant Client as requestClient
  participant Backend as API
  participant Auth as useAuthStore
  participant Access as useAccessStore

  View->>Client: request (with params)
  Client->>Backend: Add Authorization + Accept-Language headers
  Backend-->>Client: 401 Unauthorized (token expired)
  Client->>Auth: doRefreshToken()
  Auth->>Backend: POST /refresh_token (refresh_token)
  Backend-->>Auth: new access_token
  Auth->>Access: setAccessToken(new token)
  Client->>Backend: retry original request with new token
  Backend-->>Client: 200 OK + data
  Client-->>View: unwrap data
```

## Notes / 说明
- Title auto-updates via `watchEffect` on route meta + app name when `dynamicTitle` enabled.
- Guards verify access codes vs route meta and redirect to `/login` with redirect query when unauthorized.
- Error interceptor surfaces backend messages via `ant-design-vue` `message.error`; customize in `utils/request.ts`.
