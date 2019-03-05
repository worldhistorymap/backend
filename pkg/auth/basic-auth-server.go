package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/worldhistorymap/backend/pkg/tools"
	"golang.org/x/crypto/scrypt"
)

var (
	host     = tools.GetEnv("users_host", "localhost")
	port     = tools.GetEnv("users_post", "5432")
	user     = tools.GetEnv("users_user", "postgres")
	password = tools.GetEnv("users_password", "postgres")
	dbname   = tools.GetEnv("users_dbname", "user_accounts")
)

var dbParams = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", host, port, user, password, dbname)

type UserAccount struct {
	gorm.Model
	Username     string `gorm:"size:255"`
	PasswordHash []byte `gorm:"type:text"`
	PasswordSalt []byte `gorm:"type:text"`
	ID           uint   `gorm:"AUTO_INCREMENT"`
	Email        string `gorm:"size:255"`
	Joined       time.Time
}

type NewUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type userLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func AuthServer() {
	db, err := gorm.Open("postgres", dbParams)
	defer db.Close()
	if err != nil {
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/login/", login(db))
	mux.HandleFunc("/signup/", signup(db))
	log.Fatal(http.ListenAndServe("localhost:8000", mux))
}

func signup(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newUser := new(NewUser)
		if err := json.NewDecoder(r.Body).Decode(newUser); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if newUser.Username == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if len(newUser.Password) < 10 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !stringsUnalike(newUser.Username, newUser.Password) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		salt := createSalt(260)
		passwordHash, err := scrypt.Key([]byte(newUser.Password), salt, tools.ScryptN, tools.ScryptR, tools.ScryptP, tools.ScryptKeyLen)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		user := UserAccount{
			Username:     newUser.Username,
			PasswordHash: passwordHash,
			PasswordSalt: salt,
			Email:        newUser.Email,
			Joined:       time.Now(),
		}

		err = db.Create(&user).Error

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

type JwtToken struct {
	Token string `json:"jwtToken"`
}

func login(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		loginReq := new(userLogin)
		if err := json.NewDecoder(r.Body).Decode(loginReq); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userInfo := new(UserAccount)

		db.Where("username = ?", loginReq.Username).First(&userInfo)
		passwordHash, err := scrypt.Key([]byte(loginReq.Password), []byte(userInfo.PasswordSalt), tools.ScryptN, tools.ScryptR, tools.ScryptP, tools.ScryptKeyLen)

		if string(userInfo.PasswordHash) != string(passwordHash) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["name"] = user
		claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
		t, err := token.SignedString(tools.JwtSecretKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		jwtToken := JwtToken{
			Token: t,
		}

		err = json.NewEncoder(w).Encode(jwtToken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func stringsUnalike(a, b string) bool {
	return true
}

func createSalt(saltLength int) []byte {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	salt := make([]byte, saltLength)
	for i := range salt {
		salt[i] = byte(rune(r1.Intn(128)))
	}
	return salt
}
