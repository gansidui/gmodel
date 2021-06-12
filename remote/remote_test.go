package remote

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestRemote(t *testing.T) {
	articleDBPath := "./article_test.db"
	tagDBPath := "./tag_test.db"
	indexDBPath := "./index_test.db"
	idDBPath := "./id_test.db"

	defer func() {
		os.RemoveAll(articleDBPath)
		os.RemoveAll(tagDBPath)
		os.RemoveAll(indexDBPath)
		os.RemoveAll(idDBPath)
	}()

	// 启动server
	go func() {
		config := &APIServerConfig{
			ArticleDBPath: articleDBPath,
			TagDBPath:     tagDBPath,
			IndexDBPath:   indexDBPath,
			IdDBPath:      idDBPath,
			ListeningAddr: ":9999",
			UseGzip:       false,
		}
		server := &APIServer{}
		server.Start(config)
	}()

	// 先等server启动，再启动client
	time.Sleep(3 * time.Second)

	// 启动client
	StartClient(t)

	// 等等
	time.Sleep(5 * time.Second)
}

func StartClient(t *testing.T) {
	// 下面的代码是从 gmodel_test.go 中拷贝过来的

	gmodel := &APIClient{}
	gmodel.Start("http://127.0.0.1:9999")

	if gmodel.GetArticleCount() != 0 {
		t.Fatal()
	}
	if gmodel.GetTagCount() != 0 {
		t.Fatal()
	}
	if gmodel.GetMaxArticleId() != 0 {
		t.Fatal()
	}

	tags := []string{"tag1", "tag2", "tag3"}
	articleId, customArticleId, err := gmodel.AddArticle(tags, "data_id_1", "custom_article_id")
	if err != nil {
		t.Fatal()
	}
	if articleId != 1 || customArticleId != "custom_article_id" {
		t.Fatal()
	}
	if gmodel.GetTagCount() != 3 {
		t.Fatal()
	}
	if gmodel.GetArticleCountByTag("tag2") != 1 {
		t.Fatal()
	}
	if gmodel.GetArticleCountByTag("tag3") != 1 {
		t.Fatal()
	}
	if gmodel.GetArticleCountByTag("null") != 0 {
		t.Fatal()
	}

	article, err := gmodel.GetArticle(articleId, "")
	if err != nil {
		t.Fatal()
	}
	if article.Id != 1 || articleId != 1 {
		t.Fatal()
	}

	// test customArticleId
	article, err = gmodel.GetArticle(0, "custom_article_id")
	if err != nil {
		t.Fatal()
	}
	if article.Id != 1 || article.CustomArticleId != "custom_article_id" {
		t.Fatal()
	}

	curTags := make([]string, 0)
	for _, tagId := range article.TagIds {
		if tag, err := gmodel.GetTagById(tagId); err == nil {
			curTags = append(curTags, tag.Name)
		}
	}
	if !isEqual(curTags, tags) {
		t.Fatal()
	}

	articleId, _, err = gmodel.AddArticle([]string{"tag3", "tag4"}, "data_id_2", "")
	if articleId != 2 || err != nil {
		t.Fatal()
	}
	if gmodel.GetTagCount() != 4 {
		t.Fatal()
	}

	articles := gmodel.GetNextArticles(0, "", 2)
	if len(articles) != 2 {
		t.Fatal()
	}
	if articles[0].Id != 1 || articles[1].Id != 2 {
		t.Fatal()
	}

	gmodel.AddArticle([]string{"tag4", "tag2"}, "data_id_3", "")

	articles = gmodel.GetNextArticlesByTag("tag2", 0, "", 10)
	if len(articles) != 2 {
		t.Fatal()
	}

	articles = gmodel.GetNextArticlesByTag("tag2", 2, "", 10)
	if len(articles) != 1 {
		t.Fatal()
	}

	if gmodel.UpdateArticle(1, "", []string{"tag2", "tag99"}, "new_data_id_1") != nil {
		t.Fatal()
	}
	if gmodel.GetTagCount() != 4 {
		t.Fatal()
	}

	if gmodel.DeleteArticle(2, "") != nil {
		t.Fatal()
	}
	if gmodel.DeleteArticle(4, "") == nil {
		t.Fatal()
	}
	if gmodel.GetArticleCount() != 2 {
		t.Fatal()
	}

	gmodel.AddArticle([]string{"tag222", "tag111"}, "hello", "")
	if gmodel.RenameTag("tag2", "tag1024") != nil {
		t.Fatal()
	}

	articles = gmodel.GetNextArticles(0, "", 10)
	for _, article := range articles {
		fmt.Println(article.Id, convertTagIds(gmodel, article.TagIds), article.Data)
	}

}

func isEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := 0; i < len(left); i++ {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func convertTagIds(gmodel *APIClient, ids []uint64) []string {
	names := make([]string, 0)
	for _, id := range ids {
		if tag, err := gmodel.GetTagById(id); err == nil {
			names = append(names, tag.Name)
		}
	}
	return names
}
