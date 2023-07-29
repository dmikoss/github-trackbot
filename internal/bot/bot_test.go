package bot

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

var (
	mux           *http.ServeMux
	fetcherClient *Fetcher
	server        *httptest.Server
)

// start fake http server to work with our client
func start() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	fetcherClient = NewFetcher(context.Background(), http.DefaultClient)
	fetcherClient.BaseURL, _ = url.Parse(server.URL)
}

// teardown
func finish() {
	server.Close()
}

// return readed file content
func getFileContent(file string) []byte {
	content, err := os.ReadFile(file)
	if err != nil {
		return []byte{}
	}
	return content
}

func TestFetchLanguagesCase(t *testing.T) {
	start()
	defer finish()

	mux.HandleFunc("/trending", func(w http.ResponseWriter, r *http.Request) {
		website := getFileContent("../../testdata/github-trending.html")
		fmt.Fprint(w, string(website))
	})

	languages, err := fetcherClient.FetchLanguagesList()
	if err != nil {
		t.Errorf("Cant FetchLanguagesList")
	}

	if len(languages) != 708 {
		t.Errorf("FetchLanguagesList returned %+v languages, expexted 708", len(languages))
	}

	if languages[0].Name != "Unknown languages" {
		t.Errorf("languages[0] == %s, expexted Unknown languages", languages[0].Name)
	}
}

func TestFetchReposCase(t *testing.T) {
	start()
	defer finish()

	// empty html (404) - no registered http handlers
	repos, err := fetcherClient.FetchRepos(TimeDaily, Language{})
	if err != nil {
		t.Errorf("Cant FetchRepos")
	}
	if len(repos) != 0 {
		t.Errorf("FetchRepos returned %+v repos, expexted 0", len(repos))
	}

	// all repos (no specific language)
	mux.HandleFunc("/trending", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, string(getFileContent("../../testdata/github-trending.html")))
	})
	// only go language
	mux.HandleFunc("/trending/go", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, string(getFileContent("../../testdata/github-trending-go.html")))
	})

	// all lang repos
	repos, err = fetcherClient.FetchRepos(TimeDaily, Language{})
	if err != nil {
		t.Errorf("Cant FetchRepos")
	}
	if len(repos) != 22 {
		t.Errorf("FetchRepos returned %+v repos, expexted 22", len(repos))
	}

	// golang repos
	repos, err = fetcherClient.FetchRepos(TimeDaily, Language{Name: "go"})
	if err != nil {
		t.Errorf("Cant FetchRepos")
	}
	if len(repos) != 25 {
		t.Errorf("FetchRepos returned %+v golang repos, expexted 25", len(repos))
	}
}
