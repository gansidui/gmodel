package remote

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gansidui/gmodel"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

const (
	ErrCodeSuccess = 0
	ErrCodeFailed  = -1
	ErrMsgSuccess  = "success"
)

type APIServerConfig struct {
	// 数据库路径
	ArticleDBPath string
	TagDBPath     string
	IndexDBPath   string

	// model中的文章ID都是uint64类型，需要增加字符串ID的扩展功能
	// 保存字符串ID和整型ID的映射，也可以用于自定义ID（字符串）
	// 如果不使用自定义ID，则随机产生一个字符串ID
	// 这样做的原因避免ID为整数，然后导致网站被遍历抓取
	// 用自定义ID还有一个好处，比如将文章的标题作为文章的自定义ID，自带去重效果
	IdDBPath string

	// 监听地址
	ListeningAddr string

	// 是否开启Gzip
	// 建议：如果服务器之间走内网（内网带宽非常大），则无需开启
	UseGzip bool
}

type APIServer struct {
	model   *gmodel.GModel
	idMgr   *gmodel.IdMgr
	useGzip bool
}

func (this *APIServer) Start(config *APIServerConfig) {
	// 打开数据库
	this.model = &gmodel.GModel{}
	if err := this.model.Open(config.ArticleDBPath, config.TagDBPath, config.IndexDBPath); err != nil {
		log.Fatal(err)
	}

	this.idMgr = &gmodel.IdMgr{}
	if err := this.idMgr.Open(config.IdDBPath); err != nil {
		log.Fatal(err)
	}
	this.useGzip = config.UseGzip

	// 执行退出逻辑，用于保存数据
	defer func() {
		this.model.Close()
		this.idMgr.Close()
	}()

	// Gin包设置Release模式
	gin.SetMode(gin.ReleaseMode)

	// 创建服务器并监听端口
	server := &http.Server{
		Addr:         config.ListeningAddr,
		Handler:      this.newHandler(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	go func() {
		log.Println("start listening:", config.ListeningAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	log.Println("signal: ", <-quit)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Println("server shutdowm:", err)
		return
	}

	// catching ctx.Done(). timeout of 5 seconds.
	// select {
	// case <-ctx.Done():
	// 	log.Println("timeout of 5 seconds")
	// }

	log.Println("server exiting")
}

func (this *APIServer) newHandler() *gin.Engine {
	router := gin.Default()

	// gzip
	if this.useGzip {
		router.Use(gzip.Gzip(gzip.DefaultCompression))
	}

	// api
	router.POST(APIGetModelInfo, this.getModelInfoHandler)
	router.POST(APIAddArticle, this.addArticleHandler)
	router.POST(APIDeleteArticle, this.deleteArticleHandler)
	router.POST(APIGetArticle, this.getArticleHandler)
	router.POST(APIGetNextArticles, this.getNextArticlesHandler)
	router.POST(APIGetPrevArticles, this.getPrevArticlesHandler)
	router.POST(APIGetNextArticlesByTag, this.getNextArticlesByTagHandler)
	router.POST(APIGetPrevArticlesByTag, this.getPrevArticlesByTagHandler)
	router.POST(APIUpdateArticle, this.updateArticleHandler)
	router.POST(APIGetTagById, this.getTagByIdHandler)
	router.POST(APIGetTagByName, this.getTagByNameHandler)
	router.POST(APIRenameTag, this.renameTagHandler)
	router.POST(APIGetArticleCountByTag, this.getArticleCountByTagHandler)

	return router
}

func (this *APIServer) getModelInfoHandler(c *gin.Context) {
	resp := &GetModelInfoResp{}
	resp.ErrCode = ErrCodeSuccess
	resp.ErrMsg = ErrMsgSuccess
	resp.ArticleCount = this.model.GetArticleCount()
	resp.TagCount = this.model.GetTagCount()
	resp.MaxArticleId = this.model.GetMaxArticleId()

	c.JSON(http.StatusOK, resp)
}

func (this *APIServer) addArticleHandler(c *gin.Context) {
	resp := &AddArticleResp{}
	resp.ErrCode = ErrCodeSuccess
	resp.ErrMsg = ErrMsgSuccess

	var req AddArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = err.Error()
		c.JSON(http.StatusOK, resp)
		return
	}

	// 判断自定义ID是否已经存在，如果已经存在，则返回
	if req.CustomArticleId != "" {
		if _, exist := this.idMgr.GetIntId(req.CustomArticleId); exist {
			resp.ErrCode = ErrCodeFailed
			resp.ErrMsg = "CustomArticleId is exist"
			c.JSON(http.StatusOK, resp)
			return
		}
	}

	// 保存新文章
	articleId, err := this.model.AddArticle(req.Tags, req.Data)
	if err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = "AddArticle failed: " + err.Error()
		c.JSON(http.StatusOK, resp)
		return
	}

	// 自定义ID
	if req.CustomArticleId != "" {
		resp.CustomArticleId = req.CustomArticleId
		if err := this.idMgr.SetIdMap(articleId, req.CustomArticleId); err != nil {
			resp.ErrCode = ErrCodeFailed
			resp.ErrMsg = "SetIdMap failed: " + err.Error()
		}
	} else {
		customId, err := this.idMgr.AddIntId(articleId)
		if err != nil {
			resp.ErrCode = ErrCodeFailed
			resp.ErrMsg = "AddIntId failed: " + err.Error()
		} else {
			resp.CustomArticleId = customId
		}
	}

	resp.ArticleId = articleId
	c.JSON(http.StatusOK, resp)
}

func (this *APIServer) deleteArticleHandler(c *gin.Context) {
	resp := &DeleteArticleResp{}
	resp.ErrCode = ErrCodeSuccess
	resp.ErrMsg = ErrMsgSuccess

	var req DeleteArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = err.Error()
		c.JSON(http.StatusOK, resp)
		return
	}

	articleId := req.ArticleId
	if req.CustomArticleId != "" {
		if intId, ok := this.idMgr.GetIntId(req.CustomArticleId); ok {
			articleId = intId
		}
	}

	if err := this.model.DeleteArticle(articleId); err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = "DeleteArticle failed: " + err.Error()
	}

	c.JSON(http.StatusOK, resp)
}

