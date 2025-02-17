package controller

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"matchx/dbServer"
	"matchx/models"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PropertyWithDistance struct {
	Property models.RegisterProperty `json:"property"`
	Distance float64                 `json:"distance"`
}

func GetNearbyProps(ctx *gin.Context) {
	latitude, longitude := ctx.Query("lat"), ctx.Query("lng")
	log.Println(latitude, longitude)
	lat, err1 := strconv.ParseFloat(latitude, 64)
	lng, err2 := strconv.ParseFloat(longitude, 64)

	if err1 != nil || err2 != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude or longitude"})
		return
	}

	properties, responseErr := getLocations(lat, lng, dbServer.Dbhandler)
	if responseErr != nil {
		ctx.AbortWithStatusJSON(responseErr.Status, responseErr)
		return
	}

	ctx.JSON(http.StatusOK, properties)

}

func getLocations(lat float64, lng float64, dbHandler *sql.DB) ([]PropertyWithDistance, *models.ResponseError) {
	query := `
		SELECT id, user_id, property_type, longitude, latitude, locality, lease_type, 
		       furnished_status, property_area, internet, ac, ro, kitchen, geezer,
		       (6371 * acos(
		           cos(radians($1)) * cos(radians(latitude)) * 
		           cos(radians(longitude) - radians($2)) + 
		           sin(radians($1)) * sin(radians(latitude))
		       )) AS distance 
		FROM property
		WHERE (6371 * acos(
		          cos(radians($1)) * cos(radians(latitude)) * 
		          cos(radians(longitude) - radians($2)) + 
		          sin(radians($1)) * sin(radians(latitude))
		      )) <= 15 
		ORDER BY distance ASC;`

	rows, err := dbHandler.Query(query, lat, lng)
	if err != nil {
		return nil, &models.ResponseError{
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		}
	}
	defer rows.Close()

	var properties []PropertyWithDistance
	for rows.Next() {
		var p models.RegisterProperty
		var d float64

		// Scan values from the row
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.PropertyType, &p.Location.Longitude, &p.Location.Latitude, &p.Locality,
			&p.LeaseType, &p.FurnishedStatus, &p.PropertyArea, &p.Internet, &p.AC,
			&p.RO, &p.Kitchen, &p.Geezer, &d,
		); err != nil {
			return nil, &models.ResponseError{
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			}
		}

		// Append to slice
		properties = append(properties, PropertyWithDistance{
			Property: p,
			Distance: d,
		})
	}

	return properties, nil
}

func GetUser(ctx *gin.Context) {
	userMail := ctx.Query("email")
	userMail, err := url.QueryUnescape(userMail)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email"})
		return
	}
	if userMail == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.ResponseError{
			Message: "Invalid user mail",
			Status:  http.StatusBadRequest,
		})
		return
	}
	user, responseErr := FetchUserByMail(userMail, dbServer.Dbhandler)
	if responseErr != nil {
		ctx.AbortWithStatusJSON(responseErr.Status, responseErr)
		return
	}
	ctx.JSON(http.StatusCreated, user)
}

