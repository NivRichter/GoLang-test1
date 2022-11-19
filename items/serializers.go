package items

import (
	"github.com/gosimple/slug"
	"github.com/gothinkster/golang-gin-realworld-example-app/users"
	"github.com/gin-gonic/gin"
)

type TagSerializer struct {
	C *gin.Context
	TagModel
}

type TagsSerializer struct {
	C    *gin.Context
	Tags []TagModel
}

func (s *TagSerializer) Response() string {
	return s.TagModel.Tag
}

func (s *TagsSerializer) Response() []string {
	response := []string{}
	for _, tag := range s.Tags {
		serializer := TagSerializer{s.C, tag}
		response = append(response, serializer.Response())
	}
	return response
}

type ItemUserSerializer struct {
	C *gin.Context
	ItemUserModel
}

func (s *ItemUserSerializer) Response() users.ProfileResponse {
	response := users.ProfileSerializer{s.C, s.ItemUserModel.UserModel}
	return response.Response()
}

type ItemSerializer struct {
	C *gin.Context
	ItemModel
}

type ItemResponse struct {
	ID             uint                  `json:"-"`
	Title          string                `json:"title"`
	Slug           string                `json:"slug"`
	Description    string                `json:"description"`
	Body           string                `json:"body"`
	CreatedAt      string                `json:"createdAt"`
	UpdatedAt      string                `json:"updatedAt"`
	Seller         users.ProfileResponse `json:"seller"`
	Tags           []string              `json:"tagList"`
	Favorite       bool                  `json:"favorited"`
	FavoritesCount uint                  `json:"favoritesCount"`
}

type ItemsSerializer struct {
	C        *gin.Context
	Items []ItemModel
}

func (s *ItemSerializer) Response() ItemResponse {
	myUserModel := s.C.MustGet("my_user_model").(users.UserModel)
	sellerSerializer := ItemUserSerializer{s.C, s.Seller}
	response := ItemResponse{
		ID:          s.ID,
		Slug:        slug.Make(s.Title),
		Title:       s.Title,
		Description: s.Description,
		Body:        s.Body,
		CreatedAt:   s.CreatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
		//UpdatedAt:      s.UpdatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:      s.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
		Seller:         sellerSerializer.Response(),
		Favorite:       s.isFavoriteBy(GetItemUserModel(myUserModel)),
		FavoritesCount: s.favoritesCount(),
	}
	response.Tags = make([]string, 0)
	for _, tag := range s.Tags {
		serializer := TagSerializer{s.C, tag}
		response.Tags = append(response.Tags, serializer.Response())
	}
	return response
}

func (s *ItemsSerializer) Response() []ItemResponse {
	response := []ItemResponse{}
	for _, item := range s.Items {
		serializer := ItemSerializer{s.C, item}
		response = append(response, serializer.Response())
	}
	return response
}

type CommentSerializer struct {
	C *gin.Context
	CommentModel
}

type CommentsSerializer struct {
	C        *gin.Context
	Comments []CommentModel
}

type CommentResponse struct {
	ID        uint                  `json:"id"`
	Body      string                `json:"body"`
	CreatedAt string                `json:"createdAt"`
	UpdatedAt string                `json:"updatedAt"`
	Seller    users.ProfileResponse `json:"seller"`
}

func (s *CommentSerializer) Response() CommentResponse {
	sellerSerializer := ItemUserSerializer{s.C, s.Seller}
	response := CommentResponse{
		ID:        s.ID,
		Body:      s.Body,
		CreatedAt: s.CreatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
		UpdatedAt: s.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
		Seller:    sellerSerializer.Response(),
	}
	return response
}

func (s *CommentsSerializer) Response() []CommentResponse {
	response := []CommentResponse{}
	for _, comment := range s.Comments {
		serializer := CommentSerializer{s.C, comment}
		response = append(response, serializer.Response())
	}
	return response
}
