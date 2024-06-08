package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ton-danai/go-web-server/internal/database"
)

type apiConfig struct {
	fileserverHits int
	currentId      int
	db             *database.DB
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	instant, serverError := database.New("./database.json")
	if serverError != nil {
		log.Println("Cannot init database")
		return
	}

	apiCfg := apiConfig{
		fileserverHits: 0,
		currentId:      1,
		db:             instant,
	}

	mux := http.NewServeMux()
	//Namespace : app
	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))

	// Namesapce : api
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /api/reset", apiCfg.handlerReset)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerPostChirps)

	//Namespace : admin
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerAdminMetrics)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}
