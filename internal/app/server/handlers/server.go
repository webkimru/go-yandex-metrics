package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/models"
	"github.com/webkimru/go-yandex-metrics/internal/utils"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

const (
	Gauge   = "gauge"
	Counter = "counter"

	ContentTypeJSON = "application/json"
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

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	err = t.Execute(w, res)
	if err != nil {
		log.Println("template execution error, Execute() = ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// PostMetrics обрабатывае входящие метрики
func (m *Repository) PostMetrics(w http.ResponseWriter, r *http.Request) {
	var metrics models.Metrics
	// application/json
	if r.Header.Get("Content-Type") == ContentTypeJSON {
		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		// text/plain
		metrics.MType = chi.URLParam(r, "metric")
		metrics.ID = chi.URLParam(r, "name")
		switch metrics.MType {
		case Counter:
			value, err := utils.GetInt64ValueFromSting(chi.URLParam(r, "value"))
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			metrics.Delta = &value
		case Gauge:
			value, err := utils.GetFloat64ValueFromSting(chi.URLParam(r, "value"))
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			metrics.Value = &value
		}
	}

	// При попытке передать запрос с некорректным типом метрики возвращать `http.StatusBadRequest`.
	if metrics.MType != Counter && metrics.MType != Gauge {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// При попытке передать запрос с некорректным значением возвращать `http.StatusBadRequest`.
	switch metrics.MType {
	case Gauge:
		res, err := m.Store.UpdateGauge(metrics.ID, *metrics.Value)
		if err != nil {
			log.Println("failed to update the data from storage, UpdateGauge() = ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := m.WriteResponseGauge(w, r, metrics, res); err != nil {
			log.Println("failed to write the data to the connection, WriteResponseGauge() =", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

	case Counter:
		res, err := m.Store.UpdateCounter(metrics.ID, *metrics.Delta)
		if err != nil {
			log.Println("failed to update the data from storage, UpdateCounter() = ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := m.WriteResponseCounter(w, r, metrics, res); err != nil {
			log.Println("failed to write the data to the connection, WriteResponseCounter() =", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (m *Repository) GetMetric(w http.ResponseWriter, r *http.Request) {
	var metrics models.Metrics
	// application/json
	if r.Header.Get("Content-Type") == ContentTypeJSON {
		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		// text/plain
		metrics.MType = chi.URLParam(r, "metric")
		metrics.ID = chi.URLParam(r, "name")
	}

	switch metrics.MType {
	case Counter:
		res, err := m.Store.GetCounter(metrics.ID)
		if err != nil {
			log.Println("failed to get the data from storage, GetCounter() = ", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err := m.WriteResponseCounter(w, r, metrics, res); err != nil {
			log.Println("failed to write the data to the connection, WriteResponseCounter() =", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	case Gauge:
		res, err := m.Store.GetGauge(metrics.ID)
		if err != nil {
			log.Println("failed to get the data from storage, GetGauge() = ", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err := m.WriteResponseGauge(w, r, metrics, res); err != nil {
			log.Println("failed to write the data to the connection, WriteResponseGauge() =", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	default:
		w.WriteHeader(http.StatusNotFound)
	}

}

func (m *Repository) WriteResponseCounter(w http.ResponseWriter, r *http.Request, metrics models.Metrics, value int64) error {
	// application/json
	if r.Header.Get("Content-Type") == ContentTypeJSON {
		metrics.Delta = &value

		w.Header().Set("Content-Type", ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(metrics); err != nil {
			return err
		}

		return nil
	}

	// text/plain
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(strconv.Itoa(int(value))))
	if err != nil {
		return err
	}

	return nil
}

func (m *Repository) WriteResponseGauge(w http.ResponseWriter, r *http.Request, metrics models.Metrics, value float64) error {
	// application/json
	if r.Header.Get("Content-Type") == ContentTypeJSON {
		metrics.Value = &value

		w.Header().Set("Content-Type", ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(metrics); err != nil {
			return err
		}

		return nil
	}

	// text/plain
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))
	if err != nil {
		return err
	}

	return nil
}
