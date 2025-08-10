package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	connstr := "user=postgres dbname=registapp password=ccc123 sslmode=disable"
	db, err := sql.Open("postgres", connstr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic("database e baglanamadi: " + err.Error())
	}
	fmt.Println("Database'e bağlanıldı.")

	// ملفات الواجهة
	http.Handle("/register.css", http.FileServer(http.Dir(".")))

	// عرض صفحات
	http.HandleFunc("/", serveRegisterPage)           // الصفحة الرئيسية (التحقق من الكوكي)
	http.HandleFunc("/login", serveLoginPage)         // عرض صفحة تسجيل الدخول
	http.HandleFunc("/register", handleRegister(db))  // معالجة التسجيل
	http.HandleFunc("/logout", handleLogout)
	http.Handle("/login.css", http.FileServer(http.Dir(".")))
	http.HandleFunc("/signup", serveSignupPage)
    http.HandleFunc("/signup-submit", handleSignup(db))




	// تسجيل الدخول (POST فقط)
	http.HandleFunc("/login-submit", handleLogin(db))

	fmt.Println("Server çalışıyor: http://localhost:8081")
	http.ListenAndServe(":8081", nil)
}
func serveLoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "login.html")
		return
	}
	http.Error(w, "Yöntem geçersiz", http.StatusMethodNotAllowed)
}

func serveRegisterPage(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("auth")
	if err != nil || cookie.Value != "ok" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.ServeFile(w, r, "register.html")
}

func handleLogin(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Yöntem geçersiz", http.StatusMethodNotAllowed)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		var storedPassword string
		err := db.QueryRow("SELECT password FROM users WHERE username = $1", username).Scan(&storedPassword)
		if err != nil || password != storedPassword {
			http.Error(w, "Kullanıcı adı veya şifre hatalı", http.StatusUnauthorized)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "auth",
			Value:    "ok",
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(5 * time.Minute),
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
func handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0), // انتهت الآن
		MaxAge:   -1,              // حذف فوري
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
func serveSignupPage(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        http.ServeFile(w, r, "signup.html")
        return
    }
    http.Error(w, "Yöntem geçersiz", http.StatusMethodNotAllowed)
}

func handleSignup(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Yöntem geçersiz", http.StatusMethodNotAllowed)
            return
        }

        username := r.FormValue("username")
        password := r.FormValue("password")

        // تحقق إذا الاسم موجود مسبقاً
        var exists bool
        err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)", username).Scan(&exists)
        if err != nil {
            http.Error(w, "Veritabanı hatası", http.StatusInternalServerError)
            return
        }
        if exists {
            http.Error(w, "Kullanıcı adı zaten var", http.StatusBadRequest)
            return
        }

        // إضافة للمستخدمين (مبدئياً بدون Hash)
        _, err = db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", username, password)
        if err != nil {
            http.Error(w, "Kayıt başarısız", http.StatusInternalServerError)
            return
        }

        http.Redirect(w, r, "/login", http.StatusSeeOther)
    }
}




func handleRegister(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Yöntem geçersiz", http.StatusMethodNotAllowed)
			return
		}

		cookie, err := r.Cookie("auth")
		if err != nil || cookie.Value != "ok" {
			http.Error(w, "Önce giriş yapmalısınız", http.StatusUnauthorized)
			return
		}

		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		phone := r.FormValue("phone")
		email := r.FormValue("email")
		nationality := r.FormValue("nationality")
		motivation := r.FormValue("motivation")

		if firstName == "" || lastName == "" {
			http.Error(w, "Ad ve soyad zorunludur", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(`
			INSERT INTO register (first_name, last_name, phone, email, nationality, motivation, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW())
		`, firstName, lastName, phone, email, nationality, motivation)

		if err != nil {
			http.Error(w, "Kayıt başarısız: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "Kayıt başarıyla tamamlandı!")
	}
}
