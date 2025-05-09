package archive

import (
	"encoding/json"
	"fmt"
	"github.com/imsat-spb/go-apkdk-core"
	"time"
)

type eventItemInfo struct {
	Time     int64  `json:"time"`
	RawData  string `json:"rawData"`
	Stations []int  `json:"stations"`
	DeviceId int32  `json:"deviceId"`
	Format   byte   `json:"format"`
}

type objectFullStateUpdateEventInfo struct {
	processingTime time.Time
	packageInfo    *core.DataPackage
	stations       []int
}

func (update *objectFullStateUpdateEventInfo) getArchiveServerRequest() []*RequestItem {
	if len(update.stations) == 0 {
		return nil
	}

	// Ддя записи в архив должны получить число миллисекунд
	processingTime := core.GetUnixMillisecondsFromTime(update.processingTime)

	id := fmt.Sprintf("%d_%d_%d", update.packageInfo.DeviceId, update.packageInfo.Format, processingTime)

	rq := map[string]*createRequest{"create": {DocType: "_doc", Index: "events", Id: id}}

	buf, err := json.Marshal(rq)

	if err != nil {
		return nil
	}

	item := &eventItemInfo{
		Time:     processingTime,
		Stations: update.stations,
		DeviceId: update.packageInfo.DeviceId,
		Format:   update.packageInfo.Format,
		RawData:  update.packageInfo.GetBase64String()}

	itemBuf, err := json.Marshal(item)
	if err != nil {
		return nil
	}

	return []*RequestItem{{string(buf), string(itemBuf)}}
}
