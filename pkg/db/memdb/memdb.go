package memdb

import (
	"fmt"
	"skillfactory/aggrnews/pkg/db"
)

// Структура - объект БД
type DBAggrNews struct {
	id  int
	con []db.Publication
	uq  map[string]int // для устранения дублирования публикаций
}

// Конструктор объекта БД
func NewDBAggrNews(pconn string) (*DBAggrNews, error) {
	return &DBAggrNews{uq: map[string]int{}, con: []db.Publication{}, id: 1}, nil
}

// Функция регистрации масива публикаций в БД
func (dba *DBAggrNews) New(p []db.Publication) error {
	for _, ent := range p {
		if _, ok := dba.uq[ent.Title]; !ok {
			ent.ID = dba.id
			dba.con = append(dba.con, ent)
			dba.uq[ent.Title] = ent.ID
			dba.id++
		}
	}

	return nil
}

// Функция возвращает публикации в количестве n
func (dba *DBAggrNews) Last(n int) ([]db.Publication, error) {
	l := make([]db.Publication, 0)
	if n < 0 {
		return l, fmt.Errorf("Отрицательное число публикаций (%s)", n)
	}

	if len(dba.con) <= n {
		l = append(l, dba.con...)
	} else {
		for i := len(dba.con); i > len(dba.con)-n; i-- {
			l = append(l, dba.con[i-1])
		}
	}
	return l, nil

}
