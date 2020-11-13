module github.com/datawire/k8s_initializer_sample_app

go 1.15

require (
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/prometheus/client_golang v1.7.1
	go.opentelemetry.io/collector v0.14.0
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.13.0
	go.opentelemetry.io/contrib/propagators v0.13.0
	go.opentelemetry.io/otel v0.13.0
	go.opentelemetry.io/otel/exporters/otlp v0.13.0
	go.opentelemetry.io/otel/exporters/stdout v0.13.0 // indirect
	go.opentelemetry.io/otel/sdk v0.13.0
)
