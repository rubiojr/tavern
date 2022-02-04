package middleware

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// Accepts a list of Charm server hosts allowed for publishing in this
// Tavern instance.
// If the list is empty, any charm host is allowed.
func JWKS(allowedServers map[string]struct{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := getClaims(c.GetHeader("Authorization"))
		if err != nil {
			log.Printf("JWT parsing error: %s", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		issuer, err := url.Parse(claims.Issuer)
		if err != nil {
			log.Printf("valid issuer not found: %s", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if claims.Subject == "" {
			log.Printf("invalid CharmID found: %s", err)
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		c.Set("charm_id", claims.Subject)

		if len(allowedServers) > 0 {
			if _, ok := allowedServers[issuer.Hostname()]; !ok {
				log.Printf("issuer %s not accepted", issuer.Hostname())
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "charm server cannot publish"})
				return
			}
		}

		p := jwks.NewCachingProvider(issuer, 1*time.Hour)
		jwtValidator, err := validator.New(
			p.KeyFunc,
			validator.EdDSA,
			issuer.String(),
			[]string{"tavern"},
		)
		if err != nil {
			log.Printf("could not create validator: %s", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		valid := false
		var handler http.HandlerFunc
		handler = func(http.ResponseWriter, *http.Request) { valid = true }
		middleware := jwtmiddleware.New(jwtValidator.ValidateToken)
		middleware.CheckJWT(handler).ServeHTTP(c.Writer, c.Request)
		if valid {
			c.Next()
		} else {
			log.Printf("JWT validation failed")
			c.Abort()
		}
	}
}

func getClaims(auth string) (*jwt.RegisteredClaims, error) {
	tMinLen := len("Bearer ")
	if len(auth) <= tMinLen {
		return nil, fmt.Errorf("invalid header token")
	}

	encodedToken := auth[tMinLen:]
	p := jwt.Parser{}
	t, _, err := p.ParseUnverified(encodedToken, &jwt.RegisteredClaims{})
	if err != nil {
		return nil, err
	}

	return t.Claims.(*jwt.RegisteredClaims), nil
}
