package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	otelpyroscope "github.com/grafana/otel-profiling-go"
	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

var tracer = otel.Tracer("demo-service")

func main() {
	// Get pod info from environment
	podName := os.Getenv("POD_NAME")
	podNamespace := os.Getenv("POD_NAMESPACE")
	if podName == "" {
		podName = "unknown-pod"
	}
	if podNamespace == "" {
		podNamespace = "default"
	}

	pyroscopeServer := os.Getenv("PYROSCOPE_SERVER_ADDRESS")
	if pyroscopeServer == "" {
		pyroscopeServer = "http://pyroscope:4040"
	}

	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otlpEndpoint == "" {
		otlpEndpoint = "tempo:4317"
	}

	log.Printf("Starting trace-profile-demo")
	log.Printf("Pod: %s/%s", podNamespace, podName)
	log.Printf("Pyroscope: %s", pyroscopeServer)
	log.Printf("Tempo OTLP: %s", otlpEndpoint)

	// 1. Initialize Pyroscope Profiler FIRST
	_, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: "trace-profile-demo",
		ServerAddress:   pyroscopeServer,
		Logger:          pyroscope.StandardLogger,
		Tags: map[string]string{
			"service":       "demo",
			"env":           "test",
			"namespace":     podNamespace,
			"pod":           podName,
			"k8s_pod":       podName,
			"k8s_namespace": podNamespace,
		},
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
		},
	})
	if err != nil {
		log.Fatal("Failed to start Pyroscope:", err)
	}

	// 2. Initialize OpenTelemetry Trace Exporter
	ctx := context.Background()
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(otlpEndpoint),
	)
	if err != nil {
		log.Fatal("Failed to create trace exporter:", err)
	}

	// 3. Configure the base Tracer Provider
	baseTp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("trace-profile-demo"),
			semconv.ServiceVersion("1.0.0"),
			attribute.String("k8s.namespace.name", podNamespace),
			attribute.String("k8s.pod.name", podName),
			attribute.String("service.name", "trace-profile-demo"),
			attribute.String("service.namespace", podNamespace),
		)),
	)

	// 4. CRITICAL: Wrap the tracer provider with otelpyroscope
	otel.SetTracerProvider(otelpyroscope.NewTracerProvider(baseTp))

	// Set propagator for trace context
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// 5. Setup HTTP Handlers - Regular endpoints
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/fib", fibHandler)
	http.HandleFunc("/cpu-heavy", cpuHeavyHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/ready", readyHandler)

	// 6. Error endpoints for testing
	http.HandleFunc("/error/panic", panicHandler)
	http.HandleFunc("/error/timeout", timeoutHandler)
	http.HandleFunc("/error/memory", memoryLeakHandler)
	http.HandleFunc("/error/cpu", cpuSpikeHandler)
	http.HandleFunc("/error/database", databaseErrorHandler)
	http.HandleFunc("/error/validation", validationErrorHandler)
	http.HandleFunc("/error/notfound", notFoundHandler)
	http.HandleFunc("/error/unauthorized", unauthorizedHandler)
	http.HandleFunc("/error/slow", slowHandler)

	// 7. Background error generator (random errors every few seconds)
	go randomErrorGenerator()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// ============= NORMAL HANDLERS =============

func helloHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, span := tracer.Start(ctx, "helloHandler")
	defer span.End()

	time.Sleep(10 * time.Millisecond)

	span.SetAttributes(attribute.String("message", "Hello from demo"))
	fmt.Fprintf(w, "Hello from trace-profile-demo! (Pod: %s)\n", os.Getenv("POD_NAME"))
}

func fibHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, span := tracer.Start(ctx, "fibHandler")
	defer span.End()

	n := 40
	result := fibonacci(n)

	span.SetAttributes(attribute.Int("fib.n", n))
	fmt.Fprintf(w, "Fibonacci(%d) = %d\n", n, result)
}

func cpuHeavyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, span := tracer.Start(ctx, "cpuHeavyHandler")
	defer span.End()

	result := 0.0
	for i := 0; i < 10_000_000; i++ {
		result += rand.Float64()
	}

	span.SetAttributes(attribute.Float64("computation.result", result))
	fmt.Fprintf(w, "CPU heavy computation complete: %f (Pod: %s)\n", result, os.Getenv("POD_NAME"))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Ready")
}

// ============= ERROR HANDLERS =============

// panicHandler - Causes the service to crash (will be restarted by K8s)
func panicHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, span := tracer.Start(ctx, "panicHandler")
	defer span.End()

	log.Printf("💥 PANIC endpoint triggered!")
	span.SetAttributes(attribute.String("error.type", "panic"))
	span.RecordError(errors.New("intentional panic"))
	span.SetStatus(codes.Error, "panic occurred")

	// This will crash the pod
	panic("Intentional panic triggered by /error/panic endpoint")
}

