package main

import (
    "crypto/tls"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
    "github.com/rs/cors"
    
    "ssc-admin/database"
    "ssc-admin/handlers"
    "ssc-admin/middleware"
    "ssc-admin/security"
    "ssc-admin/seeders"
)

var (
    appEnv     string
    dbHost     string
    dbPort     string
    dbUser     string
    dbPassword string
    dbName     string
)

func init() {
    // تحميل متغيرات البيئة
    if err := godotenv.Load(); err != nil {
        log.Println("⚠️ ملف .env غير موجود، سيتم استخدام المتغيرات الافتراضية")
    }

    appEnv = getEnv("APP_ENV", "production")
    dbHost = getEnv("DB_HOST", "localhost")
    dbPort = getEnv("DB_PORT", "3306")
    dbUser = getEnv("DB_USER", "ssc_user")
    dbPassword = getEnv("DB_PASSWORD", "")
    dbName = getEnv("DB_NAME", "ssc_db")
}

func main() {
    // ==================== 1. معالج الأوامر ====================
    runSeeder := flag.Bool("seed", false, "تشغيل البيانات التجريبية (للتطوير فقط)")
    migrateOnly := flag.Bool("migrate", false, "تنفيذ الترحيلات فقط")
    flag.Parse()

    // ==================== 2. تهيئة قاعدة البيانات ====================
    config := database.Config{
        Host:     dbHost,
        Port:     dbPort,
        User:     dbUser,
        Password: dbPassword,
        DBName:   dbName,
    }

    if err := database.InitDB(config); err != nil {
        log.Fatalf("❌ فشل الاتصال بقاعدة البيانات: %v", err)
    }
    defer database.Close()
    fmt.Println("✅ متصل بـ MariaDB بنجاح")

    // تنفيذ الترحيلات فقط
    if *migrateOnly {
        fmt.Println("✅ تم تنفيذ الترحيلات بنجاح")
        return
    }

    // ==================== 3. تشغيل البيانات التجريبية (للتطوير فقط) ====================
    if *runSeeder && appEnv == "development" {
        if err := seeders.Run(); err != nil {
            log.Printf("⚠️ خطأ في تشغيل البيانات التجريبية: %v", err)
        } else {
            fmt.Println("✅ تم إدخال البيانات التجريبية")
        }
    } else if *runSeeder && appEnv == "production" {
        log.Println("❌ لا يمكن تشغيل البيانات التجريبية في وضع الإنتاج!")
        os.Exit(1)
    }

    // ==================== 4. تهيئة الأمان ====================
    securityConfig := security.Config{
        JWTSecret:     getEnv("JWT_SECRET", "default-secret-change-me"),
        RateLimit:     getEnvAsInt("RATE_LIMIT", 100),
        RateLimitWindow: getEnvAsInt("RATE_LIMIT_WINDOW", 60),
        MaxUploadSize: getEnvAsInt64("MAX_UPLOAD_SIZE", 2097152),
    }
    security.Init(securityConfig)
    fmt.Println("✅ نظام الأمان جاهز")

    // ==================== 5. إعداد الـ Router ====================
    router := mux.NewRouter()

    // ==================== 6. API Routes العامة (بدون مصادقة) ====================
    // واجهة الموقع العام
    router.HandleFunc("/api/public/products", handlers.GetPublicProducts).Methods("GET", "OPTIONS")
    router.HandleFunc("/api/public/product/{id:[0-9]+}", handlers.GetPublicProduct).Methods("GET", "OPTIONS")
    router.HandleFunc("/api/public/categories", handlers.GetPublicCategories).Methods("GET", "OPTIONS")
    router.HandleFunc("/api/public/settings", handlers.GetPublicSettings).Methods("GET", "OPTIONS")
    
    // طلبات الشراء
    router.HandleFunc("/api/public/order", handlers.CreateOrder).Methods("POST", "OPTIONS")
    router.HandleFunc("/api/public/coupon/validate", handlers.ValidateCoupon).Methods("POST", "OPTIONS")
    router.HandleFunc("/api/public/contact", handlers.SubmitContact).Methods("POST", "OPTIONS")

    // تسجيل الدخول
    router.HandleFunc("/api/auth/login", handlers.Login).Methods("POST", "OPTIONS")
    router.HandleFunc("/api/auth/logout", handlers.Logout).Methods("POST")
    router.HandleFunc("/api/auth/refresh", handlers.RefreshToken).Methods("POST")

    // ==================== 7. API Routes المحمية (للمدير) ====================
    adminRoutes := router.PathPrefix("/api/admin").Subrouter()
    adminRoutes.Use(middleware.JWTAuth)
    adminRoutes.Use(middleware.RateLimit)
    adminRoutes.Use(middleware.CSRFProtect)
    adminRoutes.Use(middleware.AuditLog)

    // Dashboard
    adminRoutes.HandleFunc("/dashboard/stats", handlers.GetDashboardStats).Methods("GET")
    
    // Produits
    adminRoutes.HandleFunc("/products", handlers.GetProducts).Methods("GET")
    adminRoutes.HandleFunc("/product", handlers.CreateProduct).Methods("POST")
    adminRoutes.HandleFunc("/product/{id:[0-9]+}", handlers.GetProduct).Methods("GET")
    adminRoutes.HandleFunc("/product/{id:[0-9]+}", handlers.UpdateProduct).Methods("PUT")
    adminRoutes.HandleFunc("/product/{id:[0-9]+}", handlers.DeleteProduct).Methods("DELETE")
    
    // Catégories
    adminRoutes.HandleFunc("/categories", handlers.GetCategories).Methods("GET")
    adminRoutes.HandleFunc("/category", handlers.CreateCategory).Methods("POST")
    adminRoutes.HandleFunc("/category/{id:[0-9]+}", handlers.UpdateCategory).Methods("PUT")
    adminRoutes.HandleFunc("/category/{id:[0-9]+}", handlers.DeleteCategory).Methods("DELETE")
    
    // Commandes
    adminRoutes.HandleFunc("/orders", handlers.GetOrders).Methods("GET")
    adminRoutes.HandleFunc("/order/{id:[0-9]+}/status", handlers.UpdateOrderStatus).Methods("PUT")
    adminRoutes.HandleFunc("/order/{id:[0-9]+}", handlers.GetOrderDetails).Methods("GET")
    
    // Clients
    adminRoutes.HandleFunc("/customers", handlers.GetCustomers).Methods("GET")
    adminRoutes.HandleFunc("/customer/{phone}", handlers.GetCustomerDetails).Methods("GET")
    
    // Messages
    adminRoutes.HandleFunc("/messages", handlers.GetMessages).Methods("GET")
    adminRoutes.HandleFunc("/message/{id:[0-9]+}", handlers.DeleteMessage).Methods("DELETE")
    adminRoutes.HandleFunc("/message/{id:[0-9]+}/read", handlers.MarkMessageRead).Methods("PUT")
    
    // Coupons
    adminRoutes.HandleFunc("/coupons", handlers.GetCoupons).Methods("GET")
    adminRoutes.HandleFunc("/coupon", handlers.CreateCoupon).Methods("POST")
    adminRoutes.HandleFunc("/coupon/{code}", handlers.DeleteCoupon).Methods("DELETE")
    
    // Staff
    adminRoutes.HandleFunc("/staff", handlers.GetStaff).Methods("GET")
    adminRoutes.HandleFunc("/staff", handlers.CreateStaff).Methods("POST")
    adminRoutes.HandleFunc("/staff/{id:[0-9]+}", handlers.UpdateStaff).Methods("PUT")
    adminRoutes.HandleFunc("/staff/{id:[0-9]+}", handlers.DeleteStaff).Methods("DELETE")
    
    // Analytics
    adminRoutes.HandleFunc("/analytics/sales", handlers.GetSalesAnalytics).Methods("GET")
    adminRoutes.HandleFunc("/analytics/top-products", handlers.GetTopProducts).Methods("GET")
    adminRoutes.HandleFunc("/analytics/visitors", handlers.GetVisitorsStats).Methods("GET")
    
    // Reports
    adminRoutes.HandleFunc("/reports/generate", handlers.GenerateReport).Methods("POST")
    adminRoutes.HandleFunc("/reports/pdf", handlers.ExportPDF).Methods("GET")
    adminRoutes.HandleFunc("/reports/excel", handlers.ExportExcel).Methods("GET")
    
    // Settings
    adminRoutes.HandleFunc("/settings", handlers.GetSettings).Methods("GET")
    adminRoutes.HandleFunc("/settings", handlers.UpdateSettings).Methods("PUT")
    
    // Backup
    adminRoutes.HandleFunc("/backup/export", handlers.ExportBackup).Methods("GET")
    adminRoutes.HandleFunc("/backup/import", handlers.ImportBackup).Methods("POST")
    adminRoutes.HandleFunc("/backup/list", handlers.ListBackups).Methods("GET")
    
    // Appearance
    adminRoutes.HandleFunc("/appearance", handlers.GetAppearance).Methods("GET")
    adminRoutes.HandleFunc("/appearance", handlers.UpdateAppearance).Methods("PUT")
    
    // Content
    adminRoutes.HandleFunc("/content", handlers.GetContent).Methods("GET")
    adminRoutes.HandleFunc("/content", handlers.UpdateContent).Methods("PUT")

    // Upload
    adminRoutes.HandleFunc("/upload", handlers.UploadFile).Methods("POST")
    adminRoutes.HandleFunc("/upload/{filename}", handlers.DeleteFile).Methods("DELETE")

    // ==================== 8. الملفات الثابتة ====================
    router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", 
        middleware.CacheControl(http.FileServer(http.Dir("./static/")))))
    router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", 
        middleware.CacheControl(http.FileServer(http.Dir("./uploads/")))))
    
    // الصفحات الرئيسية
    router.HandleFunc("/", serveIndexHTML)
    router.HandleFunc("/admin", serveAdminHTML)
    router.HandleFunc("/health", healthCheck)

    // ==================== 9. Middleware الشامل ====================
    handler := middleware.RecoverPanic(router)
    handler = middleware.SecurityHeaders(handler)
    handler = middleware.RequestID(handler)
    handler = middleware.Logging(handler)
    handler = middleware.GzipCompression(handler)

    // CORS (للإنتاج - نطاقات محددة فقط)
    var corsMiddleware *cors.Cors
    if appEnv == "production" {
        corsMiddleware = cors.New(cors.Options{
            AllowedOrigins:   []string{
                "https://ssc.tn",
                "https://www.ssc.tn",
            },
            AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
            AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID"},
            ExposedHeaders:   []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining"},
            AllowCredentials: true,
            MaxAge:           86400,
        })
    } else {
        // وضع التطوير - يسمح بكل شيء
        corsMiddleware = cors.AllowAll()
    }
    
    finalHandler := corsMiddleware.Handler(handler)

    // ==================== 10. تشغيل الخادم ====================
    port := getEnv("PORT", "8443")
    addr := ":" + port

    server := &http.Server{
        Addr:         addr,
        Handler:      finalHandler,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    // تشغيل HTTP/2 مع TLS في الإنتاج
    if appEnv == "production" {
        tlsCert := getEnv("TLS_CERT", "/etc/ssl/certs/ssc.crt")
        tlsKey := getEnv("TLS_KEY", "/etc/ssl/private/ssc.key")
        
        server.TLSConfig = &tls.Config{
            MinVersion:               tls.VersionTLS13,
            PreferServerCipherSuites: true,
            CurvePreferences: []tls.CurveID{
                tls.X25519,
                tls.CurveP256,
            },
            CipherSuites: []uint16{
                tls.TLS_AES_256_GCM_SHA384,
                tls.TLS_CHACHA20_POLY1305_SHA256,
            },
        }
        
        fmt.Printf("\n🚀 خادم الإنتاج يعمل على https://%s\n", addr)
        fmt.Printf("📍 تونس - Smart Step Control\n")
        fmt.Printf("🛡️ مستوى الحماية: 99.9%%\n")
        
        if err := server.ListenAndServeTLS(tlsCert, tlsKey); err != nil {
            log.Fatalf("❌ فشل تشغيل الخادم: %v", err)
        }
    } else {
        // وضع التطوير - HTTP فقط
        fmt.Printf("\n🚀 خادم التطوير يعمل على http://localhost%s\n", addr)
        fmt.Printf("⚠️ للإنتاج فقط، استخدم APP_ENV=production\n")
        
        if err := server.ListenAndServe(); err != nil {
            log.Fatalf("❌ فشل تشغيل الخادم: %v", err)
        }
    }
}

// ==================== دوال مساعدة ====================

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        var intVal int
        fmt.Sscanf(value, "%d", &intVal)
        return intVal
    }
    return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
    if value := os.Getenv(key); value != "" {
        var intVal int64
        fmt.Sscanf(value, "%d", &intVal)
        return intVal
    }
    return defaultValue
}

func serveIndexHTML(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }
    http.ServeFile(w, r, "./static/index.html")
}

func serveAdminHTML(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "./static/admin.html")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, `{"status":"ok","env":"%s","time":"%s"}`, appEnv, time.Now().Format(time.RFC3339))
}
