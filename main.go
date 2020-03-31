package main

import (
	"github.com/faeelol/multi-tenant-service/database"
	"github.com/faeelol/multi-tenant-service/routes"
)

func main() {
	router := routes.InitRoutes()
	db, err := database.InitDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = router.Run(":8080")
	if err != nil {
		panic(err)
	}
}
