package main

import (
	"skillfactory/aggrnews/pkg/api"
	"skillfactory/aggrnews/pkg/db"
	"strings"
	"sync"

	//"aggrnews/pkg/db/memdb"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"skillfactory/aggrnews/pkg/db/postgres"
	"time"
)

// Структура хранения конфигурации
type config struct {
	TimePeriodScan uint
	Sites          []struct {
		Url string
	}
	DBUser     string
	DBPassword string
}

// Загрузка конфигурации
func load_conf(fn string) config {
	f, err := os.OpenFile(fn, os.O_RDONLY, 0111)
	if err != nil {
		log.Fatalf("Не найден файл конфигурации%s", fn)
	}
	b, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("Ошибка чтения конфигурации %s", fn)
	}

	cfg := config{DBUser: "postgres", DBPassword: "06041972"}
	if json.Unmarshal(b, &cfg) != nil {
		log.Fatalf("Ошибка чтения конфигурации %s", fn)
	}

	//fmt.Printf("%v", cfg)
	return cfg
}

// Обработка RSS-ссылки
func processRSSLinksChan(l string, tps uint, chdb chan<- []api.RSSPublication, chlog chan<- error, w *sync.WaitGroup) {
	defer w.Done()
	for {
		rnf := api.NewRSSNewsFeed(l)
		err := rnf.ProcessLink()
		if err != nil {
			chlog <- err
		}
		chdb <- rnf.Channel.Publications
		<-time.After(time.Duration(tps) * time.Second)
	}
}

// Сохранение постов в БД
func processRSSPostToDatabase(dbc db.DBInterface, chdb <-chan []api.RSSPublication, chlog chan<- error, w *sync.WaitGroup) {
	defer w.Done()
	for v := range chdb {
		dbp := []db.Publication{}
		for _, ent := range v {
			p := db.Publication{
				Title:   ent.Title,
				Content: ent.Description,
				Link:    ent.Link,
			}

			t := time.Now()

			if strings.Contains(ent.PubTime, "GMT") {
				t, _ = time.Parse(time.RFC1123, ent.PubTime)
			} else {
				t, _ = time.Parse(time.RFC1123Z, ent.PubTime)
			}
			p.PubTime = t.UnixMilli()
			dbp = append(dbp, p)
		}
		if len(dbp) == 0 {
			continue
		}
		err := dbc.New(dbp)
		if err != nil {
			chlog <- err
		}
	}
}

// Вывод ошибок в консоль
func processLog(cherr <-chan error) {
	for err := range cherr {
		log.Println(err.Error())
	}
}

func main() {
	cfg := load_conf("config.json")
	connstr := fmt.Sprintf("postgres://%s:%s@%s/postgres", cfg.DBUser, cfg.DBPassword, "localhost")
	dbc, _ := postgres.NewDBAggrNews(connstr)
	//dbc, _ := memdb.NewDBAggrNews("")
	chdb := make(chan []api.RSSPublication)
	cherr := make(chan error)
	w := new(sync.WaitGroup)
	go processLog(cherr)
	for _, v := range cfg.Sites {
		w.Add(1)
		go processRSSLinksChan(v.Url, cfg.TimePeriodScan, chdb, cherr, w)
	}
	w.Add(1)
	go processRSSPostToDatabase(dbc, chdb, cherr, w)
	api := api.NewAPIAggrNews(dbc, cherr)
	w.Add(1)
	err := http.ListenAndServe("localhost:80", api.Router())
	if err != nil {
		cherr <- err
	}
	w.Done()
	w.Wait()
	close(chdb)
	close(cherr)
}
