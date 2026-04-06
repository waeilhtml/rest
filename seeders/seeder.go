package seeders

import (
    "log"
    "time"

    "ssc-admin/database"
    "golang.org/x/crypto/bcrypt"
)

// Run - تشغيل البيانات التجريبية (للتطوير فقط)
func Run() error {
    log.Println("🌱 جاري إدخال البيانات التجريبية...")

    // التحقق من وجود بيانات بالفعل
    var count int
    database.DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
    if count > 0 {
        log.Println("📦 البيانات موجودة بالفعل، تخطي السيدر")
        return nil
    }

    // ==================== 1. إدخال الفئات ====================
    categories := []struct {
        Name string
        Slug string
    }{
        {"Caméras", "cameras"},
        {"Alarmes", "alarms"},
        {"Contrôle Accès", "access"},
        {"Stockage", "storage"},
        {"Accessoires", "accessories"},
    }

    for _, c := range categories {
        _, err := database.DB.Exec(`INSERT INTO categories (name, slug) VALUES (?, ?)`, c.Name, c.Slug)
        if err != nil {
            log.Printf("⚠️ خطأ في إضافة فئة %s: %v", c.Name, err)
        }
    }
    log.Println("✅ تم إدخال 5 فئات")

    // ==================== 2. إدخال المنتجات ====================
    products := []struct {
        Name        string
        Category    string
        Price       float64
        OldPrice    float64
        Image       string
        Description string
        Stock       string
        Features    string
        Sales       int
    }{
        {
            Name:        "Caméra HD 4MP Night Vision",
            Category:    "cameras",
            Price:       299.99,
            OldPrice:    349.99,
            Image:       "/uploads/camera1.jpg",
            Description: "Caméra extérieure 4MP avec vision nocturne infrarouge et détection de mouvement intelligente. Idéale pour la surveillance extérieure jour et nuit.",
            Stock:       "available",
            Features:    "4MP,Infrarouge,IP66,Détection IA",
            Sales:       28,
        },
        {
            Name:        "Kit Alarme Sans Fil",
            Category:    "alarms",
            Price:       499.99,
            OldPrice:    599.99,
            Image:       "/uploads/alarme1.jpg",
            Description: "Kit complet d'alarme sans fil avec 3 détecteurs, sirène 110dB et application mobile. Contrôle à distance et notifications en temps réel.",
            Stock:       "available",
            Features:    "Sans fil,Application mobile,110dB,3 détecteurs",
            Sales:       15,
        },
        {
            Name:        "Contrôleur Biométrique",
            Category:    "access",
            Price:       399.99,
            OldPrice:    449.99,
            Image:       "/uploads/biometric.jpg",
            Description: "Contrôle d'accès par empreinte digitale et carte RFID pour 1000 utilisateurs. Interface intuitive et historique des accès.",
            Stock:       "low",
            Features:    "Empreinte,RFID,1000 utilisateurs,IP54",
            Sales:       8,
        },
        {
            Name:        "DVR 8 Canaux 4K",
            Category:    "storage",
            Price:       349.99,
            OldPrice:    399.99,
            Image:       "/uploads/dvr.jpg",
            Description: "Enregistreur 8 canaux 4K avec disque dur 2TB inclus et compression H.265. Visionnage à distance sur mobile et PC.",
            Stock:       "available",
            Features:    "8 canaux,4K,2TB,H.265",
            Sales:       12,
        },
        {
            Name:        "Détecteur Mouvement PIR",
            Category:    "alarms",
            Price:       49.99,
            OldPrice:    0,
            Image:       "/uploads/detector.jpg",
            Description: "Détecteur de mouvement infrarouge passif avec portée 12m et angle 90°. Idéal pour les systèmes d'alarme.",
            Stock:       "available",
            Features:    "Infrarouge,12m,90°,Anti-masque",
            Sales:       35,
        },
        {
            Name:        "Caméra WiFi 360°",
            Category:    "cameras",
            Price:       159.99,
            OldPrice:    199.99,
            Image:       "/uploads/camera360.jpg",
            Description: "Caméra intérieure panoramique avec suivi automatique, vision nocturne et audio bidirectionnel. Contrôle via application.",
            Stock:       "available",
            Features:    "WiFi,360°,Suivi auto,Audio bidirectionnel",
            Sales:       22,
        },
        {
            Name:        "Sirène Extérieure",
            Category:    "alarms",
            Price:       79.99,
            OldPrice:    0,
            Image:       "/uploads/sirene.jpg",
            Description: "Sirène extérieure puissante 120dB avec flash LED et boîtier IP65. Résiste aux intempéries.",
            Stock:       "available",
            Features:    "120dB,LED,IP65,Anti-arrachement",
            Sales:       18,
        },
        {
            Name:        "Caméra PTZ 20X",
            Category:    "cameras",
            Price:       899.99,
            OldPrice:    1099.99,
            Image:       "/uploads/ptz.jpg",
            Description: "Caméra motorisée PTZ 20x zoom optique, autotracking et WDR. Idéale pour la surveillance de grandes surfaces.",
            Stock:       "available",
            Features:    "PTZ,20x zoom,Autotracking,WDR",
            Sales:       5,
        },
        {
            Name:        "Lecteur RFID",
            Category:    "access",
            Price:       129.99,
            OldPrice:    0,
            Image:       "/uploads/rfid.jpg",
            Description: "Lecteur de badges RFID pour contrôle d'accès, compatible Wiegand. Installation facile.",
            Stock:       "available",
            Features:    "RFID,Wiegand,IP65,125kHz/13.56MHz",
            Sales:       10,
        },
        {
            Name:        "Batterie Alarme",
            Category:    "accessories",
            Price:       29.99,
            OldPrice:    0,
            Image:       "/uploads/battery.jpg",
            Description: "Batterie de secours pour système d'alarme, autonomie 24h. Sans entretien.",
            Stock:       "low",
            Features:    "12V,7Ah,Sans entretien",
            Sales:       45,
        },
    }

    for _, p := range products {
        _, err := database.DB.Exec(`
            INSERT INTO products (name, category, price, old_price, image, description, stock, features, sales) 
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
            p.Name, p.Category, p.Price, p.OldPrice, p.Image, p.Description, p.Stock, p.Features, p.Sales)
        if err != nil {
            log.Printf("⚠️ خطأ في إضافة منتج %s: %v", p.Name, err)
        }
    }
    log.Println("✅ تم إدخال 10 منتجات")

    // ==================== 3. إدخال المستخدمين ====================
    // admin123 (مشفرة)
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
    
    users := []struct {
        Name        string
        Phone       string
        Password    string
        Permissions string
    }{
        {
            Name:        "Administrateur",
            Phone:       "admin",
            Password:    string(hashedPassword),
            Permissions: "products,categories,orders,customers,messages,coupons,staff,reports,content,appearance,backup,settings",
        },
        {
            Name:        "Ahmed Ben Ali",
            Phone:       "20123456",
            Password:    string(hashedPassword),
            Permissions: "products,orders,customers,messages",
        },
        {
            Name:        "Sami Trabelsi",
            Phone:       "22123456",
            Password:    string(hashedPassword),
            Permissions: "products,reports",
        },
    }

    for _, u := range users {
        _, err := database.DB.Exec(`
            INSERT INTO users (name, phone, password, permissions) VALUES (?, ?, ?, ?)`,
            u.Name, u.Phone, u.Password, u.Permissions)
        if err != nil {
            log.Printf("⚠️ خطأ في إضافة مستخدم %s: %v", u.Name, err)
        }
    }
    log.Println("✅ تم إدخال 3 مستخدمين (admin/admin123)")

    // ==================== 4. إدخال الطلبات ====================
    orders := []struct {
        FullName string
        Phone    string
        Email    string
        Address  string
        Total    float64
        Status   string
        Date     string
        Items    string
    }{
        {
            FullName: "Jean Dupont",
            Phone:    "0612345678",
            Email:    "jean.dupont@email.com",
            Address:  "12 rue de Paris, 75001 Paris, France",
            Total:    299.99,
            Status:   "delivered",
            Date:     "2024-01-15",
            Items:    `[{"name":"Caméra HD 4MP Night Vision","price":299.99,"quantity":1}]`,
        },
        {
            FullName: "Marie Martin",
            Phone:    "0623456789",
            Email:    "marie.martin@email.com",
            Address:  "45 avenue des Lys, 69002 Lyon, France",
            Total:    499.99,
            Status:   "confirmed",
            Date:     "2024-01-20",
            Items:    `[{"name":"Kit Alarme Sans Fil","price":499.99,"quantity":1}]`,
        },
        {
            FullName: "Pierre Bernard",
            Phone:    "0634567890",
            Email:    "pierre.bernard@email.com",
            Address:  "8 rue des Oliviers, 13001 Marseille, France",
            Total:    399.99,
            Status:   "processing",
            Date:     "2024-01-25",
            Items:    `[{"name":"Contrôleur Biométrique","price":399.99,"quantity":1}]`,
        },
        {
            FullName: "Sophie Dubois",
            Phone:    "0645678901",
            Email:    "sophie.dubois@email.com",
            Address:  "23 boulevard de la Mer, 33000 Bordeaux, France",
            Total:    549.98,
            Status:   "delivered",
            Date:     "2024-02-01",
            Items:    `[{"name":"DVR 8 Canaux 4K","price":349.99,"quantity":1},{"name":"Caméra WiFi 360°","price":199.99,"quantity":1}]`,
        },
        {
            FullName: "Thomas Leroy",
            Phone:    "0656789012",
            Email:    "thomas.leroy@email.com",
            Address:  "67 rue Centrale, 59000 Lille, France",
            Total:    129.98,
            Status:   "confirmed",
            Date:     "2024-02-05",
            Items:    `[{"name":"Détecteur Mouvement PIR","price":49.99,"quantity":2},{"name":"Sirène Extérieure","price":79.99,"quantity":1}]`,
        },
        {
            FullName: "Julie Moreau",
            Phone:    "0667890123",
            Email:    "julie.moreau@email.com",
            Address:  "15 place du Marché, 31000 Toulouse, France",
            Total:    799.98,
            Status:   "processing",
            Date:     "2024-02-08",
            Items:    `[{"name":"Kit Alarme Sans Fil","price":499.99,"quantity":1},{"name":"Caméra HD 4MP Night Vision","price":299.99,"quantity":1}]`,
        },
        {
            FullName: "Nicolas Petit",
            Phone:    "0678901234",
            Email:    "nicolas.petit@email.com",
            Address:  "34 rue des Écoles, 44000 Nantes, France",
            Total:    449.98,
            Status:   "delivered",
            Date:     "2024-02-10",
            Items:    `[{"name":"Caméra WiFi 360°","price":199.99,"quantity":2},{"name":"Détecteur Mouvement PIR","price":49.99,"quantity":1}]`,
        },
        {
            FullName: "Mohamed Ali",
            Phone:    "20123456",
            Email:    "med.ali@email.tn",
            Address:  "Centre Urbain Nord, Tunis, Tunisie",
            Total:    299.99,
            Status:   "delivered",
            Date:     "2024-02-15",
            Items:    `[{"name":"Caméra HD 4MP Night Vision","price":299.99,"quantity":1}]`,
        },
        {
            FullName: "Yasmine Ben Salah",
            Phone:    "22123456",
            Email:    "yasmine.bensalah@email.tn",
            Address:  "Les Berges du Lac, Tunis, Tunisie",
            Total:    899.99,
            Status:   "confirmed",
            Date:     "2024-02-18",
            Items:    `[{"name":"Caméra PTZ 20X","price":899.99,"quantity":1}]`,
        },
        {
            FullName: "Karim Jaziri",
            Phone:    "23123456",
            Email:    "karim.jaziri@email.tn",
            Address:  "Sousse, Tunisie",
            Total:    479.98,
            Status:   "processing",
            Date:     "2024-02-20",
            Items:    `[{"name":"Contrôleur Biométrique","price":399.99,"quantity":1},{"name":"Détecteur Mouvement PIR","price":79.99,"quantity":1}]`,
        },
    }

    for _, o := range orders {
        _, err := database.DB.Exec(`
            INSERT INTO orders (full_name, phone, email, address, total, status, date, items) 
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
            o.FullName, o.Phone, o.Email, o.Address, o.Total, o.Status, o.Date, o.Items)
        if err != nil {
            log.Printf("⚠️ خطأ في إضافة طلب لـ %s: %v", o.FullName, err)
        }
    }
    log.Println("✅ تم إدخال 10 طلبات")

    // ==================== 5. إدخال الكوبونات ====================
    coupons := []struct {
        Code     string
        Discount float64
        Type     string
        MaxUses  int
        Uses     int
        Expires  string
    }{
        {"WELCOME10", 10, "percent", 50, 12, "2024-12-31"},
        {"SSC20", 20, "percent", 100, 5, "2024-06-30"},
        {"SUMMER50", 50, "fixed", 30, 8, "2024-08-31"},
        {"TUNISIA25", 25, "percent", 200, 0, "2024-12-31"},
        {"FREESHIP", 15, "fixed", 500, 127, "2024-12-31"},
    }

    for _, c := range coupons {
        _, err := database.DB.Exec(`
            INSERT INTO coupons (code, discount, type, max_uses, uses, expires) 
            VALUES (?, ?, ?, ?, ?, ?)`,
            c.Code, c.Discount, c.Type, c.MaxUses, c.Uses, c.Expires)
        if err != nil {
            log.Printf("⚠️ خطأ في إضافة كوبون %s: %v", c.Code, err)
        }
    }
    log.Println("✅ تم إدخال 5 كوبونات")

    // ==================== 6. إدخال الرسائل ====================
    messages := []struct {
        Name    string
        Email   string
        Message string
        Date    string
        IsRead  int
    }{
        {
            Name:    "Jean Dupont",
            Email:   "jean.dupont@email.com",
            Message: "Bonjour, je souhaite un devis pour 5 caméras extérieures avec installation. Pouvez-vous me contacter ?",
            Date:    "2024-02-10",
            IsRead:  0,
        },
        {
            Name:    "Marie Martin",
            Email:   "marie.martin@email.com",
            Message: "Quel est le délai de livraison pour le kit alarme ? J'ai besoin de le recevoir avant le 15 mars.",
            Date:    "2024-02-12",
            IsRead:  0,
        },
        {
            Name:    "Sophie Dubois",
            Email:   "sophie.dubois@email.com",
            Message: "Est-ce que les produits sont garantis ? Combien de temps ?",
            Date:    "2024-02-14",
            IsRead:  1,
        },
        {
            Name:    "Mohamed Ali",
            Email:   "med.ali@email.tn",
            Message: "Est-ce que vous livrez à Sfax ? J'ai besoin de 2 caméras.",
            Date:    "2024-02-15",
            IsRead:  0,
        },
        {
            Name:    "Thomas Leroy",
            Email:   "thomas.leroy@email.com",
            Message: "Problème avec ma commande #5, le détecteur ne fonctionne pas correctement. J'aimerais un remplacement.",
            Date:    "2024-02-16",
            IsRead:  0,
        },
        {
            Name:    "Yasmine Ben Salah",
            Email:   "yasmine.bensalah@email.tn",
            Message: "Merci pour la livraison rapide ! Les produits sont conformes. Je recommande vivement.",
            Date:    "2024-02-18",
            IsRead:  1,
        },
    }

    for _, m := range messages {
        _, err := database.DB.Exec(`
            INSERT INTO messages (name, email, message, date, is_read) 
            VALUES (?, ?, ?, ?, ?)`,
            m.Name, m.Email, m.Message, m.Date, m.IsRead)
        if err != nil {
            log.Printf("⚠️ خطأ في إضافة رسالة من %s: %v", m.Name, err)
        }
    }
    log.Println("✅ تم إدخال 6 رسائل")

    // ==================== 7. إدخال الإعدادات ====================
    _, err := database.DB.Exec(`
        INSERT INTO settings (id, site_name, site_phone, site_email, site_address, 
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
        "Smart Step Control est votre partenaire de confiance pour les solutions de sécurité professionnelle en Tunisie. Nous proposons une large gamme de produits : caméras de surveillance, alarmes, contrôle d'accès et bien plus encore. Notre engagement : qualité, fiabilité et service après-vente réactif. Avec plus de 5000 clients satisfaits, nous sommes le leader en Tunisie dans le domaine de la sécurité électronique.",
        "Vous pouvez nous contacter :\n\n📞 Téléphone : +216 70 000 000\n📧 Email : contact@ssc.tn\n📍 Adresse : Centre Urbain Nord, Tunis, Tunisie\n💬 WhatsApp : +216 20 000 000\n\nNos horaires :\nLundi-Vendredi : 9h-18h\nSamedi : 9h-13h\nDimanche : Fermé",
        "Conditions générales de vente :\n\n1. Livraison sous 48h en Tunisie\n2. Garantie 2 ans pièces et main d'œuvre\n3. Paiement sécurisé (CB, PayPal, virement bancaire)\n4. Droit de rétractation 14 jours\n5. Service client disponible 7j/7\n6. Installation possible par nos techniciens\n7. Support technique gratuit pendant 1 an")
    
    if err != nil {
        log.Printf("⚠️ خطأ في إضافة الإعدادات: %v", err)
    } else {
        log.Println("✅ تم إدخال إعدادات الموقع")
    }

    log.Println("\n🎉 تم إدخال جميع البيانات التجريبية بنجاح!")
    log.Println("📊 ملخص البيانات:")
    log.Println("   - 5 فئات")
    log.Println("   - 10 منتجات")
    log.Println("   - 3 مستخدمين (admin/admin123)")
    log.Println("   - 10 طلبات")
    log.Println("   - 5 كوبونات")
    log.Println("   - 6 رسائل")
    log.Println("   - إعدادات الموقع")

    return nil
}
