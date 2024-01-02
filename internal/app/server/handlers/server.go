package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/webkimru/go-yandex-metrics/internal/utils"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

// Default задет дефолтный маршрут
func (m *Repository) Default(w http.ResponseWriter, _ *http.Request) {
	stringHTML := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Metrics</title>
</head>
<body>
    {{range $k, $v := .counter}}
    	{{$k}} {{$v}}<br>
	{{end}}
    {{range $k, $v := .gauge}}
    	{{$k}} {{$v}}<br>
	{{end}}
</body>
</html>
`

	res, err := m.Store.GetAllMetrics()
	if err != nil {
		log.Println("failed to get the data from storage, GetAllMetrics() = ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t := template.New("Metrics")
	t, err = t.Parse(stringHTML)
	if err != nil {
		log.Println("HTML template is not parsed, Parse() = ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "Content-Type")
	err = t.Execute(w, res)
	if err != nil {
		log.Println("template execution error, Execute() = ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// PostMetrics обрабатывае входящие метрики
func (m *Repository) PostMetrics(w http.ResponseWriter, r *http.Request) {
	// Парсим маршрут и получаем мапку: метрика, значение
	metric := make(map[string]string, 3)
	metric["type"] = chi.URLParam(r, "metric")
	metric["name"] = chi.URLParam(r, "name")
	metric["value"] = chi.URLParam(r, "value")

	// При попытке передать запрос с некорректным типом метрики возвращать `http.StatusBadRequest`.
	if metric["type"] != Counter && metric["type"] != Gauge {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// При попытке передать запрос с некорректным значением возвращать `http.StatusBadRequest`.
	switch metric["type"] {
	case Gauge:
		value, err := utils.GetFloat64ValueFromSting(metric["value"])
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = m.Store.UpdateGauge(metric["name"], value)
		if err != nil {
			log.Println("failed to update the data from storage, UpdateGauge() = ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return

	case Counter:
		value, err := utils.GetInt64ValueFromSting(metric["value"])
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = m.Store.UpdateCounter(metric["name"], value)
		if err != nil {
			log.Println("failed to update the data from storage, UpdateCounter() = ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (m *Repository) GetMetric(w http.ResponseWriter, r *http.Request) {
	metric := chi.URLParam(r, "metric")
	name := chi.URLParam(r, "name")

	switch metric {
	case Counter:
		res, err := m.Store.GetCounter(name)
		if err != nil {
			log.Println("failed to get the data from storage, GetCounter() = ", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, err = w.Write([]byte(strconv.FormatInt(res, 10)))
		if err != nil {
			log.Println("failed to write the data to the connection, Write() =", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return

	case Gauge:
		res, err := m.Store.GetGauge(name)
		if err != nil {
			log.Println("failed to get the data from storage, GetGauge() = ", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, err = w.Write([]byte(strconv.FormatFloat(res, 'f', -1, 64)))
		if err != nil {
			log.Println("failed to write the data to the connection, Write() =", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}