func RegisterNP(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Println("Error while reading Register new property request body", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var property models.RegisterProperty
	err = json.Unmarshal(body, &property)
	if err != nil {
		log.Println("Error while unmarshaling Register new property request body", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	// log.Println(property)

	property_id, responseErr := createProperty(&property, dbServer.Dbhandler)
	if responseErr != nil {
		ctx.AbortWithStatusJSON(responseErr.Status, responseErr)
		return
	}
	ctx.JSON(http.StatusCreated, property_id)
}

func Login(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Println("Error while reading Login user request body", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	var user models.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		log.Println("Error while unmarshaling Login user request body", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if user.Username == "" {
		responseErr := models.ResponseError{
			Message: "Invalid username",
			Status:  http.StatusBadRequest,
		}
		ctx.AbortWithStatusJSON(responseErr.Status, responseErr)
		return
	}
	if user.Email == "" {
		responseErr := models.ResponseError{
			Message: "Invalid email",
			Status:  http.StatusBadRequest,
		}
		ctx.AbortWithStatusJSON(responseErr.Status, responseErr)
		return
	}
	if user.Name == "" {
		responseErr := models.ResponseError{
			Message: "Invalid name",
			Status:  http.StatusBadRequest,
		}
		ctx.AbortWithStatusJSON(responseErr.Status, responseErr)
		return
	}
	if user.Role == "" {
		responseErr := models.ResponseError{
			Message: "Invalid role",
			Status:  http.StatusBadRequest,
		}
		ctx.AbortWithStatusJSON(responseErr.Status, responseErr)
		return
	}

	present, responseErr := checkUser(&user, dbServer.Dbhandler)
	if responseErr != nil {
		ctx.AbortWithStatusJSON(responseErr.Status, responseErr)
		return
	} else if present {
		ctx.JSON(http.StatusFound, user.Username)
		return
	} else {
		usnm, responseErr := createUser(&user, dbServer.Dbhandler)
		if responseErr != nil {
			ctx.AbortWithStatusJSON(responseErr.Status, responseErr)
			return
		}
		ctx.JSON(http.StatusCreated, usnm)
		return
	}
}

func checkUser(user *models.User, dbHandler *sql.DB) (bool, *models.ResponseError) {
	query := `select username from users where username=$1`
	rows, err := dbHandler.Query(query, user.Username)
	if err != nil {
		return false, &models.ResponseError{
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		}
	}
	defer rows.Close()
	var username_user string
	for rows.Next() {
		err := rows.Scan(&username_user)
		if err != nil {
			return false, &models.ResponseError{
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			}
		}
	}
	if rows.Err() != nil {
		return false, &models.ResponseError{
			Message: "Error while reading rows",
			Status:  http.StatusInternalServerError,
		}
	}
	if username_user == "" {
		return false, nil
	} else {
		return true, nil
	}
}

func createUser(user *models.User, dbHandler *sql.DB) (string, *models.ResponseError) {
	query := `insert into users(username, full_name, user_role, email, picture)
				values($1,$2,$3,$4,$5)
				returning username`
	rows, err := dbHandler.Query(query, user.Username, user.Name, user.Role, user.Email, user.Picture)
	if err != nil {
		return "", &models.ResponseError{
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		}
	}
	defer rows.Close()
	var username string
	for rows.Next() {
		err := rows.Scan(&username)
		if err != nil {
			return "", &models.ResponseError{
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			}
		}
	}
	if rows.Err() != nil {
		return "", &models.ResponseError{
			Message: "Error while reading rows",
			Status:  http.StatusInternalServerError,
		}
	}
	return username, nil
}

func createProperty(property *models.RegisterProperty, dbHandler *sql.DB) (string, *models.ResponseError) {
	query := `insert into property(user_id, property_type, longitude, latitude, locality, lease_type, furnished_status, property_area, internet, ac, ro, kitchen, geezer)
				values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
				returning id`
	rows, err := dbHandler.Query(query, property.UserID, property.PropertyType, property.Location.Longitude, property.Location.Latitude, property.Locality, property.LeaseType, property.FurnishedStatus, property.PropertyArea, property.Internet, property.AC, property.RO, property.Kitchen, property.Geezer)
	if err != nil {
		return "", &models.ResponseError{
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		}
	}
	defer rows.Close()
	var id string
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return "", &models.ResponseError{
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			}
		}
	}
	if rows.Err() != nil {
		return "", &models.ResponseError{
			Message: "Error while reading rows",
			Status:  http.StatusInternalServerError,
		}
	}
	return id, nil
}

func FetchUserByMail(userMail string, dbHandler *sql.DB) (*models.User, *models.ResponseError) {
	query := `select *
				 from users 
				 where email=$1`
	rows, err := dbHandler.Query(query, userMail)
	if err != nil {
		return nil, &models.ResponseError{
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		}
	}
	defer rows.Close()

	var id, username, full_name, user_role, email, picture string
	for rows.Next() {
		err := rows.Scan(&id, &username, &full_name, &user_role, &email, &picture)
		if err != nil {
			return nil, &models.ResponseError{
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			}
		}
	}
	if rows.Err() != nil {
		return nil, &models.ResponseError{
			Message: "Error while reading rows",
			Status:  http.StatusInternalServerError,
		}
	}
	if id == "" {
		return nil, &models.ResponseError{
			Message: "User not found",
			Status:  http.StatusNotFound,
		}
	} else {
		return &models.User{
			ID:       id,
			Username: username,
			Name:     full_name,
			Role:     user_role,
			Email:    email,
			Picture:  picture,
		}, nil
	}
}
