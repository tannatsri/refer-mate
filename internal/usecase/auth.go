package usecase

import (
	"context"
	"errors"
	"fmt"
	"refer-mate/internal/domain"
	"refer-mate/internal/infrastrcture/gmail"
	"refer-mate/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type AuthUseCase struct {
	userRepo    repository.UserRepository
	tokenRepo   repository.TokenRepository
	gmailClient *gmail.Client
	jwtSecret   string
	jwtExpiry   time.Duration
}

func NewAuthUseCase(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	gmailClient *gmail.Client,
	jwtSecret string,
	jwtExpiryHours int,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		gmailClient: gmailClient,
		jwtSecret:   jwtSecret,
		jwtExpiry:   time.Duration(jwtExpiryHours) * time.Hour,
	}
}

func (uc *AuthUseCase) GetAuthURL(state string) string {
	return uc.gmailClient.AuthCodeURL(state)
}

func (uc *AuthUseCase) HandleCallback(ctx context.Context, code string) (string, *domain.User, error) {
	token, err := uc.gmailClient.Exchange(ctx, code)
	if err != nil {
		return "", nil, fmt.Errorf("token exchange failed: %w", err)
	}

	info, err := uc.gmailClient.GetUserInfo(ctx, token)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get user info: %w", err)
	}

	user, err := uc.userRepo.FindByGoogleID(ctx, info.ID)
	if err != nil {
		return "", nil, err
	}

	if user == nil {
		user, err = uc.userRepo.Create(ctx, &domain.User{
			GoogleID:       info.ID,
			Email:          info.Email,
			Name:           info.Name,
			ProfilePicture: info.Picture,
		})
		if err != nil {
			return "", nil, err
		}
	} else {
		user.Name = info.Name
		user.ProfilePicture = info.Picture
		user, err = uc.userRepo.Update(ctx, user)
		if err != nil {
			return "", nil, err
		}
	}

	oauthToken := &domain.OAuthToken{
		UserID:       user.ID,
		Provider:     "google",
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
	}
	if _, err := uc.tokenRepo.Upsert(ctx, oauthToken); err != nil {
		return "", nil, err
	}

	jwtToken, err := uc.generateJWT(user.ID, user.Email)
	if err != nil {
		return "", nil, err
	}

	return jwtToken, user, nil
}

func (uc *AuthUseCase) GetMe(ctx context.Context, userID int64) (*domain.User, error) {
	return uc.userRepo.FindByID(ctx, userID)
}

func (uc *AuthUseCase) ValidateJWT(tokenStr string) (int64, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(uc.jwtSecret), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("invalid user_id claim")
	}
	return int64(userIDFloat), nil
}

func (uc *AuthUseCase) GetOAuthToken(ctx context.Context, userID int64) (*oauth2.Token, error) {
	t, err := uc.tokenRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, errors.New("no OAuth token found for user")
	}
	return uc.gmailClient.TokenFromDomain(t), nil
}

func (uc *AuthUseCase) generateJWT(userID int64, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(uc.jwtExpiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.jwtSecret))
}
