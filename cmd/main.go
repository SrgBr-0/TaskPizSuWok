package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/SrgBr-0/TaskPizSuWok/internal/funcs"
)

func main() {
	r := chi.NewRouter()
	r.Get("/container-info/{name}", funcs.GetContainerInfoHandler)

	log.Println("Starting server on :8080")
	http.ListenAndServe(":8080", r)
}
