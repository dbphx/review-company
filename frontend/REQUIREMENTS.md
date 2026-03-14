# Frontend Requirements - ReactJS

Phần này mô tả chi tiết yêu cầu kỹ thuật và tính năng của phần giao diện (Frontend) cho Nền Tảng Review Công Ty.

*(Lưu ý: Trang Admin đã được tách ra một project độc lập ở thư mục `/admin-frontend`)*

## 1. Công Nghệ Sử Dụng (Tech Stack)
- **Framework Core:** ReactJS 18+.
- **Ngôn ngữ:** TypeScript.
- **Build Tool:** Vite.
- **Styling:** Tailwind CSS.
- **Routing:** React Router DOM.
- **Form:** react-hook-form.
- **Icons:** lucide-react.
- **Auth user:** Google OAuth (`@react-oauth/google`) cho profile portal.

## 2. Các Tính Năng Cho Người Dùng (Public Web)
Hệ thống cho phép người dùng ẩn danh hoàn toàn, không cần đăng ký.
### 2.1. Trang Chủ (Home)
- Hiển thị Search Bar autocomplete theo kết quả từ Elasticsearch.
- Có trạng thái `Đang tìm kiếm...`, empty-state `Không tìm thấy công ty phù hợp.`.
- Không crash route khi search lỗi hoặc không có dữ liệu.
- Hiển thị `Công ty nổi bật` và `Review mới nhất`.

### 2.2. Trang Chi Tiết Công Ty
- Hiển thị thông tin công ty (Tên, Logo, Ngành nghề, Quy mô, Website).
- Hiển thị Điểm trung bình (Average Rating) tổng thể và theo số sao.
- Danh sách review và comment count theo review.

### 2.3. Đánh Giá Công Ty (Review)
- Cho phép chọn số sao (1-5).
- Tùy chọn nhập "Tên hiển thị" (nếu để trống, mặc định là "Ẩn danh").
- Nhập "Tiêu đề" review.
- Các trường nhập đánh giá chi tiết:
  - Ý kiến tổng quan (Môi trường, Quản lý, v.v...).
  - Ưu điểm (Pros).
  - Nhược điểm (Cons).
  - Tùy chọn chia sẻ **Mức lương** (Gross).
  - Tùy chọn chia sẻ **Kinh nghiệm phỏng vấn**.
- Nút Gửi Review.

### 2.4. Tính Năng Bình Luận (Comment)
- Người dùng bình luận/reply trực tiếp theo thread.
- Dữ liệu comment/reply lưu bền vững và không mất khi refresh.
- Nút `Bình luận (x)` luôn hiển thị số lượng ngay cả khi collapse.
- Khi mở thread, mặc định hiển thị 3 bình luận đầu; có `Xem thêm bình luận`.

## 3. Quản Lý Đăng Nhập & Hồ Sơ Người Dùng (Profile)
- **Xác Thực (Authentication):** Đăng nhập thông qua Google Native (sử dụng `@react-oauth/google`).
- **Trang Profile:**
  - Hiển thị thông tin người dùng (Avatar, Tên, Email).
  - Hiển thị danh sách các công ty mà người dùng được "Full Access" (do Admin cấp quyền).
- **Lưu ý hiện trạng:** Cơ chế giới hạn xem theo quyền đã có schema + UI profile; enforcement chi tiết ở mọi endpoint chưa bật full cho production.

## 4. Error UX
- Có route-level `errorElement` riêng để hiển thị trang lỗi thân thiện thay vì màn hình lỗi mặc định của React Router.
