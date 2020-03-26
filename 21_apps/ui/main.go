package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
)

var (
	firestoreClient *firestore.Client
	storageClient   *storage.Client
	tmpl            *template.Template
)

func main() {
	ctx := context.Background()
	var err error

	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	firestoreClient, err = firestore.NewClient(ctx, "aerobic-botany-270918")
	if err != nil {
		log.Fatal(err)
	}

	tmpl, err = template.New("div").Parse(imageTmpl)
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler)

	r.HandleFunc("/images/{key}", imageHandler)
	log.Fatal(http.ListenAndServe(":8080", r))

}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, page)

	iter := firestoreClient.Collection("khappygo").Documents(r.Context())
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}
		s := newDisplayable(doc.Data())
		tmpl.Execute(w, s)
	}
	io.WriteString(w, "</html>")
}
func imageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Println(vars)
	image := vars["key"]
	log.Println(image)
	bucketRdr, err := storageClient.Bucket("aerobic-botany-270918").Object("processed/" + image).NewReader(r.Context())
	if err != nil {
		log.Println(err)
		return
	}
	defer bucketRdr.Close()
	io.Copy(w, bucketRdr)
}

type sentiments []sentiment
type sentiment struct {
	Sentiment string
	Value     float64
}

func newSentiments(elements map[string]interface{}) sentiments {
	output := make([]sentiment, 0, len(elements))
	for k, v := range elements {
		if val, ok := v.(float64); ok {
			output = append(output, sentiment{k, val})
		}
	}
	return sentiments(output)
}

func (s sentiment) String() string {
	return fmt.Sprintf("%v:%2.2f", s.Sentiment, s.Value)
}

func (s sentiments) Len() int           { return len(s) }
func (s sentiments) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sentiments) Less(i, j int) bool { return s[i].Value > s[j].Value }

const imageTmpl = `
<div class="gallery">
  <a target="_blank" href="{{.Image}}">
    <img src="{{.Image}}" alt="{{.Image}}" width="600" height="400">
  </a>
  <div class="desc">{{.Desc}}</div>
</div>
`

type displayable struct {
	Image string
	Desc  string
}

func newDisplayable(elements map[string]interface{}) displayable {
	s := newSentiments(elements)
	sort.Sort(s)
	return displayable{
		Image: strings.Replace(strings.Replace(elements["Src"].(string), "gs://aerobic-botany-270918/processed", "images", 1), `"`, ``, -1),
		Desc:  fmt.Sprintf("%v\n%v", []sentiment(s)[0], []sentiment(s)[1]),
	}
}
