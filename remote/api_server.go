package admin

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

	// 监听地址
	ListeningAddr string

	// 是否开启Gzip
	// 建议：如果服务器之间走内网（内网带宽非常大），则无需开启
	UseGzip bool
}

type APIServer struct {
	model   *gmodel.GModel
	useGzip bool
}

func (this *APIServer) Start(config *APIServerConfig) {
	// 打开数据库
	this.model = &gmodel.GModel{}
	if err := this.model.Open(config.ArticleDBPath, config.TagDBPath, config.IndexDBPath); err != nil {
		log.Fatal(err)
	}
	this.useGzip = config.UseGzip

	// 执行退出逻辑，用于保存数据
	defer func() {
		this.model.Close()
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

	return router
}

func (this *APIServer) getModelInfoHandler(c *gin.Context) {
	resp := &GetModelInfoResp{
		ArticleCount: this.model.GetArticleCount(),
		TagCount:     this.model.GetTagCount(),
		MaxArticleId: this.model.GetMaxArticleId(),
	}
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

	articleId, err := this.model.AddArticle(req.Tags, req.Data)
	if err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = err.Error()
		c.JSON(http.StatusOK, resp)
		return
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

	err := this.model.DeleteArticle(req.ArticleId)
	if err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = err.Error()
		c.JSON(http.StatusOK, resp)
		return
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

	article, err := this.model.GetArticle(req.ArticleId)
	if err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = err.Error()
		c.JSON(http.StatusOK, resp)
		return
	}

	resp.Article = article
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

	resp.Articles = this.model.GetNextArticles(req.ArticleId, req.N)
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

	resp.Articles = this.model.GetPrevArticles(req.ArticleId, req.N)
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

	resp.Articles = this.model.GetNextArticlesByTag(req.Tag, req.ArticleId, req.N)
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

	resp.Articles = this.model.GetPrevArticlesByTag(req.Tag, req.ArticleId, req.N)
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

	err := this.model.UpdateArticle(req.ArticleId, req.NewTags, req.NewData)
	if err != nil {
		resp.ErrCode = ErrCodeFailed
		resp.ErrMsg = err.Error()
		c.JSON(http.StatusOK, resp)
		return
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
		resp.ErrMsg = err.Error()
		c.JSON(http.StatusOK, resp)
		return
	}

	resp.Tag = tag
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
		resp.ErrMsg = err.Error()
		c.JSON(http.StatusOK, resp)
		return
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
		resp.ErrMsg = err.Error()
		c.JSON(http.StatusOK, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
}
