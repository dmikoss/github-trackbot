package model

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

const (
	TimeToday = "daily"
	TimeWeek  = "weekly"
	TimeMonth = "monthly"

	urlBase     = "https://github.com"
	urlTrending = "/trending"
)

type Time string

type Language struct {
	Name string
}

type TimeInterval struct {
	Type     string
	Repolist []Repo
}

type Repo struct {
	Name string
	Desc string
	Url  string
}

type GHTrends struct {
	client  *http.Client
	BaseURL *url.URL
}

func New(httpclient *http.Client) *GHTrends {
	serverUrl, _ := url.Parse(urlBase)
	return &GHTrends{
		client:  httpclient,
		BaseURL: serverUrl}
}

func isElementMenuLanguage(token html.Token) bool {
	if token.Data != "details" {
		return false
	}
	for _, a := range token.Attr {
		if a.Key == "id" && a.Val == "select-menu-language" {
			return true
		}
	}
	return false
}

func isElementLanguage(token html.Token) bool {
	if token.Data != "span" {
		return false
	}
	for _, a := range token.Attr {
		if a.Key == "class" && a.Val == "select-menu-item-text" {
			return true
		}
	}
	return false
}

func (t *GHTrends) FetchLanguagesList() ([]Language, error) {
	var languages []Language

	resp, err := t.client.Get(t.BaseURL.String() + urlTrending)
	if err != nil {
		return languages, err
	}

	tokenizer := html.NewTokenizer(resp.Body)
	depth := 0
	languageListDepth := 0
	inLanguageNode := false
	for {
		tokenType := tokenizer.Next()

		if tokenType == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				break
			}
			return languages, err
		}

		if tokenType == html.StartTagToken {
			depth++
		}
		if tokenType == html.EndTagToken {
			depth--
		}
		if languageListDepth > 0 && depth < languageListDepth {
			inLanguageNode = false
		}

		token := tokenizer.Token()

		if isElementMenuLanguage(token) {
			languageListDepth = depth
			inLanguageNode = true
		}
		if isElementLanguage(token) {
			tokenizer.Next()
			if inLanguageNode {
				langName := strings.Trim(tokenizer.Token().Data, "\n\r\t ")
				languages = append(languages, Language{Name: langName})
			}
		}

	}
	defer resp.Body.Close()
	return languages, nil
}

func (*GHTrends) FetchRepos(timeframe Time, lang Language) []Repo {
	var projects []Repo
	return projects
}
