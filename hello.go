package main

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/gothinkster/golang-gin-realworld-example-app/items"
	"github.com/gothinkster/golang-gin-realworld-example-app/common"
	"./users"
	"github.com/jinzhu/gorm"
)

func Migrate(db *gorm.DB) {
	users.AutoMigrate()
	db.AutoMigrate(&items.ItemModel{})
	db.AutoMigrate(&items.TagModel{})
	db.AutoMigrate(&items.FavoriteModel{})
	db.AutoMigrate(&items.ItemUserModel{})
	db.AutoMigrate(&items.CommentModel{})
}

func main() {

	db := common.Init()
	Migrate(db)
	defer db.Close()

	r := gin.Default()

	v1 := r.Group("/api")
	users.UsersRegister(v1.Group("/users"))
	v1.Use(users.AuthMiddleware(false))
	items.ItemsAnonymousRegister(v1.Group("/items"))
	items.TagsAnonymousRegister(v1.Group("/tags"))

	v1.Use(users.AuthMiddleware(true))
	users.UserRegister(v1.Group("/user"))
	users.ProfileRegister(v1.Group("/profiles"))

	items.ItemsRegister(v1.Group("/items"))

	testAuth := r.Group("/api/ping")

	testAuth.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// test 1 to 1
	tx1 := db.Begin()
	userA := users.UserModel{
		Username: "AAAAAAAAAAAAAAAA",
		Email:    "aaaa@g.cn",
		Bio:      "hehddeda",
		Image:    nil,
	}
	tx1.Save(&userA)
	tx1.Commit()
	fmt.Println(userA)

	//db.Save(&ItemUserModel{
	//    UserModelID:userA.ID,
	//})
	//var userAA ItemUserModel
	//db.Where(&ItemUserModel{
	//    UserModelID:userA.ID,
	//}).First(&userAA)
	//fmt.Println(userAA)

	r.Run(":3000") // listen and serve on 0.0.0.0:8080
}
