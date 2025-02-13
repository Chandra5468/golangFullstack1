package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Todo struct {
	ID        bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"` // if value is false or 0 donot add it to response
	Completed bool          `json:"completed"`
	Body      string        `json:"body"`
}

var collection *mongo.Collection

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
}

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("error loading .env file", err)
	}

	MONGODB_URI := os.Getenv("MONGODB_URI")

	clientOptions := options.Client().ApplyURI(MONGODB_URI)

	client, err := mongo.Connect(clientOptions) // In this case context.background is useful for some cancellations, timeouts etc..
	// NOT using context.Background as v2 of mongo driver does not need it.

	if err != nil {
		log.Fatal("Error connecting to Mongodb", err)
	}
	defer client.Disconnect(context.Background())
	err = client.Ping(context.Background(), nil)

	if err != nil {
		log.Fatal("Error while pinging", err)
	}

	log.Println("Connected to mongoDb")

	collection = client.Database("golang_DB").Collection("todos")

	server := http.NewServeMux()

	server.HandleFunc("GET /v1/api/todos", getTodos)
	server.HandleFunc("POST /v1/api/todos", createTodo)
	server.HandleFunc("PATCH /v1/api/todos", updateTodo)
	server.HandleFunc("DELETE /v1/api/todos", deleteTodo)
	http.ListenAndServe("0.0.0.0:"+os.Getenv("PORT"), server)
}

func getTodos(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	var todos []Todo

	cursor, err := collection.Find(context.Background(), bson.M{}) //bson.M is passing filter while find
	// defer cursor.Close(context.Background()) // Check this error
	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode("There is some error while finding the documents")
		cursor.Close(context.Background())
		return
	}

	for cursor.Next(context.Background()) {
		var todo Todo

		if err := cursor.Decode(&todo); err != nil {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode("struct malformed to decode mongo")
			cursor.Close(context.Background())
			return
		}

		todos = append(todos, todo)
	}

	cursor.Close(context.Background())
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(&todos)
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	todo := new(Todo)

	err := json.NewDecoder(r.Body).Decode(&todo)

	if err != nil || err == io.EOF {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode("struct malformed from req.body or req.body empty")
		return
	}

	insertResult, err := collection.InsertOne(context.Background(), &todo)

	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode("Insertion in mongo failed")
		log.Println(err.Error())
		return
	}

	todo.ID = insertResult.InsertedID.(bson.ObjectID) // this is type conversion from interface to objectId

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(todo)

}

func updateTodo(w http.ResponseWriter, r *http.Request) {
	// todo := new(Todo)
	// // err := json.NewDecoder(r.Body).Decode(&todo)
	// // if err != nil {
	// // 	w.WriteHeader(400)
	// // 	json.NewEncoder(w).Encode("Updation failed due to decoding")
	// // 	log.Println(err.Error())
	// // 	return
	// // }
	enableCors(&w)
	id := r.URL.Query().Get("id")

	_id, err := bson.ObjectIDFromHex(id)

	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode("Conversion error of id")
		log.Println(err.Error())
		return
	}
	filter := bson.M{"_id": _id}
	update := bson.M{"$set": bson.M{"completed": true}}
	_, err = collection.UpdateOne(context.Background(), filter, update)

	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode("Updation failed")
		log.Println(err.Error())
		return
	}

	w.WriteHeader(202)
	json.NewEncoder(w).Encode("Successfully updated")
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	id := r.URL.Query().Get("id")

	_id, err := bson.ObjectIDFromHex(id)

	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode("Conversion error of id")
		log.Println(err.Error())
		return
	}

	filter := bson.M{"_id": _id}
	_, err = collection.DeleteOne(context.Background(), filter)

	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode("Deletion failed")
		log.Println(err.Error())
		return
	}

	w.WriteHeader(202)
	json.NewEncoder(w).Encode("Successfully deleted")
}
