package main

// todo webframe work

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/pinecone-io/go-pinecone/v3/pinecone"
	"gopkg.in/src-d/go-git.v4"
)

// ROUTES
//
//	There should probably be a serverside component for chatbot respones. also throbbers are needed

type repoSession struct {
	url   string
	token string
	// maybe some other characteristics? we'll see as we need.
	conversation string
}

// Global Variables
var active_repos []repoSession // this needs some kind of threading or something

func remove(s []int, i int) []int {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func queryRepo(w http.ResponseWriter, r *http.Request) {
	// TODO
	// we'll write text in
	// we should give them a cookie

	f, err := os.Create("vectors.json")
	if err != nil {
		panic(err)
	}

	output := []byte("hello!")
	f.Write(output)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Convert the body to a string
	text := string(body)

	// Now you can use the 'text' variable as needed
	fmt.Println("Received text:", text)

	// query the pineconedb

	// if err := idxConnection.Upsert(ctx, records); err != nil {
	// 	log.Fatalf("Failed to upsert records: %v", err)
	// 	w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
	// 	w.WriteHeader(http.Status)                                  // aw yep
	// 	w.Write([]byte("Failed to create IndexConnection for Host: " + err.Error()))
	// 	return
	// }

	// // now we have the index connection, we need to determine the uuid for this conversation no?

	// command = exec.Command("python3", fmt.Sprintf("../llm_scripts/run_llm.py --vectorfile %s", filename))
}

func cleanUpRepo(token string) {
	// now we will do all of this via the tokens as opposed to the string names
	os.RemoveAll(fmt.Sprintf("./working/%s", token))
	for index, session := range active_repos {
		index = index
		if session.token == token {
			// remove()
			log.Println("hihihi")
		}
	}
}

func cloneGithub(url string) (uid string) {

	log.Println("cloning " + url)

	cloneOptions := &git.CloneOptions{
		URL: url,
	}

	_, err := git.PlainClone(fmt.Sprintf("./working/%s", uid), true, cloneOptions)

	token := uuid.Must(uuid.NewRandom())
	idtoken := fmt.Sprintf("%x", token)

	session := repoSession{url: url, token: idtoken}
	active_repos = append(active_repos, session) // we can parse the url later

	if err != nil {
		print("[ERR] FAILED TO CLONE: ", url)
		return ""
	}

	return idtoken

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

	token := cloneGithub(string(url))

	// upsert to pinecone db
	idxConnection, err := pineconeClient.Index(pinecone.NewIndexConnParams{Host: "https://debug-index-g9pn9ot.svc.aped-4627-b74a.pinecone.io", Namespace: string(token)})
	if err != nil {
		log.Fatalf("Failed to create IndexConnection for Host: %v", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
		w.WriteHeader(http.StatusInternalServerError)               // aw yep
		w.Write([]byte("Failed to create IndexConnection for Host: " + err.Error()))
	}

	chunks := chunk_files(token)

	records := *pinecone.IntegratedRecordsFromMap(chunks, pinecone.IntegratedRecordParams{
		Namespace: token,
	})
	println("Records: ", records)

	// upsert the records to the index
	if err := idxConnection.Upsert(ctx, records); err != nil {
		log.Fatalf("Failed to upsert records: %v", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
		w.WriteHeader(http.Status)                                  // aw yep
		w.Write([]byte("Failed to create IndexConnection for Host: " + err.Error()))
		return
	} // check what the fuck is inside the response once we do this. i would like json

	// etc
	w.WriteHeader(http.StatusOK) // aw yep
	response := map[string]string{
		"status": "success",
		"token":  token,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("Successfully cloned and indexed repository: %s with token: %s", url, token)

	// TODO run the model on the chunks.

	return

}

func chunk_files(uuid string) (chunks map[string]string) {
	cmd := exec.Command("python3", fmt.Sprintf("../llm_scripts/queryRepo.py --workingdir %s", uuid))
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run chunking script: %v", err)
		return map[string]string{}
	}
	// this creates a json with the uuid
	defer os.Remove(fmt.Sprintf("./working/%s/temp.json", uuid)) // don't get rid of it yet
	file, err := os.Open(fmt.Sprintf("./working/%s/temp.json", uuid))
	if err != nil {
		log.Fatalf("Failed to open temp.json: %v", err)
		return map[string]string{}
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&chunks); err != nil {
		log.Fatalf("Failed to decode temp.json: %v", err)
		return map[string]string{}
	}
	return chunks
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
