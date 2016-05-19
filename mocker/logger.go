package mocker

import (
	"github.com/Sirupsen/logrus"
	"github.com/tsaikd/KDGoLib/logutil"
	"github.com/tsaikd/gin"
)

var logger = logutil.DefaultLogger

func init() {
	if gin.Mode() == gin.DebugMode {
		logger.Level = logrus.DebugLevel
	}
}
