package ExampleJson

import (
	"fmt"
	"github.com/jcbowen/jcbaseGo/helper"
)

type itemStruct struct {
	Id    *int   `json:"id"`
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

var data = dataStruct{
	Name:        "test computers",
	Description: "List of computer products",
	Prices:      []float64{2400, 2100, 1200, 400.87, 89.90, 150.10},
	Names:       []string{"John Doe", "Jane Doe", "Tom", "Jerry", "Nicolas", "Abby"},
	Items:       []itemStruct{{Id: helper.Int(1), Name: "MacBook Pro 13 inch retina", Price: 1350}, {Id: helper.Int(2), Name: "MacBook Pro 15 inch retina", Price: 1700}, {Id: helper.Int(3), Name: "Sony VAIO", Price: 1200}, {Id: helper.Int(4), Name: "Fujitsu", Price: 850}, {Id: nil, Name: "HP core i3 SSD", Price: 850}},
}

func ExampleJsonFileToStruct() {
	testData := dataStruct{}

	result := helper.JsonFile("../example/json/example.json").ToStruct(&testData)
	if result.HasError() {
		fmt.Println(result.Errors())
		return
	}

	fmt.Println("testData:", testData)
}

func ExampleJsonFileToString() {
	testData := ""

	result := helper.JsonFile("../example/json/example.json").ToString(&testData)
	if result.HasError() {
		fmt.Println(result.Errors())
		return
	}

	fmt.Println("testData:", testData)
}
