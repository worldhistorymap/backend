package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"github.com/ory/dockertest"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Connect to Docker fail: %s", err)
	}

	resource, err :=  pool.Run("localhost:5000/postgres-auth", "latest", []string{})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err = pool.Retry(func () error {
		hostPort := resource.GetPort("5432/tcp")
		dbParams := fmt.Sprintf("host=localhost port=%s user=%s sslmode=disable", hostPort, user)
		db, err = gorm.Open("postgres", dbParams)
		if err != nil {
			log.Fatalf("Could not connect to database: %s", err)
		}
		return db.DB().Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	if err = pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}


func TestBasicAccountCreation(t *testing.T) {
	username := "test"
	password := "12345678910"
	email := "test@historymap.io"
	account := map[string]string{"username": username, "password": password, "email": email}
	jsString, _ := json.Marshal(account)
	req, err := http.NewRequest("POST", "/signup", bytes.NewBuffer(jsString))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(signup(db))
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("incorrect status code: recieved:%d, expected: %d", status, http.StatusOK)
	}

	expected := ""
	if recorder.Body.String() != expected {
		t.Errorf("unexpected body: recieved: %s expected: %s", recorder.Body.String(), expected)
	}


	req, err = http.NewRequest("POST", "/login", bytes.NewBuffer(jsString))
	req.Header.Set("Content-Type", "application/json")

	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(login(db))
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("incorrect status code: recieved:%d, expected: %d", status, http.StatusOK)
	}

	token := new(JwtToken)

	err = json.Unmarshal(recorder.Body.Bytes(), &token)
	/*
	if err != nil {
		log.Fatal("Json Unmarshal Error: %s ", err)
	} */

	if token.Token == "" {
		t.Errorf("Jwt Token is empty")
	}
}
