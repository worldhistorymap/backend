package tileserver

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

var TILEROOT string = "/tiles/"
var regions []string = []string{"iberia", "mediaeval_middle_east", "northern_europe"}
var BELOWLOW string = "BL"
var ABOVEHIGH string = "UH"
var NOTILE string = "NT"

func TileServer() {
	c := cache.New(48*60*time.Minute, 60*time.Minute)
	for _, region := range regions {
		createRegion(c, region)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handle(c))
	log.Fatal(http.ListenAndServe(":8000", mux))
}

func createRegion(c *cache.Cache, region string) {
	dirs, err := ioutil.ReadDir(TILEROOT + region)
	fmt.Printf("Loading: %s", region)
	if err != nil {
		return
	}

	var years []int

	for _, year := range dirs {
		yearNum, err := strconv.Atoi(year.Name())
		if err == nil {
			/**Avoid things that end in BC **/
			years = append(years, yearNum)
		}
	}

	sort.Ints(years)
	c.Set(region, years, cache.NoExpiration)
	return
}

func modifiedBinarySearch(years *[]int, year int, low int, high int) (int, int, error) {
	/**get range between year that map to same tile**/
	if year < (*years)[0] {
		return 0, (*years)[low], errors.New(BELOWLOW)
	}

	if year > (*years)[len(*years)-1] {
		return (*years)[high], 0, errors.New(ABOVEHIGH)
	}

	if low == high || high == low+1 {
		return (*years)[low], (*years)[high], nil
	}
	med := (low + high) / 2
	if (*years)[med] == year {
		return (*years)[med], (*years)[med+1], nil
	}

	if (*years)[med] < year {
		return modifiedBinarySearch(years, year, med, high)
	} else {
		return modifiedBinarySearch(years, year, low, med)
	}
}

func findYear(years *[]int, year string) (string, string, error) {
	yr, err := strconv.Atoi(year)
	if err != nil {
		return "", "", err
	}

	lyr, ryr, err := modifiedBinarySearch(years, yr, 0, len(*years)-1)
	if err != nil {
		return strconv.Itoa(lyr), strconv.Itoa(ryr), err
	}
	return strconv.Itoa(lyr), strconv.Itoa(ryr), nil
}

func searchfs(c *cache.Cache, region string, year string) (string, string, error) {
	lst, found := c.Get(region)
	if !found {
		return "", "", errors.New("region lst not found")
	}
	var years []int = lst.([]int)
	lYear, rYear, err := findYear(&years, year)

	if err != nil {
		return lYear, rYear, err
	}

	return lYear, rYear, nil
}

func updateCache(c *cache.Cache, region string, year string) (string, error) {
	updatedYear := ""
	yr, err := strconv.Atoi(year)
	if err != nil {
		return "", err
	}

	if yr > 2019 || yr < 0 {
		/**So that nobody puts a small or large number and makes us a ton of unnecessary dates **/
		return "", err
	}

	lYear, rYear, err := searchfs(c, region, year)
	if err != nil {
		if err.Error() == BELOWLOW {
			lYear = "0"
			updatedYear = "NA"
		} else if err.Error() == ABOVEHIGH {
			rYear = "2019"
			updatedYear = "NA"
		} else {
			return "", err
		}
	} else {
		updatedYear = lYear
	}

	start, err := strconv.Atoi(lYear)
	if err != nil {
		log.Fatal(err)
	}
	end, err := strconv.Atoi(rYear)
	for i := start; i < end; i++ {
		year := strconv.Itoa(i)
		val := region + year
		c.Set(val, updatedYear, cache.DefaultExpiration)
	}
	return updatedYear, nil
}

func handle(c *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var updatedYear string
		var err error
		path := strings.Split(r.URL.Path, "/")
		region := path[1]
		year := path[2]
		val := region + year
		resolvedYear, found := c.Get(val)
		if !found {
			updatedYear, err = updateCache(c, region, year)
			if err != nil {
				/**Insert Log **/
				w.WriteHeader(http.StatusOK)
				return
			}
		} else {
			updatedYear = resolvedYear.(string)
		}
		if updatedYear != NOTILE {
			/** enhancement - fix to not include double /. **/
			url := TILEROOT + strings.Replace(r.URL.Path, year, updatedYear, 1)
			http.ServeFile(w, r, url)
		}
		w.WriteHeader(http.StatusOK)
		return
	}
}
