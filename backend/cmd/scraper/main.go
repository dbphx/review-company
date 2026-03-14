package main

import (
	"log"

	"github.com/review/backend/internal/config"
	"github.com/review/backend/internal/db"
	"github.com/review/backend/internal/model"
	"github.com/review/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.LoadConfig()
	db.ConnectPostgres(cfg)
	db.ConnectElasticsearch(cfg)

	repo := repository.NewCompanyRepository(db.DB, db.ES)

	// Since live scraping real sites like TopCV/ITviec requires bypassing Cloudflare/Captcha,
	// we will seed some dummy data here to ensure the MVP is functional out of the box.

	dummyCompanies := []model.Company{
		{
			Name:        "FPT Software",
			LogoURL:     "https://cdn.haitrieu.com/wp-content/uploads/2022/01/Logo-FPT-Software-FPTS-1.png",
			Website:     "https://fptsoftware.com",
			Industry:    "IT / Phần mềm",
			Size:        "10000+",
			Description: "Công ty phần mềm lớn nhất Việt Nam, chuyên outsource.",
		},
		{
			Name:        "VNG Corporation",
			LogoURL:     "https://upload.wikimedia.org/wikipedia/commons/thumb/c/cd/VNG_logo.svg/2560px-VNG_logo.svg.png",
			Website:     "https://vng.com.vn",
			Industry:    "Game / Tech",
			Size:        "3000+",
			Description: "Kỳ lân công nghệ của Việt Nam, nổi tiếng với Zalo, Zing.",
		},
		{
			Name:        "Shopee VietNam",
			LogoURL:     "https://upload.wikimedia.org/wikipedia/commons/thumb/f/fe/Shopee.svg/1200px-Shopee.svg.png",
			Website:     "https://shopee.vn",
			Industry:    "Thương mại điện tử",
			Size:        "5000+",
			Description: "Nền tảng thương mại điện tử hàng đầu Đông Nam Á.",
		},
		{
			Name:        "Tiki",
			LogoURL:     "https://salt.tikicdn.com/ts/upload/ae/f5/15/2228f38cf84d1b8451bb49e2c4537081.png",
			Website:     "https://tiki.vn",
			Industry:    "Thương mại điện tử",
			Size:        "1000+",
			Description: "Website mua sắm trực tuyến uy tín tại Việt Nam.",
		},
		{
			Name:        "Momo",
			LogoURL:     "https://upload.wikimedia.org/wikipedia/vi/f/fe/MoMo_Logo.png",
			Website:     "https://momo.vn",
			Industry:    "Fintech",
			Size:        "2000+",
			Description: "Ví điện tử số 1 Việt Nam.",
		},
	}

	log.Println("Starting to seed dummy companies...")

	for _, c := range dummyCompanies {
		err := repo.Create(&c)
		if err != nil {
			log.Printf("Failed to seed company %s: %v", c.Name, err)
		} else {
			log.Printf("Successfully seeded company: %s", c.Name)
		}
	}

	// Seed Admin User
	log.Println("Seeding Admin User...")
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
	adminUser := model.AdminUser{
		Email:    "admin@review.com",
		Password: string(hash),
		Name:     "Super Admin",
		Role:     "ADMIN",
	}
	if err := db.DB.Where("email = ?", "admin@review.com").FirstOrCreate(&adminUser).Error; err != nil {
		log.Printf("Failed to seed admin user: %v", err)
	}

	log.Println("Seeding complete.")
}
