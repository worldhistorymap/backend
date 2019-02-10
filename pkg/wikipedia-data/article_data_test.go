package wikipediadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/ory/dockertest"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")

	if err != nil {
		log.Fatalf("Connect to Docker fail: %s", err)
	}

	resource, err :=  pool.Run("localhost:5000/postgres-wikipedia-data", "latest", []string{})

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

func TestBasicArticleDataAddNoUserNoConccurency(t *testing.T) {
	url := "https://en.wikipedia.org/wiki/Alhambra"
	title := "Alhambra"
	lat := 37.17695
	lon := -3.59001
	article := map[string] interface {} {
		"url": url,
		"title": title,
		"lat": lat,
		"lon": lon,
		"articleInteraction": GENERATED,
	}

	jsString, _ := json.Marshal(article)
	req, err := http.NewRequest("POST", "/wikidata", bytes.NewBuffer(jsString))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatal(err)
	}

	articles := make(chan ArticleData, 5000)
	users := make(chan UserArticleData, 5000)

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(dataPipeline(articles, users))
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("incorrect status code: recieved: %d, expected: %d",
			status, http.StatusOK)
	}

	article = map[string] interface {} {
		"url": url,
		"title": title,
		"lat": lat,
		"lon": lon,
		"articleInteraction": HOVERED_OVER,
	}

	jsString, _ = json.Marshal(article)
	req, err = http.NewRequest("POST", "/wikidata", bytes.NewBuffer(jsString))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatal(err)
	}

	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("incorrect status code: recieved: %d, expected: %d",
			status, http.StatusOK)
	}

	/* Running calls that would ordinarily be called as goroutines */
	processArticleData(db, articles)
	processUserData(db, users)

	articleData := new(ArticleData)
	db.Where("url = ?", url).First(&articleData)

	if articleData.Title != title {
		t.Errorf("Article title incorrect")
	}

	if articleData.Lat != lat {
		t.Errorf("Article Lat incorrect: Expected %f, Recieved: %f", lat, articleData.Lat)
	}

	if articleData.Lon != lon {
		t.Errorf("Article Lon incorrect: Expected %f, Recieved: %f", lon, articleData.Lon)
	}

	if articleData.HoveredOver != 1 {
		t.Errorf("Article Data Hovered Over Incorrect: Expected : 1, Received %d", articleData.HoveredOver)
	}

	if articleData.Generated != 1 {
		t.Errorf("Article Data Hovered Over Incorrect: Expected : 1, Received %d", articleData.Generated)
	}

	if articleData.Clicked != 0 {
		t.Errorf("Article Data Clicked Incorrect: Expected : 0, Received %d", articleData.Clicked)
	}
}

