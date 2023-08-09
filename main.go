package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type apiConfig struct {
	fileserverHits int
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	r := chi.NewRouter()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	r.Handle("/app", fsHandler)
	r.Handle("/app/*", fsHandler)

	api := chi.NewRouter()
	api.Get("/healthz", handlerReadiness)
	api.Post("/validate_chirp", validateChirp)

	admin := chi.NewRouter()
	admin.Get("/metrics", apiCfg.handlerMetrics)

	r.Mount("/api", api)
	r.Mount("/admin", admin)
	corsMux := middlewareCors(r)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
<html>

<body>
	<h1>Welcome, Chirpy Admin</h1>
	<p>Chirpy has been visited %d times!</p>
</body>

</html>
	`, cfg.fileserverHits)))
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type errorVals struct {
		Error string `json:"error"`
	}
	type goodVal struct {
		Good bool `json:"valid"`
	}

	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	log.Printf("Length of request body: %v", len(params.Body))
	if err != nil {
		errval := errorVals{Error: "Something went wrong"}
		dat, _ := json.Marshal(errval)
		log.Printf("Error reading JSON")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(dat)
	} else if len(params.Body) > 140 {
		errval := errorVals{Error: "Chirp is too long"}
		dat, _ := json.Marshal(errval)
		log.Printf("Chirp too long")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(dat)
	} else {
		valid := goodVal{Good: true}
		dat, _ := json.Marshal(valid)
		log.Printf("OK!")
		w.WriteHeader(http.StatusOK)
		w.Write(dat)
	}
}
