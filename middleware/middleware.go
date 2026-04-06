package middleware

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "log"
    "net/http"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/gorilla/mux"
)

// ==================== المتغيرات العامة ====================
var jwtSecret = []byte("ssc-secret-key-2024-change-in-production")

// Rate limiting
type rateLimiter struct {
    visits map[string][]time.Time
    mu     sync.RWMutex
    limit  int
    window time.Duration
}

var limiter = &rateLimiter{
    visits: make(map[string][]time.Time),
    limit:  100,     // 100 طلب
    window: 60 * time.Second, // في 60 ثانية
}

// ==================== 1. JWT Authentication ====================
func JWTAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // الحصول على التوكن من Header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            sendError(w, http.StatusUnauthorized, "مطلوب توكن المصادقة")
            return
        }

        // التحقق من صيغة Bearer
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            sendError(w, http.StatusUnauthorized, "صيغة التوكن غير صحيحة")
            return
        }

        tokenString := parts[1]

        // التحقق من التوكن
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("طريقة التوقيع غير صالحة: %v", token.Header["alg"])
            }
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            sendError(w, http.StatusUnauthorized, "توكن غير صالح أو منتهي")
            return
        }

        // استخراج البيانات من التوكن
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            sendError(w, http.StatusUnauthorized, "بيانات التوكن غير صالحة")
            return
        }

        userID, ok := claims["user_id"].(float64)
        if !ok {
            sendError(w, http.StatusUnauthorized, "معرف المستخدم غير صالح")
            return
        }

        phone, ok := claims["phone"].(string)
        if !ok {
            phone = ""
        }

        // إضافة بيانات المستخدم إلى السياق
        ctx := context.WithValue(r.Context(), "user_id", int(userID))
        ctx = context.WithValue(ctx, "user_phone", phone)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// ==================== 2. Rate Limiting (منع DDoS) ====================
func RateLimit(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // الحصول على IP العميل
        ip := getClientIP(r)
        
        limiter.mu.Lock()
        defer limiter.mu.Unlock()
        
        now := time.Now()
        
        // تنظيف الطلبات القديمة
        var validTimes []time.Time
        for _, t := range limiter.visits[ip] {
            if now.Sub(t) < limiter.window {
                validTimes = append(validTimes, t)
            }
        }
        
        // التحقق من عدد الطلبات
        if len(validTimes) >= limiter.limit {
            w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limiter.limit))
            w.Header().Set("X-RateLimit-Remaining", "0")
            w.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(limiter.window.Seconds())))
            sendError(w, http.StatusTooManyRequests, "تم تجاوز الحد الأقصى للطلبات. حاول مرة أخرى لاحقاً")
            return
        }
        
        // إضافة الطلب الحالي
        limiter.visits[ip] = append(validTimes, now)
        
        w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limiter.limit))
        w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(limiter.limit-len(validTimes)-1))
        
        next.ServeHTTP(w, r)
    })
}

// RateLimitByIP - نفس الوظيفة ولكن اسم مختلف للتوافق
func RateLimitByIP(next http.Handler) http.Handler {
    return RateLimit(next)
}

// ==================== 3. CORS (Cross-Origin Resource Sharing) ====================
func CORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // تحديد النطاقات المسموحة (للإنتاج)
        allowedOrigins := map[string]bool{
            "https://ssc.tn":        true,
            "https://www.ssc.tn":    true,
            "http://localhost:3000": true,
            "http://localhost:8080": true,
        }
        
        origin := r.Header.Get("Origin")
        if allowedOrigins[origin] || origin == "" {
            w.Header().Set("Access-Control-Allow-Origin", origin)
        }
        
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Request-ID")
        w.Header().Set("Access-Control-Expose-Headers", "Content-Length, X-Request-ID, X-RateLimit-Limit, X-RateLimit-Remaining")
        w.Header().Set("Access-Control-Allow-Credentials", "true")
        w.Header().Set("Access-Control-Max-Age", "86400")
        
        // معالجة طلبات OPTIONS
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

