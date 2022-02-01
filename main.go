package main

import (
	"fmt"
	"net/http"

	"github.com/dmikoss/GithubTrackBot/model"
)

func main() {

	trends := model.New(http.DefaultClient)
	fmt.Println(trends.FetchLanguagesList())
}
