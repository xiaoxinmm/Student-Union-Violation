# 学生会违纪管理系统

学校学生会内部使用的违纪记录管理系统，用于日常查寝、纪律检查等场景的违纪信息录入、查询、公示和导出。

原来是 PHP 写的，问题太多（SQL 注入、明文密码、按日期建表……），索性用 Go 重写了一版。

## 功能

- **违纪录入** — 学生会成员登录后录入违纪信息（宿舍号、姓名、班级、时间段、原因、部门、执勤人），支持上传胸卡照片
- **今日公示** — 当天违纪记录一览，20 秒自动刷新，适合投屏展示
- **审查管理** — 管理员查看全部记录，支持按日期/关键词筛选、删除记录、查看照片
- **数据导出** — 按日期导出 CSV，Excel 可以直接打开
- **用户管理** — 管理员可添加/删除用户、重置密码

## 技术栈

- Go + Gin
- MySQL 8.0
- 前端原生 HTML/CSS/JS，没用框架

## 部署

### Docker（推荐）

```bash
docker-compose up -d
```

启动后访问 `http://localhost:8080`，默认管理员账号 `admin`，密码 `admin123`。

MySQL 数据持久化在 Docker volume 里，不会因为重启丢数据。

### 手动部署

需要：
- Go 1.22+
- MySQL 8.0

```bash
# 建库
mysql -u root -p -e "CREATE DATABASE suv CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 编译
CGO_ENABLED=0 go build -o server ./cmd/server

# 配置环境变量
export DB_HOST=127.0.0.1
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=你的密码
export DB_NAME=suv
export JWT_SECRET=随便写一个长字符串
export PORT=8080

# 启动
./server
```

首次启动会自动建表和创建默认管理员。

## 目录结构

```
cmd/server/       程序入口
internal/
  config/         环境变量读取
  database/       数据库连接和建表
  handler/        请求处理
  middleware/     JWT 认证、CSRF、权限控制
  model/          数据结构定义
web/
  static/css/     样式
  static/js/      前端逻辑
  templates/      HTML 页面
```

## 相比旧版改了什么

- 修了 SQL 注入（全部用 prepared statement）
- 密码改成 bcrypt 哈希，不再明文存
- 文件上传加了类型白名单和大小限制（5MB）
- 不再按日期建表，统一一张 violations 表
- 加了 JWT 认证和 CSRF 防护
- 删除操作要二次确认
- 去掉了 PHP 探针（x.php）

## 注意事项

- 生产环境记得改 `JWT_SECRET` 和数据库密码
- 上传的照片存在 `uploads/` 目录（Docker 部署时是 volume）
- 导出的 CSV 带 BOM 头，Windows 下 Excel 打开不会乱码

## License

[AGPL-3.0](LICENSE)
