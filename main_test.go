package main_test

import (
	main "a21hc3NpZ25tZW50"
	"a21hc3NpZ25tZW50/db/filebased"
	"a21hc3NpZ25tZW50/middleware"
	"a21hc3NpZ25tZW50/model"
	repo "a21hc3NpZ25tZW50/repository"
	"a21hc3NpZ25tZW50/service"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var html string

func parseHTML() *goquery.Document {
	html = strings.Replace(html, `{{template "general/header"}}`, "", 1)

	tmpl, err := template.New("index").Parse(html)
	Expect(err).NotTo(HaveOccurred())

	var buf strings.Builder
	err = tmpl.Execute(&buf, nil)
	Expect(err).NotTo(HaveOccurred())

	strReader := strings.NewReader(buf.String())
	doc, err := goquery.NewDocumentFromReader(strReader)
	Expect(err).NotTo(HaveOccurred())

	return doc
}

func SetCookie(mux *gin.Engine) *http.Cookie {
	login := model.UserLogin{
		Email:    "test@mail.com",
		Password: "testing123",
	}

	body, _ := json.Marshal(login)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/user/login", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(w, r)

	var cookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "session_token" {
			cookie = c
		}
	}

	return cookie
}

var _ = Describe("Task Tracker Plus", Ordered, func() {
	var apiServer *gin.Engine

	var userRepo repo.UserRepository
	var sessionRepo repo.SessionRepository
	var categoryRepo repo.CategoryRepository
	var taskRepo repo.TaskRepository

	var userService service.UserService
	var sessionService service.SessionService
	var categoryService service.CategoryService
	var taskService service.TaskService

	var insertCategories []model.Category
	var insertTasks []model.Task
	var expectedUserTask []model.UserTaskCategory

	var filebasedDb *filebased.Data

	var err error

	BeforeEach(func() {
		gin.SetMode(gin.ReleaseMode) //release

		os.Remove("file.db")

		filebasedDb, err = filebased.InitDB()

		userRepo = repo.NewUserRepo(filebasedDb)
		sessionRepo = repo.NewSessionsRepo(filebasedDb)
		categoryRepo = repo.NewCategoryRepo(filebasedDb)
		taskRepo = repo.NewTaskRepo(filebasedDb)

		userService = service.NewUserService(userRepo, sessionRepo)
		sessionService = service.NewSessionService(sessionRepo)
		categoryService = service.NewCategoryService(categoryRepo)
		taskService = service.NewTaskService(taskRepo)

		Expect(err).ShouldNot(HaveOccurred())

		apiServer = gin.New()
		apiServer = main.RunServer(apiServer, filebasedDb)

		expectedUserTask = []model.UserTaskCategory{
			{
				ID:       1,
				Fullname: "test",
				Email:    "test@mail.com",
				Task:     "Task 2",
				Deadline: "2023-06-01",
				Priority: 1,
				Status:   "Completed",
				Category: "Category 2",
			},
			{
				ID:       1,
				Fullname: "test",
				Email:    "test@mail.com",
				Task:     "Task 5",
				Deadline: "2023-06-07",
				Priority: 5,
				Status:   "In Progress",
				Category: "Category 3",
			},
		}

		// Init test data:
		insertCategories = []model.Category{
			{ID: 1, Name: "Category 1"},
			{ID: 2, Name: "Category 2"},
			{ID: 3, Name: "Category 3"},
			{ID: 4, Name: "Category 4"},
			{ID: 5, Name: "Category 5"},
		}

		for _, v := range insertCategories {
			err := categoryRepo.Store(&v)
			Expect(err).ShouldNot(HaveOccurred())
		}

		insertTasks = []model.Task{
			{
				ID:         1,
				Title:      "Task 1",
				Deadline:   "2023-05-30",
				Priority:   2,
				Status:     "In Progress",
				CategoryID: 1,
				UserID:     2,
			},
			{
				ID:         2,
				Title:      "Task 2",
				Deadline:   "2023-06-01",
				Priority:   1,
				Status:     "Completed",
				CategoryID: 2,
				UserID:     1,
			},
			{
				ID:         3,
				Title:      "Task 3",
				Deadline:   "2023-06-02",
				Priority:   4,
				Status:     "Completed",
				CategoryID: 1,
				UserID:     3,
			},
			{
				ID:         4,
				Title:      "Task 4",
				Deadline:   "2023-06-02",
				Priority:   3,
				Status:     "Completed",
				CategoryID: 1,
				UserID:     4,
			},
			{
				ID:         5,
				Title:      "Task 5",
				Deadline:   "2023-06-07",
				Priority:   5,
				Status:     "In Progress",
				CategoryID: 3,
				UserID:     1,
			},
		}

		for _, v := range insertTasks {
			err := taskRepo.Store(&v)
			Expect(err).ShouldNot(HaveOccurred())
		}

		reqRegister := model.UserRegister{
			Fullname: "test",
			Email:    "test@mail.com",
			Password: "testing123",
		}

		reqBody, _ := json.Marshal(reqRegister)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/v1/user/register", bytes.NewReader(reqBody))
		r.Header.Set("Content-Type", "application/json")
		apiServer.ServeHTTP(w, r)

		Expect(w.Result().StatusCode).To(Equal(http.StatusCreated))
	})

	Describe("Auth Middleware", func() {
		var (
			router *gin.Engine
			w      *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			router = gin.Default()
			w = httptest.NewRecorder()
		})

		When("valid token is provided", func() {
			It("should set user Email in context and call next middleware", func() {
				claims := &model.Claims{Email: "aditira@gmail.com"}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				signedToken, _ := token.SignedString(model.JwtKey)
				req, _ := http.NewRequest(http.MethodGet, "/", nil)
				req.AddCookie(&http.Cookie{Name: "session_token", Value: signedToken})

				router.Use(middleware.Auth())
				router.GET("/", func(ctx *gin.Context) {
					Email := ctx.MustGet("email").(string)
					Expect(Email).To(Equal("aditira@gmail.com"))
				})

				router.ServeHTTP(w, req)
				Expect(w.Code).To(Equal(http.StatusOK))
			})
		})

		When("session token is missing", func() {
			It("should return unauthorized error response", func() {
				req, _ := http.NewRequest(http.MethodGet, "/", nil)

				router.Use(middleware.Auth())

				router.ServeHTTP(w, req)
				Expect(w.Code).To(Equal(http.StatusSeeOther))
			})
		})

		When("session token is invalid", func() {
			It("should return unauthorized error response", func() {
				req, _ := http.NewRequest(http.MethodGet, "/", nil)
				req.AddCookie(&http.Cookie{Name: "session_token", Value: "invalid_token"})

				router.Use(middleware.Auth())

				router.ServeHTTP(w, req)
				Expect(w.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("Repository", func() {
		Describe("Sessions repository", func() {
			When("add session data to sessions table database postgres", func() {
				It("should save data session to sessions table database postgres", func() {
					session := model.Session{
						Token:  "cc03dbea-4085-47ba-86fe-020f5d01a9d8",
						Email:  "aditira@gmail.com",
						Expiry: time.Date(2022, 11, 17, 20, 34, 58, 651387237, time.UTC),
					}
					err := sessionRepo.AddSessions(session)
					Expect(err).ShouldNot(HaveOccurred())

					result, err := filebasedDb.GetFirstSession()
					Expect(result.Token).To(Equal(session.Token))
					Expect(result.Email).To(Equal(session.Email))

					Expect(err).ShouldNot(HaveOccurred())
				})
			})

			When("delete selected session to sessions table database postgres", func() {
				It("should delete data session target from sessions table database postgres", func() {
					session := model.Session{
						Token:  "cc03dbea-4085-47ba-86fe-020f5d01a9d8",
						Email:  "aditira@gmail.com",
						Expiry: time.Date(2022, 11, 17, 20, 34, 58, 651387237, time.UTC),
					}
					err := sessionRepo.AddSessions(session)
					Expect(err).ShouldNot(HaveOccurred())

					result, err := filebasedDb.GetFirstSession()
					Expect(result.Token).To(Equal(session.Token))
					Expect(result.Email).To(Equal(session.Email))

					err = sessionRepo.DeleteSession("cc03dbea-4085-47ba-86fe-020f5d01a9d8")
					Expect(err).ShouldNot(HaveOccurred())

					result, err = filebasedDb.GetFirstSession()
					Expect(result).To(Equal(model.Session{}))
				})
			})

			When("update selected session to sessions table database postgres", func() {
				It("should update data session target the username field from sessions table database postgres", func() {
					session := model.Session{
						Token:  "cc03dbea-4085-47ba-86fe-020f5d01a9d8",
						Email:  "aditira@gmail.com",
						Expiry: time.Date(2022, 11, 17, 20, 34, 58, 651387237, time.UTC),
					}
					err := sessionRepo.AddSessions(session)
					Expect(err).ShouldNot(HaveOccurred())

					result, err := filebasedDb.GetFirstSession()
					Expect(result.Token).To(Equal(session.Token))
					Expect(result.Email).To(Equal(session.Email))

					sessionUpdate := model.Session{
						Token:  "cc03dbac-4085-22ba-75fe-103f9a01b6d5",
						Email:  "aditira@gmail.com",
						Expiry: time.Date(2022, 11, 17, 20, 34, 58, 651387237, time.UTC),
					}
					err = sessionRepo.UpdateSessions(sessionUpdate)
					Expect(err).ShouldNot(HaveOccurred())

					result, err = filebasedDb.GetFirstSession()
					Expect(result.Token).To(Equal(sessionUpdate.Token))

					Expect(err).ShouldNot(HaveOccurred())
				})
			})

			When("check session availability with name", func() {
				It("return data session with target name", func() {
					_, err := sessionRepo.SessionAvailEmail("aditira@gmail.com")
					Expect(err).Should(HaveOccurred())

					session := model.Session{
						Token:  "cc03dbea-4085-47ba-86fe-020f5d01a9d8",
						Email:  "aditira@gmail.com",
						Expiry: time.Now().Add(5 * time.Hour),
					}
					err = sessionRepo.AddSessions(session)
					Expect(err).ShouldNot(HaveOccurred())

					result, err := filebasedDb.GetFirstSession()
					Expect(result.Token).To(Equal(session.Token))
					Expect(result.Email).To(Equal(session.Email))

					res, err := sessionRepo.SessionAvailEmail("aditira@gmail.com")
					Expect(err).ShouldNot(HaveOccurred())
					Expect(res.Token).To(Equal(session.Token))
					Expect(result.Email).To(Equal(session.Email))
				})
			})

			When("check session availability with token", func() {
				It("return data session with target token", func() {
					_, err := sessionRepo.SessionAvailToken("cc03dbea-4085-47ba-86fe-020f5d01a9d8")
					Expect(err).Should(HaveOccurred())

					session := model.Session{
						Token:  "cc03dbea-4085-47ba-86fe-020f5d01a9d8",
						Email:  "aditira@gmail.com",
						Expiry: time.Now().Add(5 * time.Hour),
					}
					err = sessionRepo.AddSessions(session)
					Expect(err).ShouldNot(HaveOccurred())

					result, err := filebasedDb.GetFirstSession()
					Expect(result.Token).To(Equal(session.Token))
					Expect(result.Email).To(Equal(session.Email))

					res, err := sessionRepo.SessionAvailToken("cc03dbea-4085-47ba-86fe-020f5d01a9d8")
					Expect(err).ShouldNot(HaveOccurred())
					Expect(res.Token).To(Equal(session.Token))
					Expect(result.Email).To(Equal(session.Email))
				})
			})

		})

		Describe("User", func() {
			When("fetching a single user data by email from users table in the database", func() {
				It("should return a single task data", func() {
					expectUser := model.User{
						Fullname: "test",
						Email:    "test@mail.com",
						Password: "testing123",
					}

					resUser, err := userRepo.GetUserByEmail("test@mail.com")
					Expect(err).ShouldNot(HaveOccurred())
					Expect(resUser.Fullname).To(Equal(expectUser.Fullname))
					Expect(resUser.Email).To(Equal(expectUser.Email))
					Expect(resUser.Password).To(Equal(expectUser.Password))
				})
			})

			When("retrieving user task categories from user repository", func() {
				It("should return the expected user task categories", func() {
					resUserTask, err := userRepo.GetUserTaskCategory()
					Expect(err).ShouldNot(HaveOccurred())
					Expect(resUserTask).To(Equal(expectedUserTask))
				})
			})
		})

		Describe("Task", func() {
			When("updating a category with valid ID and new category data", func() {
				It("should update the existing category with the new category data", func() {
					newCategory := model.Category{
						ID:   1,
						Name: "Updated with Repository Category 1",
					}
					err = categoryRepo.Update(1, newCategory)
					Expect(err).ShouldNot(HaveOccurred())

					result, err := categoryRepo.GetByID(1)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(result.Name).To(Equal(newCategory.Name))

					Expect(err).ShouldNot(HaveOccurred())
				})
			})

			When("deleting a category with a valid category ID", func() {
				It("should delete the category from the database without returning an error", func() {
					err = categoryRepo.Delete(2)
					Expect(err).ShouldNot(HaveOccurred())

					result, err := categoryRepo.GetByID(2)
					Expect(err.Error()).To(Equal("record not found"))
					Expect(result).To(BeNil())
				})
			})

			When("retrieving the list of categories from the database", func() {
				It("should return the list of categories without any errors and the list should contain the expected number of categories", func() {
					results, err := categoryRepo.GetList()
					Expect(err).ShouldNot(HaveOccurred())
					Expect(results).To(HaveLen(5))

					Expect(results).To(Equal(insertCategories))
				})
			})

		})

		Describe("Task", func() {
			When("updating task data in tasks table in the database", func() {
				It("should update the existing task data in tasks table in the database", func() {
					newTask := model.Task{
						ID:         1,
						Title:      "Updated with Repository Task 1",
						Deadline:   "2023-05-30",
						Priority:   2,
						CategoryID: 1,
						Status:     "In Progress",
					}
					err = taskRepo.Update(newTask.ID, &newTask)
					Expect(err).ShouldNot(HaveOccurred())

					result, err := taskRepo.GetByID(1)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(result.Title).To(Equal(newTask.Title))
					Expect(result.Deadline).To(Equal(newTask.Deadline))
					Expect(result.Priority).To(Equal(newTask.Priority))
					Expect(result.CategoryID).To(Equal(newTask.CategoryID))
					Expect(result.Status).To(Equal(newTask.Status))
				})
			})

			When("deleting a task with a valid task ID from the database", func() {
				It("should delete the task without any errors", func() {
					err = taskRepo.Delete(2)
					Expect(err).ShouldNot(HaveOccurred())

					result, err := taskRepo.GetByID(2)
					Expect(err.Error()).To(Equal("record not found"))
					Expect(result).To(BeNil())
				})
			})

			When("retrieving the list of tasks from the database", func() {
				It("should return the list of tasks without any errors", func() {
					results, err := taskRepo.GetList()
					Expect(err).ShouldNot(HaveOccurred())
					Expect(results).To(HaveLen(5))

					Expect(results).To(Equal(insertTasks))
				})
			})

			When("retrieving the list of tasks for a specific category from the database", func() {
				It("should return the list of tasks for the specified category without any errors", func() {
					taskCategory, err := taskRepo.GetTaskCategory(1)
					Expect(err).ShouldNot(HaveOccurred())

					Expect(taskCategory).To(Equal([]model.TaskCategory{
						{ID: 1, Title: "Task 1", Category: "Category 1"},
						{ID: 3, Title: "Task 3", Category: "Category 1"},
						{ID: 4, Title: "Task 4", Category: "Category 1"},
					}))
				})
			})

		})
	})

	Describe("Service", func() {
		Describe("Session Service", func() {
			Describe("GetSessionByEmail", func() {
				When("retrieving user session by email from session repository", func() {
					It("should return the expected user session", func() {
						session := model.Session{
							Token:  "cc03dbea-4085-47ba-86fe-020f5d01a9d8",
							Email:  "aditira@gmail.com",
							Expiry: time.Now().Add(5 * time.Hour),
						}
						err = sessionRepo.AddSessions(session)
						Expect(err).ShouldNot(HaveOccurred())

						res, err := sessionService.GetSessionByEmail("aditira@gmail.com")
						Expect(err).ShouldNot(HaveOccurred())
						Expect(res.Token).To(Equal(session.Token))
						Expect(res.Email).To(Equal(session.Email))
					})
				})
			})
		})

		Describe("User Service", func() {
			Describe("GetUserTaskCategory", func() {
				When("retrieving user task categories from user repository", func() {
					It("should return the expected user task categories", func() {
						resUserTask, err := userService.GetUserTaskCategory()
						Expect(err).ShouldNot(HaveOccurred())
						Expect(resUserTask).To(Equal(expectedUserTask))
					})
				})
			})
		})

		Describe("Category Service", func() {
			Describe("Update", func() {
				When("updating a category in the database", func() {
					It("should update the category without any errors", func() {
						category := model.Category{
							ID:   1,
							Name: "Updated with Service Category 1",
						}

						err := categoryService.Update(1, category)
						Expect(err).ShouldNot(HaveOccurred())
					})
				})
			})

			Describe("Delete", func() {
				When("deleting a category from the database", func() {
					It("should delete the category without any errors", func() {
						err := categoryService.Delete(3)
						Expect(err).ShouldNot(HaveOccurred())
					})
				})
			})

			Describe("GetList", func() {
				When("retrieving the list of categories from the database", func() {
					It("should return the list of categories without any errors", func() {
						categories, err := categoryService.GetList()
						Expect(err).ShouldNot(HaveOccurred())
						Expect(categories).To(HaveLen(5))

						Expect(categories).To(Equal(insertCategories))
					})
				})
			})
		})

		Describe("Task Service", func() {
			Describe("Update", func() {
				When("updating a task in the database", func() {
					It("should update the task without any errors", func() {
						task := &model.Task{
							ID:         1,
							Title:      "Updated with Service Task 1",
							Deadline:   "2023-05-30",
							Priority:   5,
							CategoryID: 1,
							Status:     "In Progress",
						}

						err := taskService.Update(task.ID, task)
						Expect(err).ShouldNot(HaveOccurred())
					})
				})
			})

			Describe("Delete", func() {
				When("deleting a task from the database", func() {
					It("should delete the task without any errors", func() {
						err := taskService.Delete(3)
						Expect(err).ShouldNot(HaveOccurred())
					})
				})
			})

			Describe("GetList", func() {
				When("retrieving the list of tasks from the database", func() {
					It("should return the list of tasks without any errors", func() {
						tasks, err := taskService.GetList()
						Expect(err).ShouldNot(HaveOccurred())
						Expect(tasks).To(Equal(insertTasks))
					})
				})
			})

			Describe("GetTaskCategory", func() {
				When("retrieving the category of a task from the database", func() {
					It("should return the task category without any errors", func() {
						taskCategories, err := taskService.GetTaskCategory(1)
						Expect(err).ShouldNot(HaveOccurred())
						Expect(taskCategories).To(Equal([]model.TaskCategory{
							{ID: 1, Title: "Task 1", Category: "Category 1"},
							{ID: 3, Title: "Task 3", Category: "Category 1"},
							{ID: 4, Title: "Task 4", Category: "Category 1"},
						}))
					})
				})
			})
		})
	})

	Describe("API", func() {
		Describe("User API", func() {
			When("send empty email and password with POST method", func() {
				It("should return a bad request", func() {
					loginData := model.UserLogin{
						Email:    "",
						Password: "",
					}

					body, _ := json.Marshal(loginData)
					w := httptest.NewRecorder()
					r := httptest.NewRequest("POST", "/api/v1/user/login", bytes.NewReader(body))
					r.Header.Set("Content-Type", "application/json")

					apiServer.ServeHTTP(w, r)

					errResp := model.ErrorResponse{}
					err := json.Unmarshal(w.Body.Bytes(), &errResp)
					Expect(err).To(BeNil())
					Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
					Expect(errResp.Error).To(Equal("invalid decode json"))
				})
			})

			When("send email and password with POST method", func() {
				It("should return a success", func() {
					loginData := model.UserLogin{
						Email:    "test@mail.com",
						Password: "testing123",
					}
					body, _ := json.Marshal(loginData)
					w := httptest.NewRecorder()
					r := httptest.NewRequest("POST", "/api/v1/user/login", bytes.NewReader(body))
					r.Header.Set("Content-Type", "application/json")
					apiServer.ServeHTTP(w, r)

					var resp = map[string]interface{}{}
					err = json.Unmarshal(w.Body.Bytes(), &resp)
					Expect(err).To(BeNil())
					Expect(w.Result().StatusCode).To(Equal(http.StatusOK))
					Expect(resp["message"]).To(Equal("login success"))
				})
			})

			Describe("GetUserTaskCategory", func() {
				When("sending without cookie", func() {
					It("should return status code 401", func() {
						r, _ := http.NewRequest("GET", "/api/v1/user/tasks", nil)
						w := httptest.NewRecorder()
						r.Header.Set("Content-Type", "application/json")
						apiServer.ServeHTTP(w, r)
						Expect(w.Code).To(Equal(http.StatusUnauthorized))
					})
				})

				When("retrieving user list by task and category", func() {
					It("should return status code 200 and task list", func() {
						r, _ := http.NewRequest("GET", "/api/v1/user/tasks", nil)
						w := httptest.NewRecorder()

						r.AddCookie(SetCookie(apiServer))
						apiServer.ServeHTTP(w, r)
						Expect(w.Code).To(Equal(http.StatusOK))

						var userTasks []model.UserTaskCategory
						Expect(json.Unmarshal(w.Body.Bytes(), &userTasks)).Should(Succeed())
						Expect(userTasks).To(Equal(expectedUserTask))
					})
				})
			})
		})

		Describe("Category API", func() {
			Describe("UpdateCategory", func() {
				When("sending without cookie", func() {
					It("should return status code 401", func() {
						updatedCategory := model.Category{
							ID:   1,
							Name: "Updated with API Category 1",
						}

						requestBody, _ := json.Marshal(updatedCategory)
						r, _ := http.NewRequest("PUT", "/api/v1/category/update/1", bytes.NewReader(requestBody))
						r.Header.Set("Content-Type", "application/json")
						w := httptest.NewRecorder()
						apiServer.ServeHTTP(w, r)
						Expect(w.Code).To(Equal(http.StatusUnauthorized))
					})
				})

				When("updating an existing category", func() {
					It("should update the category and return status code 200", func() {
						updatedCategory := model.Category{
							ID:   1,
							Name: "Updated with API Category 1",
						}

						requestBody, _ := json.Marshal(updatedCategory)
						r, _ := http.NewRequest("PUT", "/api/v1/category/update/1", bytes.NewReader(requestBody))
						r.Header.Set("Content-Type", "application/json")
						w := httptest.NewRecorder()

						r.AddCookie(SetCookie(apiServer))
						apiServer.ServeHTTP(w, r)
						Expect(w.Code).To(Equal(http.StatusOK))

						var response model.SuccessResponse
						Expect(json.Unmarshal(w.Body.Bytes(), &response)).Should(Succeed())
						Expect(response.Message).To(Equal("category update success"))
					})
				})

				When("updating a non-existing category", func() {
					It("should return status code 400", func() {
						updatedCategory := model.Category{
							ID:   1,
							Name: "Updated with API Category 1",
						}

						requestBody, _ := json.Marshal(updatedCategory)
						r, _ := http.NewRequest("PUT", "/api/v1/category/update/abc", bytes.NewReader(requestBody))
						r.Header.Set("Content-Type", "application/json")
						w := httptest.NewRecorder()

						r.AddCookie(SetCookie(apiServer))
						apiServer.ServeHTTP(w, r)
						Expect(w.Code).To(Equal(http.StatusBadRequest))

						var response model.ErrorResponse
						Expect(json.Unmarshal(w.Body.Bytes(), &response)).Should(Succeed())
						Expect(response.Error).To(Equal("invalid Category ID"))
					})
				})

				When("sending an invalid request", func() {
					It("should return status code 400", func() {
						reqBody := []byte("invalid request body")

						r, _ := http.NewRequest("PUT", "/api/v1/category/update/1", bytes.NewReader(reqBody))
						r.Header.Set("Content-Type", "application/json")
						w := httptest.NewRecorder()

						r.AddCookie(SetCookie(apiServer))
						apiServer.ServeHTTP(w, r)
						Expect(w.Code).To(Equal(http.StatusBadRequest))

						var response model.ErrorResponse
						Expect(json.Unmarshal(w.Body.Bytes(), &response)).Should(Succeed())
						Expect(response.Error).NotTo(BeNil())
					})
				})
			})

			Describe("DeleteCategory", func() {
				When("sending without cookie", func() {
					It("should return status code 401", func() {
						r, _ := http.NewRequest("DELETE", "/api/v1/category/delete/4", nil)
						w := httptest.NewRecorder()
						r.Header.Set("Content-Type", "application/json")
						apiServer.ServeHTTP(w, r)
						Expect(w.Code).To(Equal(http.StatusUnauthorized))
					})
				})

				When("deleting a category", func() {
					It("should delete the category and return status code 200", func() {
						r, _ := http.NewRequest("DELETE", "/api/v1/category/delete/4", nil)
						w := httptest.NewRecorder()

						r.AddCookie(SetCookie(apiServer))
						apiServer.ServeHTTP(w, r)
						Expect(w.Code).To(Equal(http.StatusOK))

						var response model.SuccessResponse
						Expect(json.Unmarshal(w.Body.Bytes(), &response)).Should(Succeed())
						Expect(response.Message).To(Equal("category delete success"))
					})
				})
			})

			Describe("GetCategoryList", func() {
				When("sending without cookie", func() {
					It("should return status code 401", func() {
						r, _ := http.NewRequest("GET", "/api/v1/category/list", nil)
						w := httptest.NewRecorder()
						r.Header.Set("Content-Type", "application/json")
						apiServer.ServeHTTP(w, r)
						Expect(w.Code).To(Equal(http.StatusUnauthorized))
					})
				})

				When("retrieving the list of categories", func() {
					It("should return the list of categories and status code 200", func() {
						r, _ := http.NewRequest("GET", "/api/v1/category/list", nil)
						w := httptest.NewRecorder()

						r.AddCookie(SetCookie(apiServer))
						apiServer.ServeHTTP(w, r)
						Expect(w.Code).To(Equal(http.StatusOK))

						var response []model.Category
						Expect(json.Unmarshal(w.Body.Bytes(), &response)).Should(Succeed())
						Expect(response).To(Equal(insertCategories))
					})
				})
			})
		})

		Describe("Task API", func() {
			Describe("UpdateTask", func() {
				When("sending without cookie", func() {
					It("should return status code 401", func() {
						updatedTask := model.Task{
							ID:         1,
							Title:      "Updated with API Task 1",
							Deadline:   "2023-05-30",
							Priority:   5,
							CategoryID: 1,
							Status:     "In Progress",
						}
						reqBody, _ := json.Marshal(updatedTask)

						r, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/task/update/%d", 1), bytes.NewReader(reqBody))
						w := httptest.NewRecorder()
						r.Header.Set("Content-Type", "application/json")
						apiServer.ServeHTTP(w, r)
						Expect(w.Code).To(Equal(http.StatusUnauthorized))
					})
				})

				When("updating existing task", func() {
					It("should return status code 200", func() {
						updatedTask := model.Task{
							ID:         1,
							Title:      "Updated with API Task 1",
							Deadline:   "2023-05-30",
							Priority:   5,
							CategoryID: 1,
							Status:     "In Progress",
						}
						reqBody, _ := json.Marshal(updatedTask)

						r, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/task/update/%d", 1), bytes.NewReader(reqBody))
						w := httptest.NewRecorder()

						r.AddCookie(SetCookie(apiServer))
						apiServer.ServeHTTP(w, r)
						Expect(w.Code).To(Equal(http.StatusOK))

						var response model.SuccessResponse
						Expect(json.Unmarshal(w.Body.Bytes(), &response)).Should(Succeed())
						Expect(response.Message).To(Equal("update task success"))
					})
				})

				When("sending invalid request", func() {
					It("should return status code 400", func() {
						reqBody := []byte("invalid request body")
						r, _ := http.NewRequest("PUT", "/api/v1/task/update/1", bytes.NewReader(reqBody))
						w := httptest.NewRecorder()

						r.AddCookie(SetCookie(apiServer))
						apiServer.ServeHTTP(w, r)

						Expect(w.Code).To(Equal(http.StatusBadRequest))

						var response model.ErrorResponse
						Expect(json.Unmarshal(w.Body.Bytes(), &response)).Should(Succeed())
						Expect(response.Error).NotTo(BeNil())
					})
				})
			})

			Describe("DeleteTask", func() {
				When("sending without cookie", func() {
					It("should return status code 401", func() {
						r, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/task/delete/%d", 1), nil)
						w := httptest.NewRecorder()
						r.Header.Set("Content-Type", "application/json")
						apiServer.ServeHTTP(w, r)
						Expect(w.Code).To(Equal(http.StatusUnauthorized))
					})
				})

				When("deleting existing task", func() {
					It("should return status code 200", func() {
						r, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/task/delete/%d", 1), nil)
						w := httptest.NewRecorder()

						r.AddCookie(SetCookie(apiServer))
						apiServer.ServeHTTP(w, r)
						Expect(w.Code).To(Equal(http.StatusOK))

						var response model.SuccessResponse
						Expect(json.Unmarshal(w.Body.Bytes(), &response)).Should(Succeed())
						Expect(response.Message).To(Equal("delete task success"))
					})
				})

				When("deleting non-existing task", func() {
					It("should return status code 400", func() {
						taskID := "abc"
						r, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/task/delete/%s", taskID), nil)
						w := httptest.NewRecorder()

						r.AddCookie(SetCookie(apiServer))
						apiServer.ServeHTTP(w, r)
						Expect(w.Code).To(Equal(http.StatusBadRequest))

						var response model.ErrorResponse
						Expect(json.Unmarshal(w.Body.Bytes(), &response)).Should(Succeed())
						Expect(response.Error).To(Equal("Invalid task ID"))
					})
				})
			})

			Describe("GetTaskList", func() {
				When("sending without cookie", func() {
					It("should return status code 401", func() {
						r, _ := http.NewRequest("GET", "/api/v1/task/list", nil)
						w := httptest.NewRecorder()
						r.Header.Set("Content-Type", "application/json")
						apiServer.ServeHTTP(w, r)
						Expect(w.Code).To(Equal(http.StatusUnauthorized))
					})
				})

				When("retrieving task list", func() {
					It("should return status code 200 and task list", func() {
						r, _ := http.NewRequest("GET", "/api/v1/task/list", nil)
						w := httptest.NewRecorder()

						r.AddCookie(SetCookie(apiServer))
						apiServer.ServeHTTP(w, r)
						Expect(w.Code).To(Equal(http.StatusOK))

						var response []model.Task
						Expect(json.Unmarshal(w.Body.Bytes(), &response)).Should(Succeed())
						Expect(response).To(Equal(insertTasks))
					})
				})
			})

			Describe("GetTaskListByCategory", func() {
				When("sending without cookie", func() {
					It("should return status code 401", func() {
						r, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/task/category/%d", 1), nil)
						w := httptest.NewRecorder()
						r.Header.Set("Content-Type", "application/json")
						apiServer.ServeHTTP(w, r)
						Expect(w.Code).To(Equal(http.StatusUnauthorized))
					})
				})

				When("retrieving task list by category", func() {
					It("should return status code 200 and task list", func() {
						r, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/task/category/%d", 1), nil)
						w := httptest.NewRecorder()

						r.AddCookie(SetCookie(apiServer))
						apiServer.ServeHTTP(w, r)
						Expect(w.Code).To(Equal(http.StatusOK))

						var response []model.TaskCategory
						Expect(json.Unmarshal(w.Body.Bytes(), &response)).Should(Succeed())
						Expect(response).To(Equal([]model.TaskCategory{
							{ID: 1, Title: "Task 1", Category: "Category 1"},
							{ID: 3, Title: "Task 3", Category: "Category 1"},
							{ID: 4, Title: "Task 4", Category: "Category 1"},
						}))
					})
				})
			})
		})

		Describe("HTML", func() {
			Describe("views/main/index.html", func() {
				var (
					doc *goquery.Document
				)

				BeforeEach(func() {
					data, err := ioutil.ReadFile("views/main/index.html")
					Expect(err).NotTo(HaveOccurred())
					html = string(data)
					doc = parseHTML()
				})

				Describe("Render", func() {
					It("should have at least 3 sections or divs", func() {
						sections := doc.Find("section").Length()
						divs := doc.Find("div").Length()
						Expect(sections + divs).To(BeNumerically(">=", 3))
					})

					It("should have at least one types of heading tags", func() {
						headings := []string{"h1", "h2", "h3", "h4", "h5", "h6"}
						count := 0
						for _, heading := range headings {
							if doc.Find(heading).Length() > 0 {
								count++
							}
						}
						Expect(count).To(BeNumerically(">=", 1))
					})

					It("should have at least one paragraph", func() {
						Expect(doc.Find("p").Length()).To(BeNumerically(">=", 1))
					})

					It("should have a heading with text 'Task Tracker Plus'", func() {
						Expect(doc.Find("h1").Text()).To(ContainSubstring("Task Tracker Plus"))
					})

					It("should have a login link that navigates to /login", func() {
						Expect(doc.Find("a[href='/client/login']").AttrOr("href", "")).To(Equal("/client/login"))
					})

					It("should have a register link that navigates to /register", func() {
						Expect(doc.Find("a[href='/client/register']").AttrOr("href", "")).To(Equal("/client/register"))
					})
				})

				Describe("Style", func() {
					It("should have at least one component use tailwind for section or div", func() {
						found := false
						doc.Find("div, section").Each(func(_ int, s *goquery.Selection) {
							classes := strings.Split(s.AttrOr("class", ""), " ")
							if model.RepresentsTailwind(classes) {
								found = true
							}
						})
						Expect(found).To(BeTrue(), "No Tailwind Classes found in div or section")
					})

					It("should have at least one component use tailwind for h1, h2, h3, h4, h5, h6", func() {
						found := false
						doc.Find("h1, h2, h3, h4, h5, h6").Each(func(_ int, s *goquery.Selection) {
							classes := strings.Split(s.AttrOr("class", ""), " ")
							if model.RepresentsTailwind(classes) {
								found = true
							}
						})
						Expect(found).To(BeTrue(), "No Tailwind Classes found in h1 to h6")
					})

					It("should have at least one component use tailwind for responsive design", func() {
						found := false
						doc.Find("[class*='sm:'] ,[class*='md:'], [class*='lg:'], [class*='xl:'], [class*='2xl:']").Each(func(_ int, s *goquery.Selection) {
							classes := strings.Split(s.AttrOr("class", ""), " ")
							if model.RepresentsTailwind(classes) {
								found = true
							}
						})
						Expect(found).To(BeTrue(), "No Tailwind Classes found in sm, md, lg, xl or 2x1")
					})
				})
			})

			Describe("views/auth/login.html", func() {
				var (
					doc *goquery.Document
				)

				BeforeEach(func() {
					data, err := ioutil.ReadFile("views/auth/login.html")
					Expect(err).NotTo(HaveOccurred())
					html = string(data)
					doc = parseHTML()
				})

				Describe("Render", func() {
					It("should have at least one types of heading tags", func() {
						headings := []string{"h1", "h2", "h3", "h4", "h5", "h6"}
						count := 0
						for _, heading := range headings {
							if doc.Find(heading).Length() > 0 {
								count++
							}
						}
						Expect(count).To(BeNumerically(">=", 1))
					})

					It("should have an email input field", func() {
						emailInput := doc.Find(`input[type="email"]`)
						Expect(len(emailInput.Nodes)).NotTo(Equal(0))
					})

					It("should have a password input field", func() {
						passwordInput := doc.Find(`input[type="password"]`)
						Expect(len(passwordInput.Nodes)).NotTo(Equal(0))
					})

					It("should have a submit button with text 'Login'", func() {
						Expect(doc.Find("button").Text()).To(ContainSubstring("Login"))
					})

					It("should have a register link that navigates to '/client/register", func() {
						registerLink := doc.Find(`a[href="/client/register"]`)
						Expect(len(registerLink.Nodes)).NotTo(Equal(0))
					})
				})

				Describe("Style", func() {
					It("should have at least one component use tailwind for section or div", func() {
						found := false
						doc.Find("div, section").Each(func(_ int, s *goquery.Selection) {
							classes := strings.Split(s.AttrOr("class", ""), " ")
							if model.RepresentsTailwind(classes) {
								found = true
							}
						})
						Expect(found).To(BeTrue(), "No Tailwind Classes found in div or section")
					})

					It("should have at least one component use tailwind for h1, h2, h3, h4, h5, h6", func() {
						found := false
						doc.Find("h1, h2, h3, h4, h5, h6").Each(func(_ int, s *goquery.Selection) {
							classes := strings.Split(s.AttrOr("class", ""), " ")
							if model.RepresentsTailwind(classes) {
								found = true
							}
						})
						Expect(found).To(BeTrue(), "No Tailwind Classes found in h1 to h6")
					})
				})
			})

			Describe("views/auth/register.html", func() {
				var (
					doc *goquery.Document
				)

				BeforeEach(func() {
					data, err := ioutil.ReadFile("views/auth/register.html")
					Expect(err).NotTo(HaveOccurred())
					html = string(data)
					doc = parseHTML()
				})

				Describe("Render", func() {
					It("should have at least one types of heading tags", func() {
						headings := []string{"h1", "h2", "h3", "h4", "h5", "h6"}
						count := 0
						for _, heading := range headings {
							if doc.Find(heading).Length() > 0 {
								count++
							}
						}
						Expect(count).To(BeNumerically(">=", 1))
					})

					It("should have a fullname input field", func() {
						fullnameInput := doc.Find(`input[type="text"]`)
						Expect(len(fullnameInput.Nodes)).NotTo(Equal(0))
					})

					It("should have an email input field", func() {
						emailInput := doc.Find(`input[type="email"]`)
						Expect(len(emailInput.Nodes)).NotTo(Equal(0))
					})

					It("should have a password input field", func() {
						passwordInput := doc.Find(`input[type="password"]`)
						Expect(len(passwordInput.Nodes)).NotTo(Equal(0))
					})

					It("should have a submit button with text 'Register'", func() {
						Expect(doc.Find("button").Text()).To(ContainSubstring("Register"))
					})

					It("should have a register link that navigates to '/client/login", func() {
						registerLink := doc.Find(`a[href="/client/login"]`)
						Expect(len(registerLink.Nodes)).NotTo(Equal(0))
					})
				})

				Describe("Style", func() {
					It("should have at least one component use tailwind for section or div", func() {
						found := false
						doc.Find("div, section").Each(func(_ int, s *goquery.Selection) {
							classes := strings.Split(s.AttrOr("class", ""), " ")
							if model.RepresentsTailwind(classes) {
								found = true
							}
						})
						Expect(found).To(BeTrue(), "No Tailwind Classes found in div or section")
					})

					It("should have at least one component use tailwind for h1, h2, h3, h4, h5, h6", func() {
						found := false
						doc.Find("h1, h2, h3, h4, h5, h6").Each(func(_ int, s *goquery.Selection) {
							classes := strings.Split(s.AttrOr("class", ""), " ")
							if model.RepresentsTailwind(classes) {
								found = true
							}
						})
						Expect(found).To(BeTrue(), "No Tailwind Classes found in h1 to h6")
					})
				})
			})
		})
	})
})
