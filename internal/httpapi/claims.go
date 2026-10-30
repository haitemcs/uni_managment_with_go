package httpapi

import "github.com/haitemcs/uni_managment_with_go/internal/auth"

type ClaimsAlias = auth.Claims

func hasRole(value any, roles []string) bool {
	claims, ok := value.(*auth.Claims)
	if !ok {
		return false
	}
	for _, role := range roles {
		if claims.Role == role {
			return true
		}
	}
	return false
}

func currentClaims(value any) (*auth.Claims, bool) {
	claims, ok := value.(*auth.Claims)
	return claims, ok
}
