# Ginmill

Create Gin server with pre-define routes.

It is useful to create compatible endpoint/API set with well-known request patha and focus on implementing gin.Context handlers.

## Example

Implement Features from pre-defined routes.

```go
```

Let your gin server be with the Features

```go

```

Publish your website interface to be known as Features.
Something like:

```go
package mastodon

import (
	"github.com/gin-gonic/gin"
	"github.com/ginmills/ginmill"
)

// IMastodon defines functions must to be done
type IMastodon interface {
	// OAuthAuthorize for GET /oauth/authorize
	OAuthAuthorize(c *gin.Context)
	// OAuthObtainToken for POST /oauth/token
	OAuthObtainToken(c *gin.Context)
	// OAuthRevokeToken for POST /oauth/revoke
	OAuthRevokeToken(c *gin.Context)

}

// Features let you to be a mastodon
func Features(m IMastodon) (features *ginmill.Features) {
	r := gin.New()

	oauth := r.Group("/oauth")

	oauth.GET("/authorize", m.OAuthAuthorize)
	oauth.POST("/token", m.OAuthObtainToken)
	oauth.POST("/revoke", m.OAuthRevokeToken)

	// TODO: add more known api routes

	features = ginmill.NewFeatures(r.Routes())

	return features
}

```

So that people have ginmills with your features.

Full IMastodon Features can be found [here](https://github.com/ginmills/mastodon)
