# Infrastructure Requirements - Docker

Thiết lập cơ sở hạ tầng cho môi trường phát triển (Development) bằng Docker Compose, chuẩn bị sẵn sàng cho việc triển khai (Deployment) trong tương lai.

## 1. Môi Trường Phát Triển (Local / Dev)
Dự án sử dụng **Docker Compose** để quản lý các thành phần cơ sở dữ liệu và search engine, giúp developer (hoặc máy chủ chạy lần đầu) có thể setup mọi thứ chỉ với 1 lệnh:
`docker-compose up -d`

## 2. Các Dịch Vụ (Services) Cần Thiết
Bao gồm 3 thành phần đang dùng ở local:

### 2.1. PostgreSQL (Cơ Sở Dữ Liệu Chính)
- **Version:** 15-alpine (hoặc mới nhất).
- **Mục đích:** Lưu trữ toàn bộ dữ liệu (Companies, Reviews, Comments, Users/Admins).
- **Cấu hình (Volumes):** Lưu trữ data ra thư mục `.postgres-data` để không mất dữ liệu khi tắt container.
- **Mật khẩu & User:** Cấu hình qua biến môi trường (Environment Variables) trong file `docker-compose.yml`.
- **Khởi tạo ban đầu:** Có script `init.sql` tự động tạo Database `review_cong_ty_db` khi container chạy lần đầu tiên.
  - *Trạng thái hiện tại:* schema được tạo/migrate bởi backend GORM.

### 2.2. Elasticsearch (Search Engine)
- **Version:** 8.x (Single-node cluster cho MVP).
- **Mục đích:** Index dữ liệu công ty (Tên, Ngành nghề) để phục vụ tính năng tìm kiếm (Search Bar) siêu tốc trên Web.
- **Cấu hình:** Tắt tính năng bảo mật (Security) tạm thời cho môi trường Dev (`xpack.security.enabled=false` hoặc cấu hình pass đơn giản) để dễ kết nối từ backend Golang.
- **Lưu trữ (Volumes):** Lưu trữ data ES ra thư mục `.es-data`.

### 2.3. (Tùy Chọn) Kibana
- Kibana đang được bật trong docker-compose để kiểm tra index/search dữ liệu nhanh.

## 3. Kiến Trúc Triển Khai (Deployment Architecture - Tương Lai)
Khi Deploy lên server thật (VPS như DigitalOcean, AWS EC2):
- **Frontend (Vite):** Build ra static files (HTML, CSS, JS) và phục vụ qua Nginx hoặc Vercel/Netlify.
- **Backend (Go Fiber):** Chạy như 1 service (Systemd hoặc Docker Container).
- **Database & ES:** Có thể dùng managed services (AWS RDS, Elastic Cloud) hoặc tự deploy Docker trên 1 con VPS riêng để đảm bảo an toàn.

## 4. Các Biến Môi Trường (Environment Variables - `.env`)
Tất cả các service (Frontend, Backend, Bot) đều cần cấu hình kết nối thông qua file `.env`. `docker-compose.yml` sẽ đọc các thông số từ đây.
Ví dụ:
- `DB_HOST=localhost`
- `DB_PORT=5432`
- `DB_USER=postgres`
- `DB_PASSWORD=secret`
- `ES_URL=http://localhost:9200`
- `GOOGLE_OAUTH_CLIENT_ID=...`
- `GOOGLE_OAUTH_SECRET=...`
- `ADMIN_JWT_SECRET=...`

## 5. Ports local
- PostgreSQL: `5432`
- Elasticsearch: `9200`
- Kibana: `5601`
- Backend API: `3000`
- Portal frontend: `5173`
- Admin frontend: `5174`
