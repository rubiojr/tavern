package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func CharmAuth(charmURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		charmID, err := charmIDFromRequest(c.Request)
		log.Printf("request from charm ID: %s", charmID)
		if err != nil {
			log.Printf("charmID claim not found: %s", err)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Set("charm_id", charmID)

		reqPath := fmt.Sprintf("%s/v1/id/%s", charmURL, charmID)
		req, err := http.NewRequest("GET", reqPath, nil)
		if err != nil {
			log.Printf("error creating request: %s", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		req.Header.Add("Authorization", c.Request.Header.Get("Authorization"))

		log.Printf("auth against %s", charmURL)
		httpc := &http.Client{}
		resp, err := httpc.Do(req)
		if err != nil {
			log.Printf("sending request to %s failed: %s", charmURL, err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if resp.StatusCode != 200 {
			log.Printf("invalid status code %s from server %s", resp.Status, charmURL)
			c.AbortWithStatus(http.StatusForbidden)
		}
	}

}

func charmIDFromRequest(r *http.Request) (string, error) {
	user := strings.Split(r.Header.Get("Authorization"), " ")[1]
	if user == "" {
		return "", fmt.Errorf("missing user key in context")
	}

	var id string
	jwt.Parse(user, func(t *jwt.Token) (interface{}, error) {
		cl := t.Claims.(jwt.MapClaims)
		var ok bool
		id, ok = cl["sub"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid charmID in token")
		}
		var raw interface{}
		return raw, nil
	})

	if id == "" {
		return "", fmt.Errorf("missing charmID in token")
	}

	return id, nil
}
