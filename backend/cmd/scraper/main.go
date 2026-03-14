package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/review-company/backend/internal/config"
	"github.com/review-company/backend/internal/db"
	"github.com/review-company/backend/internal/handler"
	"github.com/review-company/backend/internal/model"
	"github.com/review-company/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var numericNamePattern = regexp.MustCompile(`^[0-9][0-9,./-]*$`)
var citationPattern = regexp.MustCompile(`\[[0-9]+\]`)

type companyReviewSeed struct {
	Company model.Company
	Review  model.Review
}

type ensureCompanyResult struct {
	Company *model.Company
	Created bool
	Updated bool
}

type crawlOptions struct {
	MaxListPages    int
	MaxCompanyPages int
	MaxCompanies    int
}

func main() {
	mode := flag.String("mode", "companies", "seeding mode: companies|reviews-1900|reviews-1900-full|all")
	maxListPages := flag.Int("max-list-pages", 200, "max listing pages to crawl in reviews-1900-full mode")
	maxCompanyPages := flag.Int("max-company-pages", 10, "max review pages per company in reviews-1900-full mode")
	maxCompanies := flag.Int("max-companies", 0, "max companies to crawl in reviews-1900-full mode, 0 means unlimited")
	flag.Parse()

	cfg := config.LoadConfig()
	db.ConnectPostgres(cfg)
	db.ConnectElasticsearch(cfg)

	repo := repository.NewCompanyRepository(db.DB, db.ES)
	reviewRepo := repository.NewReviewRepository(db.DB)

	seedFallback := []model.Company{
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

	switch *mode {
	case "companies":
		seedCompanies(repo, seedFallback)
	case "reviews-1900":
		seed1900Reviews(repo, reviewRepo)
	case "reviews-1900-full":
		seed1900ReviewsFull(repo, reviewRepo, crawlOptions{
			MaxListPages:    *maxListPages,
			MaxCompanyPages: *maxCompanyPages,
			MaxCompanies:    *maxCompanies,
		})
	case "all":
		seedCompanies(repo, seedFallback)
		seed1900Reviews(repo, reviewRepo)
		seed1900ReviewsFull(repo, reviewRepo, crawlOptions{
			MaxListPages:    *maxListPages,
			MaxCompanyPages: *maxCompanyPages,
			MaxCompanies:    *maxCompanies,
		})
	default:
		log.Fatalf("invalid mode %q (use companies|reviews-1900|reviews-1900-full|all)", *mode)
	}

	// Seed Admin User
	log.Println("Seeding Admin User...")
	transportHash := handler.HashPasswordForTransport("admin123")
	hash, _ := bcrypt.GenerateFromPassword([]byte(transportHash), 10)
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

func seedCompanies(repo repository.CompanyRepository, fallback []model.Company) {
	companiesFromWeb, err := crawlCompaniesFromWikipedia()
	if err != nil {
		log.Printf("Failed to crawl from internet, fallback to local seed: %v", err)
	}

	if len(companiesFromWeb) == 0 {
		companiesFromWeb = fallback
		log.Printf("No crawled companies found, using fallback list (%d)", len(companiesFromWeb))
	} else {
		log.Printf("Crawled %d companies from internet source", len(companiesFromWeb))
	}

	existingMap := loadExistingCompanyMap()
	created := 0
	skipped := 0
	updated := 0

	log.Println("Starting to import companies...")

	for _, c := range companiesFromWeb {
		nameKey := normalize(c.Name)
		websiteKey := normalize(c.Website)

		if nameKey == "" {
			skipped++
			continue
		}

		existing := findExistingCompany(existingMap, nameKey, websiteKey)
		if existing != nil {
			changed := false
			if existing.LogoURL == "" && c.LogoURL != "" {
				existing.LogoURL = c.LogoURL
				changed = true
			}
			if existing.Industry == "" && c.Industry != "" {
				existing.Industry = c.Industry
				changed = true
			}
			if existing.Description == "" && c.Description != "" {
				existing.Description = c.Description
				changed = true
			}

			if changed {
				if err := repo.Update(existing); err != nil {
					log.Printf("Failed to update existing company %s: %v", existing.Name, err)
				} else {
					updated++
				}
			}

			skipped++
			continue
		}

		err := repo.Create(&c)
		if err != nil {
			log.Printf("Failed to import company %s: %v", c.Name, err)
		} else {
			created++
			registerCompany(existingMap, &c)
			log.Printf("Imported company: %s", c.Name)
		}
	}

	log.Printf("Import summary: created=%d updated=%d skipped=%d", created, updated, skipped)
}

func seed1900Reviews(repo repository.CompanyRepository, reviewRepo repository.ReviewRepository) {
	seeds := []companyReviewSeed{
		{
			Company: model.Company{
				Name:        "Công Ty Thương Nghiệp Cà Mau",
				Website:     "https://1900.com.vn/danh-gia-dn/cong-ty-co-phan-thuong-nghiep-ca-mau-1726",
				Industry:    "Bán hàng/ Kinh doanh",
				Size:        "500 - 1.000 nhân viên",
				Description: "Dữ liệu tham khảo từ 1900.com.vn",
				LogoURL:     "https://1900.com.vn/storage/uploads/companies/logo/28/unnamed-1-1693299054.png",
			},
			Review: model.Review{
				AuthorName: "Kế toán nội bộ",
				Rating:     3,
				Title:      "Công ty tốt đồng nghiệp tốt",
				Content:    "công ty tốt đồng nghiệp tốt",
				Pros:       "Môi trường và đồng nghiệp khá ổn",
				Cons:       "Nhược điểm: công ty, đồng nghiệp tốt",
				Status:     model.StatusApproved,
			},
		},
		{
			Company: model.Company{
				Name:        "BOSCH",
				Website:     "https://1900.com.vn/danh-gia-dn/cong-ty-tnhh-bosch-global-software-technologies-793",
				Industry:    "CNTT - Phần mềm",
				Size:        "500 - 1.000 nhân viên",
				Description: "Dữ liệu tham khảo từ 1900.com.vn",
				LogoURL:     "https://1900.com.vn/storage/uploads/companies/logo/58/1656790100428-1694772299.jfif",
			},
			Review: model.Review{
				AuthorName: "Java Developer",
				Rating:     4,
				Title:      "Cơ hội phát triển, cách quản lý, hệ thống, tổ chức và con người",
				Content:    "Phúc lợi tốt, công việc bình thường, ko có gì đặc biệt",
				Pros:       "Phúc lợi tốt",
				Cons:       "Không có cơ hội phát triển, công ty quá hierarchy và khó làm việc chung",
				Status:     model.StatusApproved,
			},
		},
		{
			Company: model.Company{
				Name:        "FPT Software",
				Website:     "https://1900.com.vn/danh-gia-dn/cong-ty-tnhh-phan-mem-fpt-11",
				Industry:    "CNTT - Phần mềm",
				Size:        "Trên 10.000 nhân viên",
				Description: "Dữ liệu tham khảo từ 1900.com.vn",
				LogoURL:     "https://1900.com.vn/storage/uploads/companies/logo/1/fpt-software-logo-1691137261.png",
			},
			Review: model.Review{
				AuthorName: "Tester",
				Rating:     4,
				Title:      "Review cty FPT software",
				Content:    "Môi trường làm việc khá chuyên nghiệp, nhiều bộ phận, nhiều vị trí",
				Pros:       "Môi trường chuyên nghiệp, nhiều vị trí",
				Cons:       "OT nhiều lương tăng chậm chỉ phù hợp nhảy vào",
				Status:     model.StatusApproved,
			},
		},
		{
			Company: model.Company{
				Name:        "VMO",
				Website:     "https://1900.com.vn/danh-gia-dn/cong-ty-co-phan-cong-nghe-vmo-holdings-1535",
				Industry:    "CNTT - Phần mềm",
				Size:        "500 - 1.000 nhân viên",
				Description: "Dữ liệu tham khảo từ 1900.com.vn",
				LogoURL:     "https://1900.com.vn/storage/uploads/companies/logo/67/1-1693280696.png",
			},
			Review: model.Review{
				AuthorName: "Web Developer",
				Rating:     4,
				Title:      "Văn phòng công ty tại IDMC",
				Content:    "Xịn thoáng, thấy có nhân viên lau từng cái lá cây",
				Pros:       "Văn phòng sạch, thoáng",
				Cons:       "Mùi người, chờ thang máy lâu",
				Status:     model.StatusApproved,
			},
		},
		{
			Company: model.Company{
				Name:        "Giáo dục Jaxtina",
				Website:     "https://1900.com.vn/danh-gia-dn/cong-ty-co-phan-giao-duc-jaxtina-3084",
				Industry:    "Trung tâm Ngoại ngữ",
				Size:        "200 - 500 nhân viên",
				Description: "Dữ liệu tham khảo từ 1900.com.vn",
				LogoURL:     "https://1900.com.vn/storage/uploads/companies/logo/56/1574093899998-1574048585652-1566314865500-jaxtina-1694576324.png",
			},
			Review: model.Review{
				AuthorName: "Nhân viên tư vấn tuyển sinh",
				Rating:     1,
				Title:      "Công việc khá nhiều nhưng mức lương chưa tương xứng",
				Content:    "Công việc khá nhiều nhưng mức lương chưa tương xứng. Chương trình đào tạo và học tập chưa được xây dựng bài bản.",
				Pros:       "Có cơ hội tiếp xúc nhiều học viên",
				Cons:       "Khối lượng việc nhiều, đào tạo chưa bài bản",
				Status:     model.StatusApproved,
			},
		},
	}

	existing := loadExistingCompanyMap()
	createdCompanies := 0
	updatedCompanies := 0
	createdReviews := 0
	skippedReviews := 0

	for _, seed := range seeds {
		result := ensureCompany(repo, existing, &seed.Company)
		if result.Company == nil {
			continue
		}
		if result.Created {
			createdCompanies++
		}
		if result.Updated {
			updatedCompanies++
		}

		review := seed.Review
		review.CompanyID = result.Company.ID

		var dupCount int64
		db.DB.Model(&model.Review{}).
			Where("company_id = ? AND title = ? AND author_name = ?", result.Company.ID, review.Title, review.AuthorName).
			Count(&dupCount)
		if dupCount > 0 {
			skippedReviews++
			continue
		}

		if err := reviewRepo.Create(&review); err != nil {
			log.Printf("Failed to import review for %s: %v", result.Company.Name, err)
			continue
		}
		createdReviews++
	}

	log.Printf("1900 seed summary: created_companies=%d updated_companies=%d created_reviews=%d skipped_reviews=%d", createdCompanies, updatedCompanies, createdReviews, skippedReviews)
}

func seed1900ReviewsFull(repo repository.CompanyRepository, reviewRepo repository.ReviewRepository, opts crawlOptions) {
	if opts.MaxListPages <= 0 {
		opts.MaxListPages = 1
	}
	if opts.MaxCompanyPages <= 0 {
		opts.MaxCompanyPages = 1
	}

	client := &http.Client{Timeout: 25 * time.Second}
	existing := loadExistingCompanyMap()
	visitedCompanies := make(map[string]bool)

	createdCompanies := 0
	updatedCompanies := 0
	createdReviews := 0
	skippedReviews := 0

	for page := 1; page <= opts.MaxListPages; page++ {
		listURL := fmt.Sprintf("https://1900.com.vn/review-cong-ty?page=%d", page)
		doc, err := fetchDoc(client, listURL)
		if err != nil {
			log.Printf("Failed to fetch 1900 list page %d: %v", page, err)
			continue
		}

		links := make([]string, 0, 24)
		doc.Find(".company-list #recent-update .company-item > a.company-info[href*='/danh-gia-dn/']").Each(func(_ int, a *goquery.Selection) {
			href := strings.TrimSpace(a.AttrOr("href", ""))
			if href == "" {
				return
			}
			full := absolute1900URL(href)
			if full == "" {
				return
			}
			if !visitedCompanies[full] {
				visitedCompanies[full] = true
				links = append(links, full)
			}
		})

		if len(links) == 0 {
			continue
		}

		for _, companyURL := range links {
			if opts.MaxCompanies > 0 && len(visitedCompanies) > opts.MaxCompanies {
				break
			}

			companyDoc, err := fetchDoc(client, companyURL)
			if err != nil {
				log.Printf("Failed to fetch company page %s: %v", companyURL, err)
				continue
			}

			company := parse1900Company(companyDoc)
			if company.Name == "" {
				continue
			}
			if company.Website == "" {
				company.Website = companyURL
			}

			ensured := ensureCompany(repo, existing, &company)
			if ensured.Company == nil {
				continue
			}
			if ensured.Created {
				createdCompanies++
			}
			if ensured.Updated {
				updatedCompanies++
			}

			for detailPage := 1; detailPage <= opts.MaxCompanyPages; detailPage++ {
				pagedURL := companyURL
				if detailPage > 1 {
					pagedURL = appendPageParam(companyURL, detailPage)
				}

				reviewDoc, err := fetchDoc(client, pagedURL)
				if err != nil {
					log.Printf("Failed to fetch review page %s: %v", pagedURL, err)
					break
				}

				reviews := parse1900ReviewCards(reviewDoc)
				if len(reviews) == 0 {
					break
				}

				for _, rv := range reviews {
					if rv.Title == "" || rv.Rating < 1 || rv.Rating > 5 {
						skippedReviews++
						continue
					}

					var dupCount int64
					db.DB.Model(&model.Review{}).
						Where("company_id = ? AND title = ? AND author_name = ?", ensured.Company.ID, rv.Title, rv.AuthorName).
						Count(&dupCount)
					if dupCount > 0 {
						skippedReviews++
						continue
					}

					rv.CompanyID = ensured.Company.ID
					rv.Status = model.StatusApproved
					if err := reviewRepo.Create(&rv); err != nil {
						log.Printf("Failed to import review for %s: %v", ensured.Company.Name, err)
						continue
					}
					createdReviews++
				}

				hasNext := reviewDoc.Find(".pagination-area a.page-link[rel='next']").Length() > 0
				if !hasNext {
					break
				}
			}
		}

		if opts.MaxCompanies > 0 && len(visitedCompanies) > opts.MaxCompanies {
			break
		}
	}

	log.Printf("1900 full seed summary: companies_visited=%d created_companies=%d updated_companies=%d created_reviews=%d skipped_reviews=%d", len(visitedCompanies), createdCompanies, updatedCompanies, createdReviews, skippedReviews)
}

func fetchDoc(client *http.Client, rawURL string) (*goquery.Document, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ReviewBot/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return goquery.NewDocumentFromReader(strings.NewReader(string(body)))
}

func parse1900Company(doc *goquery.Document) model.Company {
	name := sanitizeForDB(doc.Find("h1.company-detail-name > span").First().Text(), 255)
	logo := sanitizeForDB(doc.Find(".company-logo img").First().AttrOr("src", ""), 255)
	website := sanitizeForDB(doc.Find(".company-subdetail-info.website a.company-subdetail-info-text").First().AttrOr("href", ""), 255)
	industry := sanitizeForDB(doc.Find(".company-info-tags .chips").First().Text(), 255)

	size := ""
	doc.Find(".company-subdetail .company-subdetail-info").Each(func(_ int, s *goquery.Selection) {
		if size != "" {
			return
		}
		if s.Find("i.fa-users").Length() > 0 {
			size = strings.TrimSpace(s.Find(".company-subdetail-info-text").Text())
		}
	})

	descriptionParts := make([]string, 0, 2)
	highlights := strings.TrimSpace(doc.Find("#ReviewHighlights .top-summary-section .top-summary-section__title").First().Text())
	if highlights != "" {
		descriptionParts = append(descriptionParts, "Nguon: 1900.com.vn")
	}

	return model.Company{
		Name:        name,
		LogoURL:     logo,
		Website:     website,
		Industry:    industry,
		Size:        size,
		Description: strings.Join(descriptionParts, " | "),
	}
}

func parse1900ReviewCards(doc *goquery.Document) []model.Review {
	reviews := make([]model.Review, 0, 24)

	doc.Find(".ReviewsList > .ReviewItem").Each(func(_ int, card *goquery.Selection) {
		ratingText := strings.TrimSpace(card.Find(".ratingNumber").First().Text())
		rating, err := strconv.ParseFloat(ratingText, 64)
		if err != nil {
			return
		}
		ratingInt := int(rating)
		if ratingInt < 1 {
			ratingInt = 1
		}
		if ratingInt > 5 {
			ratingInt = 5
		}

		title := strings.TrimSpace(card.Find("h2.ReviewTitle").First().Text())
		authorRole := strings.TrimSpace(card.Find(".mb-1.font-14.ReviewCandidateSubtext a").First().Text())
		if authorRole == "" {
			authorRole = "Ẩn danh"
		}

		pros := normalizeBlockText(card.Find("strong.greenColor").First().Next().Find(".expandable-content").First().Text())
		cons := normalizeBlockText(card.Find("strong.text-danger").First().Next().Find(".expandable-content").First().Text())

		content := pros
		if content == "" {
			content = cons
		}
		if content == "" {
			content = strings.TrimSpace(title)
		}

		reviews = append(reviews, model.Review{
			AuthorName: sanitizeForDB(authorRole, 100),
			Rating:     ratingInt,
			Title:      sanitizeForDB(title, 255),
			Content:    content,
			Pros:       pros,
			Cons:       cons,
		})
	})

	return reviews
}

func normalizeBlockText(s string) string {
	text := strings.TrimSpace(strings.ReplaceAll(s, "\r", ""))
	text = strings.Join(strings.Fields(text), " ")
	return text
}

func appendPageParam(rawURL string, page int) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	q.Set("page", strconv.Itoa(page))
	u.RawQuery = q.Encode()
	return u.String()
}

func absolute1900URL(href string) string {
	href = strings.TrimSpace(href)
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	if strings.Contains(href, "http://") || strings.Contains(href, "https://") {
		return ""
	}
	if strings.HasPrefix(href, "/") {
		return "https://1900.com.vn" + href
	}
	return "https://1900.com.vn/" + href
}

func ensureCompany(repo repository.CompanyRepository, existing map[string]*model.Company, input *model.Company) ensureCompanyResult {
	input.Name = sanitizeForDB(input.Name, 255)
	input.LogoURL = sanitizeForDB(input.LogoURL, 255)
	input.Website = sanitizeForDB(input.Website, 255)
	input.Industry = sanitizeForDB(input.Industry, 255)
	input.Size = sanitizeForDB(input.Size, 50)
	input.Description = strings.TrimSpace(input.Description)

	nameKey := normalize(input.Name)
	websiteKey := normalize(input.Website)

	found := findExistingCompany(existing, nameKey, websiteKey)
	if found == nil {
		newCompany := *input
		if newCompany.LogoURL == "" {
			newCompany.LogoURL = defaultLogoURL(newCompany.Name)
		}
		if err := repo.Create(&newCompany); err != nil {
			log.Printf("Failed to create company %s: %v", input.Name, err)
			return ensureCompanyResult{}
		}
		registerCompany(existing, &newCompany)
		return ensureCompanyResult{Company: &newCompany, Created: true}
	}

	updated := false
	if shouldEnrichCompany(found, input) {
		updated = true
		if found.LogoURL == "" && input.LogoURL != "" {
			found.LogoURL = input.LogoURL
		}
		if found.Website == "" && input.Website != "" {
			found.Website = input.Website
		}
		if found.Industry == "" && input.Industry != "" {
			found.Industry = input.Industry
		}
		if found.Size == "" && input.Size != "" {
			found.Size = input.Size
		}
		if found.Description == "" && input.Description != "" {
			found.Description = input.Description
		}

		if err := repo.Update(found); err != nil {
			log.Printf("Failed to enrich company %s: %v", found.Name, err)
		}
		registerCompany(existing, found)
	}

	return ensureCompanyResult{Company: found, Updated: updated}
}

func shouldEnrichCompany(existing *model.Company, incoming *model.Company) bool {
	if existing.LogoURL == "" && incoming.LogoURL != "" {
		return true
	}
	if existing.Website == "" && incoming.Website != "" {
		return true
	}
	if existing.Industry == "" && incoming.Industry != "" {
		return true
	}
	if existing.Size == "" && incoming.Size != "" {
		return true
	}
	if existing.Description == "" && incoming.Description != "" {
		return true
	}
	return false
}

func sanitizeForDB(value string, max int) string {
	v := strings.TrimSpace(strings.Join(strings.Fields(value), " "))
	if max > 0 {
		r := []rune(v)
		if len(r) > max {
			v = string(r[:max])
		}
	}
	return v
}

func crawlCompaniesFromWikipedia() ([]model.Company, error) {
	collector := colly.NewCollector(
		colly.AllowedDomains("en.wikipedia.org", "www.en.wikipedia.org"),
	)

	companies := make([]model.Company, 0, 100)
	seen := make(map[string]bool)

	collector.OnHTML("table.wikitable", func(e *colly.HTMLElement) {
		if !isNotableFirmsTable(e) {
			return
		}

		e.ForEach("tbody tr", func(_ int, row *colly.HTMLElement) {
			cells := row.DOM.Find("td")
			if cells.Length() == 0 {
				return
			}

			name := getCellText(cells, 0)
			if isInvalidCompanyName(name) {
				return
			}

			industry := getCellText(cells, 1)
			sector := getCellText(cells, 2)
			headquarters := getCellText(cells, 3)
			notes := getCellText(cells, 5)

			key := normalize(name)
			if seen[key] {
				return
			}
			seen[key] = true

			descriptionParts := make([]string, 0, 4)
			if industry != "" {
				descriptionParts = append(descriptionParts, fmt.Sprintf("Industry: %s", industry))
			}
			if sector != "" {
				descriptionParts = append(descriptionParts, fmt.Sprintf("Sector: %s", sector))
			}
			if headquarters != "" {
				descriptionParts = append(descriptionParts, fmt.Sprintf("HQ: %s", headquarters))
			}
			if notes != "" {
				descriptionParts = append(descriptionParts, notes)
			}

			companies = append(companies, model.Company{
				Name:        name,
				LogoURL:     defaultLogoURL(name),
				Industry:    industry,
				Description: strings.Join(descriptionParts, " | "),
			})
		})
	})

	collector.OnError(func(_ *colly.Response, err error) {
		log.Printf("Crawler error: %v", err)
	})

	err := collector.Visit("https://en.wikipedia.org/wiki/List_of_companies_of_Vietnam")
	if err != nil {
		return nil, err
	}

	return companies, nil
}

func loadExistingCompanyMap() map[string]*model.Company {
	rows := make([]model.Company, 0)
	db.DB.Select("id", "name", "website", "logo_url", "industry", "description").Find(&rows)

	m := make(map[string]*model.Company, len(rows)*2)
	for _, c := range rows {
		copyCompany := c
		registerCompany(m, &copyCompany)
	}
	return m
}

func findExistingCompany(existing map[string]*model.Company, nameKey, websiteKey string) *model.Company {
	if nameKey != "" {
		if c, ok := existing[nameKey]; ok {
			return c
		}
	}
	if websiteKey != "" {
		if c, ok := existing[websiteKey]; ok {
			return c
		}
	}
	return nil
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func getCellText(cells *goquery.Selection, index int) string {
	text := strings.TrimSpace(cells.Eq(index).Text())
	text = strings.TrimSuffix(text, "[edit]")
	text = citationPattern.ReplaceAllString(text, "")
	return strings.TrimSpace(text)
}

func isInvalidCompanyName(name string) bool {
	if name == "" {
		return true
	}
	if numericNamePattern.MatchString(name) {
		return true
	}
	return false
}

func isNotableFirmsTable(e *colly.HTMLElement) bool {
	headers := make([]string, 0, 8)
	e.ForEach("tr th", func(_ int, h *colly.HTMLElement) {
		headers = append(headers, normalize(h.Text))
	})

	hasName := false
	hasIndustry := false
	for _, h := range headers {
		if h == "name" {
			hasName = true
		}
		if strings.Contains(h, "industry") {
			hasIndustry = true
		}
	}

	return hasName && hasIndustry
}

func registerCompany(existing map[string]*model.Company, company *model.Company) {
	nameKey := normalize(company.Name)
	websiteKey := normalize(company.Website)
	if nameKey != "" {
		existing[nameKey] = company
	}
	if websiteKey != "" {
		existing[websiteKey] = company
	}
}

func defaultLogoURL(name string) string {
	encodedName := url.QueryEscape(strings.TrimSpace(name))
	if encodedName == "" {
		encodedName = "Company"
	}
	return fmt.Sprintf("https://ui-avatars.com/api/?name=%s&background=0F766E&color=FFFFFF&bold=true", encodedName)
}