func (this *APIServer) getArticleHandler(c *gin.Context) {
	resp := &GetArticleResp{}
	resp.ErrCode = ErrCodeSuccess
	resp.ErrMsg = ErrMsgSuccess

	var req GetArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = err.Error()
		c.JSON(http.StatusOK, resp)
		return
	}

	articleId := req.ArticleId
	customArticleId := req.CustomArticleId
	if customArticleId != "" {
		if intId, ok := this.idMgr.GetIntId(customArticleId); ok {
			articleId = intId
		}
	} else {
		if stringId, ok := this.idMgr.GetStringId(articleId); ok {
			customArticleId = stringId
		}
	}

	article, err := this.model.GetArticle(articleId)
	if err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = "GetArticle failed: " + err.Error()
	}

	// 获取分类名称
	tagNameArray := make([]string, 0)
	for _, tagId := range article.TagIds {
		if tag, err := this.model.GetTagById(tagId); err == nil {
			tagNameArray = append(tagNameArray, tag.Name)
		}
	}

	resp.RemoteArticle = &RemoteArticle{
		Article:         article,
		CustomArticleId: customArticleId,
		TagNameArray:    tagNameArray,
	}

	c.JSON(http.StatusOK, resp)
}

func (this *APIServer) getNextArticlesHandler(c *gin.Context) {
	resp := &GetNextArticlesResp{}
	resp.ErrCode = ErrCodeSuccess
	resp.ErrMsg = ErrMsgSuccess

	var req GetNextArticlesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = err.Error()
		c.JSON(http.StatusOK, resp)
		return
	}

	articleId := req.ArticleId
	if req.CustomArticleId != "" {
		if intId, ok := this.idMgr.GetIntId(req.CustomArticleId); ok {
			articleId = intId
		}
	}

	remoteArticles := make([]*RemoteArticle, 0)

	articles := this.model.GetNextArticles(articleId, req.N)
	for _, article := range articles {
		// 获取分类名称
		tagNameArray := make([]string, 0)
		for _, tagId := range article.TagIds {
			if tag, err := this.model.GetTagById(tagId); err == nil {
				tagNameArray = append(tagNameArray, tag.Name)
			}
		}

		// 获取自定义文章ID
		if stringId, ok := this.idMgr.GetStringId(article.Id); ok {
			remoteArticles = append(remoteArticles, &RemoteArticle{
				Article:         article,
				CustomArticleId: stringId,
				TagNameArray:    tagNameArray,
			})
		}
	}

	resp.RemoteArticles = remoteArticles
	c.JSON(http.StatusOK, resp)
}

