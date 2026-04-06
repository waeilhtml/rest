package database

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "time"

    _ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// ==================== الإعدادات ====================
type Config struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
}

// ==================== هياكل البيانات ====================
type Product struct {
    ID          int     `json:"id"`
    Name        string  `json:"name"`
    Category    string  `json:"category"`
    Price       float64 `json:"price"`
    OldPrice    float64 `json:"oldPrice"`
    Image       string  `json:"image"`
    Description string  `json:"description"`
    Stock       string  `json:"stock"`
    Features    string  `json:"features"`
    Sales       int     `json:"sales"`
    CreatedAt   string  `json:"createdAt"`
}

type Order struct {
    ID       int     `json:"id"`
    FullName string  `json:"fullName"`
    Phone    string  `json:"phone"`
    Email    string  `json:"email"`
    Address  string  `json:"address"`
    Total    float64 `json:"total"`
    Status   string  `json:"status"`
    Date     string  `json:"date"`
    Items    string  `json:"items"`
}

type User struct {
    ID          int    `json:"id"`
    Name        string `json:"name"`
    Phone       string `json:"phone"`
    Password    string `json:"-"`
    Permissions string `json:"permissions"`
    CreatedAt   string `json:"createdAt"`
}

type Coupon struct {
    Code     string  `json:"code"`
    Discount float64 `json:"discount"`
    Type     string  `json:"type"`
    MaxUses  int     `json:"maxUses"`
    Uses     int     `json:"uses"`
    Expires  string  `json:"expires"`
}

type Message struct {
    ID      int    `json:"id"`
    Name    string `json:"name"`
    Email   string `json:"email"`
    Message string `json:"message"`
    Date    string `json:"date"`
    IsRead  int    `json:"isRead"`
}

type Settings struct {
    ID           int    `json:"id"`
    SiteName     string `json:"siteName"`
    SitePhone    string `json:"sitePhone"`
    SiteEmail    string `json:"siteEmail"`
    SiteAddress  string `json:"siteAddress"`
    FacebookUrl  string `json:"facebookUrl"`
    InstagramUrl string `json:"instagramUrl"`
    LogoUrl      string `json:"logoUrl"`
    HeroBg       string `json:"heroBg"`
    HeroTitle    string `json:"heroTitle"`
    HeroDesc     string `json:"heroDesc"`
    ThemeColor   string `json:"themeColor"`
    AboutText    string `json:"aboutText"`
    ContactText  string `json:"contactText"`
    TermsText    string `json:"termsText"`
    WhatsApp     string `json:"whatsapp"`
}

// ==================== تهيئة قاعدة البيانات ====================
func InitDB(cfg Config) error {
    // التحقق من وجود كلمة المرور
    if cfg.Password == "" && os.Getenv("DB_PASSWORD") != "" {
        cfg.Password = os.Getenv("DB_PASSWORD")
    }

    // إنشاء رابط الاتصال
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=30s",
        cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

    var err error
    DB, err = sql.Open("mysql", dsn)
    if err != nil {
        return fmt.Errorf("خطأ في فتح الاتصال: %v", err)
    }

    // إعدادات Connection Pool
    DB.SetMaxOpenConns(100)
    DB.SetMaxIdleConns(25)
    DB.SetConnMaxLifetime(30 * time.Minute)
    DB.SetConnMaxIdleTime(10 * time.Minute)

    // اختبار الاتصال
    if err := DB.Ping(); err != nil {
        return fmt.Errorf("لا يمكن الاتصال بقاعدة البيانات: %v", err)
    }

    log.Println("✅ MariaDB متصل بنجاح")
    log.Printf("📊 Connection Pool: MaxOpen=100, MaxIdle=25")

    // إنشاء الجداول
    if err := createTables(); err != nil {
        return fmt.Errorf("خطأ في إنشاء الجداول: %v", err)
    }

    return nil
}

