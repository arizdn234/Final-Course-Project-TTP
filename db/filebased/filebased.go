package filebased

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"a21hc3NpZ25tZW50/model"

	"go.etcd.io/bbolt"
)

type Data struct {
	DB *bbolt.DB
}

func InitDB() (*Data, error) {
	db, err := bbolt.Open("file.db", 0600, &bbolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Tasks"))
		if err != nil {
			return fmt.Errorf("create tasks bucket: %v", err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte("Categories"))
		if err != nil {
			return fmt.Errorf("create categories bucket: %v", err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte("Users"))
		if err != nil {
			return fmt.Errorf("create users bucket: %v", err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte("Sessions"))
		if err != nil {
			return fmt.Errorf("create sessions bucket: %v", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Data{DB: db}, nil
}

func (data *Data) StoreTask(task model.Task) error {
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return data.DB.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Tasks"))
		return b.Put([]byte(fmt.Sprintf("%d", task.ID)), taskJSON)
	})
}

func (data *Data) StoreCategory(category model.Category) error {
	categoryJSON, err := json.Marshal(category)
	if err != nil {
		return err
	}
	return data.DB.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Categories"))
		return b.Put([]byte(fmt.Sprintf("%d", category.ID)), categoryJSON)
	})
}

func (data *Data) UpdateTask(id int, task model.Task) error {
	return data.StoreTask(task) // Reuse StoreTask as it will replace the existing entry
}

func (data *Data) UpdateCategory(id int, category model.Category) error {
	return data.StoreCategory(category) // Reuse StoreCategory as it will replace the existing entry
}

func (data *Data) DeleteTask(id int) error {
	return data.DB.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Tasks"))
		return b.Delete([]byte(fmt.Sprintf("%d", id)))
	})
}

func (data *Data) DeleteCategory(id int) error {
	return data.DB.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Categories"))
		return b.Delete([]byte(fmt.Sprintf("%d", id)))
	})
}