// timeoutHandler - Simulates a timeout
func timeoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "timeoutHandler")
	defer span.End()

	duration := 30 // seconds
	log.Printf("⏰ Starting operation that will timeout after %d seconds...", duration)
	span.SetAttributes(attribute.Int("timeout.seconds", duration))

	select {
	case <-time.After(time.Duration(duration) * time.Second):
		fmt.Fprintf(w, "Operation completed")
	case <-ctx.Done():
		log.Printf("❌ Request timeout after %d seconds", duration)
		span.RecordError(context.DeadlineExceeded)
		span.SetStatus(codes.Error, "context deadline exceeded")
		w.WriteHeader(http.StatusGatewayTimeout)
		fmt.Fprintf(w, "Request timeout exceeded after %d seconds\n", duration)
	}
}

// memoryLeakHandler - Accumulates memory to simulate leaks
var memoryLeakSlice [][]byte

func memoryLeakHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, span := tracer.Start(ctx, "memoryLeakHandler")
	defer span.End()

	// Get leak size from query param (default 10MB)
	sizeStr := r.URL.Query().Get("size")
	size := 10 // MB
	if sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil {
			size = s
		}
	}

	// Allocate memory and keep reference (LEAK!)
	leak := make([]byte, size*1024*1024)
	for i := range leak {
		leak[i] = byte(i % 256)
	}
	memoryLeakSlice = append(memoryLeakSlice, leak)

	log.Printf("💾 Memory leak increased by %d MB (Total leaked: %d MB)",
		size, len(memoryLeakSlice)*size)
	span.SetAttributes(attribute.Int("leaked.mb", size))
	span.SetAttributes(attribute.Int("total.leaked.mb", len(memoryLeakSlice)*size))

	// Record as warning, not error
	span.SetStatus(codes.Ok, "Memory leak simulated")
	fmt.Fprintf(w, "Memory leak increased by %d MB. Total leaked: %d MB\n",
		size, len(memoryLeakSlice)*size)
}

// cpuSpikeHandler - Intense CPU usage
func cpuSpikeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, span := tracer.Start(ctx, "cpuSpikeHandler")
	defer span.End()

	durationStr := r.URL.Query().Get("duration")
	duration := 10 // seconds
	if durationStr != "" {
		if d, err := strconv.Atoi(durationStr); err == nil {
			duration = d
		}
	}

	log.Printf("🔥 CPU spike for %d seconds...", duration)
	span.SetAttributes(attribute.Int("spike.duration", duration))

	// CPU intensive calculation
	deadline := time.Now().Add(time.Duration(duration) * time.Second)
	result := float64(0)
	for time.Now().Before(deadline) {
		for i := 0; i < 1000000; i++ {
			result += float64(i) * float64(i) / float64(i+1)
		}
	}

	fmt.Fprintf(w, "CPU spike completed after %d seconds. Result: %f\n", duration, result)
}

// databaseErrorHandler - Simulates various database errors
func databaseErrorHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, span := tracer.Start(ctx, "databaseErrorHandler")
	defer span.End()

	errorType := r.URL.Query().Get("type")
	if errorType == "" {
		errorType = "connection_refused"
	}

	var err error
	var statusCode int

	switch errorType {
	case "connection_refused":
		err = errors.New("dial tcp 10.0.0.1:5432: connect: connection refused")
		statusCode = http.StatusServiceUnavailable
		log.Printf("💥 Database connection refused")

	case "timeout":
		err = errors.New("pq: canceling statement due to statement timeout")
		statusCode = http.StatusGatewayTimeout
		log.Printf("⏰ Database query timeout")

	case "deadlock":
		err = errors.New("deadlock detected: transaction aborted")
		statusCode = http.StatusConflict
		log.Printf("🔒 Database deadlock detected")

	case "duplicate_key":
		err = errors.New("duplicate key value violates unique constraint")
		statusCode = http.StatusConflict
		log.Printf("📝 Duplicate key error")

	case "disk_full":
		err = errors.New("could not write to disk: no space left on device")
		statusCode = http.StatusInternalServerError
		log.Printf("💾 Disk full error")

	default:
		err = errors.New("unknown database error: failed to execute query")
		statusCode = http.StatusInternalServerError
		log.Printf("❌ Generic database error")
	}

	span.RecordError(err)
	span.SetAttributes(attribute.String("error.type", errorType))
	span.SetAttributes(attribute.Int("error.code", statusCode))
	span.SetStatus(codes.Error, err.Error())

	w.WriteHeader(statusCode)
	fmt.Fprintf(w, "Database error: %v\n", err)
}

// validationErrorHandler - Simulates input validation errors
func validationErrorHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, span := tracer.Start(ctx, "validationErrorHandler")
	defer span.End()

	field := r.URL.Query().Get("field")
	if field == "" {
		field = "email"
	}

	err := fmt.Errorf("validation failed: field '%s' has invalid format", field)
	log.Printf("⚠️ Validation error on field: %s", field)

	span.RecordError(err)
	span.SetAttributes(attribute.String("error.field", field))
	span.SetAttributes(attribute.String("error.type", "validation"))
	span.SetStatus(codes.Error, err.Error())

	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "Validation error: %v\n", err)
}

