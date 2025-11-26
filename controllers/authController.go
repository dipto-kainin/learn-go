package controllers

import (
	"basic-backend/database"
	"basic-backend/helpers"
	"basic-backend/models"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getUserCollection() *mongo.Collection {
	return database.GetCollection(database.Client, "users")
}

var validate = validator.New()

// @Summary User Signup
// @Description Register a new user account with email, password, and profile information
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body models.SignupRequest true "User Registration Details"
// @Success 201 {object} models.SignupResponse "User created successfully with authentication token"
// @Failure 400 {object} models.ErrorResponse "Invalid request body or validation error"
// @Failure 409 {object} models.ErrorResponse "Email already exists in the system"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /auth/signup [post]
func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, err := getUserCollection().CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking email"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}

		hashedPassword, err := helpers.HashPassword(user.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
			return
		}
		user.Password = hashedPassword

		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		user.ID = primitive.NewObjectID()

		token, refreshToken, err := helpers.GenerateAllTokens(user.Email, user.FirstName, user.LastName, user.UserType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating tokens"})
			return
		}

		user.Token = token
		user.RefreshToken = refreshToken

		_, insertErr := getUserCollection().InsertOne(ctx, user)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User was not created"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "User created successfully",
			"token":   token,
			"user":    user,
		})
	}
}

// @Summary User Login
// @Description Authenticate user with email and password, returns JWT access token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "Email and Password"
// @Success 200 {object} models.LoginResponse "Login successful with authentication token and user details"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 401 {object} models.ErrorResponse "Invalid email or password"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /auth/login [post]
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var loginReq models.LoginRequest
		var foundUser models.User

		if err := c.BindJSON(&loginReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := getUserCollection().FindOne(ctx, bson.M{"email": loginReq.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		passwordIsValid := helpers.VerifyPassword(foundUser.Password, loginReq.Password)
		if !passwordIsValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		token, refreshToken, err := helpers.GenerateAllTokens(foundUser.Email, foundUser.FirstName, foundUser.LastName, foundUser.UserType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating tokens"})
			return
		}

		update := bson.M{
			"$set": bson.M{
				"token":         token,
				"refresh_token": refreshToken,
				"updated_at":    time.Now(),
			},
		}

		_, updateErr := getUserCollection().UpdateOne(ctx, bson.M{"_id": foundUser.ID}, update)
		if updateErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating tokens"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"token":   token,
			"user": gin.H{
				"id":         foundUser.ID,
				"email":      foundUser.Email,
				"first_name": foundUser.FirstName,
				"last_name":  foundUser.LastName,
				"user_type":  foundUser.UserType,
			},
		})
	}
}

// @Summary Get Current User
// @Description Get authenticated user profile details (requires valid JWT token)
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.UserSummary "User profile details (password field will be empty)"
// @Failure 401 {object} models.ErrorResponse "Missing or invalid authentication token"
// @Failure 404 {object} models.ErrorResponse "User not found in database"
// @Router /auth/user [get]
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.GetString("email")
		
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var user models.User
		err := getUserCollection().FindOne(ctx, bson.M{"email": email}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		user.Password = "" // Don't send password
		c.JSON(http.StatusOK, user)
	}
}
