# Portal Frontend (User)

React + Vite + TypeScript frontend cho người dùng cuối.

## Chạy local
- `npm install`
- `npm run dev`

Mặc định chạy tại `http://localhost:5173`.

## Màn hình chính
- Home: search autocomplete + công ty nổi bật + review mới nhất.
- Company detail: xem review, viết review, bình luận/reply theo thread.
- Profile: đăng nhập Google (portal user) và thông tin tài khoản.

## UX quan trọng đã triển khai
- Search có loading/empty-state, không đẩy user vào route error khi không có kết quả.
- Route-level error boundary để hiển thị lỗi thân thiện.
- Comment count hiển thị ngay cả khi collapse (`Bình luận (x)`).
- Mở thread mặc định hiển thị 3 bình luận đầu, có nút `Xem thêm bình luận`.

## Cấu hình môi trường
- `VITE_API_URL` (mặc định `http://localhost:3000/api`)
- `VITE_GOOGLE_CLIENT_ID`
