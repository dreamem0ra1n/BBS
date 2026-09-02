package admin

import (
	"fmt"

	"bbs-go/controllers/api"
	"bbs-go/model"
	"bbs-go/model/constants"
	"bbs-go/pkg/bbsurls"
	"bbs-go/pkg/config"
	"bbs-go/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

// FileController exposes uploaded file metadata to administrators. The
// object itself remains in MinIO; this endpoint makes its application source
// (topic or unattached upload) visible alongside the preview URL.
type FileController struct {
	Ctx iris.Context
}

func (c *FileController) AnyList() *web.JsonResult {
	services.FileService.ClassifyReferencedFiles()
	query := params.NewQueryParams(c.Ctx)
	query.EqByReq("id").EqByReq("user_id").EqByReq("topic_id").EqByReq("comment_id").
		EqByReq("source_type").LikeByReq("file_name").PageByReq().Desc("id")
	query.Cnd.Where("source_type <> ?", "avatar")
	files, paging := services.FileService.FindPageByParams(query)
	results := make([]map[string]interface{}, 0, len(files))
	for index := range files {
		file := &files[index]
		item := web.NewRspBuilder(file).
			Put("previewUrl", fmt.Sprintf("%s/api/file/preview/%s", config.Instance.BaseUrl, file.FileUUID))
		if file.TopicId > 0 {
			if topic := services.TopicService.Get(file.TopicId); topic != nil {
				item.Put("sourceLabel", "帖子："+topic.Title).
					Put("sourceUrl", bbsurls.TopicUrl(topic.Id)).
					Put("topicTitle", topic.Title).Put("topicStatus", topic.Status)
			} else {
				item.Put("sourceLabel", fmt.Sprintf("帖子 #%d", file.TopicId))
			}
		} else if file.CommentId > 0 {
			item.Put("sourceLabel", fmt.Sprintf("评论 #%d", file.CommentId))
			if comment := services.CommentService.Get(file.CommentId); comment != nil {
				topicId := comment.EntityId
				rootType := comment.EntityType
				for comment.EntityType == constants.EntityComment {
					parent := services.CommentService.Get(topicId)
					if parent == nil {
						break
					}
					topicId = parent.EntityId
					comment = parent
					rootType = parent.EntityType
				}
				if rootType == constants.EntityTopic {
					if topic := services.TopicService.Get(topicId); topic != nil {
						item.Put("sourceLabel", fmt.Sprintf("帖子：%s · 评论 #%d", topic.Title, file.CommentId)).
						Put("sourceUrl", bbsurls.TopicUrl(topic.Id)+"#comment-"+fmt.Sprint(file.CommentId))
					}
				} else if rootType == constants.EntityArticle {
					if article := services.ArticleService.Get(topicId); article != nil {
						item.Put("sourceLabel", fmt.Sprintf("文章：%s · 评论 #%d", article.Title, file.CommentId)).
						Put("sourceUrl", bbsurls.ArticleUrl(article.Id)+"#comment-"+fmt.Sprint(file.CommentId))
					}
				}
			}
		}
		if file.SourceType != "unattached" && file.SourceType != "topic" && file.SourceType != "comment" {
			labels := map[string]string{
				"background": "用户背景图",
				"node_logo":  "节点图标",
				"link_logo":  "友情链接图标",
			}
			if label, ok := labels[file.SourceType]; ok {
				item.Put("sourceLabel", label)
			}
		}
		results = append(results, item.Build())
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *FileController) GetBy(id int64) *web.JsonResult {
	file := services.FileService.Get(id)
	if file == nil {
		return web.JsonErrorMsg("Not found, id=" + fmt.Sprint(id))
	}
	return web.JsonData(file)
}

func (c *FileController) PostDelete() *web.JsonResult {
	id := c.Ctx.PostValueInt64Default("id", 0)
	if err := services.FileService.Delete(id, apiDeleteObject); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func apiDeleteObject(file *model.FileRecord) error {
	return api.RemoveObject(file.BucketName, file.ObjectName, file.FileUUID)
}
