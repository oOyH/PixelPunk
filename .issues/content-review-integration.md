# Content Review 前后端集成任务

## 发现的问题

### 1. 后端缺失的路由端点 ✅ 已修复
前端定义了以下API但后端未实现：
- `POST /admin/content-review/batch-hard-delete` - 批量硬删除
- `POST /admin/content-review/files/:fileId/restore` - 恢复已删除文件
- `POST /admin/content-review/batch-restore` - 批量恢复已删除文件

**修复措施：**
- 在 `content_review_controller.go` 中添加 `BatchHardDeleteReviewedFiles`、`RestoreReviewedFile`、`BatchRestoreReviewedFiles` 函数
- 在 `review_service.go` 中添加 `RestoreSoftDeletedFile` 函数和 `sendFileRestoreNotification` 通知函数
- 在 `admin_content_review_routes.go` 中注册新路由

### 2. 前后端类型不一致 ✅ 已修复
- `BatchReviewResult.results` 前端定义为 `string[]`，但后端返回 `map[string]string`

**修复措施：**
- 将前端类型修改为 `Record<string, string>`

### 3. GetFileDetail逻辑修复 ✅ 已修复
- 原代码检查 `file.NSFW` 标志，应该检查文件状态

**修复措施：**
- 修改为检查 `pending_review`、`deleted`、`pending_deletion` 等审核相关状态

## 完成的修改

### 后端文件
1. `internal/controllers/admin/content_review_controller.go`
   - 添加 `BatchFileIDsDTO` 结构体
   - 添加 `BatchHardDeleteReviewedFiles` 函数
   - 添加 `RestoreReviewedFile` 函数
   - 添加 `BatchRestoreReviewedFiles` 函数
   - 修复 `GetFileDetail` 的状态检查逻辑

2. `internal/services/review/review_service.go`
   - 添加 `RestoreSoftDeletedFile` 函数
   - 添加 `sendFileRestoreNotification` 通知函数

3. `internal/routes/admin_content_review_routes.go`
   - 注册 `/batch-hard-delete` 路由
   - 注册 `/files/:fileId/restore` 路由
   - 注册 `/batch-restore` 路由

### 前端文件
1. `web/src/api/admin/content-review/index.ts`
   - 修复 `BatchReviewResult.results` 类型为 `Record<string, string>`

## 状态: ✅ 已完成

---

## 追加修复: Docker构建问题

### 问题描述
Docker构建的镜像前端请求指向 `http://localhost:9520/api/v1`，导致连接失败。

### 根因分析
`.dockerignore` 中排除了所有 `.env.*` 文件（第90-91行），导致 `web/.env.production` 在Docker构建时不被复制，Vite使用默认配置构建。

### 修复措施
修改 `.dockerignore`，只排除本地环境文件：
```diff
- .env
- .env.*
+ .env
+ .env.local
+ .env.*.local
+ # 保留 web/.env.production 用于前端构建
```

### 验证方式
重新构建Docker镜像后，前端应使用相对路径 `/api/v1` 发起请求。


