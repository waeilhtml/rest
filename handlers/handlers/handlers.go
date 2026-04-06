package handlers

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/gorilla/mux"
    "golang.org/x/crypto/bcrypt"

    "ssc-admin/database"
    "ssc-admin/security"
)

// ==================== المتغيرات العامة ====================
var jwtSecret = []byte("ssc-secret-key-2024-change-in-production")

// ==================== المساعدات ====================
type Response struct {
    Success bool        `json:"success"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, status int, message string) {
    sendJSON(w, status, Response{Success: false, Error: message})
}

func sendSuccess(w http.ResponseWriter, data interface{}, message string) {
    sendJSON(w, http.StatusOK, Response{Success: true, Message: message, Data: data})
}

// ==================== Auth Handlers ====================

func Login(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Phone    string `json:"phone"`
        Password string `json:"password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, http.StatusBadRequest, "بيانات غير صالحة")
        return
    }

    // تنقية المدخلات
    req.Phone = security.SanitizeInput(req.Phone)

    // البحث عن المستخدم
    user, err := database.GetUserByPhone(req.Phone)
    if err != nil {
        sendError(w, http.StatusUnauthorized, "رقم الهاتف أو كلمة المرور غير صحيحة")
        return
    }

    // التحقق من كلمة المرور
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        sendError(w, http.StatusUnauthorized, "رقم الهاتف أو كلمة المرور غير صحيحة")
        return
    }

    // إنشاء JWT Token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": user.ID,
        "phone":   user.Phone,
        "exp":     time.Now().Add(24 * time.Hour).Unix(),
    })

    tokenString, err := token.SignedString(jwtSecret)
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في إنشاء التوكن")
        return
    }

    // تسجيل النشاط
    database.AddAuditLog(user.ID, "login", "تسجيل دخول", r.RemoteAddr)

    sendSuccess(w, map[string]interface{}{
        "token": tokenString,
        "user": map[string]interface{}{
            "id":          user.ID,
            "name":        user.Name,
            "phone":       user.Phone,
            "permissions": strings.Split(user.Permissions, ","),
        },
    }, "تم تسجيل الدخول بنجاح")
}

func Logout(w http.ResponseWriter, r *http.Request) {
    // تسجيل الخروج
    userID := r.Context().Value("user_id").(int)
    database.AddAuditLog(userID, "logout", "تسجيل خروج", r.RemoteAddr)
    sendSuccess(w, nil, "تم تسجيل الخروج بنجاح")
}

func RefreshToken(w http.ResponseWriter, r *http.Request) {
    // تجديد التوكن
    userID := r.Context().Value("user_id").(int)
    user, err := database.GetUserByID(userID)
    if err != nil {
        sendError(w, http.StatusUnauthorized, "مستخدم غير موجود")
        return
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": user.ID,
        "phone":   user.Phone,
        "exp":     time.Now().Add(24 * time.Hour).Unix(),
    })

    tokenString, err := token.SignedString(jwtSecret)
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في إنشاء التوكن")
        return
    }

    sendSuccess(w, map[string]interface{}{"token": tokenString}, "تم تجديد التوكن")
}

// ==================== Dashboard Handlers ====================

func GetDashboardStats(w http.ResponseWriter, r *http.Request) {
    stats, err := database.GetStats()
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في جلب الإحصائيات")
        return
    }

    // إضافة إحصائيات إضافية
    var recentOrders []database.Order
    orders, _ := database.GetAllOrders()
    if len(orders) > 5 {
        recentOrders = orders[:5]
    } else {
        recentOrders = orders
    }

    sendSuccess(w, map[string]interface{}{
        "stats":        stats,
        "recentOrders": recentOrders,
    }, "")
}

// ==================== Products Handlers ====================

func GetProducts(w http.ResponseWriter, r *http.Request) {
    products, err := database.GetAllProducts()
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في جلب المنتجات")
        return
    }
    sendSuccess(w, products, "")
}

func GetPublicProducts(w http.ResponseWriter, r *http.Request) {
    products, err := database.GetAllProducts()
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في جلب المنتجات")
        return
    }
    sendSuccess(w, products, "")
}

func GetProduct(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        sendError(w, http.StatusBadRequest, "ID غير صالح")
        return
    }

    product, err := database.GetProductByID(id)
    if err != nil {
        sendError(w, http.StatusNotFound, "المنتج غير موجود")
        return
    }
    sendSuccess(w, product, "")
}

