package auth

import (
	"context"
	"strings"
)

type contextKey string

const claimsContextKey contextKey = "flowbit.auth.claims"

// Claims is the subset of Clerk JWT claims Flowbit stores/uses.
type Claims struct {
	Subject   string
	Email     string
	FirstName string
	LastName  string
	ImageURL  string
}

func ContextWithClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(Claims)
	if !ok || strings.TrimSpace(claims.Subject) == "" {
		return Claims{}, false
	}
	return claims, true
}
