package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type product struct {
	Product_id   string
	Product_name string
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func GetProduct(response http.ResponseWriter, request *http.Request) {
	product := product{Product_id: "P001", Product_name: "Sandal"}
	res, err := json.Marshal(product)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(res))

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	response.Write(res)
}

func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

var totalRequest = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Number of get requests.",
	},
	[]string{"path"},
)

var responseStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "response_status",
		Help: "Status of HTTP response",
	},
	[]string{"status"},
)

var httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_response_time_seconds",
	Help: "Duration of HTTP requests",
}, []string{"path"})

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := NewResponseWriter(w)
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
		next.ServeHTTP(w, r)
		statusCode := rw.statusCode
		responseStatus.WithLabelValues(strconv.Itoa(statusCode)).Inc()
		totalRequest.WithLabelValues(path).Inc()
		timer.ObserveDuration()
	})
}

func init() {
	prometheus.Register(totalRequest)
	prometheus.Register(responseStatus)
	prometheus.Register(httpDuration)
}
func main() {
	router := mux.NewRouter()
	router.Use(prometheusMiddleware)

	// endpoint prometheus
	router.Path("/metrics").Handler(promhttp.Handler())
	// endpoint app
	router.HandleFunc("/product", GetProduct)

	fmt.Println("Serving requests on port 9000")
	err := http.ListenAndServe(":9000", router)
	if err != nil {
		log.Fatal(err)
	}
}
