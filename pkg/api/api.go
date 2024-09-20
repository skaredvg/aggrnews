package api

import (
	"encoding/json"
	"net/http"
	"skillfactory/aggrnews/pkg/db"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// Объект-обработчик запросов
type APIAggrNews struct {
	dbcon  db.DBInterface
	router *mux.Router
	cherr  chan<- error
}

// Объект-публикация для API
type Post struct {
	ID      int
	Title   string
	Content string
	PubTime int64
	Link    string
}

// Конструктор обработчика запросов
func NewAPIAggrNews(dbc db.DBInterface, cherr chan<- error) *APIAggrNews {
	aan := APIAggrNews{}
	aan.cherr = cherr
	aan.dbcon = dbc
	aan.router = mux.NewRouter()
	aan.endpoints()
	return &aan
}

// Функция-маршрутизатор запросов к API
func (aan *APIAggrNews) Router() *mux.Router {
	return aan.router
}

// Функция-метод API - вернуть публикации в количестве заданном параметром запроса
func (aan *APIAggrNews) posts(w http.ResponseWriter, r *http.Request) {
	url1 := r.URL.String()
	s := strings.Split(url1, "/")
	sn, err := strconv.Atoi(s[len(s)-1])
	if err != nil {
		sn = 10
	}
	publ, err := aan.dbcon.Last(sn)
	if err != nil {
		aan.cherr <- err
	}

	posts := make([]Post, 0)
	for _, v := range publ {
		posts = append(posts, Post{
			ID:      v.ID,
			Title:   v.Title,
			Content: v.Content,
			PubTime: v.PubTime,
			Link:    v.Link,
		})
	}
	b, _ := json.Marshal(posts)
	w.Header().Add("content-type", "application/json")
	w.Write(b)
}

// Привязка обработчиков REST-запросов
func (api *APIAggrNews) endpoints() {
	// получить n последних новостей
	api.router.HandleFunc("/news/{n}", api.posts).Methods(http.MethodGet, http.MethodOptions)
	// веб-приложение
	api.router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./webapp"))))
}
