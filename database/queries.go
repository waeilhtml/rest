package database

import (
    "database/sql"
    "fmt"
)

// ==================== USER QUERIES ====================

func GetUserByPhone(phone string) (*User, error) {
    var user User
    err := DB.QueryRow("SELECT id, name, phone, password, permissions FROM users WHERE phone = ?", phone).
        Scan(&user.ID, &user.Name, &user.Phone, &user.Password, &user.Permissions)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func GetUserByID(id int) (*User, error) {
    var user User
    err := DB.QueryRow("SELECT id, name, phone, password, permissions FROM users WHERE id = ?", id).
        Scan(&user.ID, &user.Name, &user.Phone, &user.Password, &user.Permissions)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func GetAllUsers() ([]User, error) {
    rows, err := DB.Query("SELECT id, name, phone, permissions FROM users ORDER BY id")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var u User
        rows.Scan(&u.ID, &u.Name, &u.Phone, &u.Permissions)
        users = append(users, u)
    }
    return users, nil
}

func CreateUser(name, phone, password, permissions string) (int64, error) {
    result, err := DB.Exec("INSERT INTO users (name, phone, password, permissions) VALUES (?, ?, ?, ?)",
        name, phone, password, permissions)
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

// ==================== PRODUCT QUERIES ====================

func GetAllProducts() ([]Product, error) {
    rows, err := DB.Query("SELECT id, name, category, price, old_price, image, description, stock, features, sales FROM products ORDER BY id")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var products []Product
    for rows.Next() {
        var p Product
        err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.Price, &p.OldPrice, &p.Image, &p.Description, &p.Stock, &p.Features, &p.Sales)
        if err != nil {
            return nil, err
        }
        products = append(products, p)
    }
    return products, nil
}

func GetProductByID(id int) (*Product, error) {
    var p Product
    err := DB.QueryRow("SELECT id, name, category, price, old_price, image, description, stock, features, sales FROM products WHERE id = ?", id).
        Scan(&p.ID, &p.Name, &p.Category, &p.Price, &p.OldPrice, &p.Image, &p.Description, &p.Stock, &p.Features, &p.Sales)
    if err != nil {
        return nil, err
    }
    return &p, nil
}

func GetProductsByCategory(category string) ([]Product, error) {
    rows, err := DB.Query("SELECT id, name, category, price, old_price, image, description, stock, features, sales FROM products WHERE category = ? ORDER BY id", category)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var products []Product
    for rows.Next() {
        var p Product
        rows.Scan(&p.ID, &p.Name, &p.Category, &p.Price, &p.OldPrice, &p.Image, &p.Description, &p.Stock, &p.Features, &p.Sales)
        products = append(products, p)
    }
    return products, nil
}

func CreateProduct(p *Product) (int64, error) {
    result, err := DB.Exec(`INSERT INTO products (name, category, price, old_price, image, description, stock, features, sales) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        p.Name, p.Category, p.Price, p.OldPrice, p.Image, p.Description, p.Stock, p.Features, 0)
    if err != nil {
        return 0, err
    }
    return result.LastInsertId()
}

func UpdateProduct(id int, p *Product) error {
    _, err := DB.Exec(`UPDATE products SET name=?, category=?, price=?, old_price=?, image=?, description=?, stock=?, features=? WHERE id=?`,
        p.Name, p.Category, p.Price, p.OldPrice, p.Image, p.Description, p.Stock, p.Features, id)
    return err
}

func DeleteProduct(id int) error {
    _, err := DB.Exec("DELETE FROM products WHERE id = ?", id)
    return err
}

func UpdateProductStock(id int, stock string) error {
    _, err := DB.Exec("UPDATE products SET stock = ? WHERE id = ?", stock, id)
    return err
}

func IncrementProductSales(id int, quantity int) error {
    _, err := DB.Exec("UPDATE products SET sales = sales + ? WHERE id = ?", quantity, id)
    return err
}

// ==================== ORDER QUERIES ====================

func GetAllOrders() ([]Order, error) {
    rows, err := DB.Query("SELECT id, full_name, phone, email, address, total, status, date, items FROM orders ORDER BY id DESC")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var orders []Order
    for rows.Next() {
        var o Order
        err := rows.Scan(&o.ID, &o.FullName, &o.Phone, &o.Email, &o.Address, &o.Total, &o.Status, &o.Date, &o.Items)
        if err != nil {
            return nil, err
        }
        orders = append(orders, o)
    }
    return orders, nil
}

func GetOrderByID(id int) (*Order, error) {
    var o Order
    err := DB.QueryRow("SELECT id, full_name, phone, email, address, total, status, date, items FROM orders WHERE id = ?", id).
        Scan(&o.ID, &o.FullName, &o.Phone, &o.Email, &o.Address, &o.Total, &o.Status, &o.Date, &o.Items)
    if err != nil {
        return nil, err
    }
    return &o, nil
}

func GetOrdersByPhone(phone string) ([]Order, error) {
    rows, err := DB.Query("SELECT id, full_name, phone, email, address, total, status, date, items FROM orders WHERE phone = ? ORDER BY id DESC", phone)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var orders []Order
    for rows.Next() {
        var o Order
        rows.Scan(&o.ID, &o.FullName, &o.Phone, &o.Email, &o.Address, &o.Total, &o.Status, &o.Date, &o.Items)
        orders = append(orders, o)
    }
    return orders, nil
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

func GetOrdersByDateRange(startDate, endDate string) ([]Order, error) {
    rows, err := DB.Query("SELECT id, full_name, phone, email, address, total, status, date, items FROM orders WHERE date BETWEEN ? AND ? ORDER BY date DESC", startDate, endDate)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var orders []Order
    for rows.Next() {
        var o Order
        rows.Scan(&o.ID, &o.FullName, &o.Phone, &o.Email, &o.Address, &o.Total, &o.Status, &o.Date, &o.Items)
        orders = append(orders, o)
    }
    return orders, nil
}

// ==================== MESSAGE QUERIES ====================

func GetAllMessages() ([]Message, error) {
    rows, err := DB.Query("SELECT id, name, email, message, date, is_read FROM messages ORDER BY id DESC")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var messages []Message
    for rows.Next() {
        var m Message
        rows.Scan(&m.ID, &m.Name, &m.Email, &m.Message, &m.Date, &m.IsRead)
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

func GetUnreadMessagesCount() (int, error) {
    var count int
    err := DB.QueryRow("SELECT COUNT(*) FROM messages WHERE is_read = 0").Scan(&count)
    return count, err
}

// ==================== COUPON QUERIES ====================

func GetAllCoupons() ([]Coupon, error) {
    rows, err := DB.Query("SELECT id, code, discount, type, max_uses, uses, expires FROM coupons ORDER BY id")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var coupons []Coupon
    for rows.Next() {
        var c Coupon
        var expires sql.NullString
        rows.Scan(&c.ID, &c.Code, &c.Discount, &c.Type, &c.MaxUses, &c.Uses, &expires)
        if expires.Valid {
            c.Expires = expires.String
        }
        coupons = append(coupons, c)
    }
    return coupons, nil
}

func GetCouponByCode(code string) (*Coupon, error) {
    var c Coupon
    var expires sql.NullString
    err := DB.QueryRow("SELECT id, code, discount, type, max_uses, uses, expires FROM coupons WHERE code = ?", code).
        Scan(&c.ID, &c.Code, &c.Discount, &c.Type, &c.MaxUses, &c.Uses, &expires)
    if err != nil {
        return nil, err
    }
    if expires.Valid {
        c.Expires = expires.String
    }
    return &c, nil
}

func CreateCoupon(c *Coupon) error {
    _, err := DB.Exec(`INSERT INTO coupons (code, discount, type, max_uses, uses, expires) 
        VALUES (?, ?, ?, ?, 0, ?)`,
        c.Code, c.Discount, c.Type, c.MaxUses, c.Expires)
    return err
}

func UpdateCouponUses(code string) error {
    _, err := DB.Exec("UPDATE coupons SET uses = uses + 1 WHERE code = ?", code)
    return err
}

func DeleteCoupon(code string) error {
    _, err := DB.Exec("DELETE FROM coupons WHERE code = ?", code)
    return err
}

// ==================== CATEGORY QUERIES ====================

func GetAllCategories() ([]Category, error) {
    rows, err := DB.Query("SELECT id, name, slug FROM categories ORDER BY id")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var categories []Category
    for rows.Next() {
        var c Category
        rows.Scan(&c.ID, &c.Name, &c.Slug)
        categories = append(categories, c)
    }
    return categories, nil
}

func GetCategoryByID(id int) (*Category, error) {
    var c Category
    err := DB.QueryRow("SELECT id, name, slug FROM categories WHERE id = ?", id).
        Scan(&c.ID, &c.Name, &c.Slug)
    if err != nil {
        return nil, err
    }
    return &c, nil
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

// ==================== SETTINGS QUERIES ====================

func GetSettings() (*Settings, error) {
    var s Settings
    err := DB.QueryRow(`SELECT site_name, site_phone, site_email, site_address, 
        facebook_url, instagram_url, logo_url, hero_bg, theme_color,
        about_text, contact_text, terms_text 
        FROM settings WHERE id = 1`).Scan(
        &s.SiteName, &s.SitePhone, &s.SiteEmail, &s.SiteAddress,
        &s.FacebookUrl, &s.InstagramUrl, &s.LogoUrl, &s.HeroBg, &s.ThemeColor,
        &s.AboutText, &s.ContactText, &s.TermsText)
    if err != nil {
        return nil, err
    }
    return &s, nil
}

func UpdateSettings(s *Settings) error {
    _, err := DB.Exec(`UPDATE settings SET 
        site_name = ?, site_phone = ?, site_email = ?, site_address = ?,
        facebook_url = ?, instagram_url = ? 
        WHERE id = 1`,
        s.SiteName, s.SitePhone, s.SiteEmail, s.SiteAddress,
        s.FacebookUrl, s.InstagramUrl)
    return err
}

func UpdateAppearance(logoUrl, heroBg, themeColor string) error {
    _, err := DB.Exec(`UPDATE settings SET logo_url = ?, hero_bg = ?, theme_color = ? WHERE id = 1`,
        logoUrl, heroBg, themeColor)
    return err
}

func UpdateContent(about, contact, terms string) error {
    _, err := DB.Exec(`UPDATE settings SET about_text = ?, contact_text = ?, terms_text = ? WHERE id = 1`,
        about, contact, terms)
    return err
}

// ==================== STATS QUERIES ====================

func GetStats() (map[string]interface{}, error) {
    stats := make(map[string]interface{})

    var productsCount int
    DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&productsCount)
    stats["products"] = productsCount

    var ordersCount int
    DB.QueryRow("SELECT COUNT(*) FROM orders").Scan(&ordersCount)
    stats["orders"] = ordersCount

    var totalSales float64
    DB.QueryRow("SELECT COALESCE(SUM(total), 0) FROM orders WHERE status != 'cancelled'").Scan(&totalSales)
    stats["total_sales"] = totalSales

    var unreadMessages int
    DB.QueryRow("SELECT COUNT(*) FROM messages WHERE is_read = 0").Scan(&unreadMessages)
    stats["unread_messages"] = unreadMessages

    var totalCustomers int
    DB.QueryRow("SELECT COUNT(DISTINCT phone) FROM orders").Scan(&totalCustomers)
    stats["total_customers"] = totalCustomers

    return stats, nil
}

func GetMonthlySales(year int) ([]float64, error) {
    monthlySales := make([]float64, 12)
    
    rows, err := DB.Query(`
        SELECT MONTH(date) as month, SUM(total) as total 
        FROM orders 
        WHERE status != 'cancelled' AND YEAR(date) = ?
        GROUP BY MONTH(date)`, year)
    if err != nil {
        return monthlySales, err
    }
    defer rows.Close()

    for rows.Next() {
        var month int
        var total float64
        rows.Scan(&month, &total)
        if month >= 1 && month <= 12 {
            monthlySales[month-1] = total
        }
    }
    return monthlySales, nil
}

// ==================== AUDIT LOG QUERIES ====================

func AddAuditLog(userID int, action, details, ip string) error {
    _, err := DB.Exec("INSERT INTO audit_logs (user_id, action, details, ip) VALUES (?, ?, ?, ?)",
        userID, action, details, ip)
    return err
}

func GetAuditLogs(limit int) ([]AuditLog, error) {
    rows, err := DB.Query("SELECT id, user_id, action, details, ip, created_at FROM audit_logs ORDER BY id DESC LIMIT ?", limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var logs []AuditLog
    for rows.Next() {
        var l AuditLog
        var createdAt string
        rows.Scan(&l.ID, &l.UserID, &l.Action, &l.Details, &l.IP, &createdAt)
        l.CreatedAt = createdAt
        logs = append(logs, l)
    }
    return logs, nil
}

// ==================== DASHBOARD QUERIES ====================

func GetRecentOrders(limit int) ([]Order, error) {
    rows, err := DB.Query("SELECT id, full_name, phone, total, status, date FROM orders ORDER BY id DESC LIMIT ?", limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var orders []Order
    for rows.Next() {
        var o Order
        rows.Scan(&o.ID, &o.FullName, &o.Phone, &o.Total, &o.Status, &o.Date)
        orders = append(orders, o)
    }
    return orders, nil
}

func GetTopProducts(limit int) ([]Product, error) {
    rows, err := DB.Query("SELECT id, name, price, sales FROM products ORDER BY sales DESC LIMIT ?", limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var products []Product
    for rows.Next() {
        var p Product
        rows.Scan(&p.ID, &p.Name, &p.Price, &p.Sales)
        products = append(products, p)
    }
    return products, nil
}
