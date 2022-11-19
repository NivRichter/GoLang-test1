package items

import (
	_ "fmt"
	"github.com/jinzhu/gorm"
	"github.com/gothinkster/golang-gin-realworld-example-app/common"
	"github.com/gothinkster/golang-gin-realworld-example-app/users"
	"strconv"
)

type ItemModel struct {
	gorm.Model
	Slug        string `gorm:"unique_index"`
	Title       string
	Description string `gorm:"size:2048"`
	Body        string `gorm:"size:2048"`
	Seller      ItemUserModel
	SellerID    uint
	Tags        []TagModel     `gorm:"many2many:item_tags;"`
	Comments    []CommentModel `gorm:"ForeignKey:ItemID"`
}

type ItemUserModel struct {
	gorm.Model
	UserModel      users.UserModel
	UserModelID    uint
	ItemModels  []ItemModel  `gorm:"ForeignKey:SellerID"`
	FavoriteModels []FavoriteModel `gorm:"ForeignKey:FavoriteByID"`
}

type FavoriteModel struct {
	gorm.Model
	Favorite     ItemModel
	FavoriteID   uint
	FavoriteBy   ItemUserModel
	FavoriteByID uint
}

type TagModel struct {
	gorm.Model
	Tag           string         `gorm:"unique_index"`
	ItemModels []ItemModel `gorm:"many2many:item_tags;"`
}

type CommentModel struct {
	gorm.Model
	Item   ItemModel
	ItemID uint
	Seller    ItemUserModel
	SellerID  uint
	Body      string `gorm:"size:2048"`
}

func GetItemUserModel(userModel users.UserModel) ItemUserModel {
	var itemUserModel ItemUserModel
	if userModel.ID == 0 {
		return itemUserModel
	}
	db := common.GetDB()
	db.Where(&ItemUserModel{
		UserModelID: userModel.ID,
	}).FirstOrCreate(&itemUserModel)
	itemUserModel.UserModel = userModel
	return itemUserModel
}

func (item ItemModel) favoritesCount() uint {
	db := common.GetDB()
	var count uint
	db.Model(&FavoriteModel{}).Where(FavoriteModel{
		FavoriteID: item.ID,
	}).Count(&count)
	return count
}

func (item ItemModel) isFavoriteBy(user ItemUserModel) bool {
	db := common.GetDB()
	var favorite FavoriteModel
	db.Where(FavoriteModel{
		FavoriteID:   item.ID,
		FavoriteByID: user.ID,
	}).First(&favorite)
	return favorite.ID != 0
}

func (item ItemModel) favoriteBy(user ItemUserModel) error {
	db := common.GetDB()
	var favorite FavoriteModel
	err := db.FirstOrCreate(&favorite, &FavoriteModel{
		FavoriteID:   item.ID,
		FavoriteByID: user.ID,
	}).Error
	return err
}

func (item ItemModel) unFavoriteBy(user ItemUserModel) error {
	db := common.GetDB()
	err := db.Where(FavoriteModel{
		FavoriteID:   item.ID,
		FavoriteByID: user.ID,
	}).Delete(FavoriteModel{}).Error
	return err
}

func SaveOne(data interface{}) error {
	db := common.GetDB()
	err := db.Save(data).Error
	return err
}

func FindOneItem(condition interface{}) (ItemModel, error) {
	db := common.GetDB()
	var model ItemModel
	tx := db.Begin()
	tx.Where(condition).First(&model)
	tx.Model(&model).Related(&model.Seller, "Seller")
	tx.Model(&model.Seller).Related(&model.Seller.UserModel)
	tx.Model(&model).Related(&model.Tags, "Tags")
	err := tx.Commit().Error
	return model, err
}

func (self *ItemModel) getComments() error {
	db := common.GetDB()
	tx := db.Begin()
	tx.Model(self).Related(&self.Comments, "Comments")
	for i, _ := range self.Comments {
		tx.Model(&self.Comments[i]).Related(&self.Comments[i].Seller, "Seller")
		tx.Model(&self.Comments[i].Seller).Related(&self.Comments[i].Seller.UserModel)
	}
	err := tx.Commit().Error
	return err
}

func getAllTags() ([]TagModel, error) {
	db := common.GetDB()
	var models []TagModel
	err := db.Find(&models).Error
	return models, err
}

