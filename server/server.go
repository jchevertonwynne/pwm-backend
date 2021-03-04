package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	jwtKey []byte
	users  = make(map[string]string)
	mu     = new(sync.Mutex)
)

func init() {
	rand.Seed(time.Now().UnixNano())
	chars := "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM"
	var builder []rune
	for i := 0; i < 64; i++ {
		builder = append(builder, rune(chars[rand.Intn(len(chars))]))
	}
	jwtKey = []byte(string(builder))
}

type claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", registerHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/check", checkHandler)

	return http.ListenAndServe(":8080", mux)
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r registerRequest) validate() bool {
	if len(r.Username) < 8 {
		return false
	}
	if len(r.Password) < 8 {
		return false
	}
	return true
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Request is not a POST", http.StatusMethodNotAllowed)
		return
	}

	var user registerRequest
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if !user.validate() {
		http.Error(w, "Username and password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	_, ok := users[user.Username]

	if ok {
		http.Error(w, "Username is taken", http.StatusBadRequest)
		return
	}
	users[user.Username] = user.Password

	expiration := time.Now().Add(5 * time.Minute)
	c := &claims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiration.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt",
		Value:   tokenString,
		Expires: expiration,
		Path:    "/",
	})
	w.WriteHeader(http.StatusOK)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Request is not a POST", http.StatusMethodNotAllowed)
		return
	}

	var user registerRequest
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	password, ok := users[user.Username]
	if !ok {
		http.Error(w, "User does not exist", http.StatusBadRequest)
		return
	}

	if user.Password != password {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	err = insertJWT(user, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "username",
		Value:   "",
		Expires: time.Now(),
		Path:    "/",
	})

	w.WriteHeader(http.StatusOK)
}

func checkHandler(w http.ResponseWriter, r *http.Request) {
	c, err := extractJWT(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "Welcome, %s", c.Username)
}

func extractJWT(r *http.Request) (claims, error) {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		return claims{}, err
	}
	contents := cookie.Value
	c := claims{}

	tkn, err := jwt.ParseWithClaims(contents, &c, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return claims{}, err
	}
	if !tkn.Valid {
		return claims{}, errors.New("JWT is not valid")
	}

	return c, nil
}

func insertJWT(user registerRequest, w http.ResponseWriter) error {
	expiration := time.Now().Add(5 * time.Minute)
	c := &claims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiration.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt",
		Value:   tokenString,
		Expires: expiration,
		Path:    "/",
	})

	return nil
}
