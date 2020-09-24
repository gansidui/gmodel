package gmodel

import (
	"fmt"
	"os"
	"testing"
)

func TestGModel(t *testing.T) {
	articleDBPath := "test_article.db"
	tagDBPath := "test_tag.db"
	indexDBPath := "test_index.db"

	defer func() {
		os.RemoveAll(articleDBPath)
		os.RemoveAll(tagDBPath)
		os.RemoveAll(indexDBPath)
	}()

	gmodel := &GModel{}
	err := gmodel.Open(articleDBPath, tagDBPath, indexDBPath)
	if err != nil {
		t.Fatal()
	}
	defer gmodel.Close()

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
	articleId, err := gmodel.AddArticle(tags, "data_id_1")
	if err != nil {
		t.Fatal()
	}
	if gmodel.GetTagCount() != 3 {
		t.Fatal()
	}

	article, err := gmodel.GetArticle(articleId)
	if err != nil {
		t.Fatal()
	}
	if article.Id != 1 || articleId != 1 {
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

	articleId, err = gmodel.AddArticle([]string{"tag3", "tag4"}, "data_id_2")
	if articleId != 2 || err != nil {
		t.Fatal()
	}
	if gmodel.GetTagCount() != 4 {
		t.Fatal()
	}

	articles := gmodel.GetNextArticles(0, 2)
	if len(articles) != 2 {
		t.Fatal()
	}
	if articles[0].Id != 1 || articles[1].Id != 2 {
		t.Fatal()
	}

	gmodel.AddArticle([]string{"tag4", "tag2"}, "data_id_3")

	articles = gmodel.GetNextArticlesByTag("tag2", 0, 10)
	if len(articles) != 2 {
		t.Fatal()
	}

	articles = gmodel.GetNextArticlesByTag("tag2", 2, 10)
	if len(articles) != 1 {
		t.Fatal()
	}

	if gmodel.UpdateArticle(1, []string{"tag2", "tag99"}, "new_data_id_1") != nil {
		t.Fatal()
	}
	if gmodel.GetTagCount() != 5 {
		t.Fatal()
	}

	if gmodel.DeleteArticle(2) != nil {
		t.Fatal()
	}
	if gmodel.DeleteArticle(4) == nil {
		t.Fatal()
	}
	if gmodel.GetArticleCount() != 2 {
		t.Fatal()
	}

	gmodel.AddArticle([]string{"tag222", "tag111"}, "hello")
	if gmodel.RenameTag("tag2", "tag1024") != nil {
		t.Fatal()
	}

	articles = gmodel.GetNextArticles(0, 10)
	for _, article := range articles {
		fmt.Println(article.Id, convertTagIds(gmodel, article.TagIds), article.Data)
	}

}

func TestNextGModel(t *testing.T) {
	articleDBPath := "test_article.db"
	tagDBPath := "test_tag.db"
	indexDBPath := "test_index.db"

	defer func() {
		os.RemoveAll(articleDBPath)
		os.RemoveAll(tagDBPath)
		os.RemoveAll(indexDBPath)
	}()

	gmodel := &GModel{}
	err := gmodel.Open(articleDBPath, tagDBPath, indexDBPath)
	if err != nil {
		t.Fatal()
	}
	defer gmodel.Close()

	gmodel.AddArticle([]string{"tag1", "tag2"}, "data_id_1")
	gmodel.AddArticle([]string{"tag3", "tag4"}, "data_id_2")
	gmodel.AddArticle([]string{"tag1", "tag2"}, "data_id_3")
	gmodel.AddArticle([]string{"tag3", "tag4"}, "data_id_4")
	gmodel.AddArticle([]string{"tag1", "tag2"}, "data_id_5")
	gmodel.AddArticle([]string{"tag3", "tag4"}, "data_id_6")
	gmodel.AddArticle([]string{"tag3", "tag4"}, "data_id_7")
	gmodel.AddArticle([]string{"tag3", "tag4"}, "data_id_8")

	fmt.Println("----------")

	// articles := gmodel.GetNextArticles(0, 100)

	articles := gmodel.GetPrevArticlesByTag("tag1", gmodel.GetMaxArticleId()+1, 2)
	// articles := gmodel.GetNextArticlesByTag("tag1", 1, 2)

	for _, article := range articles {
		fmt.Println(article.Id, convertTagIds(gmodel, article.TagIds), article.Data)
	}
	fmt.Println("----------")

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

func convertTagIds(gmodel *GModel, ids []uint64) []string {
	names := make([]string, 0)
	for _, id := range ids {
		if tag, err := gmodel.GetTagById(id); err == nil {
			names = append(names, tag.Name)
		}
	}
	return names
}
