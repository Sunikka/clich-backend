package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sunikka/clich-backend/internal/database"
	"github.com/sunikka/clich-backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

var jwt_key string = os.Getenv("JWT_SECRET")

func GenerateToken(user database.User) (string, error) {
	// TODO: What claims are necessary?
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.UserID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(jwt_key))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func validateJWT(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwt_key), nil
	})
}

// User endpoint JWT middleware
func ProtectedEndpointMW(handlerFunc http.HandlerFunc, store database.Queries, adminOnly bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr, err := GetTokenString(r)
		if err != nil {
			permissionDeniedRes(w)
			return
		}

		token, err := validateJWT(tokenStr)
		if err != nil {
			permissionDeniedRes(w)
			return
		}

		userID, err := utils.GetUserID(r)
		if err != nil {
			permissionDeniedRes(w)
			return
		}

		user, err := store.GetUserByID(r.Context(), userID)
		if err != nil {
			permissionDeniedRes(w)
			return
		}

		// Check if user ID in tokens claims matches the database user selected in the request
		if user.UserID.String() != token.Claims.(jwt.MapClaims)["user_id"].(string) {
			permissionDeniedRes(w)
			return
		}

		// Check if user has admin privileges if the endpoint is admin only
		if adminOnly && !user.Admin {
			permissionDeniedRes(w)
			return
		}

		handlerFunc(w, r)
	}
}

func CheckPassword(user database.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPw), []byte(password))
	return err == nil
}

func GetTokenString(r *http.Request) (string, error) {
	authHeaderContent := r.Header.Get("Authorization")

	if authHeaderContent == "" {
		return "", errors.New("authentication failed")
	}

	values := strings.Split(authHeaderContent, " ")
	if len(values) != 2 {
		return "", errors.New("authentication failed")
	}

	if values[0] != "JWT" {
		return "", errors.New("authentication failed")
	}

	return values[1], nil

}

func permissionDeniedRes(w http.ResponseWriter) {
	utils.RespondJSON(w, http.StatusForbidden, fmt.Sprintf("Permission denied"))
}