func FindManyItem(tag, seller, limit, offset, favorited string) ([]ItemModel, int, error) {
	db := common.GetDB()
	var models []ItemModel
	var count int

	offset_int, err := strconv.Atoi(offset)
	if err != nil {
		offset_int = 0
	}

	limit_int, err := strconv.Atoi(limit)
	if err != nil {
		limit_int = 20
	}

	tx := db.Begin()
	if tag != "" {
		var tagModel TagModel
		tx.Where(TagModel{Tag: tag}).First(&tagModel)
		if tagModel.ID != 0 {
			tx.Model(&tagModel).Offset(offset_int).Limit(limit_int).Related(&models, "ItemModels")
			count = tx.Model(&tagModel).Association("ItemModels").Count()
		}
	} else if seller != "" {
		var userModel users.UserModel
		tx.Where(users.UserModel{Username: seller}).First(&userModel)
		itemUserModel := GetItemUserModel(userModel)

		if itemUserModel.ID != 0 {
			count = tx.Model(&itemUserModel).Association("ItemModels").Count()
			tx.Model(&itemUserModel).Offset(offset_int).Limit(limit_int).Related(&models, "ItemModels")
		}
	} else if favorited != "" {
		var userModel users.UserModel
		tx.Where(users.UserModel{Username: favorited}).First(&userModel)
		itemUserModel := GetItemUserModel(userModel)
		if itemUserModel.ID != 0 {
			var favoriteModels []FavoriteModel
			tx.Where(FavoriteModel{
				FavoriteByID: itemUserModel.ID,
			}).Offset(offset_int).Limit(limit_int).Find(&favoriteModels)

			count = tx.Model(&itemUserModel).Association("FavoriteModels").Count()
			for _, favorite := range favoriteModels {
				var model ItemModel
				tx.Model(&favorite).Related(&model, "Favorite")
				models = append(models, model)
			}
		}
	} else {
		db.Model(&models).Count(&count)
		db.Offset(offset_int).Limit(limit_int).Find(&models)
	}

	for i, _ := range models {
		tx.Model(&models[i]).Related(&models[i].Seller, "Seller")
		tx.Model(&models[i].Seller).Related(&models[i].Seller.UserModel)
		tx.Model(&models[i]).Related(&models[i].Tags, "Tags")
	}
	err = tx.Commit().Error
	return models, count, err
}

func (self *ItemUserModel) GetItemFeed(limit, offset string) ([]ItemModel, int, error) {
	db := common.GetDB()
	var models []ItemModel
	var count int

	offset_int, err := strconv.Atoi(offset)
	if err != nil {
		offset_int = 0
	}
	limit_int, err := strconv.Atoi(limit)
	if err != nil {
		limit_int = 20
	}

	tx := db.Begin()
	followings := self.UserModel.GetFollowings()
	var itemUserModels []uint
	for _, following := range followings {
		itemUserModel := GetItemUserModel(following)
		itemUserModels = append(itemUserModels, itemUserModel.ID)
	}

	tx.Where("seller_id in (?)", itemUserModels).Order("updated_at desc").Offset(offset_int).Limit(limit_int).Find(&models)

	for i, _ := range models {
		tx.Model(&models[i]).Related(&models[i].Seller, "Seller")
		tx.Model(&models[i].Seller).Related(&models[i].Seller.UserModel)
		tx.Model(&models[i]).Related(&models[i].Tags, "Tags")
	}
	err = tx.Commit().Error
	return models, count, err
}

func (model *ItemModel) setTags(tags []string) error {
	db := common.GetDB()
	var tagList []TagModel
	for _, tag := range tags {
		var tagModel TagModel
		err := db.FirstOrCreate(&tagModel, TagModel{Tag: tag}).Error
		if err != nil {
			return err
		}
		tagList = append(tagList, tagModel)
	}
	model.Tags = tagList
	return nil
}

func (model *ItemModel) Update(data interface{}) error {
	db := common.GetDB()
	err := db.Model(model).Update(data).Error
	return err
}

func DeleteItemModel(condition interface{}) error {
	db := common.GetDB()
	err := db.Where(condition).Delete(ItemModel{}).Error
	return err
}

func DeleteCommentModel(condition interface{}) error {
	db := common.GetDB()
	err := db.Where(condition).Delete(CommentModel{}).Error
	return err
}