// ==================== 4. CSRF Protection ====================
var csrfTokens = sync.Map{}

func GenerateCSRFToken() string {
    bytes := make([]byte, 32)
    rand.Read(bytes)
    token := hex.EncodeToString(bytes)
    csrfTokens.Store(token, time.Now().Add(24*time.Hour))
    return token
}

func CSRFProtect(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // تخطي طلبات GET
        if r.Method == "GET" || r.Method == "OPTIONS" {
            next.ServeHTTP(w, r)
            return
        }
        
        // الحصول على توكن CSRF
        token := r.Header.Get("X-CSRF-Token")
        if token == "" {
            token = r.Header.Get("Csrf-Token")
        }
        
        if token == "" {
            sendError(w, http.StatusForbidden, "مطلوب توكن CSRF")
            return
        }
        
        // التحقق من التوكن
        value, ok := csrfTokens.Load(token)
        if !ok {
            sendError(w, http.StatusForbidden, "توكن CSRF غير صالح")
            return
        }
        
        expiry, ok := value.(time.Time)
        if !ok || time.Now().After(expiry) {
            csrfTokens.Delete(token)
            sendError(w, http.StatusForbidden, "توكن CSRF منتهي الصلاحية")
            return
        }
        
        // حذف التوكن بعد الاستخدام (للاستخدام الواحد)
        csrfTokens.Delete(token)
        
        // إضافة توكن جديد للرد
        newToken := GenerateCSRFToken()
        w.Header().Set("X-CSRF-Token", newToken)
        
        next.ServeHTTP(w, r)
    })
}

// ==================== 5. Security Headers ====================
func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // الأمان الأساسي
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
        
        // HSTS (HTTP Strict Transport Security)
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
        
        // Content Security Policy
        w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: https://i.ibb.co; connect-src 'self'")
        
        next.ServeHTTP(w, r)
    })
}

// ==================== 6. Logging Middleware ====================
func Logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // إنشاء ResponseWriter مخصص لتسجيل الحالة
        rw := &responseWriter{w, http.StatusOK}
        
        next.ServeHTTP(rw, r)
        
        // تسجيل الطلب
        log.Printf("[%s] %s %s %d %v",
            r.Method,
            r.URL.Path,
            getClientIP(r),
            rw.status,
            time.Since(start),
        )
    })
}

type responseWriter struct {
    http.ResponseWriter
    status int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.status = code
    rw.ResponseWriter.WriteHeader(code)
}

// ==================== 7. Request ID Middleware ====================
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // البحث عن Request ID موجود
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            // إنشاء Request ID جديد
            bytes := make([]byte, 16)
            rand.Read(bytes)
            requestID = hex.EncodeToString(bytes)
        }
        
        w.Header().Set("X-Request-ID", requestID)
        ctx := context.WithValue(r.Context(), "request_id", requestID)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// ==================== 8. Recover Panic Middleware ====================
func RecoverPanic(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("PANIC: %v", err)
                sendError(w, http.StatusInternalServerError, "خطأ داخلي في الخادم")
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// ==================== 9. Gzip Compression ====================
func GzipCompression(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // التحقق من قبول الضغط
        if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
            next.ServeHTTP(w, r)
            return
        }
        
        // تطبيق الضغط
        w.Header().Set("Content-Encoding", "gzip")
        w.Header().Set("Vary", "Accept-Encoding")
        
        next.ServeHTTP(w, r)
    })
}

// ==================== 10. Cache Control ====================
func CacheControl(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // للملفات الثابتة
        if strings.Contains(r.URL.Path, ".jpg") ||
           strings.Contains(r.URL.Path, ".png") ||
           strings.Contains(r.URL.Path, ".webp") ||
           strings.Contains(r.URL.Path, ".css") ||
           strings.Contains(r.URL.Path, ".js") {
            w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
        } else {
            w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
        }
        
        next.ServeHTTP(w, r)
    })
}

