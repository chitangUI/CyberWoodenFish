# 电子木鱼游戏后端 API 测试

这个文件包含了测试电子木鱼游戏后端 API 的示例请求。

## 环境设置

确保服务器在 `http://localhost:8080` 运行。

## 1. 用户注册

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "nickname": "测试用户"
  }'
```

## 2. 用户登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

保存返回的 token，用于后续认证。

## 3. 获取用户资料

```bash
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## 4. 提交分数

```bash
curl -X POST http://localhost:8080/api/v1/score \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "score": 1000,
    "game_mode": "normal",
    "duration": 120
  }'
```

## 5. 获取用户最佳分数

```bash
curl -X GET http://localhost:8080/api/v1/score \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## 6. 获取分数历史

```bash
curl -X GET "http://localhost:8080/api/v1/scores/history?page=1&limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## 7. 获取全球排行榜

```bash
curl -X GET "http://localhost:8080/api/v1/leaderboard?page=1&limit=20" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## 8. 获取今日排行榜

```bash
curl -X GET "http://localhost:8080/api/v1/leaderboard/daily?page=1&limit=20" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## 9. 获取本周排行榜

```bash
curl -X GET "http://localhost:8080/api/v1/leaderboard/weekly?page=1&limit=20" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## 10. WebSocket 连接测试 (JavaScript)

```html
<!DOCTYPE html>
<html>
<head>
    <title>电子木鱼游戏 WebSocket 测试</title>
</head>
<body>
    <div id="messages"></div>
    <button onclick="sendScore()">发送分数</button>
    <script>
        const token = 'YOUR_TOKEN_HERE';
        const ws = new WebSocket('ws://localhost:8080/api/v1/ws?room_id=test_room', 
            [], { headers: { 'Authorization': 'Bearer ' + token } });
        
        ws.onopen = function(event) {
            console.log('WebSocket connected');
            document.getElementById('messages').innerHTML += '<p>已连接</p>';
        };
        
        ws.onmessage = function(event) {
            const data = JSON.parse(event.data);
            console.log('Received:', data);
            document.getElementById('messages').innerHTML += 
                '<p>收到: ' + JSON.stringify(data, null, 2) + '</p>';
        };
        
        function sendScore() {
            const message = {
                type: 'score_update',
                data: {
                    score: Math.floor(Math.random() * 1000),
                    increment: 10
                }
            };
            ws.send(JSON.stringify(message));
        }
        
        ws.onerror = function(error) {
            console.error('WebSocket error:', error);
        };
        
        ws.onclose = function(event) {
            console.log('WebSocket closed');
        };
    </script>
</body>
</html>
```

## 健康检查

```bash
curl http://localhost:8080/health
```

## 数据库设置

在运行服务器之前，确保 PostgreSQL 已安装并运行：

```sql
-- 创建数据库
CREATE DATABASE electronic_muyu;

-- 创建用户（可选）
CREATE USER muyu_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE electronic_muyu TO muyu_user;
```

## 环境变量

复制 `.env.example` 到 `.env` 并配置：

```bash
cp .env.example .env
```

编辑 `.env` 文件设置数据库连接和其他配置。