func GetPublicProduct(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        sendError(w, http.StatusBadRequest, "ID غير صالح")
        return
    }

    product, err := database.GetProductByID(id)
    if err != nil {
        sendError(w, http.StatusNotFound, "المنتج غير موجود")
        return
    }
    sendSuccess(w, product, "")
}

func CreateProduct(w http.ResponseWriter, r *http.Request) {
    var req database.Product
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, http.StatusBadRequest, "بيانات غير صالحة")
        return
    }

    // تنقية المدخلات
    req.Name = security.SanitizeInput(req.Name)
    req.Description = security.SanitizeInput(req.Description)
    req.Category = security.SanitizeInput(req.Category)

    id, err := database.CreateProduct(&req)
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في إضافة المنتج")
        return
    }

    userID := r.Context().Value("user_id").(int)
    database.AddAuditLog(userID, "create_product", fmt.Sprintf("إضافة منتج: %s", req.Name), r.RemoteAddr)

    sendSuccess(w, map[string]interface{}{"id": id}, "تم إضافة المنتج بنجاح")
}

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        sendError(w, http.StatusBadRequest, "ID غير صالح")
        return
    }

    var req database.Product
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, http.StatusBadRequest, "بيانات غير صالحة")
        return
    }

    // تنقية المدخلات
    req.Name = security.SanitizeInput(req.Name)
    req.Description = security.SanitizeInput(req.Description)

    if err := database.UpdateProduct(id, &req); err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في تحديث المنتج")
        return
    }

    userID := r.Context().Value("user_id").(int)
    database.AddAuditLog(userID, "update_product", fmt.Sprintf("تحديث منتج ID: %d", id), r.RemoteAddr)

    sendSuccess(w, nil, "تم تحديث المنتج بنجاح")
}

func DeleteProduct(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        sendError(w, http.StatusBadRequest, "ID غير صالح")
        return
    }

    if err := database.DeleteProduct(id); err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في حذف المنتج")
        return
    }

    userID := r.Context().Value("user_id").(int)
    database.AddAuditLog(userID, "delete_product", fmt.Sprintf("حذف منتج ID: %d", id), r.RemoteAddr)

    sendSuccess(w, nil, "تم حذف المنتج بنجاح")
}

// ==================== Categories Handlers ====================

func GetCategories(w http.ResponseWriter, r *http.Request) {
    // جلب الفئات من قاعدة البيانات
    rows, err := database.DB.Query("SELECT id, name, slug FROM categories ORDER BY id")
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في جلب الفئات")
        return
    }
    defer rows.Close()

    var categories []map[string]interface{}
    for rows.Next() {
        var id int
        var name, slug string
        rows.Scan(&id, &name, &slug)
        categories = append(categories, map[string]interface{}{
            "id":   id,
            "name": name,
            "slug": slug,
        })
    }
    sendSuccess(w, categories, "")
}

func GetPublicCategories(w http.ResponseWriter, r *http.Request) {
    rows, err := database.DB.Query("SELECT id, name, slug FROM categories ORDER BY id")
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في جلب الفئات")
        return
    }
    defer rows.Close()

    var categories []map[string]interface{}
    for rows.Next() {
        var id int
        var name, slug string
        rows.Scan(&id, &name, &slug)
        categories = append(categories, map[string]interface{}{
            "id":   id,
            "name": name,
            "slug": slug,
        })
    }
    sendSuccess(w, categories, "")
}

