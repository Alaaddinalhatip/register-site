package main

import (
	"fmt"
	"net/http"
	"database/sql"
	_ "github.com/lib/pq"
)

const secretToken = "mySecret1234"

func main() {
	connstr := "user=postgres dbname=registapp password=ccc123 sslmode=disable"
	db, err := sql.Open("postgres", connstr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic("database e baglanamadi" + err.Error())
	}
	fmt.Println("database baglandi aferin")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "register.html")
	})
	http.Handle("/register.css", http.FileServer(http.Dir(".")))

	http.HandleFunc("/register", handleRegister(db))

		fmt.Println("server is runing on http://localhost:8081")
	http.ListenAndServe(":8081", nil)
}

func handleRegister(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		token := r.FormValue("token")
		if token != secretToken {
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		phone := r.FormValue("phone")
		email := r.FormValue("email")
		nationality := r.FormValue("nationality")
		motivation := r.FormValue("motivation")

		_, err := db.Exec(`
			INSERT INTO register (first_name, last_name, phone, email, nationality, motivation, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW())
		`, firstName, lastName, phone, email, nationality, motivation)

		if err != nil {
			http.Error(w, "bilgiler girilemedi:"+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "tamamen kayıtlandı!")
	}
}
