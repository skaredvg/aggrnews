package db

//Структура объекта-публикации
type Publication struct {
	ID      int
	Title   string
	Content string
	PubTime int64
	Link    string
}

type DBInterface interface {
	// добавить публикацию
	New(p []Publication) error
	// вернуть n последних публикаций
	Last(n int) ([]Publication, error)
}
