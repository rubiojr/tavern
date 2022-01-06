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
func JWKS(whitelist map[string]struct{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := getClaims(c.GetHeader("Authorization"))
		if err != nil {
			log.Printf("JWT parsing error: %s", err)
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		issuer, err := url.Parse(claims.Issuer)
		if err != nil {
			log.Printf("valid issuer not found: %s", err)
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		if claims.Subject == "" {
			log.Printf("invalid CharmID found: %s", err)
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		c.Set("charm_id", claims.Subject)

		if len(whitelist) > 0 {
			fmt.Println(whitelist)
			if _, ok := whitelist[issuer.Hostname()]; !ok {
				log.Printf("issuer %s not accepted", issuer.Hostname())
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "charm server cannot publish"})
				return
			}
		}

		fmt.Println(issuer.String())
		p := jwks.NewCachingProvider(issuer, 1*time.Hour)
		jwtValidator, err := validator.New(
			p.KeyFunc,
			validator.RS512,
			issuer.String(),
			[]string{"tavern"},
		)
		if err != nil {
			log.Printf("could not create validator: %s", err)
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		valid := false
		var handler http.HandlerFunc
		handler = func(http.ResponseWriter, *http.Request) { valid = true }
		middleware := jwtmiddleware.New(jwtValidator.ValidateToken)
		middleware.CheckJWT(handler).ServeHTTP(c.Writer, c.Request)
		if valid {
			c.Next()
		} else {
			c.Abort()
		}
	}
}

func getClaims(auth string) (*jwt.RegisteredClaims, error) {
	encodedToken := auth[len("Bearer "):]
	p := jwt.Parser{}
	t, _, err := p.ParseUnverified(encodedToken, &jwt.RegisteredClaims{})
	if err != nil {
		return nil, err
	}

	return t.Claims.(*jwt.RegisteredClaims), nil
}
