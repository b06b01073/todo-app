package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/rs/xid"

	"todo-app/hash"
	"todo-app/jwtService"
	"todo-app/models"
)

var (
	host     = "localhost"
	port     = 5432
	user     = os.Getenv("POSTGRES_USER")
	password = os.Getenv("POSTGRES_PASSWORD")
	dbname   = os.Getenv("POSTGRES_DB")
)

type Todo models.Todo
type User models.User

var DB *sql.DB
var err error

func init() {
	psqlConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	DB, err = sql.Open("postgres", psqlConn)

	if err != nil {
		panic(err)
	}

	// defer db.Close(), this line should run in main.go

	fmt.Println("db connected")
}

func SignUpHandler(c *gin.Context) {
	var user User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid inputs",
		})
		return
	}

	if user.Username == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid inputs",
		})
		return
	}

	// save to db
	hashedPassword, err := hash.HashPassword(user.Password)
	if err != nil {
		panic(err)
	}

	insertStmt, err := DB.Prepare("INSERT INTO users(username, password) VALUES($1, $2)")
	if err != nil {
		panic(err)
	}

	defer insertStmt.Close()

	_, err = insertStmt.Exec(user.Username, hashedPassword)
	if err != nil {
		panic(err)
	}

	JWT, err := jwtService.GenerateJWT(user.Username)
	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": JWT,
	})
}

func SignInHandler(c *gin.Context) {
	var user User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid inputs",
		})
		return
	}

	if user.Password == "" || user.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid inputs",
		})
		return
	}

	// check the password in db
	row, err := DB.Query("SELECT password FROM users WHERE username=$1", user.Username)
	if err != nil {
		panic(err)
	}
	defer row.Close()

	var hashedPassword string
	if row.Next() {
		err := row.Scan(&hashedPassword)
		if err != nil {
			panic(err)
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "sign in failed",
		})
		return
	}

	if hash.CheckPasswordHash(hashedPassword, user.Password) {
		JWT, err := jwtService.GenerateJWT(user.Username)
		if err != nil {
			panic(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"token": JWT,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "sign in failed",
		})
	}
}

// In the following api, username can only be extracted from gin.Context(which is verify by jwt checker)
func ListTodosHandler(c *gin.Context) {
	// an alternative way is to use c.MustGet(but in this case MustGet returns an interface, we still need to do type assertion)
	username := c.MustGet("username").(string)
	complete := c.Query("complete")

	if complete != "" && complete != "true" && complete != "false" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid query string",
		})
		return
	}

	var queryStmt string
	if complete == "" {
		queryStmt = "SELECT todo, id, complete FROM todos WHERE username = $1"
	} else if complete == "false" {
		queryStmt = "SELECT todo, id, complete FROM todos WHERE complete = 'false' AND username = $1"
	} else {
		// complete == "true"
		queryStmt = "SELECT todo, id, complete FROM todos WHERE complete = 'true' AND username = $1"
	}

	rows, err := DB.Query(queryStmt, username)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var todos = make([]Todo, 0, 10)

	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.Task, &todo.Id, &todo.Complete)
		if err != nil {
			panic(err)
		}

		todos = append(todos, todo)
	}

	c.JSON(http.StatusOK, gin.H{
		"todos": todos,
	})
}

func AddTodoHandler(c *gin.Context) {
	username := c.MustGet("username").(string)

	var todo Todo

	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "wrong todo format",
		})
		return
	}

	// By default, todo.Complete is set to false, using c.MUSTGET()
	if todo.Task == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "wrong todo format",
		})
		return
	}

	// save task to db, and give this task an id
	insertStmt, err := DB.Prepare("INSERT INTO todos(id, username, complete, todo) VALUES($1, $2, $3, $4)")
	if err != nil {
		panic(err)
	}
	defer insertStmt.Close()

	id := xid.New()

	_, err = insertStmt.Exec(id, username, todo.Complete, todo.Task)
	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": id,
	})
}

func GetTodoHandler(c *gin.Context) {
	username := c.MustGet("username").(string)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "no id specified",
		})
		return
	}

	row, err := DB.Query("SELECT id, complete, todo FROM todos WHERE id = $1 AND username=$2", id, username)
	if err != nil {
		panic(err)
	}

	defer row.Close()

	if row.Next() {
		var todo Todo
		err = row.Scan(&todo.Id, &todo.Complete, &todo.Task)
		if err != nil {
			panic(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"id":       todo.Id,
			"username": username,
			"complete": todo.Complete,
			"todo":     todo.Task,
		})
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "todo not found",
		})
	}
}

func UpdateTodoHandler(c *gin.Context) {
	username := c.MustGet("username").(string)

	var todo Todo

	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "wrong task format",
		})
		return
	}

	id := c.Param("id")

	// find todo by id in db
	// client can only update the todo and complete field
	if todo.Task != "" {
		_, err := DB.Query("UPDATE todos SET todo = $1 WHERE id = $2 AND username=$3", todo.Task, id, username)
		if err != nil {
			panic(err)
		}
	}

	_, err = DB.Query("UPDATE todos SET complete = $1 WHERE id = $2 AND username=$3", todo.Complete, id, username)
	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"username": username,
		"todo":     todo.Task,
		"complete": todo.Complete,
		"id":       id,
	})
}

func DeleteTodoHandler(c *gin.Context) {
	// auth
	username := c.MustGet("username").(string)

	id := c.Param("id")

	// delet task by id and username
	deleteStmt, err := DB.Prepare("DELETE FROM todos WHERE id=$1 AND username=$2")
	if err != nil {
		panic(nil)
	}
	defer deleteStmt.Close()

	_, err = deleteStmt.Exec(id, username)
	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}
