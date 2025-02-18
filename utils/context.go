package utils

import (
	"context"
)

type contextKey string

const UserClaimsKey contextKey = "user_claims"

func ContextWithUser(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, UserClaimsKey, claims)
}

func UserFromContext(ctx context.Context) *Claims {
	claims, ok := ctx.Value(UserClaimsKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}