func (this *APIServer) getPrevArticlesHandler(c *gin.Context) {
	resp := &GetPrevArticlesResp{}
	resp.ErrCode = ErrCodeSuccess
	resp.ErrMsg = ErrMsgSuccess

	var req GetPrevArticlesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = err.Error()
		c.JSON(http.StatusOK, resp)
		return
	}

	articleId := req.ArticleId
	if req.CustomArticleId != "" {
		if intId, ok := this.idMgr.GetIntId(req.CustomArticleId); ok {
			articleId = intId
		}
	}

	remoteArticles := make([]*RemoteArticle, 0)

	articles := this.model.GetPrevArticles(articleId, req.N)
	for _, article := range articles {
		// 获取分类名称
		tagNameArray := make([]string, 0)
		for _, tagId := range article.TagIds {
			if tag, err := this.model.GetTagById(tagId); err == nil {
				tagNameArray = append(tagNameArray, tag.Name)
			}
		}

		// 获取自定义文章ID
		if stringId, ok := this.idMgr.GetStringId(article.Id); ok {
			remoteArticles = append(remoteArticles, &RemoteArticle{
				Article:         article,
				CustomArticleId: stringId,
				TagNameArray:    tagNameArray,
			})
		}
	}

	resp.RemoteArticles = remoteArticles
	c.JSON(http.StatusOK, resp)
}

func (this *APIServer) getNextArticlesByTagHandler(c *gin.Context) {
	resp := &GetNextArticlesByTagResp{}
	resp.ErrCode = ErrCodeSuccess
	resp.ErrMsg = ErrMsgSuccess

	var req GetNextArticlesByTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = err.Error()
		c.JSON(http.StatusOK, resp)
		return
	}

	articleId := req.ArticleId
	if req.CustomArticleId != "" {
		if intId, ok := this.idMgr.GetIntId(req.CustomArticleId); ok {
			articleId = intId
		}
	}

	remoteArticles := make([]*RemoteArticle, 0)

	articles := this.model.GetNextArticlesByTag(req.Tag, articleId, req.N)
	for _, article := range articles {
		// 获取分类名称
		tagNameArray := make([]string, 0)
		for _, tagId := range article.TagIds {
			if tag, err := this.model.GetTagById(tagId); err == nil {
				tagNameArray = append(tagNameArray, tag.Name)
			}
		}

		// 获取自定义文章ID
		if stringId, ok := this.idMgr.GetStringId(article.Id); ok {
			remoteArticles = append(remoteArticles, &RemoteArticle{
				Article:         article,
				CustomArticleId: stringId,
				TagNameArray:    tagNameArray,
			})
		}
	}

	resp.RemoteArticles = remoteArticles
	c.JSON(http.StatusOK, resp)
}

