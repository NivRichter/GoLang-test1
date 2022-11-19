package items

import (
	"errors"
	"github.com/NivRichter/GoLang-test1/common"
	"github.com/NivRichter/GoLang-test1/users"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func ItemsRegister(router *gin.RouterGroup) {
	router.POST("/", ItemCreate)
	router.PUT("/:slug", ItemUpdate)
	router.DELETE("/:slug", ItemDelete)
	router.POST("/:slug/favorite", ItemFavorite)
	router.DELETE("/:slug/favorite", ItemUnfavorite)
	router.POST("/:slug/comments", ItemCommentCreate)
	router.DELETE("/:slug/comments/:id", ItemCommentDelete)
}

func ItemsAnonymousRegister(router *gin.RouterGroup) {
	router.GET("/", ItemList)
	router.GET("/:slug", ItemRetrieve)
	router.GET("/:slug/comments", ItemCommentList)
}

func TagsAnonymousRegister(router *gin.RouterGroup) {
	router.GET("/", TagList)
}

func ItemCreate(c *gin.Context) {
	itemModelValidator := NewItemModelValidator()
	if err := itemModelValidator.Bind(c); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		return
	}
	//fmt.Println(itemModelValidator.itemModel.Seller.UserModel)

	if err := SaveOne(&itemModelValidator.itemModel); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("database", err))
		return
	}
	serializer := ItemSerializer{c, itemModelValidator.itemModel}
	c.JSON(http.StatusCreated, gin.H{"item": serializer.Response()})
}

func ItemList(c *gin.Context) {
	//condition := ItemModel{}
	tag := c.Query("tag")
	seller := c.Query("seller")
	favorited := c.Query("favorited")
	limit := c.Query("limit")
	offset := c.Query("offset")
	itemModels, modelCount, err := FindManyItem(tag, seller, limit, offset, favorited)
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewError("items", errors.New("Invalid param")))
		return
	}
	serializer := ItemsSerializer{c, itemModels}
	c.JSON(http.StatusOK, gin.H{"items": serializer.Response(), "itemsCount": modelCount})
}

func ItemFeed(c *gin.Context) {
	limit := c.Query("limit")
	offset := c.Query("offset")
	myUserModel := c.MustGet("my_user_model").(users.UserModel)
	if myUserModel.ID == 0 {
		c.AbortWithError(http.StatusUnauthorized, errors.New("{error : \"Require auth!\"}"))
		return
	}
	itemUserModel := GetItemUserModel(myUserModel)
	itemModels, modelCount, err := itemUserModel.GetItemFeed(limit, offset)
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewError("items", errors.New("Invalid param")))
		return
	}
	serializer := ItemsSerializer{c, itemModels}
	c.JSON(http.StatusOK, gin.H{"items": serializer.Response(), "itemsCount": modelCount})
}

func ItemRetrieve(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "feed" {
		ItemFeed(c)
		return
	}
	itemModel, err := FindOneItem(&ItemModel{Slug: slug})
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewError("items", errors.New("Invalid slug")))
		return
	}
	serializer := ItemSerializer{c, itemModel}
	c.JSON(http.StatusOK, gin.H{"item": serializer.Response()})
}

func ItemUpdate(c *gin.Context) {
	slug := c.Param("slug")
	itemModel, err := FindOneItem(&ItemModel{Slug: slug})
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewError("items", errors.New("Invalid slug")))
		return
	}
	itemModelValidator := NewItemModelValidatorFillWith(itemModel)
	if err := itemModelValidator.Bind(c); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		return
	}

	itemModelValidator.itemModel.ID = itemModel.ID
	if err := itemModel.Update(itemModelValidator.itemModel); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("database", err))
		return
	}
	serializer := ItemSerializer{c, itemModel}
	c.JSON(http.StatusOK, gin.H{"item": serializer.Response()})
}

func ItemDelete(c *gin.Context) {
	slug := c.Param("slug")
	err := DeleteItemModel(&ItemModel{Slug: slug})
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewError("items", errors.New("Invalid slug")))
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": "Delete success"})
}

func ItemFavorite(c *gin.Context) {
	slug := c.Param("slug")
	itemModel, err := FindOneItem(&ItemModel{Slug: slug})
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewError("items", errors.New("Invalid slug")))
		return
	}
	myUserModel := c.MustGet("my_user_model").(users.UserModel)
	err = itemModel.favoriteBy(GetItemUserModel(myUserModel))
	serializer := ItemSerializer{c, itemModel}
	c.JSON(http.StatusOK, gin.H{"item": serializer.Response()})
}

func ItemUnfavorite(c *gin.Context) {
	slug := c.Param("slug")
	itemModel, err := FindOneItem(&ItemModel{Slug: slug})
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewError("items", errors.New("Invalid slug")))
		return
	}
	myUserModel := c.MustGet("my_user_model").(users.UserModel)
	err = itemModel.unFavoriteBy(GetItemUserModel(myUserModel))
	serializer := ItemSerializer{c, itemModel}
	c.JSON(http.StatusOK, gin.H{"item": serializer.Response()})
}

func ItemCommentCreate(c *gin.Context) {
	slug := c.Param("slug")
	itemModel, err := FindOneItem(&ItemModel{Slug: slug})
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewError("comment", errors.New("Invalid slug")))
		return
	}
	commentModelValidator := NewCommentModelValidator()
	if err := commentModelValidator.Bind(c); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		return
	}
	commentModelValidator.commentModel.Item = itemModel

	if err := SaveOne(&commentModelValidator.commentModel); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("database", err))
		return
	}
	serializer := CommentSerializer{c, commentModelValidator.commentModel}
	c.JSON(http.StatusCreated, gin.H{"comment": serializer.Response()})
}

func ItemCommentDelete(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	id := uint(id64)
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewError("comment", errors.New("Invalid id")))
		return
	}
	err = DeleteCommentModel([]uint{id})
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewError("comment", errors.New("Invalid id")))
		return
	}
	c.JSON(http.StatusOK, gin.H{"comment": "Delete success"})
}

func ItemCommentList(c *gin.Context) {
	slug := c.Param("slug")
	itemModel, err := FindOneItem(&ItemModel{Slug: slug})
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewError("comments", errors.New("Invalid slug")))
		return
	}
	err = itemModel.getComments()
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewError("comments", errors.New("Database error")))
		return
	}
	serializer := CommentsSerializer{c, itemModel.Comments}
	c.JSON(http.StatusOK, gin.H{"comments": serializer.Response()})
}
func TagList(c *gin.Context) {
	tagModels, err := getAllTags()
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewError("items", errors.New("Invalid param")))
		return
	}
	serializer := TagsSerializer{c, tagModels}
	c.JSON(http.StatusOK, gin.H{"tags": serializer.Response()})
}
