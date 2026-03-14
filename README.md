# Review Company - Monorepo Overview

Monorepo cho nền tảng review công ty gồm portal người dùng, admin portal riêng domain/port, backend API và hạ tầng local.

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
- Hiển thị số bình luận ngay cả khi thread đang collapse (`Bình luận (x)`).
- Chỉ hiển thị 3 comment đầu khi mở thread, có nút `Xem thêm bình luận`.
- Admin login bằng email/mật khẩu (`/api/admin/login`) + JWT.
- RBAC: chỉ admin được gọi API xóa review/comment.
- Xóa comment theo cascade toàn thread con.
- Xóa review kèm xử lý comments liên quan và cập nhật lại thống kê công ty.
- Dashboard admin lấy số liệu thật:
  - `/api/companies/stats/summary`
  - `/api/reviews/stats/daily?days=7`
- Route error boundary cho cả user portal và admin portal.

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
   - `go run cmd/scraper/main.go`
4. Frontend user:
   - `cd frontend && npm install && npm run dev`
5. Frontend admin:
   - `cd admin-frontend && npm install && npm run dev`

## Tài liệu chi tiết
- `frontend/REQUIREMENTS.md`
- `admin-frontend/REQUIREMENTS.md`
- `backend/REQUIREMENTS.md`
- `backend/cmd/scraper/REQUIREMENTS.md`
- `infrastructure/REQUIREMENTS.md`