func (data *Data) GetTaskByID(id int) (*model.Task, error) {
	var task model.Task
	err := data.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Tasks"))
		v := b.Get([]byte(fmt.Sprintf("%d", id)))
		if v == nil {
			return fmt.Errorf("record not found")
		}
		return json.Unmarshal(v, &task)
	})
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (data *Data) GetCategoryByID(id int) (*model.Category, error) {
	var category model.Category
	err := data.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Categories"))
		v := b.Get([]byte(fmt.Sprintf("%d", id)))
		if v == nil {
			return fmt.Errorf("record not found")
		}
		return json.Unmarshal(v, &category)
	})
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (data *Data) GetTasks() ([]model.Task, error) {
	var tasks []model.Task
	err := data.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Tasks"))
		return b.ForEach(func(k, v []byte) error {
			var task model.Task
			if err := json.Unmarshal(v, &task); err != nil {
				log.Println("Error unmarshaling task:", err)
				return nil // Continue despite error
			}
			tasks = append(tasks, task)
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching tasks: %v", err)
	}
	return tasks, nil
}

func (data *Data) GetCategories() ([]model.Category, error) {
	var categories []model.Category
	err := data.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Categories"))
		return b.ForEach(func(k, v []byte) error {
			var category model.Category
			if err := json.Unmarshal(v, &category); err != nil {
				log.Println("Error unmarshaling category:", err)
				return nil // Continue despite error
			}
			categories = append(categories, category)
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching categories: %v", err)
	}
	return categories, nil
}

func (data *Data) Reset() error {
	return data.DB.Update(func(tx *bbolt.Tx) error {
		if err := tx.DeleteBucket([]byte("Tasks")); err != nil {
			return err
		}
		if err := tx.DeleteBucket([]byte("Categories")); err != nil {
			return err
		}
		if err := tx.DeleteBucket([]byte("Users")); err != nil {
			return err
		}

		return nil
	})
}

func (data *Data) CloseDB() error {
	return data.DB.Close()
}

func (data *Data) GetTaskListByCategory(categoryID int) ([]model.TaskCategory, error) {
	var taskCategories []model.TaskCategory
	category, err := data.GetCategoryByID(categoryID)
	if err != nil {
		return nil, fmt.Errorf("error fetching category: %v", err)
	}

	err = data.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Tasks"))
		if b == nil {
			return fmt.Errorf("tasks bucket not found")
		}
		return b.ForEach(func(k, v []byte) error {
			var task model.Task
			if err := json.Unmarshal(v, &task); err != nil {
				log.Printf("Error unmarshaling task: %v", err)
				return nil // Continue processing next item in case of error
			}
			if task.CategoryID == categoryID {
				taskCategories = append(taskCategories, model.TaskCategory{
					ID:       task.ID,
					Title:    task.Title,
					Category: category.Name,
				})
			}
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching tasks for category %d: %v", categoryID, err)
	}
	if len(taskCategories) == 0 {
		return nil, fmt.Errorf("no tasks found for category ID: %d", categoryID)
	}
	return taskCategories, nil
}

func (data *Data) GetUserByEmail(email string) (model.User, error) {
	var user model.User
	found := false // Flag to check if the user is found

	err := data.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Users"))
		if b == nil {
			return fmt.Errorf("users bucket not found")
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var u model.User
			if err := json.Unmarshal(v, &u); err != nil {
				continue // Skip on unmarshal error
			}
			if u.Email == email {
				user = u
				found = true
				break // Stop the loop once the user is found
			}
		}
		return nil // Return nil error from the View transaction
	})

	if err != nil {
		return model.User{}, err // Return the error if the transaction failed
	}
	if !found {
		return model.User{}, nil // Return an empty User struct and nil error if not found
	}
	return user, nil // Return the found user and nil error
}

func (data *Data) CreateUser(user model.User) (model.User, error) {
	err := data.DB.Update(func(tx *bbolt.Tx) error {
		usersBucket := tx.Bucket([]byte("Users"))
		if usersBucket == nil {
			return fmt.Errorf("users bucket not found")
		}

		// Find the highest existing ID
		maxID := 0
		err := usersBucket.ForEach(func(k, v []byte) error {
			id := btoi(k)
			if id > maxID {
				maxID = id
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("error reading user IDs: %v", err)
		}

		// Set the new user's ID to maxID + 1
		newUserID := maxID + 1
		user.ID = newUserID // Assuming User.ID is of type int

		userJSON, err := json.Marshal(user)
		if err != nil {
			return fmt.Errorf("error marshaling user: %v", err)
		}

		// Store the new user with the new ID
		return usersBucket.Put(itob(newUserID), userJSON)
	})
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

// itob converts an integer to a byte slice
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// btoi converts a byte slice to an integer
func btoi(b []byte) int {
	if len(b) < 8 {
		return 0
	}
	return int(binary.BigEndian.Uint64(b))
}

func (data *Data) GetUserTaskCategory() ([]model.UserTaskCategory, error) {
	var results []model.UserTaskCategory

	err := data.DB.View(func(tx *bbolt.Tx) error {
		usersBucket := tx.Bucket([]byte("Users"))
		tasksBucket := tx.Bucket([]byte("Tasks"))
		categoriesBucket := tx.Bucket([]byte("Categories"))

		if usersBucket == nil || tasksBucket == nil || categoriesBucket == nil {
			return fmt.Errorf("one or more required buckets do not exist")
		}

		return usersBucket.ForEach(func(_, userValue []byte) error {
			var user model.User
			if err := json.Unmarshal(userValue, &user); err != nil {
				return err // skip badly formatted user records
			}

			// Now fetch tasks for the user
			return tasksBucket.ForEach(func(_, taskValue []byte) error {
				var task model.Task
				if err := json.Unmarshal(taskValue, &task); err != nil {
					return err // skip badly formatted task records
				}

				if task.UserID == user.ID { // Check if the task belongs to the user
					var category model.Category
					catValue := categoriesBucket.Get([]byte(fmt.Sprintf("%d", task.CategoryID)))
					if catValue != nil {
						if err := json.Unmarshal(catValue, &category); err != nil {
							return err // skip badly formatted category records
						}
					}

					result := model.UserTaskCategory{
						ID:       int(user.ID),
						Fullname: user.Fullname,
						Email:    user.Email,
						Task:     task.Title,
						Deadline: task.Deadline,
						Priority: task.Priority,
						Status:   task.Status,
						Category: category.Name,
					}
					results = append(results, result)

				}
				return nil
			})
		})
	})

	if err != nil {
		return nil, err
	}
	return results, nil
}

func (data *Data) AddSession(session model.Session) error {
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return data.DB.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Sessions"))
		return b.Put([]byte(session.Token), sessionJSON)
	})
}

func (data *Data) DeleteSession(token string) error {
	return data.DB.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Sessions"))
		return b.Delete([]byte(token))
	})
}

