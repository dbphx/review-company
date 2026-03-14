# Scraper/Seeder Requirements - Golang CLI

Hiện tại command trong `cmd/scraper` đang đóng vai trò **seeder dữ liệu mẫu** cho môi trường dev (thay vì scraper production hoàn chỉnh).

## 1. Trạng thái hiện tại
- Tạo dữ liệu mẫu công ty vào PostgreSQL.
- Đồng bộ index công ty sang Elasticsearch thông qua repository.
- Seed tài khoản admin mặc định:
  - email: `admin@review.com`
  - password: `admin123`
- Chạy thủ công bằng CLI cho môi trường local/dev.

## 2. Dữ liệu được seed
- Danh sách công ty mẫu: FPT Software, VNG, Shopee, Tiki, Momo.
- Thuộc tính: tên, logo, website, industry, size, description.

## 3. Quy trình chạy seed
1. Load cấu hình `.env`.
2. Kết nối PostgreSQL + Elasticsearch.
3. Insert/upsert dữ liệu công ty mẫu.
4. Index dữ liệu công ty vào Elasticsearch.
5. Seed admin user mặc định nếu chưa tồn tại.

## 4. Cách chạy
- `go run cmd/scraper/main.go`

## 5. Lưu ý tương lai
- Có thể mở rộng lại thành scraper production (colly/goquery) và tách command `scraper` + `seeder` riêng.
- Có thể đưa vào cron để cập nhật công ty định kỳ.
