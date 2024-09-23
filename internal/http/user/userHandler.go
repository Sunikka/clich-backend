package userService

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/sunikka/clich-backend/internal/auth"
	"github.com/sunikka/clich-backend/internal/database"
	"github.com/sunikka/clich-backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

func NewUserService(listenAddr string, storage *database.Queries) *UserService {
	return &UserService{
		listenAddr: ":" + listenAddr,
		storage:    storage,
	}
}

func (userService UserService) Run(port string) {
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	routerV1 := chi.NewRouter()

	routerV1.Post("/login", userService.handleLogin)
	routerV1.Post("/register", userService.handleRegister)

	routerV1.Get("/users", auth.ProtectedEndpointMW(userService.handleGetUsers, *userService.storage, false))
	routerV1.Get("/users/{id}", auth.ProtectedEndpointMW(userService.handleGetUserByID, *userService.storage, false))
	routerV1.Delete("/users/{id}", auth.ProtectedEndpointMW(userService.handleDeleteUser, *userService.storage, false))
	routerV1.Put("/users/{id}", auth.ProtectedEndpointMW(userService.handleUpdateUser, *userService.storage, false))

	// http.HandleFunc("/health", s.handleHealth)
	router.Mount("/v1", routerV1)
	server := &http.Server{
		Handler: router,
		Addr:    ":" + port,
	}

	log.Printf("Login Service listening on port: %s ", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func (userService *UserService) handleLogin(w http.ResponseWriter, r *http.Request) {
	req := auth.LoginRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondErrJSON(w, http.StatusBadRequest, fmt.Sprintf("Login failed: %v", err))
		return
	}

	// first login version is by username, however it will be probably remade later, since this forces making usernames unique
	user, err := userService.storage.GetUserByName(r.Context(), req.Username)
	if err != nil {
		utils.RespondErrJSON(w, http.StatusBadRequest, fmt.Sprintf("Login failed"))
		log.Printf("Login error for username %s: %v ", user.Username, err)
		return
	}

	if auth.CheckPassword(user, req.Password) == false {
		utils.RespondErrJSON(w, http.StatusForbidden, fmt.Sprintf("Access denied, Wrong username or password"))
		return
	}

	token, err := auth.GenerateToken(user)
	if err != nil {
		log.Printf("token generation failed: %v", err)
		return
	}

	response := auth.LoginResponse{
		Username: user.Username,
		Token:    token,
	}

	utils.RespondJSON(w, http.StatusOK, response)
}

func (userService *UserService) handleRegister(w http.ResponseWriter, r *http.Request) {
	params := auth.RegisterRequest{}

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		utils.RespondErrJSON(w, http.StatusBadRequest, fmt.Sprintf("error parsing JSON: %v", err))
		return
	}

	hashedPW, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.RespondErrJSON(w, 400, fmt.Sprintf("Account registration failed: %v", err))
		return
	}

	user, err := userService.storage.CreateUser(r.Context(), database.CreateUserParams{
		UserID:    uuid.New(),
		Username:  params.Username,
		HashedPw:  string(hashedPW),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})

	if err != nil {
		utils.RespondErrJSON(w, http.StatusBadRequest, fmt.Sprintf("Account registration failed: %v", err))
		return
	}

	utils.RespondJSON(w, http.StatusCreated, user)
}

func (userService *UserService) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	usersDB, err := userService.storage.GetUsers(r.Context())
	if err != nil {
		utils.RespondErrJSON(w, http.StatusNotFound, fmt.Sprintf("Couldn't fetch users, error: %v", err))
		return
	}

	users := []userRes{}
	for i := range usersDB {
		users = append(users, userToUserRes(usersDB[i]))
	}

	utils.RespondJSON(w, http.StatusOK, users)
}

func (userService *UserService) handleGetUserByID(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserID(r)
	if err != nil {
		utils.RespondErrJSON(w, http.StatusBadRequest, fmt.Sprintf("Parsing user ID failed: %v", err))
		return
	}

	user, err := userService.storage.GetUserByID(r.Context(), userID)

	utils.RespondJSON(w, http.StatusOK, userToUserRes(user))
}

func (userService *UserService) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserID(r)
	if err != nil {
		utils.RespondErrJSON(w, http.StatusBadRequest, fmt.Sprintf("Parsing user ID failed: %v", err))
		return
	}

	err = userService.storage.DeleteUserByID(r.Context(), userID)
	if err != nil {
		utils.RespondErrJSON(w, http.StatusBadRequest, fmt.Sprintf("User deletion failed: %v", err))
		return
	}

	utils.RespondJSON(w, http.StatusOK, fmt.Sprintf("Succesfully deleted user: %v", userID))
}

func (userService *UserService) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	updatedInfo := UserUpdateReq{}
	err := json.NewDecoder(r.Body).Decode(&updatedInfo)
	if err != nil {
		utils.RespondErrJSON(w, http.StatusBadRequest, fmt.Sprintf("error parsing JSON: %v", err))
		return
	}

	userID, err := utils.GetUserID(r)
	if err != nil {
		utils.RespondErrJSON(w, http.StatusBadRequest, fmt.Sprintf("Parsing user ID failed: %v", err))
		return
	}

	user, err := userService.storage.GetUserByID(r.Context(), userID)
	params := database.UpdateUserParams{
		Username:  user.Username,
		UpdatedAt: time.Now().UTC(),
		UserID:    userID,
	}

	applyUserUpdates(updatedInfo, &params)

	updatedUser, err := userService.storage.UpdateUser(r.Context(), params)
	utils.RespondJSON(w, http.StatusOK, updatedUser)
}

func applyUserUpdates(req UserUpdateReq, updateParams *database.UpdateUserParams) {
	if req.Username != "" {
		updateParams.Username = req.Username
	}

	// if req.Password != "" {
	// 	updateParams.Password = req.Password
	// }
}

func userToUserRes(user database.User) userRes {
	return userRes{
		UserID:    user.UserID,
		Username:  user.Username,
		Admin:     user.Admin,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
