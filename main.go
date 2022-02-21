package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/thedevsaddam/renderer"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var rnd *renderer.Render
var db *mgo.Database

const (
	hostName		string ="localhost:27017"
	dbName			string="demo_todo"
	collectionName 	string = "todo"
	port 			string = ":9000"
)


type (
	todoModel struct {
		ID        bson.ObjectId `bson:"_id,omitempty"`
		Title     string        `bson:"title"`
		Completed bool          `bson:"completed"`
		createdAt time.Time     `bson:"createdAt"`
 }
	todo struct {
		ID        string    `json:"id,omitempty"`
		Title     string    `json:"title"`
		Completed bool    `json:"completed"`
		createdAt time.Time `json:"created_at"`
	}
)

func init () {
	rnd = renderer.New()
	sess,err :=mgo.Dial(hostName)
	checkError(err)
	sess.SetMode(mgo.Monotonic,true)
	db = sess.DB(dbName)
}

func homeHandler (w http.ResponseWriter,r *http.Request){
	err := rnd.Template(w,http.StatusOK,[]string{"/static/home.tpl"},nil)
	checkError(err)
}
func fetchTodos(w http.ResponseWriter ,r *http.Request){
todos := []todoModel{}
if err := db.C(collectionName).Find(bson.M{}).All(&todos);err!=nil{
checkError(err)
rnd.JSON(w, http.StatusProcessing,renderer.M{
	"message":"Failed to fetch todo",
	"error":err,
})
return
}
todoList := []todo{}

for _,t :=range todos {
	todoList = append(todoList, todo{
		ID: t.ID.Hex(),
		Title: t.Title,
		Completed: t.Completed,
		createdAt: t.createdAt,
	})
}
rnd.JSON(w,http.StatusOK,renderer.M{
	"data":todoList,
})
}

func createTodo( w http.ResponseWriter,r *http.Request){
var t todo
if err := json.NewDecoder(r.Body).Decode(&t);err!=nil{
rnd.JSON(w,http.StatusProcessing,err)
return
}

if t.Title == ""{
	rnd.JSON(w,http.StatusBadRequest,renderer.M{
		"messaage":"title is empty",
	})
}
tm :=todoModel{
	ID: bson.NewObjectId(),
	Title: t.Title,
	Completed: false,
	createdAt: time.Now(),
}
if err :=db.C(collectionName).Insert(&tm);err!=nil{
	rnd.JSON(w,http.StatusProcessing,renderer.M{
		"message" : "Failed to Insert To Do object",
		"err" : err,
	})
	return
}

rnd.JSON(w,http.StatusCreated,renderer.M{
	"message":"todo created successfully",
	"todo_it":tm.ID.Hex(),
})

}
func updateTodo(){

}
func deleteTodo(w http.ResponseWriter,r *http.Request){
	id:=strings.TrimSpace(chi.URLParam(r,"id"))
	if !bson.IsObjectIdHex(id){
		rnd.JSON(w,http.StatusProcessing,renderer.M{
			"message":"Todo with that Id does not exists",
		})
		return
}
if err:= db.C(collectionName).RemoveId(bson.ObjectIdHex(id));err!=nil{

		rnd.JSON(w,http.StatusProcessing,renderer.M{
			"message":"Deleting Todo Error",
			"error":err,
		})
		return
}
rnd.JSON(w,http.StatusOK,renderer.M{
	"message":"To deleted successfully",
})

}

func main () {
stopChan := make(chan os.Signal)
signal.Notify(stopChan,os.Interrupt)
r := chi.NewRouter()
r.Use(middleware.Logger)
r.Get("/",homeHandler)
r.Mount("/todo",todoHandlers())

srv := http.Server{
	Addr: port,
	Handler: r,
	ReadTimeout: 60*time.Second,
	WriteTimeout: 60*time.Second,
	IdleTimeout: 60*time.Second,
}

go func ()  {
	log.Println("listening on port , ",port)
	if err :=srv.ListenAndServe(); err!=nil{
		log.Printf("listen : %s \n",err)
	}
}()
<-stopChan
log.Println("shutting down server")
ctx,cancel :=context.WithTimeout(context.Background(),5*time.Second)
srv.Shutdown(ctx)
defer cancel(
)



}

func todoHandlers () http.Handler {
rg := chi.NewRouter()
return rg.Group(func(r chi.Router) {
	r.Get("/",fetchTodos)
	// r.Post("/",createTodo)
	// r.Put("/{id}",updateTodo)
	// r.Delete("/{id}",deleteTodo)
})
}



func checkError (err error) {
	if err!=nil{
		log.Fatal(err)
	}
}