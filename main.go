package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/takacs/go-web/internal/database"
)

type apiConfig struct {
	fileserverHits int
	jwt            string
	DB             *database.DB
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	godotenv.Load()

	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatal(err)
	}

	apiCfg := apiConfig{
		fileserverHits: 0,
		jwt:            os.Getenv("JWT_SECRET"),
		DB:             db,
	}

	router := chi.NewRouter()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	router.Handle("/app", fsHandler)
	router.Handle("/app/*", fsHandler)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", handlerReadiness)
	apiRouter.Get("/chirps", apiCfg.handlerChirpsRetrieve)
	apiRouter.Get("/chirps/{chirpID}", apiCfg.handlerChirpsGetId)
	apiRouter.Post("/chirps", apiCfg.handlerChirpsCreate)
	apiRouter.Post("/users", apiCfg.handlerUsersCreate)
	apiRouter.Post("/login", apiCfg.handlerUsersLogin)
	apiRouter.Post("/refresh", apiCfg.handlerTokenRefresh)
	apiRouter.Put("/users", apiCfg.handlerUsersUpdate)
	apiRouter.Post("/revoke", apiCfg.handlerRevokeToken)
	apiRouter.Delete("/chirps/{chirpID}", apiCfg.handlerChirpDelete)
	apiRouter.Post("/polka/webhooks", apiCfg.handlerPolkaWebooks)
	router.Mount("/api", apiRouter)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiCfg.handlerMetrics)
	router.Mount("/admin", adminRouter)

	corsMux := middlewareCors(router)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
