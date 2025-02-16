package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
	"golang.org/x/crypto/argon2"

	"github.com/UserNameShouldBeHere/AvitoTask/internal/domain"
	customErrors "github.com/UserNameShouldBeHere/AvitoTask/internal/errors"
)

type AuthStorage interface {
	CreateUser(ctx context.Context, userCreds domain.UserCredantials) error
	GetPassword(ctx context.Context, email string) (string, error)
	HasUser(ctx context.Context, name string) (bool, error)
}

type AuthService struct {
	authStorage    AuthStorage
	logger         *zap.SugaredLogger
	saltLength     int
	jwtKey         []byte
	expirationTime int
}

func NewAuthService(
	authStorage AuthStorage,
	logger *zap.SugaredLogger,
	saltLength int,
	expirationTime int) (*AuthService, error) {

	jwtKey := make([]byte, 16)
	_, err := rand.Read(jwtKey)
	if err != nil {
		return nil, fmt.Errorf("%w (service.NewAuthService): %w", customErrors.ErrFailedToGenJWTKey, err)
	}

	authService := AuthService{
		authStorage:    authStorage,
		logger:         logger,
		saltLength:     saltLength,
		jwtKey:         jwtKey,
		expirationTime: expirationTime,
	}

	return &authService, nil
}

func (authService *AuthService) LoginOrCreateUser(
	ctx context.Context,
	userCreds domain.UserCredantials) (string, error) {
	ok, err := authService.authStorage.HasUser(ctx, userCreds.UserName)
	if err != nil {
		authService.logger.Errorf("failed to check for user (service.LoginOrCreateUser): %w", err)
		return "", fmt.Errorf("(service.LoginOrCreateUser): %w", err)
	}

	if ok {
		err = authService.loginUser(ctx, userCreds)
		if err != nil {
			authService.logger.Errorf("failed to login user (service.LoginOrCreateUser): %w", err)
			return "", fmt.Errorf("(service.LoginOrCreateUser): %w", err)
		}
	} else {
		err = authService.createUser(ctx, userCreds)
		if err != nil {
			authService.logger.Errorf("failed to create user (service.LoginOrCreateUser): %w", err)
			return "", fmt.Errorf("(service.LoginOrCreateUser): %w", err)
		}
	}

	token, err := authService.createToken(userCreds.UserName)
	if err != nil {
		authService.logger.Errorf("failed to create session (service.LoginOrCreateUser): %w", err)
		return "", fmt.Errorf("(service.LoginOrCreateUser): %w", err)
	}

	return token, nil
}

func (authService *AuthService) GetNameAndCheck(ctx context.Context, token string) (string, bool) {
	claims, err := authService.getTokenClaims(token)
	if err != nil {
		authService.logger.Errorf("failed to check session (service.GetNameAndCHeck): %w", err)
		return "", false
	}

	name, ok := (*claims)["name"]

	if !ok {
		authService.logger.Errorf("failed to get name from token (service.GetNameAndCHeck): %w", err)
		return "", false
	}

	return name.(string), true
}

func (authService *AuthService) createUser(ctx context.Context, userCreds domain.UserCredantials) error {
	salt, err := genRandomSalt(authService.saltLength)
	if err != nil {
		authService.logger.Errorf("failed to generate salt (service.createUser): %w", err)
		return fmt.Errorf("%w (service.createUser): %w", customErrors.ErrInternal, err)
	}

	hash, err := hashPassword(userCreds.Password, salt)
	if err != nil {
		authService.logger.Errorf("failed to hash password (service.createUser): %w", err)
		return fmt.Errorf("%w (service.createUser): %w", customErrors.ErrInternal, err)
	}

	hashedPassword := append(salt, hash...)

	userCreds.Password = base64.RawStdEncoding.EncodeToString(hashedPassword)

	err = authService.authStorage.CreateUser(ctx, userCreds)
	if err != nil {
		authService.logger.Errorf("failed to create user (service.createUser): %w", err)
		return fmt.Errorf("(service.createUser): %w", err)
	}

	return nil
}

func (authService *AuthService) loginUser(ctx context.Context, userCreds domain.UserCredantials) error {
	expectedPassword, err := authService.authStorage.GetPassword(ctx, userCreds.UserName)
	if err != nil {
		authService.logger.Errorf("failed to get password (service.loginUser): %w", err)
		return fmt.Errorf("(service.loginUser): %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(expectedPassword)
	if err != nil {
		authService.logger.Errorf("failed to decode password (service.loginUser): %w", err)
		return fmt.Errorf("%w (service.loginUser): %w", customErrors.ErrInternal, err)
	}

	salt := expectedHash[0:authService.saltLength]
	givenHash, err := hashPassword(userCreds.Password, salt)
	if err != nil {
		authService.logger.Errorf("failed to hash password (service.loginUser): %w", err)
		return fmt.Errorf("%w (service.loginUser): %w", customErrors.ErrInternal, err)
	}

	givenPassword := append(salt, givenHash...)

	if expectedPassword != base64.RawStdEncoding.EncodeToString(givenPassword) {
		authService.logger.Errorf("passwords do not match (service.loginUser)")
		return fmt.Errorf("%w (service.loginUser)", customErrors.ErrIncorrectEmailOrPassword)
	}

	return nil
}

func hashPassword(password string, salt []byte) ([]byte, error) {
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	return hash, nil
}

func genRandomSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}

	return salt, nil
}

type myCustomClaims struct {
	Name string `json:"name"`
	jwt.StandardClaims
}

func (authService *AuthService) createToken(name string) (string, error) {
	claims := myCustomClaims{
		name,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * time.Duration(authService.expirationTime)).Unix(),
			Issuer:    "auth",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(authService.jwtKey)
	if err != nil {
		return "", fmt.Errorf("%w (redis.createToken): %w", customErrors.ErrFailedToCreateToken, err)
	}

	return signedToken, nil
}

func (authService *AuthService) getTokenClaims(token string) (*jwt.MapClaims, error) {
	parsedToken, err := jwt.Parse(token,
		func(token *jwt.Token) (interface{}, error) {
			return authService.jwtKey, nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, err
	}

	return &claims, nil
}
