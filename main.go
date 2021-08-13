package main

import (
	H "todo-app/handlers"

	"todo-app/middlewares"

	"github.com/gin-gonic/gin"
)

func main() {
	defer H.DB.Close()

	r := gin.Default()

	authorized := r.Group("/")
	authorized.Use(middlewares.JWTAuth())
	{
		// github上面沒有解釋為什麼需要加括號，應該是為了排版
		authorized.GET("/todos", H.ListTodosHandler)
		authorized.POST("/todo", H.AddTodoHandler)
		authorized.GET("/todo/:id", H.GetTodoHandler)
		authorized.PUT("/todo/:id", H.UpdateTodoHandler)
		authorized.DELETE("/todo/:id", H.DeleteTodoHandler)
	}

	r.POST("/sign-up", H.SignUpHandler)
	r.POST("/sign-in", H.SignInHandler)

	r.Run()
}
