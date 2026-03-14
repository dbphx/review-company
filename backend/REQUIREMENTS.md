# Backend API Requirements - Golang Fiber

Hệ thống API Backend quản lý dữ liệu, tìm kiếm công ty và xử lý nghiệp vụ cho Nền Tảng Review.

## 1. Công Nghệ Sử Dụng (Tech Stack)
- **Ngôn ngữ:** Golang (Phiên bản 1.21 trở lên).
- **Web Framework:** Fiber (Cú pháp gọn, hiệu suất cực cao).
- **Database Chính:** PostgreSQL.
- **ORM / Query Builder:** `gorm`.
- **Search Engine:** Elasticsearch (Thư viện `elastic/go-elasticsearch`).
- **Auth user portal:** Google token exchange endpoint.
- **Auth admin:** Email/Password hash (SHA-256 phía client) + bcrypt + JWT.
- **Session store:** Redis (TTL 1h cho phiên admin/mod).
- **Kiến trúc Thiết kế:** Mô hình 3-tier (Handler -> Service -> Repository).

## 2. API Cốt Lõi (Public)
Không yêu cầu đăng nhập đối với mọi người dùng.

### 2.1. API Tìm Kiếm (Search)
- `GET /api/search?q=abc` -> Truy vấn **Elasticsearch** trả về danh sách công ty nhanh chóng.
- Trả về thông tin cơ bản: Tên, Logo, Ngành nghề, Điểm trung bình.

### 2.2. API Công Ty (Company)
- `GET /api/companies/{id}` -> Lấy chi tiết công ty.
- `GET /api/companies/top` -> Các công ty nổi bật.
  - hỗ trợ `order=asc|desc` để lấy top thấp/cao theo rating.
- `GET /api/companies/stats/summary` -> Tổng số công ty, tổng review, avg rating toàn hệ thống.
- Cập nhật điểm trung bình + total review khi tạo/xóa review.

### 2.3. API Review & Comment
- `GET /api/companies/{id}/reviews` -> Danh sách review của công ty (Chỉ lấy review trạng thái `APPROVED`). Phân trang.
- `POST /api/companies/{id}/reviews` -> Lưu review mới vào PostgreSQL. (Lưu kèm IP, Session ID để dự phòng chống Spam, gán trạng thái `APPROVED` mặc định cho MVP).
- `GET /api/reviews/recent` -> Danh sách review mới nhất cho portal home.
- `GET /api/reviews/stats/daily?days=7` -> Số lượng review theo ngày (cho chart admin).
- `GET /api/reviews` -> Danh sách review có phân trang cho admin.
  - hỗ trợ filter `seed_version` và `created_date`.
- `GET /api/reviews/{id}` -> Chi tiết review.
- `GET /api/reviews/{id}/comments` -> Danh sách comment (Thảo luận) dưới 1 review.
- `POST /api/reviews/{id}/comments` -> Gửi bình luận reply vào review.

## 3. API Quản Trị Viên (Admin - Protected)
Phân quyền theo role `ADMIN`/`MOD`, với session Redis bắt buộc còn hiệu lực.

### 3.1. API Xác Thực (Auth)
- `POST /api/auth/google` -> Nhận Google access token từ frontend user portal, lấy userinfo và lưu/cập nhật bảng `users`.
- `POST /api/admin/login` -> Đăng nhập admin bằng email/password, trả JWT admin.

### 3.2. API Dashboard & Thống Kê
- `GET /api/companies/stats/summary`
- `GET /api/reviews/stats/daily?days=7`
- `GET /api/analytics/visits/daily?days=7`
- `GET /api/analytics/visits/monthly?months=6`
- `POST /api/analytics/visits` (portal tracking)

### 3.3. API Quản Lý (Moderation)
- `DELETE /api/reviews/{id}` -> Soft-delete review (admin only).
- `DELETE /api/comments/{id}` -> Xóa comment theo cascade thread con (admin only).
- `GET /api/admin/users` -> ADMIN/MOD xem danh sách user quản trị.
- `POST /api/admin/users` -> chỉ ADMIN tạo user ADMIN/MOD.
- `DELETE /api/admin/users/{id}` -> chỉ ADMIN xóa user quản trị.
- `GET /api/admin/sessions` -> danh sách phiên đang hoạt động.
- `DELETE /api/admin/sessions/{sessionId}` -> ADMIN thu hồi phiên MOD.
- `DELETE /api/admin/users/{id}/sessions` -> ADMIN thu hồi toàn bộ phiên của MOD.
- `GET /api/admin/data-mode` -> xem mode dữ liệu (`v1|v2|all`).
- `POST /api/admin/data-mode` -> ADMIN đổi mode dữ liệu toàn hệ thống.

## 4. Thiết Kế Cơ Sở Dữ Liệu (PostgreSQL Schema)
Thiết kế tập trung vào việc dễ dàng chống Spam trong tương lai.

### Bảng `companies`
- `id` (UUID, PK)
- `name` (Varchar, Indexed)
- `logo_url` (Varchar)
- `website` (Varchar)
- `industry` (Varchar)
- `size` (Varchar)
- `description` (Text)
- `avg_rating` (Numeric - Cache)
- `total_reviews` (Int)
- *Ghi chú: Sync qua Elasticsearch.*

### Bảng `reviews`
- `id` (UUID, PK)
- `company_id` (UUID, FK -> companies)
- `author_name` (Varchar - Mặc định "Ẩn danh")
- `rating` (Int 1-5)
- `title` (Varchar)
- `content` (Text - Ý kiến, Ưu/Nhược điểm)
- `salary_gross` (Numeric - Nullable)
- `interview_exp` (Text - Nullable)
- `status` (Enum: `APPROVED`, `HIDDEN`, `DELETED`)
- `seed_version` (`v1`, `v2`) để tách dữ liệu khởi tạo và dữ liệu mới.
- `ip_address` (Varchar - Chống Spam)
- `device_id` / `session_id` (Varchar)
- `created_at` (Timestamp)
- `deleted_at` (Timestamp - Cho Soft Delete)

### Bảng `comments`
- `id` (UUID, PK)
- `review_id` (UUID, FK -> reviews)
- `parent_comment_id` (UUID, FK -> comments, Nullable - Reply lồng nhau)
- `author_name` (Varchar)
- `content` (Text)
- `status` (Enum)
- `ip_address` (Varchar)
- `created_at` (Timestamp)
- `deleted_at` (Timestamp)

### Bảng `admin_users`
- `id` (UUID, PK)
- `email` (Varchar, Unique)
- `password` (bcrypt hash)
- `name` (Varchar)
- `role` (Enum/String: `ADMIN` hoặc `MOD`)
- `created_at`, `updated_at`

### Bảng `users` (Dành cho người dùng portal)
- `id` (UUID, PK)
- `auth0_id` (Google user id hiện tại)
- `email` (Varchar, Unique)
- `provider` (Google)
- `role` (default `MODERATOR`)
- `created_at` (Timestamp)
- `allowed_companies` qua bảng nối `user_allowed_companies`.

### Bảng `visits`
- `id` (UUID, PK)
- `path` (Varchar)
- `session_id` (Varchar)
- `ip_address` (Varchar)
- `created_at` (Timestamp)

### Bảng `system_settings`
- `key` (PK)
- `value`
- `updated_at`
