package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
)

var (
	firestoreClient *firestore.Client
	storageClient   *storage.Client
)

func main() {
	ctx := context.Background()
	var err error

	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	firestoreClient, err = firestore.NewClient(ctx, "khappygo")
	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)

	r.HandleFunc("/images/{key}", ImageHandler)
	log.Fatal(http.ListenAndServe(":8080", r))

}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, page)

	iter := firestoreClient.Collection("kahppygo").Documents(r.Context())
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}
		fmt.Fprintf(w, "%#v", doc.Data())
	}
	io.WriteString(w, "</html>")
}
func ImageHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, `{"alive": true}`)
}