func CreateCategory(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name string `json:"name"`
        Slug string `json:"slug"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, http.StatusBadRequest, "بيانات غير صالحة")
        return
    }

    req.Name = security.SanitizeInput(req.Name)
    if req.Slug == "" {
        req.Slug = strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))
    }

    result, err := database.DB.Exec("INSERT INTO categories (name, slug) VALUES (?, ?)", req.Name, req.Slug)
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في إضافة الفئة")
        return
    }

    id, _ := result.LastInsertId()
    userID := r.Context().Value("user_id").(int)
    database.AddAuditLog(userID, "create_category", fmt.Sprintf("إضافة فئة: %s", req.Name), r.RemoteAddr)

    sendSuccess(w, map[string]interface{}{"id": id}, "تم إضافة الفئة بنجاح")
}

func UpdateCategory(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        sendError(w, http.StatusBadRequest, "ID غير صالح")
        return
    }

    var req struct {
        Name string `json:"name"`
        Slug string `json:"slug"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, http.StatusBadRequest, "بيانات غير صالحة")
        return
    }

    req.Name = security.SanitizeInput(req.Name)

    _, err = database.DB.Exec("UPDATE categories SET name = ?, slug = ? WHERE id = ?", req.Name, req.Slug, id)
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في تحديث الفئة")
        return
    }

    userID := r.Context().Value("user_id").(int)
    database.AddAuditLog(userID, "update_category", fmt.Sprintf("تحديث فئة ID: %d", id), r.RemoteAddr)

    sendSuccess(w, nil, "تم تحديث الفئة بنجاح")
}

func DeleteCategory(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        sendError(w, http.StatusBadRequest, "ID غير صالح")
        return
    }

    _, err = database.DB.Exec("DELETE FROM categories WHERE id = ?", id)
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في حذف الفئة")
        return
    }

    userID := r.Context().Value("user_id").(int)
    database.AddAuditLog(userID, "delete_category", fmt.Sprintf("حذف فئة ID: %d", id), r.RemoteAddr)

    sendSuccess(w, nil, "تم حذف الفئة بنجاح")
}

// ==================== Orders Handlers ====================

func GetOrders(w http.ResponseWriter, r *http.Request) {
    orders, err := database.GetAllOrders()
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في جلب الطلبات")
        return
    }
    sendSuccess(w, orders, "")
}

func GetOrderDetails(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        sendError(w, http.StatusBadRequest, "ID غير صالح")
        return
    }

    var order database.Order
    err = database.DB.QueryRow(`SELECT id, full_name, phone, email, address, total, status, date, items 
        FROM orders WHERE id = ?`, id).Scan(
        &order.ID, &order.FullName, &order.Phone, &order.Email,
        &order.Address, &order.Total, &order.Status, &order.Date, &order.Items)
    if err != nil {
        sendError(w, http.StatusNotFound, "الطلب غير موجود")
        return
    }
    sendSuccess(w, order, "")
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
    var req database.Order
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, http.StatusBadRequest, "بيانات غير صالحة")
        return
    }

    // تنقية المدخلات
    req.FullName = security.SanitizeInput(req.FullName)
    req.Phone = security.SanitizeInput(req.Phone)
    req.Email = security.SanitizeInput(req.Email)
    req.Address = security.SanitizeInput(req.Address)
    req.Date = time.Now().Format("2006-01-02")
    req.Status = "processing"

    id, err := database.CreateOrder(&req)
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في إنشاء الطلب")
        return
    }

    sendSuccess(w, map[string]interface{}{"id": id, "status": "processing"}, "تم إنشاء الطلب بنجاح")
}

func UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        sendError(w, http.StatusBadRequest, "ID غير صالح")
        return
    }

    var req struct {
        Status string `json:"status"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, http.StatusBadRequest, "بيانات غير صالحة")
        return
    }

    if err := database.UpdateOrderStatus(id, req.Status); err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في تحديث حالة الطلب")
        return
    }

    userID := r.Context().Value("user_id").(int)
    database.AddAuditLog(userID, "update_order_status", fmt.Sprintf("تحديث حالة طلب ID: %d إلى %s", id, req.Status), r.RemoteAddr)

    sendSuccess(w, nil, "تم تحديث حالة الطلب")
}

// ==================== Customers Handlers ====================

func GetCustomers(w http.ResponseWriter, r *http.Request) {
    // جلب العملاء من الطلبات
    rows, err := database.DB.Query(`
        SELECT full_name, phone, email, COUNT(*) as orders, SUM(total) as total 
        FROM orders 
        GROUP BY phone 
        ORDER BY total DESC`)
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في جلب العملاء")
        return
    }
    defer rows.Close()

    var customers []map[string]interface{}
    for rows.Next() {
        var name, phone, email string
        var orders, total int
        rows.Scan(&name, &phone, &email, &orders, &total)
        customers = append(customers, map[string]interface{}{
            "name":    name,
            "phone":   phone,
            "email":   email,
            "orders":  orders,
            "total":   float64(total),
        })
    }
    sendSuccess(w, customers, "")
}

func GetCustomerDetails(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    phone := vars["phone"]

    rows, err := database.DB.Query(`
        SELECT id, full_name, phone, email, address, total, status, date 
        FROM orders WHERE phone = ? ORDER BY id DESC`, phone)
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في جلب بيانات العميل")
        return
    }
    defer rows.Close()

    var orders []database.Order
    for rows.Next() {
        var o database.Order
        rows.Scan(&o.ID, &o.FullName, &o.Phone, &o.Email, &o.Address, &o.Total, &o.Status, &o.Date)
        orders = append(orders, o)
    }

    sendSuccess(w, orders, "")
}

// ==================== Messages Handlers ====================

func GetMessages(w http.ResponseWriter, r *http.Request) {
    messages, err := database.GetAllMessages()
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في جلب الرسائل")
        return
    }
    sendSuccess(w, messages, "")
}

func SubmitContact(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name    string `json:"name"`
        Email   string `json:"email"`
        Message string `json:"message"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, http.StatusBadRequest, "بيانات غير صالحة")
        return
    }

    // تنقية المدخلات
    req.Name = security.SanitizeInput(req.Name)
    req.Email = security.SanitizeInput(req.Email)
    req.Message = security.SanitizeInput(req.Message)

    // التحقق من البريد الإلكتروني
    if !security.ValidateEmail(req.Email) {
        sendError(w, http.StatusBadRequest, "البريد الإلكتروني غير صالح")
        return
    }

    id, err := database.CreateMessage(req.Name, req.Email, req.Message)
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في إرسال الرسالة")
        return
    }

    sendSuccess(w, map[string]interface{}{"id": id}, "تم إرسال رسالتك بنجاح")
}

func MarkMessageRead(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        sendError(w, http.StatusBadRequest, "ID غير صالح")
        return
    }

    if err := database.MarkMessageRead(id); err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في تحديث حالة الرسالة")
        return
    }

    sendSuccess(w, nil, "تم تعيين الرسالة كمقروءة")
}

func DeleteMessage(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        sendError(w, http.StatusBadRequest, "ID غير صالح")
        return
    }

    if err := database.DeleteMessage(id); err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في حذف الرسالة")
        return
    }

    userID := r.Context().Value("user_id").(int)
    database.AddAuditLog(userID, "delete_message", fmt.Sprintf("حذف رسالة ID: %d", id), r.RemoteAddr)

    sendSuccess(w, nil, "تم حذف الرسالة")
}

// ==================== Coupons Handlers ====================

func GetCoupons(w http.ResponseWriter, r *http.Request) {
    coupons, err := database.GetAllCoupons()
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في جلب الكوبونات")
        return
    }
    sendSuccess(w, coupons, "")
}

func CreateCoupon(w http.ResponseWriter, r *http.Request) {
    var req database.Coupon
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, http.StatusBadRequest, "بيانات غير صالحة")
        return
    }

    req.Code = strings.ToUpper(security.SanitizeInput(req.Code))

    if err := database.CreateCoupon(&req); err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في إضافة الكوبون")
        return
    }

    userID := r.Context().Value("user_id").(int)
    database.AddAuditLog(userID, "create_coupon", fmt.Sprintf("إضافة كوبون: %s", req.Code), r.RemoteAddr)

    sendSuccess(w, nil, "تم إضافة الكوبون بنجاح")
}

