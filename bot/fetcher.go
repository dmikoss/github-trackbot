package bot

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

const (
	urlBase     = "https://github.com"
	urlTrending = "/trending"
)

type TrendTime int

const (
	TimeDaily TrendTime = 0 // daily
	TimeWeek  TrendTime = 1 // weekly
	TimeMonth TrendTime = 2 // monthly
)

type Fetcher struct {
	client  *http.Client
	BaseURL *url.URL
}

func NewFetcher(httpclient *http.Client) *Fetcher {
	serverUrl, _ := url.Parse(urlBase)
	return &Fetcher{
		client:  httpclient,
		BaseURL: serverUrl}
}

type Language struct {
	Name string
}

type TimeInterval struct {
	Type     string
	Repolist []Repo
}

type Repo struct {
	NameURL  string
	Desc     string
	Language string
	Stars    [4]int // daily, weekly, monthly, all
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

func getAttrValue(token html.Token, attrkey string) string {
	for _, a := range token.Attr {
		if a.Key == attrkey {
			return a.Val
		}
	}
	return ""
}

// Fetch language list from https://github.com/trending
func (t *Fetcher) FetchLanguagesList() ([]Language, error) {
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

// Fetch trending repo list from https://github.com/trending
// if lang.Name not set when fetch all trending language repos
func (t *Fetcher) FetchRepos(timeframe TrendTime, lang Language) ([]Repo, error) {
	var projects []Repo
	scopearr := [3]string{"daily", "weekly", "monthly"}
	scope := scopearr[timeframe]

	// prepare query path with parameters
	qstr := t.BaseURL.String() + urlTrending + "/" + strings.ReplaceAll(lang.Name, " ", "-")
	qstr = strings.TrimSuffix(qstr, "/")
	querypath, err := url.Parse(qstr)
	if err != nil {
		return projects, err
	}
	q := querypath.Query()
	q.Set("since", scope)
	querypath.RawQuery = q.Encode()

	// http get
	resp, err := t.client.Get(querypath.String())
	if err != nil {
		return projects, err
	}
	defer resp.Body.Close()

	beginRepoBlock := false
	// parsing body to retrieve repo information
	tokenizer := html.NewTokenizer(resp.Body)
	for {
		tokenType := tokenizer.Next()

		if tokenType == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				break
			}
		}
		token := tokenizer.Token()

		// find common repos block on the page
		if isElementMatch(token, "article", "class", "Box-row") {
			beginRepoBlock = true
		}

		if beginRepoBlock && isElementMatch(token, "h1", "class", "h3 lh-condensed") {
			var repo Repo
			// search <a> tag with href
			for tokenizer.Next() != html.ErrorToken {
				token = tokenizer.Token()
				if token.Data == "a" {
					repo.NameURL = getAttrValue(token, "href")
					break
				}
			}
			// search <p> tag with repo description
			for tokenizer.Next() != html.ErrorToken {
				if tokenizer.Token().Data == "p" {
					tokenizer.Next()
					token = tokenizer.Token()
					repo.Desc = strings.TrimSpace(token.Data)
					break
				}
			}

			// search <span> tag with language
			for tokenizer.Next() != html.ErrorToken {
				if isElementMatch(tokenizer.Token(), "span", "itemprop", "programmingLanguage") {
					tokenizer.Next()
					token = tokenizer.Token()
					repo.Language = strings.TrimSpace(token.Data)
					break
				}
			}

			// search closing svg tag
			for tokenizer.Next() != html.ErrorToken {
				token = tokenizer.Token()
				if token.Type == html.EndTagToken && token.Data == "svg" {
					break
				}
			}

			// search text node with overall stars count
			for tokenizer.Next() != html.ErrorToken {
				token = tokenizer.Token()
				if token.Type == html.TextToken {
					stars := strings.ReplaceAll(token.Data, ",", "")
					stars = strings.TrimSpace(stars)
					repo.Stars[3], _ = strconv.Atoi(stars)
					break
				}
			}

			// search <span> tag with stars since "period"
			for tokenizer.Next() != html.ErrorToken {
				if isElementMatch(tokenizer.Token(), "span", "class", "d-inline-block float-sm-right") {
					break
				}
			}

			// search closing svg tag
			for tokenizer.Next() != html.ErrorToken {
				token = tokenizer.Token()
				if token.Type == html.EndTagToken && token.Data == "svg" {
					break
				}
			}

			// before string util func
			before := func(value string, a string) string {
				// Get substring before a string.
				pos := strings.Index(value, a)
				if pos == -1 {
					return ""
				}
				return value[0:pos]
			}

			// search text node with overall stars count
			for tokenizer.Next() != html.ErrorToken {
				token = tokenizer.Token()
				if token.Type == html.TextToken {
					starssince := strings.TrimSpace(token.Data)
					starssince = strings.ReplaceAll(starssince, ",", "")
					starssince = before(starssince, " stars")
					repo.Stars[timeframe], _ = strconv.Atoi(starssince)
					break
				}
			}
			projects = append(projects, repo)
		}

		if beginRepoBlock && tokenType == html.EndTagToken && token.Data == "article" {
			beginRepoBlock = false
		}
	}
	return projects, nil
}