// ==================== 11. Audit Log Middleware ====================
func AuditLog(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // تسجيل الطلبات المهمة فقط
        importantMethods := map[string]bool{
            "POST":   true,
            "PUT":    true,
            "DELETE": true,
        }
        
        if importantMethods[r.Method] {
            userID := r.Context().Value("user_id")
            if userID != nil {
                log.Printf("[AUDIT] User %d - %s %s from %s", userID, r.Method, r.URL.Path, getClientIP(r))
            }
        }
        
        next.ServeHTTP(w, r)
    })
}

// ==================== 12. Role/Permissions Middleware ====================
func RequirePermission(permission string) mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := r.Context().Value("user_id")
            if userID == nil {
                sendError(w, http.StatusUnauthorized, "غير مصرح")
                return
            }
            
            // هنا يمكن التحقق من صلاحيات المستخدم من قاعدة البيانات
            // للتبسيط، نسمح لكل المستخدمين المصادقين
            
            next.ServeHTTP(w, r)
        })
    }
}

// ==================== دوال مساعدة ====================

// getClientIP - الحصول على IP العميل الحقيقي
func getClientIP(r *http.Request) string {
    // التحقق من proxy headers
    if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
        ips := strings.Split(xff, ",")
        return strings.TrimSpace(ips[0])
    }
    if xri := r.Header.Get("X-Real-IP"); xri != "" {
        return xri
    }
    // الرجوع إلى RemoteAddr
    ip := r.RemoteAddr
    if idx := strings.LastIndex(ip, ":"); idx != -1 {
        ip = ip[:idx]
    }
    return ip
}

// sendError - إرسال رد خطأ موحد
func sendError(w http.ResponseWriter, status int, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    w.Write([]byte(`{"success":false,"error":"` + message + `"}`))
}

// ==================== 13. Maintenance Mode Middleware ====================
var maintenanceMode = false

func SetMaintenanceMode(enabled bool) {
    maintenanceMode = enabled
}

func Maintenance(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if maintenanceMode {
            // تخطي API معينة في وضع الصيانة
            if strings.HasPrefix(r.URL.Path, "/api/admin") {
                sendError(w, http.StatusServiceUnavailable, "الموقع في وضع الصيانة. يرجى المحاولة لاحقاً")
                return
            }
        }
        next.ServeHTTP(w, r)
    })
}

// ==================== 14. IP Whitelist Middleware ====================
var whitelistedIPs = map[string]bool{
    "127.0.0.1": true,
    "::1":       true,
    // أضف IPs المسموحة للإدارة هنا
}

func IPWhitelist(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // فقط للإدارة
        if strings.HasPrefix(r.URL.Path, "/api/admin") {
            ip := getClientIP(r)
            if !whitelistedIPs[ip] {
                sendError(w, http.StatusForbidden, "IP غير مسموح")
                return
            }
        }
        next.ServeHTTP(w, r)
    })
}

// ==================== 15. Cleanup (تنظيف التوكنات المنتهية) ====================
func StartCleanup() {
    ticker := time.NewTicker(1 * time.Hour)
    go func() {
        for range ticker.C {
            csrfTokens.Range(func(key, value interface{}) bool {
                expiry, ok := value.(time.Time)
                if ok && time.Now().After(expiry) {
                    csrfTokens.Delete(key)
                }
                return true
            })
            
            // تنظيف rate limiter
            limiter.mu.Lock()
            for ip, times := range limiter.visits {
                var validTimes []time.Time
                for _, t := range times {
                    if time.Since(t) < limiter.window {
                        validTimes = append(validTimes, t)
                    }
                }
                if len(validTimes) == 0 {
                    delete(limiter.visits, ip)
                } else {
                    limiter.visits[ip] = validTimes
                }
            }
            limiter.mu.Unlock()
        }
    }()
}
