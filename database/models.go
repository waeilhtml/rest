package database

// ==================== STRUCTURES PRINCIPALES ====================

// User - structure pour les utilisateurs (admin et staff)
type User struct {
    ID          int    `json:"id"`
    Name        string `json:"name"`
    Phone       string `json:"phone"`
    Password    string `json:"-"` // Ne jamais envoyer le mot de passe en JSON
    Permissions string `json:"permissions"`
}

// Product - structure pour les produits
type Product struct {
    ID          int     `json:"id"`
    Name        string  `json:"name"`
    Category    string  `json:"category"`
    Price       float64 `json:"price"`
    OldPrice    float64 `json:"old_price,omitempty"`
    Image       string  `json:"image"`
    Description string  `json:"description,omitempty"`
    Stock       string  `json:"stock"` // available, low, unavailable
    Features    string  `json:"features,omitempty"`
    Sales       int     `json:"sales"`
    Badge       string  `json:"badge,omitempty"` // new, sale
}

// Order - structure pour les commandes
type Order struct {
    ID       int     `json:"id"`
    FullName string  `json:"full_name"`
    Phone    string  `json:"phone"`
    Email    string  `json:"email,omitempty"`
    Address  string  `json:"address"`
    Total    float64 `json:"total"`
    Status   string  `json:"status"` // processing, confirmed, delivered, cancelled
    Date     string  `json:"date"`
    Items    string  `json:"items"` // JSON string des produits
}

// OrderItem - structure pour un article dans une commande
type OrderItem struct {
    Name     string  `json:"name"`
    Price    float64 `json:"price"`
    Quantity int     `json:"quantity"`
}

// Message - structure pour les messages de contact
type Message struct {
    ID      int    `json:"id"`
    Name    string `json:"name"`
    Email   string `json:"email"`
    Message string `json:"message"`
    Date    string `json:"date"`
    IsRead  int    `json:"is_read"` // 0 = non lu, 1 = lu
}

// Coupon - structure pour les codes promo
type Coupon struct {
    ID       int     `json:"id"`
    Code     string  `json:"code"`
    Discount float64 `json:"discount"`
    Type     string  `json:"type"` // percent, fixed
    MaxUses  int     `json:"max_uses"`
    Uses     int     `json:"uses"`
    Expires  string  `json:"expires,omitempty"`
}

// Category - structure pour les categories
type Category struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Slug string `json:"slug"`
}

// Settings - structure pour les parametres du site
type Settings struct {
    SiteName     string `json:"site_name"`
    SitePhone    string `json:"site_phone"`
    SiteEmail    string `json:"site_email"`
    SiteAddress  string `json:"site_address"`
    FacebookUrl  string `json:"facebook_url"`
    InstagramUrl string `json:"instagram_url"`
    LogoUrl      string `json:"logo_url"`
    HeroBg       string `json:"hero_bg"`
    ThemeColor   string `json:"theme_color"`
    AboutText    string `json:"about_text"`
    ContactText  string `json:"contact_text"`
    TermsText    string `json:"terms_text"`
}

// AuditLog - structure pour les journaux d'audit
type AuditLog struct {
    ID        int    `json:"id"`
    UserID    int    `json:"user_id"`
    Action    string `json:"action"`
    Details   string `json:"details"`
    IP        string `json:"ip"`
    CreatedAt string `json:"created_at"`
}

// Stats - structure pour les statistiques du dashboard
type Stats struct {
    Products       int     `json:"products"`
    Orders         int     `json:"orders"`
    TotalSales     float64 `json:"total_sales"`
    UnreadMessages int     `json:"unread_messages"`
    TotalCustomers int     `json:"total_customers"`
}

// ==================== STRUCTURES POUR LES REQUETES API ====================

// LoginRequest - structure pour la requete de connexion
type LoginRequest struct {
    Phone    string `json:"phone"`
    Password string `json:"password"`
}

// CreateOrderRequest - structure pour creer une commande
type CreateOrderRequest struct {
    FullName string      `json:"full_name"`
    Phone    string      `json:"phone"`
    Email    string      `json:"email"`
    Address  string      `json:"address"`
    Items    []OrderItem `json:"items"`
    Total    float64     `json:"total"`
    Coupon   string      `json:"coupon,omitempty"`
}