func (data *Data) UpdateSession(session model.Session) error {
	return data.AddSession(session) // Reuse AddSession as it will overwrite the existing entry
}

func (data *Data) SessionByToken(token string) (model.Session, error) {
	var session model.Session
	err := data.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Sessions"))
		sessionData := b.Get([]byte(token))
		if sessionData == nil {
			return fmt.Errorf("session not found")
		}
		return json.Unmarshal(sessionData, &session)
	})
	if err != nil {
		return model.Session{}, err
	}
	return session, nil
}

func (data *Data) TokenExpired(session model.Session) bool {
	return session.Expiry.Before(time.Now())
}

func (data *Data) TokenValidity(token string) (model.Session, error) {
	session, err := data.SessionByToken(token)
	if err != nil {
		return model.Session{}, err
	}

	if data.TokenExpired(session) {
		err := data.DeleteSession(token)
		if err != nil {
			return model.Session{}, err
		}
		return model.Session{}, fmt.Errorf("session expired")
	}

	return session, nil
}

func (data *Data) GetFirstSession() (model.Session, error) {
	var session model.Session
	found := false // Flag to check if at least one session is found

	err := data.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Sessions"))
		if b == nil {
			return fmt.Errorf("sessions bucket not found")
		}

		// Use Cursor to iterate through the bucket
		c := b.Cursor()
		k, v := c.First() // Retrieve the first session
		if k != nil {
			err := json.Unmarshal(v, &session)
			if err != nil {
				return err // Return unmarshaling error
			}
			found = true // Set found true as we have at least one session
		}
		return nil
	})

	if err != nil {
		return model.Session{}, err // Return error encountered during the View transaction
	}

	if !found {
		return model.Session{}, fmt.Errorf("no sessions found") // No session was found
	}

	return session, nil // Return the first session found
}

func (data *Data) SessionAvailEmail(email string) (model.Session, error) {
	var session model.Session
	found := false // Flag to check if at least one session matches the email

	err := data.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Sessions"))
		if b == nil {
			return fmt.Errorf("sessions bucket not found")
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var s model.Session
			if err := json.Unmarshal(v, &s); err != nil {
				continue // Skip badly formatted session records
			}
			if s.Email == email {
				session = s
				found = true
				break // Stop the iteration as we found the session
			}
		}
		return nil
	})

	if err != nil {
		return model.Session{}, err // Return error encountered during the View transaction
	}

	if !found {
		return model.Session{}, fmt.Errorf("no session available for email: %s", email) // No session was found for the given email
	}

	return session, nil // Return the found session
}

func (data *Data) SessionAvailToken(token string) (model.Session, error) {
	var session model.Session
	err := data.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Sessions"))
		if b == nil {
			return fmt.Errorf("sessions bucket not found")
		}

		sessionData := b.Get([]byte(token))
		if sessionData == nil {
			return fmt.Errorf("no session available for token: %s", token) // No session was found for the given token
		}

		return json.Unmarshal(sessionData, &session)
	})

	if err != nil {
		return model.Session{}, err // Return error encountered during the View transaction or unmarshaling
	}

	return session, nil // Return the found session
}
