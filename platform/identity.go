package platform

import "net/http"

// Headers the gateway sets after it has checked the JWT. Services trust them and never see the token.
const (
	HeaderUserID    = "X-User-Id"
	HeaderProfileID = "X-Profile-Id"
	HeaderRoles     = "X-User-Roles"
	HeaderRequestID = "X-Request-Id"
)

// Identity is the caller, as the gateway described it.
type Identity struct {
	UserID    string
	ProfileID string
	Roles     string
}

// IdentityFrom reads the identity headers from a request.
func IdentityFrom(r *http.Request) Identity {
	return Identity{
		UserID:    r.Header.Get(HeaderUserID),
		ProfileID: r.Header.Get(HeaderProfileID),
		Roles:     r.Header.Get(HeaderRoles),
	}
}