// notFoundHandler - Simulates 404 errors
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, span := tracer.Start(ctx, "notFoundHandler")
	defer span.End()

	resource := r.URL.Query().Get("resource")
	if resource == "" {
		resource = "user:123"
	}

	err := fmt.Errorf("resource not found: %s", resource)
	log.Printf("🔍 Resource not found: %s", resource)

	span.RecordError(err)
	span.SetAttributes(attribute.String("resource", resource))
	span.SetAttributes(attribute.Int("http.status_code", http.StatusNotFound))
	span.SetStatus(codes.Error, err.Error())

	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "Resource not found: %s\n", resource)
}

// unauthorizedHandler - Simulates authentication errors
func unauthorizedHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, span := tracer.Start(ctx, "unauthorizedHandler")
	defer span.End()

	err := errors.New("authentication failed: invalid or expired token")
	log.Printf("🔒 Unauthorized access attempt")

	span.RecordError(err)
	span.SetAttributes(attribute.String("error.type", "unauthorized"))
	span.SetAttributes(attribute.Int("http.status_code", http.StatusUnauthorized))
	span.SetStatus(codes.Error, err.Error())

	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, "Unauthorized: %v\n", err)
}

// slowHandler - Slow but successful operation (tests latency monitoring)
func slowHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, span := tracer.Start(ctx, "slowHandler")
	defer span.End()

	delayStr := r.URL.Query().Get("delay")
	delay := 5 // seconds
	if delayStr != "" {
		if d, err := strconv.Atoi(delayStr); err == nil {
			delay = d
		}
	}

	log.Printf("🐢 Slow operation taking %d seconds...", delay)
	span.SetAttributes(attribute.Int("delay.seconds", delay))

	// Simulate work with progress updates
	for i := 0; i < delay; i++ {
		select {
		case <-ctx.Done():
			span.RecordError(ctx.Err())
			span.SetStatus(codes.Error, "request cancelled")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "Operation cancelled after %d seconds\n", i)
			return
		case <-time.After(1 * time.Second):
			// Do some work each second
			_ = fibonacci(30)
			log.Printf("  Progress: %d/%d seconds", i+1, delay)
		}
	}

	span.SetStatus(codes.Ok, "slow operation completed")
	fmt.Fprintf(w, "Slow operation completed after %d seconds\n", delay)
}

// ============= BACKGROUND ERROR GENERATOR =============

func randomErrorGenerator() {
	log.Println("🤖 Starting random error generator (errors will appear every 30-60 seconds)")

	// List of error endpoints to hit randomly
	errorEndpoints := []string{
		"/error/database",
		"/error/validation",
		"/error/notfound",
		"/error/unauthorized",
		"/error/timeout",
		"/error/memory",
		"/error/cpu",
	}

	for {
		// Random delay between 30 and 90 seconds
		delay := time.Duration(30+rand.Intn(60)) * time.Second
		time.Sleep(delay)

		// Pick random endpoint
		endpoint := errorEndpoints[rand.Intn(len(errorEndpoints))]

		// Add random parameters
		var fullURL string
		switch endpoint {
		case "/error/database":
			dbErrors := []string{"connection_refused", "timeout", "deadlock", "duplicate_key", "disk_full"}
			errType := dbErrors[rand.Intn(len(dbErrors))]
			fullURL = fmt.Sprintf("http://localhost:8080%s?type=%s", endpoint, errType)
		case "/error/validation":
			fields := []string{"email", "username", "password", "age", "phone"}
			field := fields[rand.Intn(len(fields))]
			fullURL = fmt.Sprintf("http://localhost:8080%s?field=%s", endpoint, field)
		case "/error/notfound":
			resources := []string{"user:123", "product:456", "order:789", "page:index"}
			resource := resources[rand.Intn(len(resources))]
			fullURL = fmt.Sprintf("http://localhost:8080%s?resource=%s", endpoint, resource)
		case "/error/cpu":
			durations := []int{3, 5, 8, 10}
			duration := durations[rand.Intn(len(durations))]
			fullURL = fmt.Sprintf("http://localhost:8080%s?duration=%d", endpoint, duration)
		case "/error/memory":
			sizes := []int{1, 2, 5, 10}
			size := sizes[rand.Intn(len(sizes))]
			fullURL = fmt.Sprintf("http://localhost:8080%s?size=%d", endpoint, size)
		default:
			fullURL = fmt.Sprintf("http://localhost:8080%s", endpoint)
		}

		log.Printf("🤖 Random error generator calling: %s", fullURL)

		// Make the request (async, don't wait for response to avoid blocking)
		go func(url string) {
			resp, err := http.Get(url)
			if err != nil {
				log.Printf("Error generator failed to call %s: %v", url, err)
				return
			}
			defer resp.Body.Close()
			log.Printf("Error generator got response: %s - Status: %d", url, resp.StatusCode)
		}(fullURL)
	}
}

// ============= UTILITY FUNCTIONS =============

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}
