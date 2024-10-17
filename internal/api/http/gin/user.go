package gin

import (
	"fmt"
	"net/http"
	"todo-app/domain"
	"todo-app/pkg/clients"
	"todo-app/pkg/tokenprovider"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserService interface {
	Register(data *domain.UserCreate) error
	Login(data *domain.UserLogin) (tokenprovider.Token, error)
	GetAllUser() ([]domain.User, error)
	GetUserByID(id uuid.UUID) (domain.User, error)
	UpdateUser(id uuid.UUID, user *domain.UserUpdate) error
	DeleteUser(id uuid.UUID) error
}

type userHandler struct {
	userService UserService
}

func NewUserHandler(apiVersion *gin.RouterGroup, svc UserService, middlewareAuth func(c *gin.Context), middlewareRateLimit func(c *gin.Context)) {
	userHandler := &userHandler{
		userService: svc,
	}

	users := apiVersion.Group("/users")
	users.POST("/register", userHandler.RegisterUserHandler)
	users.POST("/login", userHandler.LoginHandler)
	users.GET("", middlewareAuth, userHandler.GetAllUserHandler)
	users.GET("/:id", middlewareAuth, userHandler.GetUserHandler)
	users.PATCH("/:id", middlewareAuth, userHandler.UpdateUserHandler)
	users.DELETE("/:id", middlewareAuth, userHandler.DeleteUserHandler)
}

// RegisterUserHandler handles user registration.
//
// @Summary      Register a new user
// @Description  This endpoint allows new users to register by providing their details.
//
//	A successful registration returns the user's ID.
//
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user  body      domain.UserCreate  true  "User registration payload"
// @Success      201   {object}  clients.SuccessRes   "User successfully registered"
// @Failure      400   {object}  clients.AppError     "Bad Request - Invalid input data"
// @Failure      500   {object}  clients.AppError     "Internal Server Error - Unexpected error"
// @Router       /users/register [post]
func (h *userHandler) RegisterUserHandler(c *gin.Context) {
	var data domain.UserCreate

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusBadRequest, clients.ErrInvalidRequest(err))
		return
	}

	if err := h.userService.Register(&data); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusCreated, clients.SimpleSuccessResponse(data.ID))
}

// LoginHandler handles user login.
//
// @Summary      User login
// @Description  This endpoint allows users to log in using their credentials and receive an authentication token.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user  body      domain.UserLogin  true  "User login payload"
// @Success      200   {object}  clients.SuccessRes   "User successfully logged in"
// @Failure      400   {object}  clients.AppError     "Bad Request - Invalid login credentials"
// @Failure      500   {object}  clients.AppError     "Internal Server Error - Unexpected error"
// @Router       /users/login [post]
func (h *userHandler) LoginHandler(c *gin.Context) {
	var data domain.UserLogin

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusBadRequest, clients.ErrInvalidRequest(err))
		return
	}

	token, err := h.userService.Login(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, clients.SimpleSuccessResponse(token))
}

// GetAllUserHandler retrieves all users.
//
// @Summary      Get all users
// @Description  This endpoint retrieves a list of all registered users. It is accessible only to admin users.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Success      200  {object}  clients.SuccessRes  "List of users retrieved successfully"
// @Failure      500  {object}  clients.AppError    "Internal Server Error - Unexpected error"
// @Router       /users [get]
func (h *userHandler) GetAllUserHandler(c *gin.Context) {
	items, err := h.userService.GetAllUser()
	if err != nil {
		c.JSON(http.StatusBadRequest, clients.ErrInvalidRequest(err))
		return
	}

	c.JSON(http.StatusOK, clients.SimpleSuccessResponse(items))
}

// GetUserHandler retrieves a user by ID.
//
// @Summary      Get a user by ID
// @Description  This endpoint retrieves a user by their unique ID.
//
//	If the user is not found, an appropriate error message is returned.
//
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id   path      string                 true  "User ID"
// @Success      200  {object}  clients.SuccessRes     "User retrieved successfully"
// @Failure      400  {object}  clients.AppError       "Invalid ID format or bad request"
// @Failure      404  {object}  clients.AppError       "User not found"
// @Failure      500  {object}  clients.AppError       "Internal Server Error - Unexpected error"
// @Router       /users/{id} [get]
func (h *userHandler) GetUserHandler(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, clients.ErrInvalidRequest(err))
		return
	}

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, clients.SimpleSuccessResponse(user))
}

// UpdateUserHandler updates an existing user.
//
// @Summary      Update a user
// @Description  This endpoint allows updating the properties of an existing user by their ID.
//
//	It requires the user ID and the updated user data.
//
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id    path      string                 true  "User ID"
// @Param        user  body      domain.UserUpdate      true  "User update payload"
// @Success      200   {object}  clients.SuccessRes     "User updated successfully"
// @Failure      400   {object}  clients.AppError       "Invalid input or bad request"
// @Failure      404   {object}  clients.AppError       "User not found"
// @Failure      500   {object}  clients.AppError       "Internal Server Error - Unexpected error"
// @Router       /users/{id} [patch]
func (h *userHandler) UpdateUserHandler(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, clients.ErrInvalidRequest(err))
		return
	}

	user := domain.UserUpdate{}
	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusBadRequest, clients.ErrInvalidRequest(err))
		return
	}
	var user1 domain.User
	requester := c.MustGet(clients.CurrentUser).(clients.Requester)
	user1.ID = requester.GetUserID()

	if user1.ID != id {
		c.JSON(http.StatusUnauthorized, clients.ErrInvalidRequest(fmt.Errorf("unauthorized: ID does not match")))
		return
	}

	if err := h.userService.UpdateUser(id, &user); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, clients.SimpleSuccessResponse(true))
}

// DeleteUserHandler deletes a user by ID.
//
// @Summary      Delete a user
// @Description  This endpoint deletes a user identified by their unique ID.
//
//	If the user is not found, an appropriate error message is returned.
//
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id   path      string                 true  "User ID"
// @Success      200  {object}  clients.SuccessRes     "User deleted successfully"
// @Failure      400  {object}  clients.AppError       "Invalid ID format or bad request"
// @Failure      404  {object}  clients.AppError       "User not found"
// @Failure      500  {object}  clients.AppError       "Internal Server Error - Unexpected error"
// @Router       /users/{id} [delete]
func (h *userHandler) DeleteUserHandler(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, clients.ErrInvalidRequest(err))
		return
	}

	if err := h.userService.DeleteUser(id); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, clients.SimpleSuccessResponse(true))
}
