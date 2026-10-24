package ExampleJson

import (
	"fmt"

	"github.com/jcbowen/jcbaseGo/component/helper"
)

type itemStruct struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

type dataStruct struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Prices      []float64    `json:"prices"`
	Names       []string     `json:"names"`
	Items       []itemStruct `json:"items"`
}

func JsonFileToStruct() {
	testData := dataStruct{}

	result := helper.JsonFile("../example/json/example.json").ToStruct(&testData)
	if result.HasError() {
		fmt.Println(result.Errors())
		return
	}

	fmt.Println("testData:", testData)
}

func JsonFileToString() {
	testData := ""

	result := helper.JsonFile("../example/json/example.json").ToString(&testData)
	if result.HasError() {
		fmt.Println(result.Errors())
		return
	}

	fmt.Println("testData:", testData)
}

func JsonFileToMap() {
	testData := map[string]interface{}{}

	result := helper.JsonFile("../example/json/example.json").ToMap(&testData)
	if result.HasError() {
		fmt.Println(result.Errors())
		return
	}

	fmt.Println("testData:", testData)
}