// CreateProductRequest - structure pour creer un produit
type CreateProductRequest struct {
    Name        string  `json:"name"`
    Category    string  `json:"category"`
    Price       float64 `json:"price"`
    OldPrice    float64 `json:"old_price"`
    Image       string  `json:"image"`
    Description string  `json:"description"`
    Stock       string  `json:"stock"`
    Features    string  `json:"features"`
}

// UpdateOrderStatusRequest - structure pour mettre a jour le statut d'une commande
type UpdateOrderStatusRequest struct {
    Status string `json:"status"`
}

// CreateCouponRequest - structure pour creer un coupon
type CreateCouponRequest struct {
    Code     string  `json:"code"`
    Discount float64 `json:"discount"`
    Type     string  `json:"type"`
    MaxUses  int     `json:"max_uses"`
    Expires  string  `json:"expires"`
}

// ValidateCouponRequest - structure pour valider un coupon
type ValidateCouponRequest struct {
    Code string `json:"code"`
}

// ContactRequest - structure pour le formulaire de contact
type ContactRequest struct {
    Name    string `json:"name"`
    Email   string `json:"email"`
    Message string `json:"message"`
}

// CreateStaffRequest - structure pour creer un employe
type CreateStaffRequest struct {
    Name        string   `json:"name"`
    Phone       string   `json:"phone"`
    Password    string   `json:"password"`
    Permissions []string `json:"permissions"`
}

// UpdateSettingsRequest - structure pour mettre a jour les parametres
type UpdateSettingsRequest struct {
    SiteName     string `json:"site_name"`
    SitePhone    string `json:"site_phone"`
    SiteEmail    string `json:"site_email"`
    SiteAddress  string `json:"site_address"`
    FacebookUrl  string `json:"facebook_url"`
    InstagramUrl string `json:"instagram_url"`
}

// UpdateAppearanceRequest - structure pour mettre a jour l'apparence
type UpdateAppearanceRequest struct {
    LogoUrl    string `json:"logo_url"`
    HeroBg     string `json:"hero_bg"`
    ThemeColor string `json:"theme_color"`
}

// UpdateContentRequest - structure pour mettre a jour le contenu des pages
type UpdateContentRequest struct {
    About   string `json:"about"`
    Contact string `json:"contact"`
    Terms   string `json:"terms"`
}

// ReportRequest - structure pour generer un rapport
type ReportRequest struct {
    StartDate string `json:"start_date"`
    EndDate   string `json:"end_date"`
}

// ==================== STRUCTURES POUR LES REPONSES API ====================

// LoginResponse - structure pour la reponse de connexion
type LoginResponse struct {
    Token string `json:"token"`
    User  User   `json:"user"`
}

// DashboardResponse - structure pour la reponse du dashboard
type DashboardResponse struct {
    Stats        Stats   `json:"stats"`
    RecentOrders []Order `json:"recent_orders"`
}

// ProductsResponse - structure pour la reponse des produits
type ProductsResponse struct {
    Products   []Product `json:"products"`
    TotalCount int       `json:"total_count"`
    Page       int       `json:"page"`
    PerPage    int       `json:"per_page"`
}

// OrdersResponse - structure pour la reponse des commandes
type OrdersResponse struct {
    Orders     []Order `json:"orders"`
    TotalCount int     `json:"total_count"`
    Page       int     `json:"page"`
    PerPage    int     `json:"per_page"`
}

// CouponValidationResponse - structure pour la validation d'un coupon
type CouponValidationResponse struct {
    Valid    bool    `json:"valid"`
    Discount float64 `json:"discount"`
    Type     string  `json:"type"`
    Message  string  `json:"message,omitempty"`
}

// ReportResponse - structure pour la reponse du rapport
type ReportResponse struct {
    StartDate  string  `json:"start_date"`
    EndDate    string  `json:"end_date"`
    OrderCount int     `json:"order_count"`
    TotalSales float64 `json:"total_sales"`
    Orders     []Order `json:"orders"`
}

// BackupData - structure pour la sauvegarde complete
type BackupData struct {
    Products   []Product  `json:"products"`
    Orders     []Order    `json:"orders"`
    Coupons    []Coupon   `json:"coupons"`
    Messages   []Message  `json:"messages"`
    Settings   Settings   `json:"settings"`
    Categories []Category `json:"categories"`
    ExportedAt string     `json:"exported_at"`
}

// ErrorResponse - structure pour les erreurs
type ErrorResponse struct {
    Success bool   `json:"success"`
    Error   string `json:"error"`
}

// SuccessResponse - structure pour les succes
type SuccessResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
}
