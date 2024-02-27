package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	client *mongo.Client
}

func NewServer(c *mongo.Client) *Server {
	return &Server{
		client: c,
	}
}

func (s *Server) getJokeByID(c *gin.Context) {
	id := c.Param("id")
	coll := s.client.Database("dadjokes").Collection("jokes")
	opts := options.FindOne().SetSort(bson.D{{"id", 1}})
	var result bson.M
	err := coll.FindOne(context.TODO(), bson.D{{"id", id}}, opts).Decode(&result)
	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in
		// the collection.
		if errors.Is(err, mongo.ErrNoDocuments) {
			fmt.Println(err)
			return
		}
		log.Fatal(err)
	}
	c.IndentedJSON(http.StatusOK, result)
}

func (s *Server) postJoke(c *gin.Context) {
	var newJoke bson.M
	coll := s.client.Database("dadjokes").Collection("jokes")

	if err := c.BindJSON(&newJoke); err != nil {
		return
	}
	res, err := coll.InsertOne(context.Background(), newJoke)
	if err != nil {
		return
	}
	c.IndentedJSON(http.StatusOK, bson.M{"_id": res.InsertedID})
}

func (s *Server) getJokes(c *gin.Context) {
	coll := s.client.Database("dadjokes").Collection("jokes")

	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())

	results := []bson.M{}

	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}
	c.IndentedJSON(http.StatusOK, results)
}

func main() {
	mongodb_uri := os.Getenv("MONGODB_URI")
	clientOpts := options.Client().ApplyURI(mongodb_uri)
	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()
	server := NewServer(client)

	router := gin.Default()
	router.GET("/jokes", server.getJokes)
	router.GET("/jokes/:id", server.getJokeByID)
	router.POST("/jokes", server.postJoke)
	router.Run(":8080")
}
