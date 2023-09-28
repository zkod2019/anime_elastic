package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	elasticsearch8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	// "github.com/olivere/elastic"
	api "github.com/elastic/go-elasticsearch/v8/esapi"
)

var elasticClient *elasticsearch8.Client

func main() {
	var err error
	elasticClient, err = elasticsearch8.NewClient(elasticsearch8.Config{
		Addresses: []string{"http://localhost:9200"},
		Username:  "elastic",
		Password:  "uPn0EEjhowA=kkhY03wU",
	})

	if err != nil {
		panic(err)
	}

	router := gin.Default()
	api := router.Group("anime")
	api.POST("", func(ctx *gin.Context) {
		body := Anime{}
		err := ctx.BindJSON(&body)
		if err != nil {
			ctx.JSON(400, err)
			return
		}
		err = Create(body)
		if err != nil {
			fmt.Println(err)
			ctx.JSON(400, err)
			return
		}
		ctx.JSON(201, nil)

	})
	api.GET("", func(ctx *gin.Context) {
		list, err := Read()
		if err != nil {
			fmt.Println(err)
			ctx.JSON(400, err)
			return
		}
		ctx.JSON(200, list)
	})
	api.PUT("/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		body := Anime{}
		err := ctx.BindJSON(&body)
		if err != nil {
			ctx.JSON(400, err)
		}
		err = Update(id, body)
		if err != nil {
			fmt.Println(err)
			ctx.JSON(400, err)
			return
		}
		ctx.JSON(202, nil)
	})
	api.DELETE("/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		err := Delete(id)
		if err != nil {
			fmt.Println(err)
			ctx.JSON(400, err)
			return
		}
		ctx.JSON(204, nil)
	})
	api.POST("/search", func(ctx *gin.Context) {

	})
	router.Run(":8080")
}

func Create(cmd Anime) error {
	byts, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	req := api.IndexRequest{
		Index:      "Anime",
		DocumentID: uuid.NewString(),
		Body:       bytes.NewReader(byts),
		Refresh:    "true",
	}
	_, err = req.Do(context.TODO(), elasticClient)
	if err != nil {
		return err
	}

	return nil
}

func Read() ([]Anime, error) { // read all docs
	response, err := elasticClient.Search(
		elasticClient.Search.WithContext(context.TODO()),
		elasticClient.Search.WithIndex("Anime"),
		elasticClient.Search.WithBody(strings.NewReader(`{"query": {"match_all": {}}}`)),
		elasticClient.Search.WithTrackTotalHits(true),
		elasticClient.Search.WithPretty(),
	)
	if err != nil {
		return nil, err
	}

	res := map[string]interface{}{}
	err = json.NewDecoder(response.Body).Decode(&res)
	if err != nil {
		return nil, err
	}
	fmt.Println(res)

	return nil, nil

	// return []Anime{
	// 	{
	// 		Id:       "1",
	// 		Title:    "Jujutsu Kaisen",
	// 		Author:   "Gege Akutami",
	// 		Season:   1,
	// 		Episodes: 24,
	// 	},
	// 	{
	// 		Id:       "2",
	// 		Title:    "Attack on Titan",
	// 		Author:   "Hajime Isayama",
	// 		Season:   4,
	// 		Episodes: 88,
	// 	},
	// 	{
	// 		Id:       "3",
	// 		Title:    "Nana",
	// 		Author:   "Ai Yazawa",
	// 		Season:   1,
	// 		Episodes: 47,
	// 	},
	// }, nil
}

func Update(id string, cmd Anime) error {
	byts, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	req := api.UpdateRequest{
		Index:      "Anime",
		DocumentID: id,
		Body:       bytes.NewReader(byts),
		Refresh:    "true",
	}

	_, err = req.Do(context.TODO(), elasticClient)
	if err != nil {
		return err
	}

	return nil
}

func Delete(id string) error {
	req := api.DeleteRequest{
		Index:      "Anime",
		DocumentID: id,
	}
	_, err := req.Do(context.TODO(), elasticClient)
	if err != nil {
		return err
	}

	return nil
}

func Query(query string) ([]Anime, error) { // search for specific doc
	return nil, nil
}
