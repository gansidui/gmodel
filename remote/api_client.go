package admin

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gansidui/gmodel"
)

type APIClient struct {
	// APIServer的地址，需要带协议，比如：http://127.0.0.1:9999
	remoteAddr string
}

func (this *APIClient) Start(remoteAddr string) {
	this.remoteAddr = remoteAddr
}

func (this *APIClient) getAPIAddr(api string) string {
	return this.remoteAddr + api
}

// 包装POST请求
func (this *APIClient) post(url string, body io.Reader) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	return respBytes, err
}

// 返回文章数量、分类数量、最大的文章ID
func (this *APIClient) GetModelInfo() (uint64, uint64, uint64, error) {
	respBytes, err := this.post(this.getAPIAddr(APIGetModelInfo), nil)
	if err != nil {
		return 0, 0, 0, err
	}

	resp := &GetModelInfoResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return 0, 0, 0, err
	}

	if resp.ErrCode != ErrCodeSuccess {
		return 0, 0, 0, errors.New(resp.ErrMsg)
	}

	return resp.ArticleCount, resp.TagCount, resp.MaxArticleId, nil
}

func (this *APIClient) GetArticleCount() uint64 {
	articleCount, _, _, _ := this.GetModelInfo()
	return articleCount
}

func (this *APIClient) GetTagCount() uint64 {
	_, tagCount, _, _ := this.GetModelInfo()
	return tagCount
}

func (this *APIClient) GetMaxArticleId() uint64 {
	_, _, maxArticleId, _ := this.GetModelInfo()
	return maxArticleId
}

func (this *APIClient) AddArticle(tags []string, data string) (uint64, error) {
	req := &AddArticleReq{
		Tags: tags,
		Data: data,
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := this.post(this.getAPIAddr(APIAddArticle), bytes.NewBuffer(reqBytes))
	if err != nil {
		return 0, err
	}

	resp := &AddArticleResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return 0, err
	}

	if resp.ErrCode != ErrCodeSuccess {
		return 0, errors.New(resp.ErrMsg)
	}

	return resp.ArticleId, nil
}

func (this *APIClient) DeleteArticle(articleId uint64) error {
	req := &DeleteArticleReq{
		ArticleId: articleId,
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := this.post(this.getAPIAddr(APIDeleteArticle), bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}

	resp := &DeleteArticleResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return err
	}

	if resp.ErrCode != ErrCodeSuccess {
		return errors.New(resp.ErrMsg)
	}

	return nil
}

func (this *APIClient) GetArticle(articleId uint64) (*gmodel.Article, error) {
	req := &GetArticleReq{
		ArticleId: articleId,
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := this.post(this.getAPIAddr(APIGetArticle), bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, err
	}

	resp := &GetArticleResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return nil, err
	}

	if resp.ErrCode != ErrCodeSuccess {
		return nil, errors.New(resp.ErrMsg)
	}

	return resp.Article, nil
}

func (this *APIClient) GetNextArticles(articleId uint64, n int) []*gmodel.Article {
	req := &GetNextArticlesReq{
		ArticleId: articleId,
		N:         n,
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := this.post(this.getAPIAddr(APIGetNextArticles), bytes.NewBuffer(reqBytes))
	if err != nil {
		return []*gmodel.Article{}
	}

	resp := &GetNextArticlesResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return []*gmodel.Article{}
	}

	if resp.ErrCode != ErrCodeSuccess {
		return []*gmodel.Article{}
	}

	return resp.Articles
}

func (this *APIClient) GetPrevArticles(articleId uint64, n int) []*gmodel.Article {
	req := &GetPrevArticlesReq{
		ArticleId: articleId,
		N:         n,
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := this.post(this.getAPIAddr(APIGetPrevArticles), bytes.NewBuffer(reqBytes))
	if err != nil {
		return []*gmodel.Article{}
	}

	resp := &GetPrevArticlesResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return []*gmodel.Article{}
	}

	if resp.ErrCode != ErrCodeSuccess {
		return []*gmodel.Article{}
	}

	return resp.Articles
}

func (this *APIClient) GetNextArticlesByTag(tagName string, articleId uint64, n int) []*gmodel.Article {
	req := &GetNextArticlesByTagReq{}
	req.ArticleId = articleId
	req.N = n
	req.Tag = tagName
	reqBytes, _ := json.Marshal(req)

	respBytes, err := this.post(this.getAPIAddr(APIGetNextArticlesByTag), bytes.NewBuffer(reqBytes))
	if err != nil {
		return []*gmodel.Article{}
	}

	resp := &GetNextArticlesByTagResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return []*gmodel.Article{}
	}

	if resp.ErrCode != ErrCodeSuccess {
		return []*gmodel.Article{}
	}

	return resp.Articles
}

func (this *APIClient) GetPrevArticlesByTag(tagName string, articleId uint64, n int) []*gmodel.Article {
	req := &GetPrevArticlesByTagReq{}
	req.ArticleId = articleId
	req.N = n
	req.Tag = tagName
	reqBytes, _ := json.Marshal(req)

	respBytes, err := this.post(this.getAPIAddr(APIGetPrevArticlesByTag), bytes.NewBuffer(reqBytes))
	if err != nil {
		return []*gmodel.Article{}
	}

	resp := &GetPrevArticlesByTagResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return []*gmodel.Article{}
	}

	if resp.ErrCode != ErrCodeSuccess {
		return []*gmodel.Article{}
	}

	return resp.Articles
}

func (this *APIClient) UpdateArticle(articleId uint64, newTags []string, data string) error {
	req := &UpdateArticleReq{
		ArticleId: articleId,
		NewTags:   newTags,
		NewData:   data,
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := this.post(this.getAPIAddr(APIUpdateArticle), bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}

	resp := &UpdateArticleResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return err
	}

	if resp.ErrCode != ErrCodeSuccess {
		return errors.New(resp.ErrMsg)
	}

	return nil
}

func (this *APIClient) GetTagById(id uint64) (*gmodel.Tag, error) {
	req := &GetTagByIdReq{
		TagId: id,
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := this.post(this.getAPIAddr(APIGetTagById), bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, err
	}

	resp := &GetTagByIdResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return nil, err
	}

	if resp.ErrCode != ErrCodeSuccess {
		return nil, errors.New(resp.ErrMsg)
	}

	return resp.Tag, nil
}

func (this *APIClient) GetTagByName(name string) (*gmodel.Tag, error) {
	req := &GetTagByNameReq{
		TagName: name,
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := this.post(this.getAPIAddr(APIGetTagByName), bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, err
	}

	resp := &GetTagByNameResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return nil, err
	}

	if resp.ErrCode != ErrCodeSuccess {
		return nil, errors.New(resp.ErrMsg)
	}

	return resp.Tag, nil

}

func (this *APIClient) RenameTag(oldName, newName string) error {
	req := &RenameTagReq{
		OldName: oldName,
		NewName: newName,
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := this.post(this.getAPIAddr(APIRenameTag), bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}

	resp := &RenameTagResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return err
	}

	if resp.ErrCode != ErrCodeSuccess {
		return errors.New(resp.ErrMsg)
	}

	return nil
}
