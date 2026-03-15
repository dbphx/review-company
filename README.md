# ReviewCT (review-company) - Monorepo Overview

Monorepo cho nền tảng review công ty gồm portal người dùng, admin portal riêng domain/port, backend API và hạ tầng local.

- Thư mục project hiện tại: `.../src/github.com/review-company`
- Go module backend: `github.com/review-company/backend`

## Mục tiêu hiện tại
- Người dùng tìm kiếm công ty nhanh (Elasticsearch), đọc/viết review ẩn danh, bình luận và reply theo thread.
- Admin đăng nhập bằng email/mật khẩu, quản lý review/comment, có quyền xóa nội dung (RBAC).
- Dashboard admin hiển thị số liệu thật từ API (không hardcode), bao gồm biểu đồ review 7 ngày gần nhất.

## Kiến trúc
1. `frontend` (port `5173`): Portal người dùng React + Vite + TypeScript.
2. `admin-frontend` (port `5174`): Portal quản trị tách riêng React + Vite + TypeScript.
3. `backend` (port `3000`): API Golang Fiber + PostgreSQL + Elasticsearch.
4. `infrastructure`: Docker Compose cho PostgreSQL, Elasticsearch, Kibana.

## Tính năng đã có
- Search công ty với autocomplete và empty-state an toàn.
- Company detail + review list + review form.
- Comment/reply thread cho user thường, lưu bền vững sau refresh.
- Like/Dislike cho review và comment (mỗi session một lựa chọn, có thể đổi lựa chọn).
- Hiển thị số bình luận ngay cả khi thread đang collapse (`Bình luận (x)`).
- Chỉ hiển thị 3 comment đầu khi mở thread, có nút `Xem thêm bình luận`.
- Admin login bằng email/mật khẩu (`/api/admin/login`) + JWT.
- RBAC theo role `ADMIN`/`MOD`.
- Admin quản lý danh sách công ty (thêm/sửa/xóa, chặn xóa khi công ty đã có review).
- Admin quản lý review có filter tìm theo tên công ty.
- Quản lý tài khoản admin/mod (ADMIN mới được tạo/xóa user quản trị).
- Quản lý phiên đăng nhập Redis (TTL 1h), ADMIN có thể thu hồi phiên của MOD.
- Xóa comment theo cascade toàn thread con.
- Xóa review kèm xử lý comments liên quan và cập nhật lại thống kê công ty.
- Dashboard admin lấy số liệu thật:
  - `/api/companies/stats/summary`
  - `/api/reviews/stats/daily?days=7`
- Dashboard có top công ty điểm cao/thấp theo mode dữ liệu, top 10/20/50, chart gradient.
- Tracking lượt truy cập unique theo IP/1h (theo ngày/tháng).
- Chế độ dữ liệu toàn hệ thống `v1 | v2 | all` cho cả admin và portal.
- Route error boundary cho cả user portal và admin portal.
- Scraper CLI hỗ trợ nhiều mode:
  - `--mode=companies`
  - `--mode=reviews-1900`
  - `--mode=reviews-1900-full`
  - `--mode=all`

## Chạy nhanh local
1. Hạ tầng:
   - `cd infrastructure`
   - `cp .env.example .env`
   - `docker-compose up -d`
2. Backend:
   - `cd backend`
   - `go run cmd/api/main.go`
3. Seed dữ liệu mẫu:
   - `cd backend`
   - `go run cmd/scraper/main.go --mode=companies`
   - hoặc `go run cmd/scraper/main.go --mode=all`
4. Frontend user:
   - `cd frontend && npm install && npm run dev`
5. Frontend admin:
   - `cd admin-frontend && npm install && npm run dev`

## Deploy prod theo domain (app/admin/api)

Bạn có thể chạy theo mô hình:
- `https://app.b.c` -> user portal
- `https://admin.b.c` -> admin portal
- `https://api.b.c` -> backend API

Đã có sẵn bộ config mẫu:
- `config/prod/deploy.env.example`
- `config/prod/backend.env.example`
- `config/prod/frontend.app.env.example`
- `config/prod/frontend.admin.env.example`
- `config/prod/nginx/default.conf.template`
- `docker-compose.prod.yml`
- `start.prod.sh`

Quick start:
1. `cp config/prod/*.example` thành các file `.env` tương ứng (script sẽ auto tạo nếu thiếu)
2. chỉnh domain + secret + DB password trong file env
3. chạy `./start.prod.sh`

Bạn có thể set domain trực tiếp khi chạy script (script sẽ tự ghi vào env):
- `APP_DOMAIN=app.real.com ADMIN_DOMAIN=admin.real.com API_DOMAIN=api.real.com ./start.prod.sh`

Lưu ý CORS backend production:
- đặt `CORS_ORIGINS=https://app.b.c,https://admin.b.c` trong `config/prod/backend.env`

Lưu ý host check của Vite (tránh 403 từ reverse proxy):
- `config/prod/frontend.app.env` đặt `VITE_ALLOWED_HOSTS=<app-domain>`
- `config/prod/frontend.admin.env` đặt `VITE_ALLOWED_HOSTS=<admin-domain>`

## Tài liệu chi tiết
- `frontend/REQUIREMENTS.md`
- `admin-frontend/REQUIREMENTS.md`
- `backend/REQUIREMENTS.md`
- `backend/cmd/scraper/REQUIREMENTS.md`
- `infrastructure/REQUIREMENTS.md`