// ==================== إنشاء الجداول ====================
func createTables() error {
    queries := []string{
        // جدول المنتجات
        `CREATE TABLE IF NOT EXISTS products (
            id INT PRIMARY KEY AUTO_INCREMENT,
            name VARCHAR(255) NOT NULL,
            category VARCHAR(100) NOT NULL,
            price DECIMAL(10,2) NOT NULL,
            old_price DECIMAL(10,2) DEFAULT NULL,
            image TEXT,
            description TEXT,
            stock VARCHAR(50) DEFAULT 'available',
            features TEXT,
            sales INT DEFAULT 0,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            INDEX idx_category (category),
            INDEX idx_price (price),
            INDEX idx_sales (sales)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

        // جدول الطلبات
        `CREATE TABLE IF NOT EXISTS orders (
            id INT PRIMARY KEY AUTO_INCREMENT,
            full_name VARCHAR(255) NOT NULL,
            phone VARCHAR(50) NOT NULL,
            email VARCHAR(255),
            address TEXT,
            total DECIMAL(10,2) NOT NULL,
            status VARCHAR(50) DEFAULT 'processing',
            date DATE NOT NULL,
            items TEXT,
            INDEX idx_phone (phone),
            INDEX idx_status (status),
            INDEX idx_date (date)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

        // جدول المستخدمين (الموظفين)
        `CREATE TABLE IF NOT EXISTS users (
            id INT PRIMARY KEY AUTO_INCREMENT,
            name VARCHAR(255) NOT NULL,
            phone VARCHAR(50) UNIQUE NOT NULL,
            password VARCHAR(255) NOT NULL,
            permissions TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            INDEX idx_phone (phone)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

        // جدول الكوبونات
        `CREATE TABLE IF NOT EXISTS coupons (
            code VARCHAR(50) PRIMARY KEY,
            discount DECIMAL(10,2) NOT NULL,
            type VARCHAR(20) DEFAULT 'percent',
            max_uses INT DEFAULT 1,
            uses INT DEFAULT 0,
            expires DATE,
            INDEX idx_expires (expires)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

        // جدول الرسائل
        `CREATE TABLE IF NOT EXISTS messages (
            id INT PRIMARY KEY AUTO_INCREMENT,
            name VARCHAR(255) NOT NULL,
            email VARCHAR(255) NOT NULL,
            message TEXT NOT NULL,
            date DATE NOT NULL,
            is_read INT DEFAULT 0,
            INDEX idx_date (date),
            INDEX idx_is_read (is_read)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

        // جدول الإعدادات
        `CREATE TABLE IF NOT EXISTS settings (
            id INT PRIMARY KEY DEFAULT 1,
            site_name VARCHAR(255) DEFAULT 'Smart Step Control',
            site_phone VARCHAR(50) DEFAULT '+216 70 000 000',
            site_email VARCHAR(255) DEFAULT 'contact@ssc.tn',
            site_address TEXT,
            facebook_url VARCHAR(500),
            instagram_url VARCHAR(500),
            logo_url TEXT,
            hero_bg TEXT,
            hero_title VARCHAR(255) DEFAULT 'Solutions de Sécurité Professionnelle',
            hero_desc TEXT DEFAULT 'Caméras, alarmes, contrôle d'accès - Protégez ce qui compte pour vous',
            theme_color VARCHAR(50) DEFAULT '#0A2E68',
            about_text TEXT,
            contact_text TEXT,
            terms_text TEXT,
            whatsapp VARCHAR(50) DEFAULT '+21620000000',
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

        // جدول سجل النشاطات
        `CREATE TABLE IF NOT EXISTS audit_logs (
            id INT PRIMARY KEY AUTO_INCREMENT,
            user_id INT,
            action VARCHAR(255) NOT NULL,
            details TEXT,
            ip VARCHAR(45),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            INDEX idx_user_id (user_id),
            INDEX idx_created_at (created_at)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

        // جدول الجلسات
        `CREATE TABLE IF NOT EXISTS sessions (
            id VARCHAR(255) PRIMARY KEY,
            user_id INT NOT NULL,
            expires_at TIMESTAMP NOT NULL,
            INDEX idx_user_id (user_id),
            INDEX idx_expires_at (expires_at)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

        // جدول الفئات
        `CREATE TABLE IF NOT EXISTS categories (
            id INT PRIMARY KEY AUTO_INCREMENT,
            name VARCHAR(100) NOT NULL,
            slug VARCHAR(100) NOT NULL UNIQUE,
            INDEX idx_slug (slug)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
    }

    for _, query := range queries {
        if _, err := DB.Exec(query); err != nil {
            return fmt.Errorf("خطأ في تنفيذ الاستعلام: %v", err)
        }
    }

    // إدخال الفئات الافتراضية إذا لم توجد
    var count int
    DB.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
    if count == 0 {
        categories := []struct{ Name, Slug string }{
            {"Caméras", "cameras"},
            {"Alarmes", "alarms"},
            {"Contrôle Accès", "access"},
            {"Stockage", "storage"},
            {"Accessoires", "accessories"},
        }
        for _, c := range categories {
            DB.Exec("INSERT INTO categories (name, slug) VALUES (?, ?)", c.Name, c.Slug)
        }
        log.Println("✅ تم إدخال الفئات الافتراضية")
    }

    // التحقق من وجود الإعدادات الافتراضية
    DB.QueryRow("SELECT COUNT(*) FROM settings").Scan(&count)
    if count == 0 {
        DB.Exec(`INSERT INTO settings (id, site_name, site_phone, site_email, site_address, theme_color) VALUES 
            (1, 'Smart Step Control', '+216 70 000 000', 'contact@ssc.tn', 'Centre Urbain Nord, Tunis, Tunisie', '#0A2E68')`)
        log.Println("✅ تم إدخال الإعدادات الافتراضية")
    }

    log.Println("✅ جميع الجداول جاهزة")
    return nil
}

// ==================== دوال مساعدة ====================

func Close() {
    if DB != nil {
        DB.Close()
        log.Println("🔒 تم إغلاق اتصال قاعدة البيانات")
    }
}

func Ping() error {
    if DB == nil {
        return fmt.Errorf("قاعدة البيانات غير مهيأة")
    }
    return DB.Ping()
}

func GetStats() (map[string]interface{}, error) {
    stats := make(map[string]interface{})
    var productsCount, ordersCount, usersCount, messagesCount int
    var totalSales float64

    DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&productsCount)
    DB.QueryRow("SELECT COUNT(*) FROM orders").Scan(&ordersCount)
    DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&usersCount)
    DB.QueryRow("SELECT COUNT(*) FROM messages WHERE is_read = 0").Scan(&messagesCount)
    DB.QueryRow("SELECT COALESCE(SUM(total), 0) FROM orders WHERE status != 'cancelled'").Scan(&totalSales)

    stats["products"] = productsCount
    stats["orders"] = ordersCount
    stats["users"] = usersCount
    stats["unread_messages"] = messagesCount
    stats["total_sales"] = totalSales

    return stats, nil
}

func AddAuditLog(userID int, action, details, ip string) error {
    _, err := DB.Exec(`INSERT INTO audit_logs (user_id, action, details, ip) VALUES (?, ?, ?, ?)`,
        userID, action, details, ip)
    return err
}

// ==================== دوال المنتجات ====================

func GetAllProducts() ([]Product, error) {
    rows, err := DB.Query(`SELECT id, name, category, price, COALESCE(old_price, 0), 
        COALESCE(image, ''), COALESCE(description, ''), stock, COALESCE(features, ''), sales 
        FROM products ORDER BY id DESC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var products []Product
    for rows.Next() {
        var p Product
        err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.Price, &p.OldPrice,
            &p.Image, &p.Description, &p.Stock, &p.Features, &p.Sales)
        if err != nil {
            return nil, err
        }
        products = append(products, p)
    }
    return products, nil
}

func GetProductByID(id int) (*Product, error) {
    var p Product
    err := DB.QueryRow(`SELECT id, name, category, price, COALESCE(old_price, 0), 
        COALESCE(image, ''), COALESCE(description, ''), stock, COALESCE(features, ''), sales 
        FROM products WHERE id = ?`, id).Scan(
        &p.ID, &p.Name, &p.Category, &p.Price, &p.OldPrice,
        &p.Image, &p.Description, &p.Stock, &p.Features, &p.Sales)
    if err != nil {
        return nil, err
    }
    return &p, nil
}

func CreateProduct(p *Product) (int64, error) {
    result, err := DB.Exec(`INSERT INTO products (name, category, price, old_price, image, description, stock, features) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
        p.Name, p.Category, p.Price, p.OldPrice, p.Image, p.Description, p.Stock, p.Features)
    if err != nil {
        return 0, err
    }
    return result.LastInsertId()
}

func UpdateProduct(id int, p *Product) error {
    _, err := DB.Exec(`UPDATE products SET name=?, category=?, price=?, old_price=?, 
        image=?, description=?, stock=?, features=? WHERE id=?`,
        p.Name, p.Category, p.Price, p.OldPrice, p.Image, p.Description, p.Stock, p.Features, id)
    return err
}

func DeleteProduct(id int) error {
    _, err := DB.Exec("DELETE FROM products WHERE id = ?", id)
    return err
}

// ==================== دوال الفئات ====================

func GetAllCategories() ([]map[string]interface{}, error) {
    rows, err := DB.Query("SELECT id, name, slug FROM categories ORDER BY id")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var categories []map[string]interface{}
    for rows.Next() {
        var id int
        var name, slug string
        rows.Scan(&id, &name, &slug)
        categories = append(categories, map[string]interface{}{
            "id": id, "name": name, "slug": slug,
        })
    }
    return categories, nil
}

func CreateCategory(name, slug string) (int64, error) {
    result, err := DB.Exec("INSERT INTO categories (name, slug) VALUES (?, ?)", name, slug)
    if err != nil {
        return 0, err
    }
    return result.LastInsertId()
}

func UpdateCategory(id int, name, slug string) error {
    _, err := DB.Exec("UPDATE categories SET name = ?, slug = ? WHERE id = ?", name, slug, id)
    return err
}

func DeleteCategory(id int) error {
    _, err := DB.Exec("DELETE FROM categories WHERE id = ?", id)
    return err
}

// ==================== دوال الطلبات ====================

func GetAllOrders() ([]Order, error) {
    rows, err := DB.Query(`SELECT id, full_name, phone, COALESCE(email, ''), 
        COALESCE(address, ''), total, status, date, COALESCE(items, '[]') 
        FROM orders ORDER BY id DESC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var orders []Order
    for rows.Next() {
        var o Order
        err := rows.Scan(&o.ID, &o.FullName, &o.Phone, &o.Email, &o.Address,
            &o.Total, &o.Status, &o.Date, &o.Items)
        if err != nil {
            return nil, err
        }
        orders = append(orders, o)
    }
    return orders, nil
}

func GetOrderByID(id int) (*Order, error) {
    var o Order
    err := DB.QueryRow(`SELECT id, full_name, phone, COALESCE(email, ''), 
        COALESCE(address, ''), total, status, date, COALESCE(items, '[]') 
        FROM orders WHERE id = ?`, id).Scan(
        &o.ID, &o.FullName, &o.Phone, &o.Email, &o.Address,
        &o.Total, &o.Status, &o.Date, &o.Items)
    if err != nil {
        return nil, err
    }
    return &o, nil
}

func CreateOrder(o *Order) (int64, error) {
    result, err := DB.Exec(`INSERT INTO orders (full_name, phone, email, address, total, status, date, items) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
        o.FullName, o.Phone, o.Email, o.Address, o.Total, o.Status, o.Date, o.Items)
    if err != nil {
        return 0, err
    }
    return result.LastInsertId()
}

func UpdateOrderStatus(id int, status string) error {
    _, err := DB.Exec("UPDATE orders SET status = ? WHERE id = ?", status, id)
    return err
}

// ==================== دوال المستخدمين ====================

func GetUserByPhone(phone string) (*User, error) {
    var u User
    err := DB.QueryRow(`SELECT id, name, phone, password, COALESCE(permissions, '') 
        FROM users WHERE phone = ?`, phone).Scan(
        &u.ID, &u.Name, &u.Phone, &u.Password, &u.Permissions)
    if err != nil {
        return nil, err
    }
    return &u, nil
}

func GetUserByID(id int) (*User, error) {
    var u User
    err := DB.QueryRow(`SELECT id, name, phone, COALESCE(permissions, '') 
        FROM users WHERE id = ?`, id).Scan(
        &u.ID, &u.Name, &u.Phone, &u.Permissions)
    if err != nil {
        return nil, err
    }
    return &u, nil
}

func GetAllUsers() ([]User, error) {
    rows, err := DB.Query(`SELECT id, name, phone, COALESCE(permissions, '') FROM users ORDER BY id`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var u User
        err := rows.Scan(&u.ID, &u.Name, &u.Phone, &u.Permissions)
        if err != nil {
            return nil, err
        }
        users = append(users, u)
    }
    return users, nil
}

func CreateUser(name, phone, hashedPassword, permissions string) (int64, error) {
    result, err := DB.Exec(`INSERT INTO users (name, phone, password, permissions) VALUES (?, ?, ?, ?)`,
        name, phone, hashedPassword, permissions)
    if err != nil {
        return 0, err
    }
    return result.LastInsertId()
}

func UpdateUser(id int, name, phone, permissions string) error {
    _, err := DB.Exec("UPDATE users SET name = ?, phone = ?, permissions = ? WHERE id = ?",
        name, phone, permissions, id)
    return err
}

func DeleteUser(id int) error {
    _, err := DB.Exec("DELETE FROM users WHERE id = ?", id)
    return err
}

// ==================== دوال الكوبونات ====================

func GetAllCoupons() ([]Coupon, error) {
    rows, err := DB.Query(`SELECT code, discount, type, max_uses, uses, COALESCE(expires, '') 
        FROM coupons ORDER BY expires DESC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var coupons []Coupon
    for rows.Next() {
        var c Coupon
        err := rows.Scan(&c.Code, &c.Discount, &c.Type, &c.MaxUses, &c.Uses, &c.Expires)
        if err != nil {
            return nil, err
        }
        coupons = append(coupons, c)
    }
    return coupons, nil
}

func GetCouponByCode(code string) (*Coupon, error) {
    var c Coupon
    err := DB.QueryRow(`SELECT code, discount, type, max_uses, uses, COALESCE(expires, '') 
        FROM coupons WHERE code = ?`, code).Scan(
        &c.Code, &c.Discount, &c.Type, &c.MaxUses, &c.Uses, &c.Expires)
    if err != nil {
        return nil, err
    }
    return &c, nil
}

func CreateCoupon(c *Coupon) error {
    _, err := DB.Exec(`INSERT INTO coupons (code, discount, type, max_uses, uses, expires) VALUES (?, ?, ?, ?, ?, ?)`,
        c.Code, c.Discount, c.Type, c.MaxUses, c.Uses, c.Expires)
    return err
}

func UpdateCouponUses(code string, uses int) error {
    _, err := DB.Exec("UPDATE coupons SET uses = ? WHERE code = ?", uses, code)
    return err
}

func DeleteCoupon(code string) error {
    _, err := DB.Exec("DELETE FROM coupons WHERE code = ?", code)
    return err
}

// ==================== دوال الرسائل ====================

func GetAllMessages() ([]Message, error) {
    rows, err := DB.Query(`SELECT id, name, email, message, date, is_read FROM messages ORDER BY id DESC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var messages []Message
    for rows.Next() {
        var m Message
        err := rows.Scan(&m.ID, &m.Name, &m.Email, &m.Message, &m.Date, &m.IsRead)
        if err != nil {
            return nil, err
        }
        messages = append(messages, m)
    }
    return messages, nil
}

func CreateMessage(name, email, message string) (int64, error) {
    result, err := DB.Exec(`INSERT INTO messages (name, email, message, date, is_read) VALUES (?, ?, ?, CURDATE(), 0)`,
        name, email, message)
    if err != nil {
        return 0, err
    }
    return result.LastInsertId()
}

func MarkMessageRead(id int) error {
    _, err := DB.Exec("UPDATE messages SET is_read = 1 WHERE id = ?", id)
    return err
}

func DeleteMessage(id int) error {
    _, err := DB.Exec("DELETE FROM messages WHERE id = ?", id)
    return err
}

// ==================== دوال الإعدادات ====================

func GetSettings() (*Settings, error) {
    var s Settings
    err := DB.QueryRow(`SELECT id, site_name, site_phone, site_email, site_address, 
        COALESCE(facebook_url, ''), COALESCE(instagram_url, ''), 
        COALESCE(logo_url, ''), COALESCE(hero_bg, ''), 
        COALESCE(hero_title, ''), COALESCE(hero_desc, ''), theme_color,
        COALESCE(about_text, ''), COALESCE(contact_text, ''), COALESCE(terms_text, ''),
        COALESCE(whatsapp, '')
        FROM settings WHERE id = 1`).Scan(
        &s.ID, &s.SiteName, &s.SitePhone, &s.SiteEmail, &s.SiteAddress,
        &s.FacebookUrl, &s.InstagramUrl, &s.LogoUrl, &s.HeroBg,
        &s.HeroTitle, &s.HeroDesc, &s.ThemeColor,
        &s.AboutText, &s.ContactText, &s.TermsText, &s.WhatsApp)
    if err != nil {
        return nil, err
    }
    return &s, nil
}

func UpdateSettings(s *Settings) error {
    _, err := DB.Exec(`UPDATE settings SET 
        site_name = ?, site_phone = ?, site_email = ?, site_address = ?,
        facebook_url = ?, instagram_url = ?, logo_url = ?, hero_bg = ?,
        hero_title = ?, hero_desc = ?, theme_color = ?,
        about_text = ?, contact_text = ?, terms_text = ?, whatsapp = ?
        WHERE id = 1`,
        s.SiteName, s.SitePhone, s.SiteEmail, s.SiteAddress,
        s.FacebookUrl, s.InstagramUrl, s.LogoUrl, s.HeroBg,
        s.HeroTitle, s.HeroDesc, s.ThemeColor,
        s.AboutText, s.ContactText, s.TermsText, s.WhatsApp)
    return err
}

func UpdateAppearance(logoUrl, heroBg, themeColor string) error {
    _, err := DB.Exec(`UPDATE settings SET logo_url = ?, hero_bg = ?, theme_color = ? WHERE id = 1`,
        logoUrl, heroBg, themeColor)
    return err
}

func UpdateContent(aboutText, contactText, termsText string) error {
    _, err := DB.Exec(`UPDATE settings SET about_text = ?, contact_text = ?, terms_text = ? WHERE id = 1`,
        aboutText, contactText, termsText)
    return err
}

// ==================== دوال الجلسات ====================

func CreateSession(id, userID string, expiresAt time.Time) error {
    _, err := DB.Exec(`INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`,
        id, userID, expiresAt)
    return err
}

func GetSession(id string) (string, error) {
    var userID string
    err := DB.QueryRow(`SELECT user_id FROM sessions WHERE id = ? AND expires_at > NOW()`, id).Scan(&userID)
    if err != nil {
        return "", err
    }
    return userID, nil
}

func DeleteSession(id string) error {
    _, err := DB.Exec("DELETE FROM sessions WHERE id = ?", id)
    return err
}

func CleanExpiredSessions() {
    DB.Exec("DELETE FROM sessions WHERE expires_at < NOW()")
}
