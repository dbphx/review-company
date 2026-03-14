# Scraper/Seeder Requirements - Golang CLI

Command `cmd/scraper` hiện là CLI seed dữ liệu cho môi trường dev, hỗ trợ nhiều mode.

## 1. Trạng thái hiện tại
- Kết nối PostgreSQL + Elasticsearch.
- Seed/crawl dữ liệu công ty.
- Seed review mẫu tham khảo từ 1900.com.vn.
- Tự tạo công ty nếu thiếu khi seed review.
- Dedupe review theo `company_id + title + author_name` để tránh insert trùng.
- Seed tài khoản admin mặc định:
  - email: `admin@review.com`
  - password: `admin123`

## 2. Các mode hỗ trợ
- `companies`:
  - Crawl danh sách công ty từ Wikipedia (List of companies of Vietnam).
  - Nếu crawl lỗi/không có data thì fallback sang danh sách seed cục bộ.
  - Update các trường còn thiếu ở công ty đã có (logo/industry/description).

- `reviews-1900`:
  - Seed một batch review mẫu từ dữ liệu tham khảo 1900.com.vn.
  - Nếu công ty chưa có thì tự tạo mới trước khi insert review.

- `all`:
  - Chạy tuần tự `companies` rồi `reviews-1900`.

## 3. Cách chạy
- Chỉ seed công ty:
  - `go run cmd/scraper/main.go --mode=companies`
- Chỉ seed review 1900:
  - `go run cmd/scraper/main.go --mode=reviews-1900`
- Chạy toàn bộ:
  - `go run cmd/scraper/main.go --mode=all`

Mặc định nếu không truyền `--mode` thì chạy `companies`.

## 4. Ghi chú
- Dữ liệu review ở mode `reviews-1900` là dữ liệu seed dev tham khảo công khai.
- Có thể mở rộng thêm mode scrape production đầy đủ trong tương lai (phân trang, rate-limit, audit nguồn, retry).
