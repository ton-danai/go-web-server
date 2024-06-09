package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/ton-danai/go-web-server/internal/database"
)

type apiConfig struct {
	fileserverHits int
	db             *database.DB
	jwtSecret      string
}

func main() {
	const filepathRoot = "."
	const port = "8080"
	// by default, godotenv will look for a file named .env in the current directory
	godotenv.Load()

	instant, serverError := database.New("./database.json")
	if serverError != nil {
		log.Println("Cannot init database")
		return
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	apiCfg := apiConfig{
		fileserverHits: 0,
		db:             instant,
		jwtSecret:      jwtSecret,
	}

	mux := http.NewServeMux()
	//Namespace : app
	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))

	// Namesapce : api
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /api/reset", apiCfg.handlerReset)

	// Namesapce : api/chirps
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirpById)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerPostChirps)

	// Namesapce : api/users
	mux.HandleFunc("POST /api/users", apiCfg.handlerPostUsers)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerPutUsers)

	// Namespace : api/login
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)

	// Namespace : admin
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerAdminMetrics)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	flag.Bool("debug", false, "Enable debug mode")
	// dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

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
