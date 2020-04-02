package routes

import (
	"github.com/faeelol/multi-tenant-service/controllers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRoutes() *gin.Engine {
	router := gin.Default()

	router.Use(cors.Default())
	router.Use(gin.Logger())
	//router.Use(gin.Recovery())

	apiGroup := router.Group("/api/v1")
	authGroup := apiGroup.Group("/auth")
	{
		authController := new(controllers.AuthController)
		authMiddleware := authController.Init()
		authGroup.POST("login", authMiddleware.LoginHandler)
		authGroup.GET("/refresh_token", authMiddleware.RefreshHandler)
		//authGroup.Use(authMiddleware.MiddlewareFunc())
		apiGroup.Use(authMiddleware.MiddlewareFunc())
	}


	//router.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
	//	claims := jwt.ExtractClaims(c)
	//	log.Printf("NoRoute claims: %#v\n", claims)
	//	c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	//})

	tenants := apiGroup.Group("/tenants")
	{
		tenantsController := new(controllers.TenantsController)
		tenants.POST("/", tenantsController.CreateTenant)
		tenants.GET("/", tenantsController.FetchTenantsBatch)
		tenants.GET("/:tenant_id", tenantsController.GetTenant)
		tenants.PUT("/:tenant_id", tenantsController.UpdateTenant)
	//	tenants.DELETE("/:tenants_id", tenantsController.DeleteTenant)
	//	tenants.GET("/:tenants_id/children", tenantsController.GetTenantChildrenList)
	//	tenants.GET("/:tenants_id/users", tenantsController.GetTenantUsersList)
	}

	users := apiGroup.Group("/users")
	{
		usersController := new(controllers.UsersController)
		users.POST("/", usersController.CreateUser)
		users.GET("/", usersController.GetUsersBatch)
		//users.GET("me", usersController.GetSelfInfo)
		//users.GET("/:user_id", usersController.GetUser)
		//users.PUT(":user_id", usersController.UpdateUser)
		//users.PUT(":user_id", usersController.DeleteUser)
	}

	return router
}
