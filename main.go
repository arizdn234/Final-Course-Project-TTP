package main

import (
	"a21hc3NpZ25tZW50/client"
	"a21hc3NpZ25tZW50/db/filebased"
	"a21hc3NpZ25tZW50/handler/api"
	"a21hc3NpZ25tZW50/handler/web"
	"a21hc3NpZ25tZW50/middleware"
	repo "a21hc3NpZ25tZW50/repository"
	"a21hc3NpZ25tZW50/service"
	"embed"
	"fmt"
	"net/http"
	"sync"
	"time"

	_ "embed"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type APIHandler struct {
	UserAPIHandler     api.UserAPI
	CategoryAPIHandler api.CategoryAPI
	TaskAPIHandler     api.TaskAPI
}

type ClientHandler struct {
	AuthWeb      web.AuthWeb
	HomeWeb      web.HomeWeb
	DashboardWeb web.DashboardWeb
	TaskWeb      web.TaskWeb
	CategoryWeb  web.CategoryWeb
	ModalWeb     web.ModalWeb
}

//go:embed views/*
var Resources embed.FS

func main() {
	gin.SetMode(gin.ReleaseMode) //release

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		router := gin.New()
		router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("[%s] \"%s %s %s\"\n",
				param.TimeStamp.Format(time.RFC822),
				param.Method,
				param.Path,
				param.ErrorMessage,
			)
		}))
		router.Use(gin.Recovery())

		filebasedDb, err := filebased.InitDB()

		if err != nil {
			panic(err)
		}

		router = RunServer(router, filebasedDb)
		router = RunClient(router, Resources, filebasedDb)

		PORT := "8080"
		fmt.Printf("Server is running on port %v\n\n`http://localhost:%v`", PORT, PORT)
		err = router.Run(":" + PORT)
		if err != nil {
			panic(err)
		}

	}()

	wg.Wait()
}

func RunServer(gin *gin.Engine, filebasedDb *filebased.Data) *gin.Engine {
	userRepo := repo.NewUserRepo(filebasedDb)
	sessionRepo := repo.NewSessionsRepo(filebasedDb)
	categoryRepo := repo.NewCategoryRepo(filebasedDb)
	taskRepo := repo.NewTaskRepo(filebasedDb)

	userService := service.NewUserService(userRepo, sessionRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	taskService := service.NewTaskService(taskRepo)

	userAPIHandler := api.NewUserAPI(userService)
	categoryAPIHandler := api.NewCategoryAPI(categoryService)
	taskAPIHandler := api.NewTaskAPI(taskService)

	apiHandler := APIHandler{
		UserAPIHandler:     userAPIHandler,
		CategoryAPIHandler: categoryAPIHandler,
		TaskAPIHandler:     taskAPIHandler,
	}

	version := gin.Group("/api/v1")
	{
		user := version.Group("/user")
		{
			user.POST("/login", apiHandler.UserAPIHandler.Login)
			user.POST("/register", apiHandler.UserAPIHandler.Register)

			user.Use(middleware.Auth()) // endpoints that require tokens from this endpoint group
			user.GET("/tasks", apiHandler.UserAPIHandler.GetUserTaskCategory)
		}

		task := version.Group("/task")
		{
			task.Use(middleware.Auth()) // endpoints that require tokens from this endpoint group
			task.POST("/add", apiHandler.TaskAPIHandler.AddTask)
			task.GET("/get/:id", apiHandler.TaskAPIHandler.GetTaskByID)
			task.PUT("/update/:id", apiHandler.TaskAPIHandler.UpdateTask)
			task.DELETE("/delete/:id", apiHandler.TaskAPIHandler.DeleteTask)
			task.GET("/list", apiHandler.TaskAPIHandler.GetTaskList)
			task.GET("/category/:id", apiHandler.TaskAPIHandler.GetTaskListByCategory)
		}

		category := version.Group("/category")
		{
			category.Use(middleware.Auth()) // endpoints that require tokens from this endpoint group
			category.POST("/add", apiHandler.CategoryAPIHandler.AddCategory)
			category.GET("/get/:id", apiHandler.CategoryAPIHandler.GetCategoryByID)
			category.PUT("/update/:id", apiHandler.CategoryAPIHandler.UpdateCategory)
			category.DELETE("/delete/:id", apiHandler.CategoryAPIHandler.DeleteCategory)
			category.GET("/list", apiHandler.CategoryAPIHandler.GetCategoryList)
		}
	}

	return gin
}

func RunClient(gin *gin.Engine, embed embed.FS, filebasedDb *filebased.Data) *gin.Engine {
	sessionRepo := repo.NewSessionsRepo(filebasedDb)
	sessionService := service.NewSessionService(sessionRepo)

	userClient := client.NewUserClient()
	taskClient := client.NewTaskClient()
	categoryClient := client.NewCategoryClient()

	authWeb := web.NewAuthWeb(userClient, sessionService, embed)
	modalWeb := web.NewModalWeb(embed)
	homeWeb := web.NewHomeWeb(embed)
	dashboardWeb := web.NewDashboardWeb(userClient, sessionService, embed)
	taskWeb := web.NewTaskWeb(taskClient, sessionService, embed)
	categoryWeb := web.NewCategoryWeb(categoryClient, sessionService, embed)

	client := ClientHandler{
		authWeb, homeWeb, dashboardWeb, taskWeb, categoryWeb, modalWeb,
	}

	gin.StaticFS("/static", http.Dir("frontend/public"))

	gin.GET("/", client.HomeWeb.Index)

	user := gin.Group("/client")
	{
		user.GET("/login", client.AuthWeb.Login)
		user.POST("/login/process", client.AuthWeb.LoginProcess)
		user.GET("/register", client.AuthWeb.Register)
		user.POST("/register/process", client.AuthWeb.RegisterProcess)

		user.Use(middleware.Auth()) // endpoints that require tokens from this endpoint group
		user.GET("/logout", client.AuthWeb.Logout)
	}

	main := gin.Group("/client")
	{
		main.Use(middleware.Auth()) // endpoints that require tokens from this endpoint group
		main.GET("/dashboard", client.DashboardWeb.Dashboard)
		main.GET("/task", client.TaskWeb.TaskPage)
		user.POST("/task/add/process", client.TaskWeb.TaskAddProcess)
		main.GET("/category", client.CategoryWeb.Category)
	}

	modal := gin.Group("/client")
	{
		modal.GET("/modal", client.ModalWeb.Modal)
	}

	return gin
}
