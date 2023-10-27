package background

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type UserRepository struct {
	db    *gorm.DB
	cache *sync.Map
}

func NewUserRepository(ctx context.Context) (*UserRepository, error) {
	cache := &sync.Map{} // Создайте экземпляр sync.Map
	db, err := gorm.Open(postgres.Open(os.Getenv("PG_URL")), &gorm.Config{})
	if err != nil {
		log.Printf("Error opening gorm: %v", err)
		return nil, fmt.Errorf("Error opening gorm: %v", err)
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		log.Printf("Error migrate: %v", err)
		return nil, fmt.Errorf("Error migrate: %v", err)
	}

	var users []User
	result := db.Find(&users)
	if result.Error != nil {
		log.Printf("Error retrieving users: %v", result.Error)
		return nil, fmt.Errorf("Error retrieving users: %v", result.Error)
	}

	for _, user := range users {
		cache.Store(user.ID, user)
	}

	return &UserRepository{db, cache}, nil
}

func (us *UserRepository) GetUserById(ID int) (User, error) {
	var user User
	data, ok := us.cache.Load(ID)
	if ok {
		log.Println("------------ch------------")
		user, ok = data.(User)
		if !ok {
			log.Printf("Error type assertion from cash: %v\n", user)
			return User{}, fmt.Errorf("Error type assertion from cash: %v\n", user)
		}
	} else {
		log.Println("------------bd------------")
		result := us.db.Find(&user, ID)
		if result.Error != nil {
			log.Println("User not found")
			return User{}, fmt.Errorf("User not found")
		}
		us.cache.Store(user.ID, user)
	}
	return user, nil
}
func (us *UserRepository) GetUsers() ([]User, error) {
	var users []User
	found := false
	us.cache.Range(func(_, value interface{}) bool {
		found = true
		users = append(users, value.(User))
		return true
	})
	if found {
		return users, nil
	} else {
		result := us.db.Find(&users)
		if result.Error != nil {
			log.Printf("Error retrieving users: %v", result.Error)
			return nil, fmt.Errorf("Error retrieving users: %v", result.Error)
		}
		found = true
	}
	if found {
		return users, nil
	} else {
		log.Println("Can't get users for db")
		return nil, fmt.Errorf("Can't get users for db")
	}
}

func (us *UserRepository) GetUserWithFilter(age int, gender string, nationality string) ([]User, error) {
	users, err := us.GetUsers()
	if err != nil {
		log.Printf("Can't get users with filter %v", err)
		return nil, fmt.Errorf("Can't get users with filter %v", err)
	}
	result := make([]User, len(users))
	for _, value := range users {
		if (age == value.Age && age != 0) &&
			(gender == value.Gender && gender != "") &&
			(nationality == value.Nationality && nationality != "") {
			result = append(result, value)
		} else if (age == value.Age && age != 0) && (gender == value.Gender && gender != "") {
			result = append(result, value)
		} else if (age == value.Age && age != 0) && (nationality == value.Nationality && nationality != "") {
			result = append(result, value)
		} else if (gender == value.Gender && gender != "") && (nationality == value.Nationality && nationality != "") {
			result = append(result, value)
		} else if age == value.Age && age != 0 {
			result = append(result, value)
		} else if gender == value.Gender && gender != "" {
			result = append(result, value)
		} else if nationality == value.Nationality && nationality != "" {
			result = append(result, value)
		} else {
			log.Println("Users with filter not found")
			return nil, fmt.Errorf("Users with filter not found")
		}
	}
	return result, nil
}

func (us *UserRepository) UserPagination(users []User, pageStr string, limitStr string) ([]User, error) {
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		log.Printf("Error Atoi page: %v", err)
		return nil, fmt.Errorf("Error Atoi page: %v", err)
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		log.Printf("Error Atoi limit: %v", err)
		return nil, fmt.Errorf("Error Atoi limit: %v", err)
	}
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit
	if startIndex < 0 || startIndex >= len(users) || endIndex > len(users) {
		log.Println("Invalid page or limit")
		return nil, fmt.Errorf("Invalid page or limit")
	}
	return users[startIndex:endIndex], nil
}

func (us *UserRepository) CreateUser(user User) (User, error) {
	result := us.db.Create(&user)
	if result.Error != nil {
		log.Printf("Error creating user: %v", result.Error)
		return User{}, fmt.Errorf("Error creating user: %v", result.Error)
	}
	us.cache.Store(user.ID, user)
	return user, nil
}

func (us *UserRepository) UpdateUser(user User) (User, error) {
	var result *gorm.DB

	log.Println(user) //

	result = us.db.Model(&User{}).Where("id = ?", user.ID).Updates(user)
	if result.Error != nil {
		log.Printf("Error updating user: %v", result.Error)
		return User{}, fmt.Errorf("Error updating user: %v", result.Error)
	}

	us.cache.Delete(user.ID)
	updateUser, err := us.GetUserById(user.ID)
	if err != nil {
		log.Printf("Error get user for update in cache %v", err)
		return User{}, fmt.Errorf("Error get user for update in cache %v", err)
	}
	us.cache.Store(user.ID, updateUser)

	return updateUser, nil
}

func (us *UserRepository) DeleteUser(userID int) (User, error) {
	user, err := us.GetUserById(userID)
	if err != nil {
		log.Printf("Error get user for delete %v", err)
		return User{}, fmt.Errorf("Error get user for delete %v", err)
	}

	result := us.db.Delete(&User{}, userID)
	if result.Error != nil {
		log.Printf("Error deleting user: %v", result.Error)
		return User{}, fmt.Errorf("Error deleting user: %v", result.Error)
	}

	us.cache.Delete(userID)

	return user, nil
}
