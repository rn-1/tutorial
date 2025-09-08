package main

//GOD, homeschooled!

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/pinecone-io/go-pinecone/v3/pinecone"
	"github.com/tmc/langchaingo/textsplitter"
	"gopkg.in/src-d/go-git.v4"
)

// ** Global Variables and Types
type repostore struct {
	m        sync.Mutex
	repolist []repoSession // this needs some kind of threading or something
}

var active_repos repostore

type repoSession struct {
	url   string
	token string
	convo []map[string]string
	// maybe some other characteristics? we'll see as we need.
}

var pineconeClient *pinecone.Client

// ** Utility functions

func slice_remove(s []repoSession, i int) []repoSession {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func checkError(err error) {
	if err != nil {
		log.Printf("error: %+v", err)
		panic("fuck")
	}
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

func run_textsplitter(uuid string) (all_chunks []map[string]string) {

	files, err := filepath.Glob(fmt.Sprintf("./working/%s/*.*[^o]", uuid)) // TODO build a better way to properly pull out files.
	if err != nil || len(files) == 0 {
		panic("oh fuck bad globbing")
	}

	for _, file := range files {
		var split textsplitter.TextSplitter
		log.Printf("indexing...")
		ext := file[strings.Index(file, "."):]
		log.Printf("done indexing!")
		if ext == ".md" {
			split = textsplitter.NewMarkdownTextSplitter(textsplitter.WithChunkSize(1000), textsplitter.WithChunkOverlap(200))
		} else {
			split = textsplitter.NewRecursiveCharacter(textsplitter.WithChunkSize(1000), textsplitter.WithChunkOverlap(200))
		}

		f, err := os.OpenFile(file, os.O_RDONLY, 0644)
		if err != nil {
			log.Fatalf("poop: %+v", err)
		}
		bytes, err := io.ReadAll(f)
		if err != nil {
			log.Printf("Failure to chunk file %s: %v", file, err)
		}
		chunks, err := split.SplitText(string(bytes))
		checkError(err)
		for _, chunk := range chunks {
			chunk_info := map[string]string{
				"id":   strings.Replace(file, fmt.Sprintf("working/%s/", uuid), "", 1),
				"text": chunk,
			}
			all_chunks = append(all_chunks, chunk_info)
		}

	}

	return all_chunks

}

func retrieve_db(query string, session string) (records []pinecone.Hit, err error) {

	log.Printf("attempting to search db")

	idxConnection, err := pineconeClient.Index(pinecone.NewIndexConnParams{Host: "debug-index-g9pn9ot.svc.aped-4627-b74a.pinecone.io", Namespace: session})
	if err != nil {
		return nil, err
	}

	res, err := idxConnection.SearchRecords(context.Background(), &pinecone.SearchRecordsRequest{
		Query: pinecone.SearchRecordsQuery{
			TopK: 40, // there's a lot of chunks
			Inputs: &map[string]interface{}{
				"text": query,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	log.Printf("Found %d hits", len(res.Result.Hits))
	return res.Result.Hits, nil
}

func assemble_messages(repo *repoSession, query string, session string) (convo []map[string]string) {

	convo = repo.convo // extract existing conversation

	convo = append(convo, map[string]string{"role": "user", "content": query})

	records, err := retrieve_db(query, session)
	checkError(err)
	message := "**Additional context extracted from the codebase:**\n"
	for _, record := range records {
		message += "From file " + record.Id + ":\n\t" + fmt.Sprintf("%v", record.Fields["text"]) + "\n"
	}

	convo = append(convo, map[string]string{"role": "system", "content": message})

	repo.convo = convo

	return convo
}

func call_api_llm(convo []map[string]string) (output string) {
	return ""
}

func call_llm(convo []map[string]string) (output string) {

	payload := map[string]interface{}{
		"conversation": convo,
	}

	var buf strings.Builder
	enc := json.NewEncoder(&buf)
	err := enc.Encode(payload)
	checkError(err)
	resp, err := http.Post("http://localhost:8000/generate", "application/json", strings.NewReader(buf.String()))
	checkError(err)

	// reuse byte? may be bad idea
	raw, err := io.ReadAll(resp.Body)
	checkError(err)
	log.Printf("%+v", raw)
	var content map[string]string
	err = json.Unmarshal(raw, &content)
	checkError(err)
	log.Printf("%+v", content)

	return content["response"]
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

	active_repos.m.Lock()
	defer active_repos.m.Unlock()
	active_repos.repolist = append(active_repos.repolist, session) // we can parse the url later

	if err != nil {
		log.Fatalf("[ERR] FAILED TO CLONE: ", url)
		return ""
	}

	return idtoken

}

// ** Route Functions

func queryRepo(w http.ResponseWriter, r *http.Request) {

	log.Print("working")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print("Error reading request body: ", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var data map[string]string
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Print("Error unmarshalling request body: ", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	query := data["text"]
	session := data["sessionid"]

	fmt.Println("Received prompt:", query)

	// write these out to the tempfile, truncate the tempfile

	active_repos.m.Lock()
	var repo *repoSession
	for i := range active_repos.repolist {
		if active_repos.repolist[i].token == session {
			repo = &active_repos.repolist[i]
			break
		}
	}
	assembled := assemble_messages(repo, query, session)
	active_repos.m.Unlock()
	if repo == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	output := call_llm(assembled)

	active_repos.m.Lock()
	repo.convo = append(repo.convo, map[string]string{"role": "assistant", "content": output})
	active_repos.m.Unlock()

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"output": string(output),
	}
	json.NewEncoder(w).Encode(response)
	log.Printf("%v", response)

}

func cleanUpRepo(w http.ResponseWriter, r *http.Request) {
	// now we will do all of this via the tokens as opposed to the string names

	//TODO define some generic error functions. Make this simple.

	body, err := io.ReadAll(r.Body)
	checkError(err)

	var data map[string]string
	err = json.Unmarshal(body, &data)
	checkError(err)

	token := data["id"]

	os.RemoveAll(fmt.Sprintf("./working/%s", token))

	// now to delete the namespace

	idxconnection, err := pineconeClient.Index(pinecone.NewIndexConnParams{Host: "debug-index-g9pn9ot.svc.aped-4627-b74a.pinecone.io", Namespace: token})
	checkError(err)

	err = idxconnection.DeleteAllVectorsInNamespace(context.Background())
	checkError(err)

	active_repos.m.Lock()
	defer active_repos.m.Unlock()
	for index, session := range active_repos.repolist {
		if session.token == token {

			active_repos.repolist = slice_remove(active_repos.repolist, index) // doesn't matter ig.
		}
	}
	log.Printf("Cleaned session with id %s", token)

	w.WriteHeader(http.StatusOK)
}

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

	// DEBUG
	// w.Header().Set("Content-Type", "application/json")
	// response := map[string]string{
	// 	"token":  token,
	// 	"output": string(token),
	// }
	// json.NewEncoder(w).Encode(response)
	// log.Printf("leaving early...")
	// return
	// DEBUG

	// upsert to pinecone db
	idxConnection, err := pineconeClient.Index(pinecone.NewIndexConnParams{Host: "debug-index-g9pn9ot.svc.aped-4627-b74a.pinecone.io", Namespace: token})
	if err != nil {
		log.Fatalf("Failed to create IndexConnection for Host: %v", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
		w.WriteHeader(http.StatusInternalServerError)               // aw yep
		w.Write([]byte("Failed to create IndexConnection for Host: " + err.Error()))
	}

	log.Printf("Chunking files...")
	chunks := run_textsplitter(token)

	log.Printf("Done!")

	var records []*pinecone.IntegratedRecord

	for _, text := range chunks {
		log.Printf("id is %s", text["id"])
		log.Printf("text is %s", text["text"])
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
		// error out here.
		w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
		w.WriteHeader(http.StatusInternalServerError)               // aw yep
		w.Write([]byte("Failed to chunk file (python error): " + err.Error()))
		return
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
	// json.NewEncoder(w).Encode(response)
	log.Printf("Successfully cloned and indexed repository: %s with token: %s", url, token)

	// call the llm endpoint, first call with all chunks

	init_convo := []map[string]string{
		{"role": "system", "content": "You are an AI assistant that is an expert in reading and understanding code. Your task is to answer questions, asked by the user, about a specified code base based on the content of its files. Give a short synopsis of the following: \n\t1: What is this code meant to do? \n\t2:How does it accomplish this? Refer to specific sections of code or practices used in the codebase \n\t3: What are the basic things one must know to be able to use the codebase effectively in their own projects?\n Please refer to the code that is given as many times as is needed, and provide as much detail as you feel is needed. You may further need to ask the user what specific functionality they want out of the codebase, and adjust your later responses accordingly."},
		{"role": "user", "content": "How can I get started with using this code repository for myself?"},
	}

	assembled_rag := "**Additional context extracted from the codebase:**\n"
	for _, chunk := range chunks {
		assembled_rag += "From file " + chunk["id"] + ":\n\t" + chunk["text"]
	}

	init_convo = append(init_convo, map[string]string{"role": "system", "content": assembled_rag})

	output := call_llm(init_convo)

	log.Println("All is well")

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"token":  token,
		"output": string(output),
	}
	json.NewEncoder(w).Encode(response)
	log.Printf("%v", response)

}

// ** Main function

func main() {

	// ctx := context.Background()

	pineconeClient = initPineconeClient()

	log.Println("Starting server on port 8080...") // TODO port from env var

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

	r.Get("/pulse", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server is alive!"))
	}) // TODO maybe more info?

	// declaring our routs
	r.Post("/initialExtract", initialExtraction)
	r.Post("/queryRepo", queryRepo)
	r.Post("/cleanup", cleanUpRepo)

	http.ListenAndServe(":8080", r)
}