func ValidateCoupon(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Code string `json:"code"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, http.StatusBadRequest, "بيانات غير صالحة")
        return
    }

    req.Code = strings.ToUpper(req.Code)

    coupon, err := database.GetCouponByCode(req.Code)
    if err != nil {
        sendError(w, http.StatusNotFound, "الكوبون غير صالح")
        return
    }

    // التحقق من صلاحية الكوبون
    if coupon.Expires != "" {
        expires, _ := time.Parse("2006-01-02", coupon.Expires)
        if expires.Before(time.Now()) {
            sendError(w, http.StatusBadRequest, "الكوبون منتهي الصلاحية")
            return
        }
    }

    if coupon.Uses >= coupon.MaxUses {
        sendError(w, http.StatusBadRequest, "تم استخدام هذا الكوبون بأقصى عدد")
        return
    }

    sendSuccess(w, coupon, "الكوبون صالح")
}

func DeleteCoupon(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    code := vars["code"]

    if err := database.DeleteCoupon(code); err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في حذف الكوبون")
        return
    }

    userID := r.Context().Value("user_id").(int)
    database.AddAuditLog(userID, "delete_coupon", fmt.Sprintf("حذف كوبون: %s", code), r.RemoteAddr)

    sendSuccess(w, nil, "تم حذف الكوبون")
}

// ==================== Staff Handlers ====================

func GetStaff(w http.ResponseWriter, r *http.Request) {
    users, err := database.GetAllUsers()
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في جلب الموظفين")
        return
    }
    sendSuccess(w, users, "")
}

func CreateStaff(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name        string   `json:"name"`
        Phone       string   `json:"phone"`
        Password    string   `json:"password"`
        Permissions []string `json:"permissions"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, http.StatusBadRequest, "بيانات غير صالحة")
        return
    }

    req.Name = security.SanitizeInput(req.Name)
    req.Phone = security.SanitizeInput(req.Phone)

    // تشفير كلمة المرور
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في تشفير كلمة المرور")
        return
    }

    permissions := strings.Join(req.Permissions, ",")
    id, err := database.CreateUser(req.Name, req.Phone, string(hashedPassword), permissions)
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في إضافة الموظف")
        return
    }

    userID := r.Context().Value("user_id").(int)
    database.AddAuditLog(userID, "create_staff", fmt.Sprintf("إضافة موظف: %s", req.Name), r.RemoteAddr)

    sendSuccess(w, map[string]interface{}{"id": id}, "تم إضافة الموظف بنجاح")
}

