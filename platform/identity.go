package platform

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// Headers the gateway sets after it has checked the JWT and the account. The gateway deletes any copy a client sent,
// so a service may trust them. A service must never be reachable except through the gateway (or another service).
const (
	HeaderUserID    = "X-User-Id"
	HeaderProfileID = "X-Profile-Id"
	HeaderRoles     = "X-User-Roles"
	HeaderRequestID = "X-Request-Id"
	// HeaderCountry is the viewer's country (upper-case ISO code) as the gateway resolved it; absent when unknown.
	HeaderCountry = "X-Viewer-Country"
	// HeaderInternalToken authenticates service-to-service calls to /internal/** (see InternalOnly).
	HeaderInternalToken = "X-Internal-Token"
)

// IdentityHeaders are the headers the gateway owns. It removes them from every incoming request.
var IdentityHeaders = []string{HeaderUserID, HeaderProfileID, HeaderRoles, HeaderCountry, HeaderInternalToken}

// Roles, as the monolith names them.
const (
	RoleCustomer       = "CUSTOMER"
	RoleContentManager = "CONTENT_MANAGER"
	RoleAdmin          = "ADMIN"
)

// Identity is the caller, as the gateway described it.
type Identity struct {
	UserID    string
	ProfileID string
	Roles     string
	Country   string
}

// IdentityFrom reads the identity headers from a request.
func IdentityFrom(r *http.Request) Identity {
	return Identity{
		UserID:    r.Header.Get(HeaderUserID),
		ProfileID: r.Header.Get(HeaderProfileID),
		Roles:     r.Header.Get(HeaderRoles),
		Country:   r.Header.Get(HeaderCountry),
	}
}

// User returns the caller's id, or false when the request is not signed in (or the header is not a UUID).
func (id Identity) User() (uuid.UUID, bool) {
	if id.UserID == "" {
		return uuid.Nil, false
	}
	u, err := uuid.Parse(id.UserID)
	return u, err == nil
}

// HasRole reports whether the caller has one of the roles.
func (id Identity) HasRole(roles ...string) bool {
	for _, have := range strings.Split(id.Roles, ",") {
		have = strings.TrimSpace(have)
		for _, want := range roles {
			if have != "" && have == want {
				return true
			}
		}
	}
	return false
}
