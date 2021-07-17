package remote

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
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

// customArticleId 可以为空，如果为空，则服务器自动生成
// 返回：文章ID、自定义文章ID
func (this *APIClient) AddArticle(tags []string, data string, customArticleId string) (uint64, string, error) {
	req := &AddArticleReq{
		Tags:            tags,
		Data:            data,
		CustomArticleId: customArticleId,
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := this.post(this.getAPIAddr(APIAddArticle), bytes.NewBuffer(reqBytes))
	if err != nil {
		return 0, "", err
	}

	resp := &AddArticleResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return 0, "", err
	}

	if resp.ErrCode != ErrCodeSuccess {
		return 0, "", errors.New(resp.ErrMsg)
	}

	return resp.ArticleId, resp.CustomArticleId, nil
}

// 文章ID随便填一个，customArticleId 优先
func (this *APIClient) DeleteArticle(articleId uint64, customArticleId string) error {
	req := &DeleteArticleReq{
		ArticleId:       articleId,
		CustomArticleId: customArticleId,
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

func (this *APIClient) GetArticle(articleId uint64, customArticleId string) (*RemoteArticle, error) {
	req := &GetArticleReq{
		ArticleId:       articleId,
		CustomArticleId: customArticleId,
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

	return resp.RemoteArticle, nil
}

func (this *APIClient) GetNextArticles(articleId uint64, customArticleId string, n int) []*RemoteArticle {
	req := &GetNextArticlesReq{
		ArticleId:       articleId,
		CustomArticleId: customArticleId,
		N:               n,
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := this.post(this.getAPIAddr(APIGetNextArticles), bytes.NewBuffer(reqBytes))
	if err != nil {
		return []*RemoteArticle{}
	}

	resp := &GetNextArticlesResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return []*RemoteArticle{}
	}

	if resp.ErrCode != ErrCodeSuccess {
		return []*RemoteArticle{}
	}

	return resp.RemoteArticles
}

func (this *APIClient) GetPrevArticles(articleId uint64, customArticleId string, n int) []*RemoteArticle {
	req := &GetPrevArticlesReq{
		ArticleId:       articleId,
		CustomArticleId: customArticleId,
		N:               n,
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := this.post(this.getAPIAddr(APIGetPrevArticles), bytes.NewBuffer(reqBytes))
	if err != nil {
		return []*RemoteArticle{}
	}

	resp := &GetPrevArticlesResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return []*RemoteArticle{}
	}

	if resp.ErrCode != ErrCodeSuccess {
		return []*RemoteArticle{}
	}

	return resp.RemoteArticles
}

func (this *APIClient) GetNextArticlesByTag(tagName string, articleId uint64, customArticleId string, n int) []*RemoteArticle {
	req := &GetNextArticlesByTagReq{}
	req.ArticleId = articleId
	req.CustomArticleId = customArticleId
	req.N = n
	req.Tag = tagName
	reqBytes, _ := json.Marshal(req)

	respBytes, err := this.post(this.getAPIAddr(APIGetNextArticlesByTag), bytes.NewBuffer(reqBytes))
	if err != nil {
		return []*RemoteArticle{}
	}

	resp := &GetNextArticlesByTagResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return []*RemoteArticle{}
	}

	if resp.ErrCode != ErrCodeSuccess {
		return []*RemoteArticle{}
	}

	return resp.RemoteArticles
}

func (this *APIClient) GetPrevArticlesByTag(tagName string, articleId uint64, customArticleId string, n int) []*RemoteArticle {
	req := &GetPrevArticlesByTagReq{}
	req.ArticleId = articleId
	req.CustomArticleId = customArticleId
	req.N = n
	req.Tag = tagName
	reqBytes, _ := json.Marshal(req)

	respBytes, err := this.post(this.getAPIAddr(APIGetPrevArticlesByTag), bytes.NewBuffer(reqBytes))
	if err != nil {
		return []*RemoteArticle{}
	}

	resp := &GetPrevArticlesByTagResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return []*RemoteArticle{}
	}

	if resp.ErrCode != ErrCodeSuccess {
		return []*RemoteArticle{}
	}

	return resp.RemoteArticles
}

func (this *APIClient) UpdateArticle(articleId uint64, customArticleId string, newTags []string, newData string) error {
	req := &UpdateArticleReq{
		ArticleId:       articleId,
		CustomArticleId: customArticleId,
		NewTags:         newTags,
		NewData:         newData,
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

func (this *APIClient) GetTagById(id uint64) (*RemoteTag, error) {
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

	return resp.RemoteTag, nil
}

func (this *APIClient) GetTagByName(name string) (*RemoteTag, error) {
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

	return resp.RemoteTag, nil
}

func (this *APIClient) GetNextTags(tagName string, n int) []*RemoteTag {
	req := &GetNextTagsReq{
		TagName: tagName,
		N:       n,
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := this.post(this.getAPIAddr(APIGetNextTags), bytes.NewBuffer(reqBytes))
	if err != nil {
		return []*RemoteTag{}
	}

	resp := &GetNextTagsResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return []*RemoteTag{}
	}

	if resp.ErrCode != ErrCodeSuccess {
		return []*RemoteTag{}
	}

	return resp.RemoteTags
}

func (this *APIClient) GetPrevTags(tagName string, n int) []*RemoteTag {
	req := &GetPrevTagsReq{
		TagName: tagName,
		N:       n,
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := this.post(this.getAPIAddr(APIGetPrevTags), bytes.NewBuffer(reqBytes))
	if err != nil {
		return []*RemoteTag{}
	}

	resp := &GetPrevTagsResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return []*RemoteTag{}
	}

	if resp.ErrCode != ErrCodeSuccess {
		return []*RemoteTag{}
	}

	return resp.RemoteTags
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

func (this *APIClient) GetArticleCountByTag(tagName string) uint64 {
	req := &GetArticleCountByTagReq{
		TagName: tagName,
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := this.post(this.getAPIAddr(APIGetArticleCountByTag), bytes.NewBuffer(reqBytes))
	if err != nil {
		return 0
	}

	resp := &GetArticleCountByTagResp{}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return 0
	}

	if resp.ErrCode != ErrCodeSuccess {
		return 0
	}

	return resp.ArticleCount
}
