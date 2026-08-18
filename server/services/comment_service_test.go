package services

import (
	"testing"

	"bbs-go/model"
	"bbs-go/model/constants"
)

func TestCommentAuthorCanEditButCannotDelete(t *testing.T) {
	user := &model.User{Model: model.Model{Id: 1}}
	comment := &model.Comment{
		Model:      model.Model{Id: 10},
		UserId:     user.Id,
		EntityType: constants.EntityArticle,
		EntityId:   20,
	}

	if !CommentService.CanManage(user, comment) {
		t.Fatal("comment author should retain edit permission")
	}
	if CommentService.CanDelete(user, comment) {
		t.Fatal("comment author should not have delete permission")
	}
}

func TestCommentModeratorCanDelete(t *testing.T) {
	user := &model.User{
		Model: model.Model{Id: 1},
		Roles: model.MasterUser_NAME,
	}
	comment := &model.Comment{
		Model:      model.Model{Id: 10},
		UserId:     user.Id,
		EntityType: constants.EntityArticle,
		EntityId:   20,
	}

	if !CommentService.CanDelete(user, comment) {
		t.Fatal("moderator should retain delete permission")
	}
}
