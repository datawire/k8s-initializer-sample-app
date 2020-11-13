package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/collector/translator/conventions"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	otelglobal "go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/otlp"
	// "go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/propagators"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

const (
	LISTEN_PORT             = "LISTEN_PORT"
	OTLP_RECEIVER_ADDRESS   = "OTLP_RECEIVER_ADDRESS"
	PROMETHEUS_PATH         = "PROMETHEUS_PATH"
	JAEGER_PATH             = "JAEGER_PATH"
	OTEL_COLLECTOR_ENDPOINT = "OTEL_COLLECTOR_ENDPOINT"
	KEYCLOAK_PATH           = "KEYCLOAK_PATH"
	ARGOCD_PATH             = "ARGOCD_PATH"
)

var (
	// Custom Prometheus metrics
	pageViews = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sample_app_page_views_total",
		Help: "The total number of page views",
	})
)

func initTracer() {
	// Initialize the OpenTelemetry Tracer
	otlpReceiver := os.Getenv(OTLP_RECEIVER_ADDRESS)
	log.Printf("Initializing the the otlp exporter to %v\n", otlpReceiver)

	// Create an exporter
	exporter, err := otlp.NewExporter(otlp.WithAddress(otlpReceiver), otlp.WithInsecure())
	//exporter, err := stdout.NewExporter(stdout.WithPrettyPrint())
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}

	cfg := trace.Config{
		DefaultSampler: trace.AlwaysSample(), // Sample all calls
	}
	tp := trace.NewTracerProvider(
		trace.WithConfig(cfg),
		trace.WithSyncer(exporter),
		trace.WithResource(resource.New(
			// Add application-specific attributes to all spans
			label.String(conventions.AttributeServiceName, "sample-app"),
			label.String(conventions.AttributeServiceVersion, "1.0.0"),
		)),
	)

	// Make our trace provider the global instance
	otelglobal.SetTracerProvider(tp)

	// Handle many trace-propagation header formats, including X-B3
	otelglobal.SetTextMapPropagator(otel.NewCompositeTextMapPropagator(b3.B3{}, propagators.TraceContext{}, propagators.Baggage{}))
}

func serveHealth(w http.ResponseWriter, r *http.Request) {
	// Health always returns 200 OK
	w.WriteHeader(http.StatusOK)
}

func serveFavIcon(w http.ResponseWriter, r *http.Request) {
	// There is no favicon, return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	// Load all HTML templates
	templates, _ := template.ParseGlob("templates/*.html")

	// Extract the request base URL to to display it in the UI.
	scheme := r.URL.Scheme
	if scheme == "" {
		scheme = getRequestScheme(r)
	}
	hostname := r.Host

	// Build all template variables
	templateVariables := struct {
		Basepath                  string
		PrometheusPath            string
		JaegerPath                string
		OpenTelemetryCollectorSvc string
		KeycloakPath              string
		ArgoCDPath                string
	}{
		Basepath:                  fmt.Sprintf("%v://%v", scheme, hostname),
		PrometheusPath:            os.Getenv(PROMETHEUS_PATH),
		JaegerPath:                os.Getenv(JAEGER_PATH),
		OpenTelemetryCollectorSvc: os.Getenv(OTEL_COLLECTOR_ENDPOINT),
		KeycloakPath:              os.Getenv(KEYCLOAK_PATH),
		ArgoCDPath:                os.Getenv(ARGOCD_PATH),
	}

	// Render the template given the variables
	err := templates.ExecuteTemplate(w, "index.html", templateVariables)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	// Increase our pageViews Prometheus counter
	pageViews.Inc()
}

func getRequestScheme(r *http.Request) string {
	// Defaults to "http"
	scheme := "http"

	// Retrieve the scheme from X-Forwarded-Proto.
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = strings.ToLower(proto)
	}

	return scheme
}

func main() {
	// Initialize the OpenTelemetry tracer
	initTracer()

	// Create an HTTP router
	r := mux.NewRouter()
	r.StrictSlash(true)

	// Register the OpenTelemetry tracer as a middleware
	r.Use(otelmux.Middleware("mux-router"))

	// Serve Prometheus metrics
	r.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Serve static assets
	staticContent := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticContent))
	r.HandleFunc("/favicon.ico", serveFavIcon).Methods("GET")

	// Index handler
	r.HandleFunc("/", serveTemplate).Methods("GET")

	// Health-check handler
	r.HandleFunc("/health", serveHealth).Methods("GET")

	// Start the HTTP server
	port := os.Getenv(LISTEN_PORT)
	if port == "" {
		port = "3000"
	}
	log.Printf("Listening on :%v...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%v", port), handlers.ProxyHeaders(r))
	if err != nil {
		log.Fatal(err)
	}
}
