package items

import (
	"github.com/gosimple/slug"
	"github.com/NivRichter/GoLang-test1/common"
	"github.com/NivRichter/GoLang-test1/users"
	"github.com/gin-gonic/gin"
)

type ItemModelValidator struct {
	Item struct {
		Title       string   `form:"title" json:"title" binding:"exists,min=4"`
		Description string   `form:"description" json:"description" binding:"max=2048"`
		Body        string   `form:"body" json:"body" binding:"max=2048"`
		Tags        []string `form:"tagList" json:"tagList"`
	} `json:"item"`
	itemModel ItemModel `json:"-"`
}

func NewItemModelValidator() ItemModelValidator {
	return ItemModelValidator{}
}

func NewItemModelValidatorFillWith(itemModel ItemModel) ItemModelValidator {
	itemModelValidator := NewItemModelValidator()
	itemModelValidator.Item.Title = itemModel.Title
	itemModelValidator.Item.Description = itemModel.Description
	itemModelValidator.Item.Body = itemModel.Body
	for _, tagModel := range itemModel.Tags {
		itemModelValidator.Item.Tags = append(itemModelValidator.Item.Tags, tagModel.Tag)
	}
	return itemModelValidator
}

func (s *ItemModelValidator) Bind(c *gin.Context) error {
	myUserModel := c.MustGet("my_user_model").(users.UserModel)

	err := common.Bind(c, s)
	if err != nil {
		return err
	}
	s.itemModel.Slug = slug.Make(s.Item.Title)
	s.itemModel.Title = s.Item.Title
	s.itemModel.Description = s.Item.Description
	s.itemModel.Body = s.Item.Body
	s.itemModel.Seller = GetItemUserModel(myUserModel)
	s.itemModel.setTags(s.Item.Tags)
	return nil
}

type CommentModelValidator struct {
	Comment struct {
		Body string `form:"body" json:"body" binding:"max=2048"`
	} `json:"comment"`
	commentModel CommentModel `json:"-"`
}

func NewCommentModelValidator() CommentModelValidator {
	return CommentModelValidator{}
}

func (s *CommentModelValidator) Bind(c *gin.Context) error {
	myUserModel := c.MustGet("my_user_model").(users.UserModel)

	err := common.Bind(c, s)
	if err != nil {
		return err
	}
	s.commentModel.Body = s.Comment.Body
	s.commentModel.Seller = GetItemUserModel(myUserModel)
	return nil
}
