package security

import (
	"log"
	"net/http"

	"github.com/ahmedsaleh747/go-creative-utils/shared"
	"github.com/ahmedsaleh747/go-creative-utils/storage"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(claims shared.IdentityClaims, skipPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" && c.Request.URL.Path == skipPath {
			c.Next()
			return
		}

		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			c.Abort()
			return
		}

		tokenStr = tokenStr[len("Bearer "):]
		if err := VerifyToken(tokenStr, claims); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Save the username in the context
		c.Set("user", claims)

		//Sample code using the claims
		user := c.MustGet("user").(shared.IdentityClaims)
		log.Printf("User: [%v:%s] calling %s", user.GetUserId(), user.GetUsername(), c.Request.URL.Path)

		// Pass on to the next-in-chain
		c.Next()
	}
}

func WithRole(allowedRole string) gin.HandlerFunc {
	return WithRoles([]string{allowedRole})
}

func WithRoles(allowedRoles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userMeta, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// Cast userRole to Role struct
		meta, ok := userMeta.(*shared.UserMeta)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid role data"})
			c.Abort()
			return
		}

		// Check if the user's role is in the list of allowed roles
		roleAllowed := false
		for _, allowedRole := range allowedRoles {
			if meta.Role == allowedRole {
				roleAllowed = true
				break
			}
		}

		if !roleAllowed {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// If the role matches, proceed to the handler
		c.Next()
	}
}

func Login(c *gin.Context, user storage.Identity, claims shared.IdentityClaims) {
	if !storage.GetUserUsingNameAndPassword(c, user) {
		return
	}
	claims.SetClaims(user)
	log.Printf("Login succedded for user: %s[%s]", claims.GetUsername(), claims.GetRole())

	token, err := GenerateToken(claims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
