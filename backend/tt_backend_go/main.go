package main

// todo webframe work

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/pinecone-io/go-pinecone/v3/pinecone"
	"gopkg.in/src-d/go-git.v4"
)

// ROUTES
//
//	There should probably be a serverside component for chatbot respones. also throbbers are needed

type repoSessions struct {
	url   string
	token string
	// maybe some other characteristics
}

// Global Variables
var active_repos []string // this needs some kind of threading or something
var active_sessions []*repoSessions

func queryRepo(w http.ResponseWriter, r *http.Request) {
	// TODO
	// we'll write text in
	// we should give them a cookie
}

func cleanUpRepo(url string) {

}

func cloneGithub(url string) (status int) {

	log.Println("cloning " + url)

	cloneOptions := &git.CloneOptions{
		URL:           url,
		ReferenceName: git.ReferenceName("refs/heads/main"), // TODO
		SingleBranch:  true,
		Depth:         1,
		Progress:      nil,
	}

	_, err := git.PlainClone("./working/", true, cloneOptions)

	active_repos = append(active_repos, url) // we can parse the url later

	if err != nil {
		print("[ERR] FAILED TO CLONE: ", url)
		return -1
	}

	return 0

}

func initPineconeClient() (client *pinecone.Client) {
	client, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: "YOUR_API_KEY",
	})
	if err != nil {
		log.Fatalf("Failed to create Client: %v", err)
		return nil
	}
	return client
}

var pineconeClient *pinecone.Client

func initialExtraction(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	var url []byte // todo test urls and whatnot

	r.Body.Read(url) // should work I hope.

	print(url)

	status := cloneGithub(string(url))
	if status == -1 {
		log.Fatalf("Couldn't clone github repository")
	}

	// upsert to pinecone db
	idxConnection, err := pineconeClient.Index(pinecone.NewIndexConnParams{Host: "INDEX_HOST", Namespace: string(url)})
	if err != nil {
		log.Fatalf("Failed to create IndexConnection for Host: %v", err)
	}

	records := chunk_files(url)

	// etc

}

func chunk_files(url string) {

}

func main() {

	ctx := context.Background()
	if pineconeClient == nil {
		log.Fatalf("Failed to initialze pinecone client, shutting down")
		return
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(cors.Handler)

	// not found and bad method

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("{\"status\":\"404\"}"))
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(405)
		w.Write([]byte("{\"status\":\"405\"}"))
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	}) // TODO get rid off this later.

	// declaring our routs
	r.Get("/initialExtract", initialExtraction)
	r.Post("/queryRepo", queryRepo)

	http.ListenAndServe(":3000", r)
}
