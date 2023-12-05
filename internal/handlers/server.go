package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

func PostMetrics(w http.ResponseWriter, r *http.Request) {
	// 1. Принимать метрики по протоколу HTTP методом `POST`.
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//  2. Парсим маршрут и получаем мапку: тип, метрика, значение
	metric, err := parseURL(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 3. При попытке передать запрос без имени метрики возвращать `http.StatusNotFound`.
	// Пример: /update/counter//123
	if metric["name"] == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 4. Проверяем тип метрики и в случае корректных данных производим запись
	switch typeMetric(metric["value"]).(type) {
	case int64:
		// При попытке передать запрос с некорректным типом метрики или значением возвращать `http.StatusBadRequest`.
		// Пример: /update/gauge/speedAverage/200
		if metric["type"] == "gauge" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// TODO: 5. Запись значение метрики counter
		//log.Println(v)
	case float64:
		// При попытке передать запрос с некорректным типом метрики или значением возвращать `http.StatusBadRequest`.
		// Пример: /update/counter/allocCount/20.0003
		if metric["type"] == "counter" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// TODO: 6. Запись значения метрики gauge
		//log.Println("float64", v)
	default:
		// При попытке передать запрос с некорректным типом метрики или значением возвращать `http.StatusBadRequest`.
		// Пример: /update/counter/allocCount/text
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func parseURL(r *http.Request) (map[string]string, error) {
	// Реализовать парсинг маршрутов
	// Пример: /update/gauge/speedAverage/200 - len = 5
	// Пример: /update/counter/               - len = 4
	// Пример: /update/                       - len = 3
	// Пример: /update/another/               - len = 4
	metric := map[string]string{}
	slice := strings.Split(r.URL.String(), "/")
	// 3. Проверяем корректность маршрута
	if len(slice) < 5 {
		return metric, errors.New("Wrong URL")
	}

	metric["type"] = slice[2]
	metric["name"] = slice[3]
	metric["value"] = slice[4]

	return metric, nil
}

func typeMetric(s string) interface{} {
	i, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return i
	}
	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return f
	}
	return ""
}
