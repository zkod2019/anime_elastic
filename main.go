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

	api "github.com/elastic/go-elasticsearch/v8/esapi"
)

var elasticClient *elasticsearch8.Client

func main() {
	var err error
	elasticClient, err = elasticsearch8.NewClient(elasticsearch8.Config{
		Addresses: []string{"http://localhost:9200"},
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
		Index:      "anime",
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
		elasticClient.Search.WithIndex("anime"),
		elasticClient.Search.WithBody(strings.NewReader(`{"query": {"match_all": {}}}`)),
		elasticClient.Search.WithTrackTotalHits(true),
		elasticClient.Search.WithPretty(),
	)
	if err != nil {
		return nil, err
	}
	var animeRes []Anime
	var elasticRes ElasticResponse
	err = json.NewDecoder(response.Body).Decode(&elasticRes)
	if err != nil {
		return nil, err
	}
	for i := range elasticRes.Hits.Hits {
		fmt.Println(elasticRes.Hits.Hits[i].ID)
		animeRes = append(animeRes, elasticRes.Hits.Hits[i].Source)
	}

	return animeRes, nil
}

func Update(id string, cmd Anime) error {
	byts, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	req := api.UpdateRequest{
		Index:      "anime",
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
		Index:      "anime",
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