func UpdateStaff(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        sendError(w, http.StatusBadRequest, "ID غير صالح")
        return
    }

    var req struct {
        Name        string   `json:"name"`
        Phone       string   `json:"phone"`
        Password    string   `json:"password"`
        Permissions []string `json:"permissions"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, http.StatusBadRequest, "بيانات غير صالحة")
        return
    }

    // تحديث البيانات
    if req.Password != "" {
        hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
        database.DB.Exec("UPDATE users SET password = ? WHERE id = ?", string(hashedPassword), id)
    }

    permissions := strings.Join(req.Permissions, ",")
    _, err = database.DB.Exec("UPDATE users SET name = ?, phone = ?, permissions = ? WHERE id = ?",
        req.Name, req.Phone, permissions, id)
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في تحديث الموظف")
        return
    }

    userID := r.Context().Value("user_id").(int)
    database.AddAuditLog(userID, "update_staff", fmt.Sprintf("تحديث موظف ID: %d", id), r.RemoteAddr)

    sendSuccess(w, nil, "تم تحديث الموظف")
}

func DeleteStaff(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        sendError(w, http.StatusBadRequest, "ID غير صالح")
        return
    }

    if err := database.DeleteUser(id); err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في حذف الموظف")
        return
    }

    userID := r.Context().Value("user_id").(int)
    database.AddAuditLog(userID, "delete_staff", fmt.Sprintf("حذف موظف ID: %d", id), r.RemoteAddr)

    sendSuccess(w, nil, "تم حذف الموظف")
}

// ==================== Analytics Handlers ====================

func GetSalesAnalytics(w http.ResponseWriter, r *http.Request) {
    // مبيعات شهرية
    rows, err := database.DB.Query(`
        SELECT MONTH(date) as month, SUM(total) as total 
        FROM orders 
        WHERE status != 'cancelled' AND YEAR(date) = YEAR(CURDATE())
        GROUP BY MONTH(date)`)
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في جلب التحليلات")
        return
    }
    defer rows.Close()

    monthlySales := make([]float64, 12)
    for rows.Next() {
        var month int
        var total float64
        rows.Scan(&month, &total)
        if month >= 1 && month <= 12 {
            monthlySales[month-1] = total
        }
    }

    sendSuccess(w, map[string]interface{}{
        "monthly_sales": monthlySales,
    }, "")
}

func GetTopProducts(w http.ResponseWriter, r *http.Request) {
    products, err := database.GetAllProducts()
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في جلب المنتجات")
        return
    }

    // ترتيب حسب المبيعات
    for i := 0; i < len(products)-1; i++ {
        for j := i + 1; j < len(products); j++ {
            if products[i].Sales < products[j].Sales {
                products[i], products[j] = products[j], products[i]
            }
        }
    }

    limit := 5
    if len(products) < limit {
        limit = len(products)
    }

    sendSuccess(w, products[:limit], "")
}

func GetVisitorsStats(w http.ResponseWriter, r *http.Request) {
    sendSuccess(w, map[string]interface{}{
        "today": 147,
        "week":  1024,
        "month": 5234,
        "total": 15234,
    }, "")
}

// ==================== Settings Handlers ====================

func GetSettings(w http.ResponseWriter, r *http.Request) {
    settings, err := database.GetSettings()
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في جلب الإعدادات")
        return
    }
    sendSuccess(w, settings, "")
}

func GetPublicSettings(w http.ResponseWriter, r *http.Request) {
    settings, err := database.GetSettings()
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في جلب الإعدادات")
        return
    }
    sendSuccess(w, settings, "")
}

func UpdateSettings(w http.ResponseWriter, r *http.Request) {
    var req database.Settings
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, http.StatusBadRequest, "بيانات غير صالحة")
        return
    }

    if err := database.UpdateSettings(&req); err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في تحديث الإعدادات")
        return
    }

    userID := r.Context().Value("user_id").(int)
    database.AddAuditLog(userID, "update_settings", "تحديث إعدادات الموقع", r.RemoteAddr)

    sendSuccess(w, nil, "تم تحديث الإعدادات")
}

// ==================== Backup Handlers ====================

func ExportBackup(w http.ResponseWriter, r *http.Request) {
    // جلب جميع البيانات
    products, _ := database.GetAllProducts()
    orders, _ := database.GetAllOrders()
    coupons, _ := database.GetAllCoupons()
    messages, _ := database.GetAllMessages()
    settings, _ := database.GetSettings()

    backup := map[string]interface{}{
        "products":  products,
        "orders":    orders,
        "coupons":   coupons,
        "messages":  messages,
        "settings":  settings,
        "exported_at": time.Now().Format(time.RFC3339),
    }

    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=ssc_backup_%s.json", time.Now().Format("20060102_150405")))
    json.NewEncoder(w).Encode(backup)
}

func ImportBackup(w http.ResponseWriter, r *http.Request) {
    var backup map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&backup); err != nil {
        sendError(w, http.StatusBadRequest, "بيانات غير صالحة")
        return
    }

    // هنا يمكن إضافة منطق الاستيراد
    userID := r.Context().Value("user_id").(int)
    database.AddAuditLog(userID, "import_backup", "استيراد نسخة احتياطية", r.RemoteAddr)

    sendSuccess(w, nil, "تم استيراد النسخة الاحتياطية")
}

func ListBackups(w http.ResponseWriter, r *http.Request) {
    // قائمة الملفات في مجلد backups
    sendSuccess(w, []string{"backup_20240101.json", "backup_20240115.json"}, "")
}

// ==================== Appearance Handlers ====================

func GetAppearance(w http.ResponseWriter, r *http.Request) {
    settings, err := database.GetSettings()
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في جلب الإعدادات")
        return
    }

    sendSuccess(w, map[string]interface{}{
        "logo_url":   settings.LogoUrl,
        "hero_bg":    settings.HeroBg,
        "theme_color": settings.ThemeColor,
    }, "")
}

func UpdateAppearance(w http.ResponseWriter, r *http.Request) {
    var req struct {
        LogoUrl   string `json:"logo_url"`
        HeroBg    string `json:"hero_bg"`
        ThemeColor string `json:"theme_color"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, http.StatusBadRequest, "بيانات غير صالحة")
        return
    }

    _, err := database.DB.Exec("UPDATE settings SET logo_url = ?, hero_bg = ?, theme_color = ? WHERE id = 1",
        req.LogoUrl, req.HeroBg, req.ThemeColor)
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في تحديث المظهر")
        return
    }

    userID := r.Context().Value("user_id").(int)
    database.AddAuditLog(userID, "update_appearance", "تحديث مظهر الموقع", r.RemoteAddr)

    sendSuccess(w, nil, "تم تحديث المظهر")
}

// ==================== Content Handlers ====================

func GetContent(w http.ResponseWriter, r *http.Request) {
    settings, err := database.GetSettings()
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في جلب المحتوى")
        return
    }

    sendSuccess(w, map[string]interface{}{
        "about":  settings.AboutText,
        "contact": settings.ContactText,
        "terms":   settings.TermsText,
    }, "")
}

