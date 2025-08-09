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

func slice_remove(s []repoSession, i int) []repoSession {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

type repoSession struct {
	url   string
	token string
	// maybe some other characteristics? we'll see as we need.
	conversation string
}

// Global Variables
var active_repos []repoSession // this needs some kind of threading or something

func queryRepo(w http.ResponseWriter, r *http.Request) {
	// TODO
	// we'll write text in
	// we should give them a cookie

	ctx := context.Background()

	_, err := os.Create("vectors.json")
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Convert the body to a string
	text := string(body) // TODO we will also receive a token

	// Now you can use the 'text' variable as needed
	fmt.Println("Received text:", text)

	// query the pineconedb

	idxConnection, err := pineconeClient.Index(pinecone.NewIndexConnParams{Host: "debug-index-g9pn9ot.svc.aped-4627-b74a.pinecone.io", Namespace: "hi"})
	if err != nil {
		log.Fatalf("Failed to create IndexConnection for Host: %v", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
		w.WriteHeader(http.StatusInternalServerError)               // aw yep
		w.Write([]byte("Failed to create IndexConnection for Host: " + err.Error()))
	}

	res, err := idxConnection.SearchRecords(ctx, &pinecone.SearchRecordsRequest{
		Query: pinecone.SearchRecordsQuery{
			TopK: 5,
			Inputs: &map[string]interface{}{
				"text": text,
			},
		},
	})

	log.Printf("records: %+v", res)

	// write these out to the tempfile, truncate the tempfile
	// dir := fmt.Sprintf("./working/%s/temp.json", uuid)
	file, _ := os.OpenFile("temp.json", os.O_CREATE, os.ModePerm)
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.Encode(res)

	// // now we have the index connection, we need to determine the uuid for this conversation no?

	// command = exec.Command("python3", fmt.Sprintf("../llm_scripts/run_llm.py --workingdir %s", filename))
}

func cleanUpRepo(token string) {
	// now we will do all of this via the tokens as opposed to the string names
	os.RemoveAll(fmt.Sprintf("./working/%s", token))
	for index, session := range active_repos {
		if session.token == token {
			active_repos = slice_remove(active_repos, index)
		}
	}
	log.Printf("Cleaned session with id %s", token)
}

func cloneGithub(url string) (uid string) {

	log.Println("cloning " + url)

	cloneOptions := &git.CloneOptions{
		URL: url,
	}
	token := uuid.Must(uuid.NewRandom())
	idtoken := fmt.Sprintf("%x", token)

	log.Println(fmt.Sprintf("cloning to directory ./working/%s/", idtoken))

	_, err := git.PlainClone(fmt.Sprintf("./working/%s/", idtoken), false, cloneOptions)

	session := repoSession{url: url, token: idtoken}
	active_repos = append(active_repos, session) // we can parse the url later

	if err != nil {
		log.Fatalf("[ERR] FAILED TO CLONE: ", url)
		return ""
	}

	return idtoken

}

func initPineconeClient() (client *pinecone.Client) {
	apiKey := "pcsk_4LZnij_JbQL6KR82nhsGvnLk1PjzTwH91cMUEWwR7SpvTWNauPzGkoGomiex8rFqysZ22Z" // TODO remove this plssss
	client, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: apiKey,
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	url := string(body)
	log.Print("url: " + url)

	token := cloneGithub(string(url))

	// upsert to pinecone db
	idxConnection, err := pineconeClient.Index(pinecone.NewIndexConnParams{Host: "debug-index-g9pn9ot.svc.aped-4627-b74a.pinecone.io", Namespace: token})
	if err != nil {
		log.Fatalf("Failed to create IndexConnection for Host: %v", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
		w.WriteHeader(http.StatusInternalServerError)               // aw yep
		w.Write([]byte("Failed to create IndexConnection for Host: " + err.Error()))
	}

	log.Printf("Chunking files...")
	chunks := chunk_files(token)

	var records []*pinecone.IntegratedRecord

	for _, text := range chunks {
		// log.Printf("id is %s", text["id"])
		// log.Printf("text is %s", text["text"])
		record := &pinecone.IntegratedRecord{
			"id":   text["id"],
			"text": text["text"],
		}

		// fmt.Println("Record chunk_text:", record["chunk_text"])
		records = append(records, record)
	}

	log.Printf("Records to upsert: %d", len(records))
	if len(chunks) == 0 {
		log.Print("Warning: No chunks generated")
	}

	// upsert the records to the index
	err = idxConnection.UpsertRecords(ctx, records)
	if err != nil {
		log.Fatalf("Failed to upsert records: %v", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
		w.WriteHeader(http.StatusInternalServerError)               // aw yep
		w.Write([]byte("Failed to create IndexConnection for Host: " + err.Error()))
		return
	} // check what the fuck is inside the response once we do this. i would like json.

	// etc
	w.WriteHeader(http.StatusOK) // aw yep
	response := map[string]string{
		"status": "success",
		"token":  token,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("Successfully cloned and indexed repository: %s with token: %s", url, token)

	// model will run when this happens.

	return
}

func chunk_files(uuid string) (chunks []map[string]string) {
	// there's no temp.json?
	cmd := exec.Command("sh", "../llm_scripts/initextract.sh", fmt.Sprintf("./working/%s/", uuid)) // args?
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run chunking script: %v", err)
		return []map[string]string{}
	}
	// this creates a json with the uuid
	// defer os.Remove(fmt.Sprintf("./working/%s/temp.json", uuid)) // don't get rid of it yet
	file, err := os.Open(fmt.Sprintf("./working/%s/temp.json", uuid))
	if err != nil {
		log.Fatalf("Failed to open temp.json: %v", err)
		return []map[string]string{}
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&chunks); err != nil {
		log.Fatalf("Failed to decode temp.json: %v", err)
		return []map[string]string{}
	}
	// log.Printf("%+v", chunks)
	return chunks
}

func main() {

	// ctx := context.Background()

	pineconeClient = initPineconeClient()

	log.Println("Starting server on port 8000...")

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
	r.Post("/initialExtract", initialExtraction)
	r.Post("/queryRepo", queryRepo)

	http.ListenAndServe(":8080", r)
}
