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

func isElementMatch(token html.Token, tag string, attrkey string, attrvalue string) bool {
	if token.Data != tag {
		return false
	}
	for _, a := range token.Attr {
		if a.Key == attrkey && a.Val == attrvalue {
			return true
		}
	}
	return false
}

// Fetch language list from https://github.com/trending
func (t *GHTrends) FetchLanguagesList() ([]Language, error) {
	var languages []Language

	resp, err := t.client.Get(t.BaseURL.String() + urlTrending)
	if err != nil {
		return languages, err
	}
	defer resp.Body.Close()

	tokenizer := html.NewTokenizer(resp.Body)
	depth := 0
	languageListDepth, inLanguageNode := 0, false
	for {
		tokenType := tokenizer.Next()

		if tokenType == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				break
			}
			return languages, err
		}
		// track node depth
		switch tokenType {
		case html.StartTagToken:
			depth++
		case html.EndTagToken:
			depth--
		}
		// only single top language block node
		if languageListDepth > 0 && depth < languageListDepth {
			break
		}

		token := tokenizer.Token()

		// find common language block on the page
		if isElementMatch(token, "details", "id", "select-menu-language") {
			languageListDepth = depth
			inLanguageNode = true
		}
		// find concrete language
		if inLanguageNode && isElementMatch(token, "span", "class", "select-menu-item-text") {
			tokenizer.Next()
			langName := strings.TrimSpace(tokenizer.Token().Data)
			languages = append(languages, Language{Name: langName})
		}
	}
	return languages, nil
}

func (*GHTrends) FetchRepos(timeframe Time, lang Language) []Repo {
	var projects []Repo
	return projects
}
