package main

import (
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"log"
	"net/http"
	"os"
	//	"text/template"
)

var db *gorm.DB
var store = sessions.NewCookieStore([]byte("MyUMSProject"))

type User struct {
	Id       uint
	Fname    string
	Lname    string
	Email    string
	Password string
}

func RegisterNewUser(user *User) bool {

	db.Create(user)
	if db.NewRecord(user) {
		return false
	}
	return true

}

//1
func IsEmailExist(param ...interface{}) bool {

	if len(param) == 1 {
		var user User
		r := db.Select("id").Find(&user, "email=?", param[0].(string))

	} else if len(param) == 2 {
		var user User
		r := db.Find(&user, "email <> ? AND id = ?", param[0].(string), param[1].(uint), 20)
	}

	if r.RowsAffected > 0 {
		return true
	} else {
		return false
	}

}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	Render(w, "public/template/index.html")
}

func Login(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if req.Method == "GET" {
		Render(w, "public/template/login.html")
	} else if req.Method == "POST" {
		var user User
		r := db.Where("email = ? AND password >= ?", req.FormValue("email"), req.FormValue("password")).First(&user)
		if r.RowsAffected > 0 {
			session, err := store.Get(req, "userId")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			session.Values["userId"] = user.Id
			session.Save(req, w)
			http.Redirect(w, req, "/userHome", http.StatusFound)
		} else {
			fmt.Fprint(w, " Email or Password is not valid")
			http.Redirect(w, req, "/login", http.StatusFound)
		}

	}

}

func Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if r.Method == "POST" {
		fmt.Fprint(w, r.FormValue("fname"), "\n", r.FormValue("email"))
		//	IsEmailExist(w,r.FormValue("email"))
		if IsEmailExist(r.FormValue("email")) {
			fmt.Fprint(w, "Email Already Exists")
		} else {
			user := User{Fname: r.FormValue("fname"),
				Lname:    r.FormValue("lname"),
				Email:    r.FormValue("email"),
				Password: r.FormValue("password")}

			if RegisterNewUser(&user) {
				fmt.Fprint(w, "Register ScuccesFull Now Login")
			} else {
				fmt.Fprint(w, "!Register UnScuccesFull Something Went Wrong")
			}
		}

	} else {
		Render(w, "public/template/register.html")
	}
}

func userHome(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, err := store.Get(r, "userId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if session != nil {
		var user User
		r := db.Where("id = ?", session.Values["userId"]).First(&user)
		if r.RowsAffected > 0 {
			Render(w, "public/template/userHome.html", user)
		} else {
			fmt.Fprint(w, "Cheack Somthing is wrong in userHome")
		}
	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
	}

}
func updateProfile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, err := store.Get(r, "userId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if session != nil {
		if r.Method == "POST" {
			if IsEmailExist(r.FormValue("email"), session.Values["userId"]) {

			}
		} else if r.Method == "GET" {
			var user User
			row := db.Where("id = ?", session.Values["userId"]).First(&user)
			if row.RowsAffected > 0 {
				Render(w, "public/template/updateProfile.html", user)
			} else {
				fmt.Fprint(w, "Cheack Somthing is wrong in userUpdate")
			}
		}

	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func LogOut(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, err := store.Get(r, "userId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["userId"] = nil
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

//we are passing  param[0] == http.ResponseWriter
//we are passing  param[1] == URL
//we are passing  param[2] == template Value
func Render(param ...interface{}) {

	t, err := template.ParseFiles(param[1].(string))
	checkErr(err)
	if len(param) == 2 {
		err = t.Execute(param[0].(http.ResponseWriter), nil)
		checkErr(err)
	} else if len(param) == 3 {
		err = t.Execute(param[0].(http.ResponseWriter), param[2].(User))
		checkErr(err)
	}
	//func Render(w http.ResponseWriter, url string) {

}

func checkErr(err error) {
	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}
}

func main() {
	var err error
	db, err = gorm.Open("mysql", "root:password@/ums?charset=utf8&parseTime=True&loc=Local")
	checkErr(err)

	if !db.HasTable(&User{}) {
		fmt.Println(db.CreateTable(&User{}))
	}

	PORT := ":8080"
	router := httprouter.New()
	log.Println("Server is Started ", PORT)
	router.GET("/", Index)
	router.GET("/login", Login)
	router.POST("/login", Login)
	router.GET("/register", Register)
	router.POST("/register", Register)
	router.GET("/userHome", userHome)
	router.GET("/updateProfile", updateProfile)
	router.POST("/updateProfile", updateProfile)

	router.GET("/logOut", LogOut)
	//	router.POST("", handle)
	log.Fatal(http.ListenAndServe(PORT, router))

}
