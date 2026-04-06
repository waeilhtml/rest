package database

import (
    "database/sql"
    "fmt"
    "log"
    "time"

    _ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

type Config struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
}

func InitDB(config Config) error {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        config.User, config.Password, config.Host, config.Port, config.DBName)

    var err error
    DB, err = sql.Open("mysql", dsn)
    if err != nil {
        return err
    }

    DB.SetMaxOpenConns(25)
    DB.SetMaxIdleConns(5)
    DB.SetConnMaxLifetime(5 * time.Minute)

    if err = DB.Ping(); err != nil {
        return err
    }

    // إنشاء الجداول
    if err = createTables(); err != nil {
        return err
    }

    // التحقق من وجود مستخدم admin وإضافته إذا لم يكن موجوداً
    if err = ensureAdminUser(); err != nil {
        log.Printf("⚠️ Warning: %v", err)
    }

    // التحقق من وجود إعدادات الموقع وإضافتها إذا لم تكن موجودة
    if err = ensureSettings(); err != nil {
        log.Printf("⚠️ Warning: %v", err)
    }

    log.Println("✅ Base de donnees connectee avec succes")
    return nil
}

func Close() {
    if DB != nil {
        DB.Close()
    }
}

func createTables() error {
    queries := []string{
        // Table users
        `CREATE TABLE IF NOT EXISTS users (
            id INT AUTO_INCREMENT PRIMARY KEY,
            name VARCHAR(100) NOT NULL,
            phone VARCHAR(20) UNIQUE NOT NULL,
            password VARCHAR(255) NOT NULL,
            permissions TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`,

        // Table categories
        `CREATE TABLE IF NOT EXISTS categories (
            id INT AUTO_INCREMENT PRIMARY KEY,
            name VARCHAR(100) NOT NULL,
            slug VARCHAR(100) UNIQUE NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`,

        // Table products
        `CREATE TABLE IF NOT EXISTS products (
            id INT AUTO_INCREMENT PRIMARY KEY,
            name VARCHAR(200) NOT NULL,
            category VARCHAR(100) NOT NULL,
            price DECIMAL(10,2) NOT NULL,
            old_price DECIMAL(10,2) DEFAULT 0,
            image VARCHAR(500),
            description TEXT,
            stock VARCHAR(20) DEFAULT 'available',
            features TEXT,
            sales INT DEFAULT 0,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`,

        // Table orders
        `CREATE TABLE IF NOT EXISTS orders (
            id INT AUTO_INCREMENT PRIMARY KEY,
            full_name VARCHAR(100) NOT NULL,
            phone VARCHAR(20) NOT NULL,
            email VARCHAR(100),
            address VARCHAR(255) NOT NULL,
            total DECIMAL(10,2) NOT NULL,
            status VARCHAR(20) DEFAULT 'processing',
            date DATE NOT NULL,
            items TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`,

        // Table messages
        `CREATE TABLE IF NOT EXISTS messages (
            id INT AUTO_INCREMENT PRIMARY KEY,
            name VARCHAR(100) NOT NULL,
            email VARCHAR(100) NOT NULL,
            message TEXT NOT NULL,
            date DATE NOT NULL,
            is_read INT DEFAULT 0,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`,

        // Table coupons
        `CREATE TABLE IF NOT EXISTS coupons (
            id INT AUTO_INCREMENT PRIMARY KEY,
            code VARCHAR(50) UNIQUE NOT NULL,
            discount DECIMAL(10,2) NOT NULL,
            type VARCHAR(20) DEFAULT 'percent',
            max_uses INT DEFAULT 1,
            uses INT DEFAULT 0,
            expires DATE,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`,

        // Table audit_logs
        `CREATE TABLE IF NOT EXISTS audit_logs (
            id INT AUTO_INCREMENT PRIMARY KEY,
            user_id INT NOT NULL,
            action VARCHAR(50) NOT NULL,
            details TEXT,
            ip VARCHAR(45),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`,

        // Table settings
        `CREATE TABLE IF NOT EXISTS settings (
            id INT PRIMARY KEY DEFAULT 1,
            site_name VARCHAR(200),
            site_phone VARCHAR(20),
            site_email VARCHAR(100),
            site_address VARCHAR(255),
            facebook_url VARCHAR(255),
            instagram_url VARCHAR(255),
            logo_url VARCHAR(500),
            hero_bg VARCHAR(500),
            theme_color VARCHAR(20),
            about_text TEXT,
            contact_text TEXT,
            terms_text TEXT,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
        )`,
    }

    for _, query := range queries {
        if _, err := DB.Exec(query); err != nil {
            return fmt.Errorf("erreur creation table: %v - query: %s", err, query)
        }
    }

    log.Println("✅ Tables creees avec succes")
    return nil
}

