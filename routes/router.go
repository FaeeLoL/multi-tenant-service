package routes

import (
	"github.com/faeelol/multi-tenant-service/controllers"
	"github.com/faeelol/multi-tenant-service/controllers/oauth2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRoutes() *gin.Engine {
	router := gin.Default()

	router.Use(cors.Default())
	router.Use(gin.Logger())
	oauth2.InitOauth2()
	apiGroup := router.Group("/api/v1")
	publicGroup := router.Group("/api/v1")
	authGroup := router.Group("/api/2/idp")
	{
		//authController := new(controllers.AuthController)
		//authMiddleware := authController.Init()
		authGroup.POST("/token", oauth2.HandleTokenRequest)
		//authGroup.GET("/refresh_token", )
		//authGroup.Use(authMiddleware.Middlew areFunc())
		apiGroup.Use(oauth2.VerifyToken)
	}

	tenants := apiGroup.Group("/tenants")
	{
		tenantsController := new(controllers.TenantsController)
		tenants.POST("", tenantsController.CreateTenant)
		tenants.GET("", tenantsController.FetchTenantsBatch)
		tenants.GET("/:tenant_id", tenantsController.GetTenant)
		tenants.PUT("/:tenant_id", tenantsController.UpdateTenant)
		tenants.DELETE("/:tenant_id", tenantsController.DeleteTenant)
		publicGroup.GET("/tenants/:tenant_id/children", tenantsController.GetTenantChildrenList)
		tenants.GET("/:tenant_id/users", tenantsController.GetTenantUsersList)
	}

	users := apiGroup.Group("/users")
	{
		usersController := new(controllers.UsersController)
		users.POST("", usersController.CreateUser)
		users.GET("", usersController.GetUsersBatch)
		users.GET("/:user_id", usersController.GetUser)
		users.PUT("/:user_id", usersController.UpdateUser)
		users.DELETE("/:user_id", usersController.DeleteUser)
		apiGroup.GET("/self_info", usersController.GetSelfInfo)
	}

	aplications := apiGroup.Group("/aplications")
	{
		aplicationsController := new(controllers.ApplicationsController)
		aplications.POST("", aplicationsController.CreateApplication)
		aplications.GET("", aplicationsController.GetApplicationsBatch)
		aplications.GET("/:app_id", aplicationsController.GetApplication)
		aplications.PUT("/:app_id", aplicationsController.UpdateApplication)
		aplications.DELETE("/:app_id", aplicationsController.DeleteApplication)
		aplications.GET("/:app_id/services", aplicationsController.GetAplicationServicesList)
		//tenants.GET("/:tenant_id/:service_id", aplicationsController.GetAplicationService)
	}

	sevices := apiGroup.Group("/services")
	{
		servicesController := new(controllers.ServicesController)
		sevices.POST("", servicesController.CreateService)
		sevices.GET("", servicesController.GetServicesBatch)
		sevices.GET("/:service_id", servicesController.GetService)
		sevices.PUT("/:service_id", servicesController.UpdateService)
		sevices.DELETE("/:service_id", servicesController.DeleteService)
	}
	return router
}
