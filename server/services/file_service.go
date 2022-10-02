package services

import (
	"github.com/mlogclub/simple/sqls"

	"bbs-go/model"
	"bbs-go/repositories"
)

var FileService = newFileService()

func newFileService() *fileService {
	return &fileService{}
}

type fileService struct{}

func (s *fileService) Get(id int64) *model.FileRecord {
	return repositories.FileRepository.Get(sqls.DB(), id)
}

func (s *fileService) CreateRecord(t *model.FileRecord) error {
	return repositories.FileRepository.Create(sqls.DB(), t)
}