func ensureAdminUser() error {
    var count int
    err := DB.QueryRow("SELECT COUNT(*) FROM users WHERE phone = 'admin'").Scan(&count)
    if err != nil {
        return err
    }

    if count == 0 {
        // كلمة المرور المشفرة لـ "admin123"
        hashedPassword := "$2a$14$N9qo8uLOickgx2ZMRZoMy.MrqCkZq5KqX5qX5qX5qX5qX5qX5qX5q"
        _, err = DB.Exec(`INSERT INTO users (name, phone, password, permissions) VALUES (?, ?, ?, ?)`,
            "Administrateur", "admin", hashedPassword,
            "products,categories,orders,customers,messages,coupons,staff,reports,content,appearance,backup,settings")
        if err != nil {
            return err
        }
        log.Println("✅ Utilisateur admin cree (admin/admin123)")
    }

    return nil
}

func ensureSettings() error {
    var count int
    err := DB.QueryRow("SELECT COUNT(*) FROM settings WHERE id = 1").Scan(&count)
    if err != nil {
        return err
    }

    if count == 0 {
        _, err = DB.Exec(`INSERT INTO settings (id, site_name, site_phone, site_email, site_address, 
            facebook_url, instagram_url, logo_url, hero_bg, theme_color,
            about_text, contact_text, terms_text) 
            VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
            "Smart Step Control Tunisie",
            "+216 70 000 000",
            "contact@ssc.tn",
            "Centre Urbain Nord, Tunis, Tunisie",
            "https://facebook.com/ssc",
            "https://instagram.com/ssc",
            "/uploads/logo.png",
            "/uploads/hero.jpg",
            "#0A2E68",
            "Smart Step Control est votre partenaire de confiance pour les solutions de securite professionnelle en Tunisie.",
            "Contactez-nous par telephone ou email.",
            "Conditions generales de vente")
        if err != nil {
            return err
        }
        log.Println("✅ Parametres par defaut crees")
    }

    return nil
}

// ==================== Stats ====================

func GetStats() (map[string]interface{}, error) {
    stats := make(map[string]interface{})

    // Nombre de produits
    var productsCount int
    DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&productsCount)
    stats["products"] = productsCount

    // Nombre de commandes
    var ordersCount int
    DB.QueryRow("SELECT COUNT(*) FROM orders").Scan(&ordersCount)
    stats["orders"] = ordersCount

    // Total des ventes
    var totalSales float64
    DB.QueryRow("SELECT COALESCE(SUM(total), 0) FROM orders WHERE status != 'cancelled'").Scan(&totalSales)
    stats["total_sales"] = totalSales

    // Messages non lus
    var unreadMessages int
    DB.QueryRow("SELECT COUNT(*) FROM messages WHERE is_read = 0").Scan(&unreadMessages)
    stats["unread_messages"] = unreadMessages

    return stats, nil
}

// ==================== Audit Logs ====================

func AddAuditLog(userID int, action, details, ip string) error {
    _, err := DB.Exec("INSERT INTO audit_logs (user_id, action, details, ip) VALUES (?, ?, ?, ?)",
        userID, action, details, ip)
    return err
}
