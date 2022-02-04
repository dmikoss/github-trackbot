package bot

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var (
	mux    *http.ServeMux
	client *GHTrends
	server *httptest.Server
)

// start fake http server to work with our client
func start() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	client = New(http.DefaultClient)
	client.BaseURL, _ = url.Parse(server.URL)
}

// teardown
func finish() {
	server.Close()
}

// return readed file content
func getFileContent(file string) []byte {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return []byte{}
	}
	return content
}

func TestFetchLanguagesCase(t *testing.T) {
	start()
	defer finish()

	mux.HandleFunc("/trending", func(w http.ResponseWriter, r *http.Request) {
		website := getFileContent("../testdata/github-trending.html")
		fmt.Fprint(w, string(website))
	})

	languages, err := client.FetchLanguagesList()
	if err != nil {
		t.Errorf("Cant FetchLanguagesList")
	}

	if len(languages) != 611 {
		t.Errorf("FetchLanguagesList returned %+v languages, expexted 611", len(languages))
	}

	if languages[0].Name != "C++" {
		t.Errorf("languages[0] == %s, expexted C++", languages[0].Name)
	}
}

func TestFetchReposCase(t *testing.T) {
	start()
	defer finish()

	// empty html (404) - no registered http handlers
	repos, err := client.FetchRepos(TimeDaily, Language{})
	if err != nil {
		t.Errorf("Cant FetchRepos")
	}
	if len(repos) != 0 {
		t.Errorf("FetchRepos returned %+v repos, expexted 0", len(repos))
	}

	// all repos (no specific language)
	mux.HandleFunc("/trending", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, string(getFileContent("../testdata/github-trending.html")))
	})
	// only go language
	mux.HandleFunc("/trending/go", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, string(getFileContent("../testdata/github-trending-go.html")))
	})

	// all lang repos
	repos, err = client.FetchRepos(TimeDaily, Language{})
	if err != nil {
		t.Errorf("Cant FetchRepos")
	}
	if len(repos) != 22 {
		t.Errorf("FetchRepos returned %+v repos, expexted 22", len(repos))
	}

	// golang repos
	repos, err = client.FetchRepos(TimeDaily, Language{Name: "go"})
	if err != nil {
		t.Errorf("Cant FetchRepos")
	}
	if len(repos) != 25 {
		t.Errorf("FetchRepos returned %+v golang repos, expexted 25", len(repos))
	}
}
