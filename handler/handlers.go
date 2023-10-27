package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/TiZir/service_with_enrichment/background"
	"github.com/labstack/echo"
)

func HomePageHandler(c echo.Context) error {
	return c.String(http.StatusOK, "The service is running!")
}

func GetUsersHandler(c echo.Context, r *background.UserRepository) error {
	var err error
	ageStr := c.QueryParam("age")
	gender := c.QueryParam("gender")
	nationality := c.QueryParam("nationality")
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")

	var age int
	if ageStr != "" {
		age, err = strconv.Atoi(ageStr)
		if err != nil {
			log.Printf("Error Atoi age: %v", err)
			fmt.Errorf("Error Atoi age: %v", err)
			return c.JSON(http.StatusBadRequest, err)
		}
	}
	filteredPeople, err := r.GetUserWithFilter(age, gender, nationality)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	// Применение пагинации
	if pageStr != "" && limitStr != "" {
		filteredPeople, err = r.UserPagination(filteredPeople, pageStr, limitStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
	}
	return c.JSON(http.StatusOK, filteredPeople)
}

func GetUserByIdHandler(c echo.Context, r *background.UserRepository) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Error Atoi age: %v", err)
		fmt.Errorf("Error Atoi age: %v", err)
		return c.JSON(http.StatusBadRequest, err)
	}

	// Поиск человека по ID
	user, err := r.GetUserById(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	return c.JSON(http.StatusNotFound, user)
}

func GetExtraAge(name string) (int, error) {
	ageURL := fmt.Sprintf("https://api.agify.io/?name=%s", name)
	ageResponse, err := http.Get(ageURL)
	if err != nil {
		log.Printf("Error get age from api: %v", err)
		fmt.Errorf("Error get age from api: %v", err)
		return 0, err
	}
	defer ageResponse.Body.Close()

	ageResponseBody, err := io.ReadAll(ageResponse.Body)
	if err != nil {
		log.Printf("Error read age from api: %v", err)
		fmt.Errorf("Error read age from api: %v", err)
		return 0, err
	}

	ageData := struct {
		Age int `json:"age"`
	}{}

	err = json.Unmarshal(ageResponseBody, &ageData)
	if err != nil {
		log.Printf("Error unmarshal age from api: %v", err)
		fmt.Errorf("Error unmarshal age from api: %v", err)
		return 0, err
	}

	return ageData.Age, nil
}

func GetExtraGender(name string) (string, error) {
	genderURL := fmt.Sprintf("https://api.genderize.io/?name=%s", name)
	genderResponse, err := http.Get(genderURL)
	if err != nil {
		log.Printf("Error get gender from api: %v", err)
		fmt.Errorf("Error get gender from api: %v", err)
		return "", err
	}
	defer genderResponse.Body.Close()

	genderResponseBody, err := io.ReadAll(genderResponse.Body)
	if err != nil {
		log.Printf("Error read gender from api: %v", err)
		fmt.Errorf("Error read gender from api: %v", err)
		return "", err
	}

	genderData := struct {
		Gender string `json:"gender"`
	}{}
	err = json.Unmarshal(genderResponseBody, &genderData)
	if err != nil {
		log.Printf("Error unmarshal gender from api: %v", err)
		fmt.Errorf("Error unmarshal gender from api: %v", err)
		return "", err
	}
	return genderData.Gender, nil
}

func GetExtraGnationality(name string) (string, error) {
	nationalityURL := fmt.Sprintf("https://api.nationalize.io/?name=%s", name)
	nationalityResponse, err := http.Get(nationalityURL)
	if err != nil {
		log.Printf("Error get nationality from api: %v", err)
		fmt.Errorf("Error get nationality from api: %v", err)
		return "", err
	}
	defer nationalityResponse.Body.Close()

	nationalityResponseBody, err := io.ReadAll(nationalityResponse.Body)
	if err != nil {
		log.Printf("Error read nationality from api: %v", err)
		fmt.Errorf("Error read nationality from api: %v", err)
		return "", err
	}
	nationalityData := struct {
		Country []struct {
			CountryID   string  `json:"country_id"`
			Probability float64 `json:"probability"`
		} `json:"country"`
	}{}
	err = json.Unmarshal(nationalityResponseBody, &nationalityData)
	if err != nil {
		log.Printf("Error unmarshal nationality from api: %v", err)
		fmt.Errorf("Error unmarshal nationality from api: %v", err)
		return "", err
	}
	var result string
	if len(nationalityData.Country) > 0 {
		result = nationalityData.Country[0].CountryID
	}
	return result, nil
}

func AddUsersHandler(c echo.Context, r *background.UserRepository) error {
	var user background.User
	err := c.Bind(&user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	// Запрос на получение возраста
	user.Age, err = GetExtraAge(user.Name)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	// Запрос на получение пола
	user.Gender, err = GetExtraGender(user.Name)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	// Запрос на получение национальности
	user.Nationality, err = GetExtraGnationality(user.Name)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	result, err := r.CreateUser(user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	return c.JSON(http.StatusCreated, result)
}

func UpdateUsersHandler(c echo.Context, r *background.UserRepository) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Error Atoi id: %v", err)
		fmt.Errorf("Error Atoi id: %v", err)
		return c.JSON(http.StatusBadRequest, err)
	}

	user := new(background.User)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	// Обновление данных
	user.ID = id
	updateUser, err := r.UpdateUser(*user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	return c.JSON(http.StatusOK, updateUser)
}

func DeleteUsersHandler(c echo.Context, r *background.UserRepository) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Error Atoi id: %v", err)
		fmt.Errorf("Error Atoi id: %v", err)
		return c.JSON(http.StatusBadRequest, err)
	}
	user, err := r.DeleteUser(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	return c.JSON(http.StatusOK, user)
}