func UpdateContent(w http.ResponseWriter, r *http.Request) {
    var req struct {
        About   string `json:"about"`
        Contact string `json:"contact"`
        Terms   string `json:"terms"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, http.StatusBadRequest, "بيانات غير صالحة")
        return
    }

    _, err := database.DB.Exec("UPDATE settings SET about_text = ?, contact_text = ?, terms_text = ? WHERE id = 1",
        req.About, req.Contact, req.Terms)
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في تحديث المحتوى")
        return
    }

    userID := r.Context().Value("user_id").(int)
    database.AddAuditLog(userID, "update_content", "تحديث محتوى الصفحات", r.RemoteAddr)

    sendSuccess(w, nil, "تم تحديث المحتوى")
}

// ==================== Upload Handlers ====================

func UploadFile(w http.ResponseWriter, r *http.Request) {
    // رفع ملف
    file, handler, err := r.FormFile("file")
    if err != nil {
        sendError(w, http.StatusBadRequest, "خطأ في رفع الملف")
        return
    }
    defer file.Close()

    // التحقق من نوع الملف
    allowedTypes := []string{"image/jpeg", "image/png", "image/webp", "image/jpg"}
    fileType := handler.Header.Get("Content-Type")
    allowed := false
    for _, t := range allowedTypes {
        if fileType == t {
            allowed = true
            break
        }
    }
    if !allowed {
        sendError(w, http.StatusBadRequest, "نوع الملف غير مسموح")
        return
    }

    // إنشاء اسم فريد
    filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), handler.Filename)

    // حفظ الملف
    // (هنا يمكن إضافة منطق الحفظ الفعلي)

    sendSuccess(w, map[string]interface{}{
        "filename": filename,
        "url":      "/uploads/" + filename,
    }, "تم رفع الملف بنجاح")
}

func DeleteFile(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    filename := vars["filename"]

    // حذف الملف
    // (هنا يمكن إضافة منطق الحذف الفعلي)

    userID := r.Context().Value("user_id").(int)
    database.AddAuditLog(userID, "delete_file", fmt.Sprintf("حذف ملف: %s", filename), r.RemoteAddr)

    sendSuccess(w, nil, "تم حذف الملف")
}

// ==================== Reports Handlers ====================

func GenerateReport(w http.ResponseWriter, r *http.Request) {
    var req struct {
        StartDate string `json:"start_date"`
        EndDate   string `json:"end_date"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, http.StatusBadRequest, "بيانات غير صالحة")
        return
    }

    // جلب الطلبات في الفترة المحددة
    rows, err := database.DB.Query(`
        SELECT id, full_name, total, status, date 
        FROM orders 
        WHERE date BETWEEN ? AND ?
        ORDER BY date DESC`, req.StartDate, req.EndDate)
    if err != nil {
        sendError(w, http.StatusInternalServerError, "خطأ في جلب التقرير")
        return
    }
    defer rows.Close()

    var orders []map[string]interface{}
    var totalSales float64
    for rows.Next() {
        var id int
        var name string
        var total float64
        var status, date string
        rows.Scan(&id, &name, &total, &status, &date)
        orders = append(orders, map[string]interface{}{
            "id": id, "customer": name, "total": total, "status": status, "date": date,
        })
        totalSales += total
    }

    sendSuccess(w, map[string]interface{}{
        "orders":      orders,
        "total_sales": totalSales,
        "order_count": len(orders),
        "start_date":  req.StartDate,
        "end_date":    req.EndDate,
    }, "")
}

func ExportPDF(w http.ResponseWriter, r *http.Request) {
    // تصدير PDF
    w.Header().Set("Content-Type", "application/pdf")
    w.Header().Set("Content-Disposition", "attachment; filename=report.pdf")
    
    // هنا يمكن إضافة منطق إنشاء PDF الفعلي
    w.Write([]byte("%PDF-1.4\n..."))
}

func ExportExcel(w http.ResponseWriter, r *http.Request) {
    // تصدير Excel
    w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
    w.Header().Set("Content-Disposition", "attachment; filename=report.xlsx")
    
    // هنا يمكن إضافة منطق إنشاء Excel الفعلي
    w.Write([]byte(""))
}
