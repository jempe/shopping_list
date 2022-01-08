package main

import (
	"log"
	"net/http"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/qor/admin"
	"github.com/qor/qor"

	"github.com/jempe/shopping_list/models/lists"
)

func main() {
	gormDB, dbError := gorm.Open("sqlite3", "shopping_lists.db")

	if dbError != nil {
		log.Println(dbError)
		os.Exit(1)
	}

	gormDB.AutoMigrate(&lists.List{}, &lists.Item{})

	Admin := admin.New(&admin.AdminConfig{
		DB:       gormDB,
		SiteName: "Shopping Lists",
	})

	Admin.AddResource(&lists.List{})
	Admin.AddResource(&lists.Item{})

	API := admin.New(&qor.Config{DB: gormDB})

	API.AddResource(&lists.List{})
	item := API.AddResource(&lists.Item{})

	item.Action(&admin.Action{
		Name: "check",
		Handler: func(actionArgument *admin.ActionArgument) error {
			for _, record := range actionArgument.FindSelectedRecords() {
				actionArgument.Context.DB.Model(record.(*lists.Item)).Update("Checked", true)
			}

			return nil
		},
	})

	item.Action(&admin.Action{
		Name: "uncheck",
		Handler: func(actionArgument *admin.ActionArgument) error {
			for _, record := range actionArgument.FindSelectedRecords() {
				actionArgument.Context.DB.Model(record.(*lists.Item)).Update("Checked", false)
			}

			return nil
		},
	})

	mux := http.NewServeMux()

	Admin.MountTo("/admin", mux)
	API.MountTo("/api", mux)

	port := "3000"

	log.Println("Serve running on port:", port)
	panic(http.ListenAndServe("127.0.0.1:"+port, mux))
}
