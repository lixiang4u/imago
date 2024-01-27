package handlers

import (
	"encoding/json"
	"github.com/lixiang4u/imago/models"
	"time"
)

func prepareRequestLog(requestLog *models.RequestLog, isExist int8) error {
	requestLog.IsExist = isExist
	buf, err := json.Marshal(requestLog)
	if err != nil {
		return err
	}
	err = models.NsqProducer(models.TopicRequest, buf)
	if err != nil {
		return err
	}
	return nil
}

func prepareShrinkLog(convertedFile string, convertedSize int64, isExist int8, requestLog *models.RequestLog, localMeta *models.LocalMeta, appConfig *models.AppConfig) {
	var now = time.Now()

	_ = prepareRequestLog(requestLog, isExist)

	_ = prepareRequestStat(&models.RequestStat{
		UserId:       appConfig.UserId,
		ProxyId:      appConfig.ProxyId,
		MetaId:       localMeta.Id,
		OriginUrl:    localMeta.Raw,
		RequestCount: 1,
		ResponseByte: uint64(convertedSize),
		SavedByte:    uint64(localMeta.Size - convertedSize),
		CreatedAt:    now,
	})
	_ = prepareUserFiles(&models.UserFiles{
		UserId:      appConfig.UserId,
		ProxyId:     appConfig.ProxyId,
		MetaId:      localMeta.Id,
		OriginUrl:   localMeta.Raw,
		OriginFile:  localMeta.RemoteLocal,
		ConvertFile: convertedFile,
		OriginSize:  uint64(localMeta.Size),
		ConvertSize: uint64(convertedSize),
		CreatedAt:   now,
	})

}

func prepareRequestStat(requestStat *models.RequestStat) error {
	buf, err := json.Marshal(requestStat)
	if err != nil {
		return err
	}
	err = models.NsqProducer(models.TopicRequestStat, buf)
	if err != nil {
		return err
	}
	return nil
}

func prepareUserFiles(userFiles *models.UserFiles) error {
	buf, err := json.Marshal(userFiles)
	if err != nil {
		return err
	}
	err = models.NsqProducer(models.TopicUserFiles, buf)
	if err != nil {
		return err
	}
	return nil
}
