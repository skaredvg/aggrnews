package postgres

import (
	"context"
	"fmt"
	"log"
	"skillfactory/aggrnews/pkg/db"

	"github.com/jackc/pgx/v4"
)

// Структура - объект соединения с БД
type DBAggrNews struct {
	con *pgx.Conn
}

// Конструктор объекта соединения с БД
func NewDBAggrNews(pconn string) (*DBAggrNews, error) {
	db, err := pgx.Connect(context.Background(), pconn)
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}
	return &DBAggrNews{con: db}, err
}

// Функция регистрации масива публикаций в БД
func (dba *DBAggrNews) New(p []db.Publication) error {
	sql := `INSERT INTO public.publication(
			title, annotation, publication_time, publication_url
			) VALUES ($1, $2, $3, $4)
			ON CONFLICT (title) DO UPDATE SET title = EXCLUDED.title
			RETURNING id`
	b := new(pgx.Batch)
	for _, ent := range p {
		b.Queue(sql, ent.Title, ent.Content, ent.PubTime, ent.Link)
	}
	if b.Len() > 0 {
		br := dba.con.SendBatch(context.Background(), b)
		if _, err := br.Query(); err != nil {
			fmt.Println(err)
			return err
		}
		br.Close()
	}

	return nil
}

// Функция возвращает публикации в количестве n
func (dba *DBAggrNews) Last(n int) ([]db.Publication, error) {
	l := make([]db.Publication, 0)
	if n < 0 {
		return l, fmt.Errorf("Отрицательное число публикаций(%s)", n)
	}
	sql := `SELECT r.id, r.title, r.annotation, r.publication_time, r.publication_url FROM (
			 SELECT p.*, row_number() OVER () as rn
			 FROM postgres.public.publication p
			 ORDER BY p.publication_time DESC) r
			WHERE r.rn <= $1`
	rows, err := dba.con.Query(context.Background(), sql, n)
	if err != nil {
		return []db.Publication{}, err
	}
	for rows.Next() {
		p := db.Publication{}
		err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.PubTime, &p.Link)
		if err != nil {
			return []db.Publication{}, err
		}
		l = append(l, p)
	}
	return l, nil
}