func (this *APIServer) getPrevArticlesByTagHandler(c *gin.Context) {
	resp := &GetPrevArticlesByTagResp{}
	resp.ErrCode = ErrCodeSuccess
	resp.ErrMsg = ErrMsgSuccess

	var req GetPrevArticlesByTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = err.Error()
		c.JSON(http.StatusOK, resp)
		return
	}

	articleId := req.ArticleId
	if req.CustomArticleId != "" {
		if intId, ok := this.idMgr.GetIntId(req.CustomArticleId); ok {
			articleId = intId
		}
	}

	remoteArticles := make([]*RemoteArticle, 0)

	articles := this.model.GetPrevArticlesByTag(req.Tag, articleId, req.N)
	for _, article := range articles {
		// 获取分类名称
		tagNameArray := make([]string, 0)
		for _, tagId := range article.TagIds {
			if tag, err := this.model.GetTagById(tagId); err == nil {
				tagNameArray = append(tagNameArray, tag.Name)
			}
		}

		// 获取自定义文章ID
		if stringId, ok := this.idMgr.GetStringId(article.Id); ok {
			remoteArticles = append(remoteArticles, &RemoteArticle{
				Article:         article,
				CustomArticleId: stringId,
				TagNameArray:    tagNameArray,
			})
		}
	}

	resp.RemoteArticles = remoteArticles
	c.JSON(http.StatusOK, resp)
}

func (this *APIServer) updateArticleHandler(c *gin.Context) {
	resp := &UpdateArticleResp{}
	resp.ErrCode = ErrCodeSuccess
	resp.ErrMsg = ErrMsgSuccess

	var req UpdateArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = err.Error()
		c.JSON(http.StatusOK, resp)
		return
	}

	articleId := req.ArticleId
	if req.CustomArticleId != "" {
		if intId, ok := this.idMgr.GetIntId(req.CustomArticleId); ok {
			articleId = intId
		}
	}

	err := this.model.UpdateArticle(articleId, req.NewTags, req.NewData)
	if err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = "UpdateArticle failed: " + err.Error()
	}

	c.JSON(http.StatusOK, resp)
}

func (this *APIServer) getTagByIdHandler(c *gin.Context) {
	resp := &GetTagByIdResp{}
	resp.ErrCode = ErrCodeSuccess
	resp.ErrMsg = ErrMsgSuccess

	var req GetTagByIdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = err.Error()
		c.JSON(http.StatusOK, resp)
		return
	}

	tag, err := this.model.GetTagById(req.TagId)
	if err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = "GetTagById failed: " + err.Error()
	}

	resp.RemoteTag = &RemoteTag{
		Tag: tag,
	}

	c.JSON(http.StatusOK, resp)
}

func (this *APIServer) getTagByNameHandler(c *gin.Context) {
	resp := &GetTagByNameResp{}
	resp.ErrCode = ErrCodeSuccess
	resp.ErrMsg = ErrMsgSuccess

	var req GetTagByNameReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = err.Error()
		c.JSON(http.StatusOK, resp)
		return
	}

	tag, err := this.model.GetTagByName(req.TagName)
	if err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = "GetTagByName failed: " + err.Error()
	}

	resp.Tag = tag
	c.JSON(http.StatusOK, resp)
}

func (this *APIServer) renameTagHandler(c *gin.Context) {
	resp := &RenameTagResp{}
	resp.ErrCode = ErrCodeSuccess
	resp.ErrMsg = ErrMsgSuccess

	var req RenameTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = err.Error()
		c.JSON(http.StatusOK, resp)
		return
	}

	err := this.model.RenameTag(req.OldName, req.NewName)
	if err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = "RenameTag failed: " + err.Error()
	}

	c.JSON(http.StatusOK, resp)
}

func (this *APIServer) getArticleCountByTagHandler(c *gin.Context) {
	resp := &GetArticleCountByTagResp{}
	resp.ErrCode = ErrCodeSuccess
	resp.ErrMsg = ErrMsgSuccess

	var req GetArticleCountByTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = err.Error()
		c.JSON(http.StatusOK, resp)
		return
	}

	resp.ArticleCount = this.model.GetArticleCountByTag(req.TagName)
	c.JSON(http.StatusOK, resp)
}
