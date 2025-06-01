# 电子木鱼游戏后端 (Electronic Wooden Fish Backend)

一个功能完整的电子木鱼游戏后端服务，使用Go语言构建，支持用户认证、分数管理、实时排行榜和多人游戏功能。

## 🎯 功能特性

### ✅ 已完成功能
- **用户系统**
  - 用户注册/登录系统
  - JWT身份认证
  - 用户资料管理
  - 密码哈希存储

- **第三方登录(SSO)**
  - ✅ Google登录集成
  - ✅ Apple登录集成
  - ✅ 自动用户账户关联

- **刷新令牌机制**
  - ✅ 安全的令牌轮换
  - ✅ 7天有效期管理
  - ✅ 数据库存储与验证

- **分数系统**
  - ✅ 分数上传和获取
  - ✅ 个人最佳记录
  - ✅ 历史分数查询
  - ✅ 游戏模式支持

- **排行榜系统**
  - ✅ 全球排行榜
  - ✅ 每日/周排行榜
  - ✅ Redis缓存优化
  - ✅ 实时更新支持

- **实时功能**
  - ✅ WebSocket实时多人游戏
  - ✅ 房间系统
  - ✅ 实时分数同步
  - ✅ 排行榜实时更新

- **API限流**
  - ✅ Redis滑动窗口算法
  - ✅ 基于IP和用户的限流
  - ✅ 不同接口的差异化限制

- **日志与监控**
  - ✅ 结构化日志系统
  - ✅ 多级别日志输出
  - ✅ 错误追踪和调试

- **容器化部署**
  - ✅ Docker容器支持
  - ✅ Docker Compose编排
  - ✅ Nginx反向代理配置

### 🚧 开发中功能
- 单元测试覆盖
- API文档生成
- 性能监控仪表板

## 技术栈

- **框架**: Gin (HTTP 框架)
- **数据库**: PostgreSQL + GORM
- **认证**: JWT
- **实时通信**: WebSocket (Gorilla WebSocket)
- **加密**: bcrypt

## API 接口

### 认证相关
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/google` - Google SSO（待实现）
- `POST /api/v1/auth/apple` - Apple SSO（待实现）
- `POST /api/v1/auth/refresh` - 刷新令牌（待实现）

### 用户相关
- `GET /api/v1/profile` - 获取用户资料
- `PUT /api/v1/profile` - 更新用户资料

### 分数相关
- `POST /api/v1/score` - 提交分数
- `GET /api/v1/score` - 获取用户最佳分数
- `GET /api/v1/scores/history` - 获取分数历史

### 排行榜
- `GET /api/v1/leaderboard` - 全球排行榜
- `GET /api/v1/leaderboard/daily` - 每日排行榜
- `GET /api/v1/leaderboard/weekly` - 周排行榜

### 实时游戏
- `GET /api/v1/ws?room_id=房间ID` - WebSocket 连接

## 环境变量

创建 `.env` 文件并配置以下环境变量：

```env
PORT=8080
DATABASE_URL=postgres://username:password@localhost/electronic_muyu?sslmode=disable
JWT_SECRET=your-super-secret-jwt-key
REDIS_URL=redis://localhost:6379

# Google SSO
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret

# Apple SSO
APPLE_TEAM_ID=your-apple-team-id
APPLE_KEY_ID=your-apple-key-id
APPLE_CLIENT_ID=your-apple-client-id
APPLE_PRIVATE_KEY=your-apple-private-key
```

## 快速开始

### 1. 安装依赖
```bash
go mod download
```

### 2. 设置数据库
确保 PostgreSQL 正在运行，并创建数据库：
```sql
CREATE DATABASE electronic_muyu;
```

### 3. 运行服务
```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动。

### 4. 健康检查
```bash
curl http://localhost:8080/health
```

## WebSocket 实时游戏

### 连接
```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws?room_id=room1');
```

### 消息格式

#### 发送分数更新
```json
{
  "type": "score_update",
  "data": {
    "score": 1000,
    "increment": 10
  }
}
```

#### 接收其他玩家分数更新
```json
{
  "type": "score_update",
  "data": {
    "user_id": 123,
    "username": "player1",
    "current_score": 1000,
    "action": "score_update",
    "timestamp": 1622547800
  }
}
```

## 数据库表结构

### users 表
- id (主键)
- username (用户名，唯一)
- email (邮箱，唯一)
- password (加密密码)
- nickname (昵称)
- avatar (头像URL)
- google_id, apple_id (SSO ID)
- total_score, highest_score (统计信息)
- games_played (游戏次数)
- created_at, updated_at

### scores 表
- id (主键)
- user_id (外键)
- score (分数)
- game_mode (游戏模式)
- duration (游戏时长)
- created_at

### game_sessions 表
- id (会话ID)
- room_id (房间ID)
- user_id (用户ID)
- current_score (当前分数)
- is_active (是否活跃)
- joined_at, left_at

### refresh_tokens 表
- id (主键)
- user_id (用户ID)
- token (刷新令牌)
- expires_at (过期时间)

## 开发计划

- [ ] 完善 Google/Apple SSO 集成
- [ ] 实现 Refresh Token 机制
- [ ] 添加 Redis 缓存支持
- [ ] 实现游戏房间管理
- [ ] 添加用户好友系统
- [ ] 实现成就系统
- [ ] 添加 API 限流
- [ ] 完善错误处理和日志
- [ ] 添加单元测试
- [ ] Docker 容器化部署

## 许可证

MIT License
