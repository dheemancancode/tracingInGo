package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/openzipkin/zipkin-go/model"

	"github.com/gorilla/mux"

	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/middleware/http"
	reporterhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

const endpointURL = "http://10.41.120.226:9411/api/v2/spans"

// Example Example code
func Example() {

	reporterURL := reporterhttp.NewReporter(endpointURL)
	localEndpoint := &model.Endpoint{ServiceName: "tracing-root-go", Port: 8080}

	// initialize our tracer
	tracer, err := zipkin.NewTracer(reporterURL, zipkin.WithLocalEndpoint(localEndpoint))
	if err != nil {
		log.Fatalf("unable to create tracer: %+v\n", err)
	}

	// create global zipkin http server middleware
	serverMiddleware := zipkinhttp.NewServerMiddleware(
		tracer, zipkinhttp.TagResponseSize(true),
	)

	// create global zipkin traced http client
	client, err := zipkinhttp.NewClient(tracer, zipkinhttp.ClientTrace(true))
	if err != nil {
		log.Fatalf("unable to create client: %+v\n", err)
	}

	// initialize router
	router := mux.NewRouter()

	router.Use(serverMiddleware)

	// set-up handlers
	router.Methods("GET").Path("/tracinggo").HandlerFunc(someFunc(client))
	router.Methods("GET").Path("/health").HandlerFunc(healthFunc(client))

	log.Fatal(http.ListenAndServe(":8080", router))
}

func healthFunc(client *zipkinhttp.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"response": "all healthy!"})
	}
}

func someFunc(client *zipkinhttp.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span := zipkin.SpanFromContext(r.Context())
		span.Tag("custom_key", "some value")

		// doing some expensive calculations....
		time.Sleep(25 * time.Millisecond)
		span.Annotate(time.Now(), "expensive_calc_done")

		newRequest, err := http.NewRequest("GET", "https://gotracing.dev.target.com/some_function", nil)
		if err != nil {
			log.Printf("unable to create client: %+v\n", err)
			http.Error(w, err.Error(), 500)
			return
		}

		ctx := zipkin.NewContext(newRequest.Context(), span)

		newRequest = newRequest.WithContext(ctx)

		res, err := client.DoWithAppSpan(newRequest, "other_function")
		if err != nil {
			log.Printf("call to spring boot application returned error: %+v\n", err)
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Println(res.Body)
		res.Body.Close()
		json.NewEncoder(w).Encode(map[string]string{"response": "all good main function - root!"})
	}
}

func main() {
	Example()
}
