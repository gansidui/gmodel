package remote

import (
	"github.com/gansidui/gmodel"
)

const (
	APIGetModelInfo         = "/admin/get-model-info"
	APIAddArticle           = "/admin/add-article"
	APIDeleteArticle        = "/admin/delete-article"
	APIGetArticle           = "/admin/get-article"
	APIGetNextArticles      = "/admin/get-next-articles"
	APIGetPrevArticles      = "/admin/get-prev-articles"
	APIGetNextArticlesByTag = "/admin/get-next-articles-by-tag"
	APIGetPrevArticlesByTag = "/admin/get-prev-articles-by-tag"
	APIUpdateArticle        = "/admin/update-article"
	APIGetTagById           = "/admin/get-tag-by-id"
	APIGetTagByName         = "/admin/get-tag-by-name"
	APIRenameTag            = "/admin/rename-tag"
)

// CustomArticleId 优先，CustomArticleId为空时才使用 Article.Id，下同
type RemoteArticle struct {
	*gmodel.Article
	CustomArticleId string `json:"custom_article_id"` // 自定义文章ID，优先级高于 Article.Id
}

type RemoteTag struct {
	*gmodel.Tag
}

type BaseResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type GetModelInfoResp struct {
	BaseResp
	ArticleCount uint64 `json:"article_count"`
	TagCount     uint64 `json:"tag_count"`
	MaxArticleId uint64 `json:"max_article_id"`
}

type AddArticleReq struct {
	Tags            []string `json:"tags"`
	Data            string   `json:"data"`
	CustomArticleId string   `json:"custom_article_id"`
}

type AddArticleResp struct {
	BaseResp
	ArticleId       uint64 `json:"article_id"`
	CustomArticleId string `json:"custom_article_id"`
}

type DeleteArticleReq struct {
	ArticleId       uint64 `json:"article_id"`
	CustomArticleId string `json:"custom_article_id"`
}

type DeleteArticleResp = BaseResp

type GetArticleReq = DeleteArticleReq

type GetArticleResp struct {
	BaseResp
	*RemoteArticle
}

type GetNextArticlesReq struct {
	ArticleId       uint64 `json:"article_id"`
	CustomArticleId string `json:"custom_article_id"`
	N               int    `json:"n"`
}

type GetNextArticlesResp struct {
	BaseResp
	RemoteArticles []*RemoteArticle `json:"remote_articles"`
}

type GetPrevArticlesReq = GetNextArticlesReq
type GetPrevArticlesResp = GetNextArticlesResp

type GetNextArticlesByTagReq struct {
	GetNextArticlesReq
	Tag string `json:"tag"`
}

type GetNextArticlesByTagResp = GetNextArticlesResp

type GetPrevArticlesByTagReq = GetNextArticlesByTagReq
type GetPrevArticlesByTagResp = GetNextArticlesByTagResp

type UpdateArticleReq struct {
	ArticleId       uint64   `json:"article_id"`
	CustomArticleId string   `json:"custom_article_id"`
	NewTags         []string `json:"new_tags"`
	NewData         string   `json:"new_data"`
}

type UpdateArticleResp = BaseResp

type GetTagByIdReq struct {
	TagId uint64 `json:"tag_id"`
}

type GetTagByIdResp struct {
	BaseResp
	*RemoteTag
}

type GetTagByNameReq struct {
	TagName string `json:"tag_name"`
}

type GetTagByNameResp = GetTagByIdResp

type RenameTagReq struct {
	OldName string `json:"old_name"`
	NewName string `json:"new_name"`
}

type RenameTagResp = BaseResp
