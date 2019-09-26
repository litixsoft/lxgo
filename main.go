package main

import (
	"fmt"
	"regexp"
)

func main() {

	//h1 := "mongodb://127.0.0.1:27017"
	//h2 := "mongodb://127.0.0.1"
	//h3 := "mongodb://localhost:27017"
	//h4 := "mongodb://localhost"
	//
	//cut1 := strings.ReplaceAll(h1, "mongo")

	re := regexp.MustCompile(`urogister-backend-test`)
	fmt.Println(re.FindAllString("mongodb://127.0.0.1", -1))
	//fmt.Printf("%q\n", re.FindAll([]byte(`mongodb://localhost  mongodb://localhost:27017 mongodb://127.0.0.1 mongodb://127.0.0.1:27017`), -1))
}
