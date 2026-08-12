package api

import (
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

const OLD_BBS_PREFIX = "OLD"

// parse string id into int id
// and return whether it is for OLD BBS
// id==-1 if parse failed
func parseIdStr(idStr string) (id int64, isOld bool) {
	isOld = false
	if strings.HasPrefix(idStr, OLD_BBS_PREFIX) {
		isOld = true
		idStr = strings.TrimPrefix(idStr, OLD_BBS_PREFIX)
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logrus.Errorf("bad id(%t): %s", isOld, err.Error())
		id = -1
	}
	return
}
