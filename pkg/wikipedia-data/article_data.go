package wikipediadata

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"historymap-microservices/pkg/middleware"
	"historymap-microservices/pkg/tools"
	"log"
	"net/http"
	"time"
)

var (
	host = tools.GetEnv("wikipedia_data_host", "oilspill.ocf.berkeley.edu")
	port = tools.GetEnv("wikipedia_data_post",  "5000")
	user = tools.GetEnv("wikipedia_data_user", "postgres")
	password = tools.GetEnv("wikipedia_data_password", "docker")
	dbname = tools.GetEnv("wikipedia_data_dbname" , "historymap_wikipedia")
)

var dbParams = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", host, port, user, password, dbname)


func ArticleDataServer() {
	db, err := gorm.Open("postgres", dbParams)
	defer db.Close()
	if err != nil {
		return
	}

	articles := make(chan ArticleData, 5000)
	users := make(chan UserArticleData, 5000)

	go processArticleData(db, articles)
	go processUserData(db, users)

	mux := http.NewServeMux()
	mux.HandleFunc("/wikidata", dataPipeline(articles, users))
	log.Fatal(http.ListenAndServe("localhost:8000", mux))
}

func dataPipeline (articles chan ArticleData, users chan UserArticleData) http.HandlerFunc {
	authChain := middleware.Auth(recordData(false, articles, nil))
	return authChain(recordData(true, articles, users))
}


var articleData = new(ArticleData)
func articleDatabaseCall (db * gorm.DB, data ArticleData) {
	notFound := db.Where("url = ? AND title = ?", data.Url, data.Title).First(&articleData)

	switch data.ArticleInteraction {
		case GENERATED:
		data.Generated += 1
		case HOVERED_OVER:
		data.HoveredOver += 1
		case CLICKED:
		data.Clicked += 1
		case SEARCHED:
		data.Searched += 1
	}
	data.UpdatedAt = time.Now()
	db.Save(&data)
}

func userDatabaseCall(db * gorm.DB, data UserArticleData) {
	if db.NewRecord(data) {
		data.CreatedAt = time.Now()
		data.HoveredOver = 0
		data.Clicked = 0
		data.Generated = 0
		data.Searched = 0
		db.Create(&data)
	}

	db.First(&data)

	switch data.ArticleInteraction {
	case GENERATED:
		data.Generated += 1
	case HOVERED_OVER:
		data.HoveredOver += 1
	case CLICKED:
		data.Clicked += 1
	case SEARCHED:
		data.Searched += 1
	}
	data.UpdatedAt = time.Now()
	db.Save(data)
}

func processArticleData (db *gorm.DB, in <-chan ArticleData) {
	for data :=  range in  {
		articleDatabaseCall(db, data)
	}
}

func processUserData (db *gorm.DB, in <-chan UserArticleData) {
	for data := range in {
		userDatabaseCall(db, data)
	}
}

func recordData(userAuth bool, articles chan<- ArticleData, users chan<- UserArticleData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := new(articleRequest)
		err := json.NewDecoder(r.Body).Decode(request)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		data := ArticleData {
			Url: request.Url,
			Title: request.Title,
			Lat: request.Lat,
			Lon: request.Lon,
			ArticleInteraction: request.ArticleInteraction,
		}

		articles <- data

		if userAuth {
			userData := UserArticleData{
				UserId: request.UserId,
				Url: request.Url,
				Title: request.Title,
				Lat: request.Lat,
				Lon: request.Lon,
				ArticleInteraction:request.ArticleInteraction,
			}

			users <- userData
		}
		w.WriteHeader(http.StatusOK)
	}
}

type articleRequest struct {
	Url string `json: "url"`
	Lat float64 `json: "lat"`
	Lon float64 `json: "lon"`
	Title string `json: "title"`
	ArticleInteraction int `json: "articleInteraction"`
	UserId uint `json:name`
}

type UserArticleData struct {
	gorm.Model
	UserId uint `gorm:"primary_key"`
	Url string `gorm:"primary_key"`
	Title string `gorm:"primary_key"`
	Lat float64
	Lon float64
	HoveredOver int
	Generated int
	Clicked int
	Searched int
	ArticleInteraction int `gorm:"-"`
}

type ArticleData struct {
	gorm.Model
	Url string `gorm:"primary_key"`
	Title string `gorm:"primary_key"`
	Lat float64
	Lon float64
	HoveredOver int
	Generated int
	Clicked int
	Searched int
	ArticleInteraction int `gorm:"-"`
}

