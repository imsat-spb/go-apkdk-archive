package archive

import (
	"encoding/json"
	"fmt"
	"github.com/imsat-spb/go-apkdk-core"
	"time"
)

type eventMeasuresUpdateInfo struct {
	Time       int64               `json:"time"`
	RawData    string              `json:"rawData"`
	Stations   []int               `json:"stations"`
	DeviceId   int32               `json:"deviceId"`
	Format     byte                `json:"format"`
	Measures   []measureItemInfo   `json:"measures,omitempty"`
	Attributes []attributeItemInfo `json:"attributes,omitempty"`
}

type measureOrAttributeItemInfo struct {
	Value        *float32 `json:"value,omitempty"`
	ObjectId     int      `json:"objectId"`
	Unit         string   `json:"unit"`
	ObjectTypeId int      `json:"objectTypeId"`
}

type measureItemInfo struct {
	measureOrAttributeItemInfo
	MeasureId int `json:"measureId"`
}

type attributeItemInfo struct {
	measureOrAttributeItemInfo
	AttributeId int `json:"attributeId"`
}

type updatedMeasures struct {
	value    float32
	measures []*archiveMeasureOrAttributeInfo
}

type measuresUpdateEventInfo struct {
	processingTime time.Time
	changedValues  map[uint16]*updatedMeasures
	packageInfo    *core.DataPackage
	stations       []int
}

func (update *measuresUpdateEventInfo) getArchiveServerRequest() []*RequestItem {
	if len(update.changedValues) == 0 || len(update.stations) == 0 {
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

	item := &eventMeasuresUpdateInfo{
		Time:     processingTime,
		Stations: update.stations,
		DeviceId: update.packageInfo.DeviceId,
		Format:   update.packageInfo.Format,
		RawData:  update.packageInfo.GetBase64String()}

	for _, itemWithValue := range update.changedValues {
		if len(itemWithValue.measures) == 0 {
			continue
		}

		for _, measureInfo := range itemWithValue.measures {

			commonInfo := measureOrAttributeItemInfo{
				ObjectId:     measureInfo.objectId,
				Unit:         measureInfo.unitOfMeasure,
				ObjectTypeId: measureInfo.objectTypeId}

			if !core.IsNaN(itemWithValue.value) {
				commonInfo.Value = &itemWithValue.value
			}

			if measureInfo.isAttribute {
				item.Attributes = append(item.Attributes, attributeItemInfo{
					commonInfo,
					measureInfo.measureOrAttributeId})

			} else {
				item.Measures = append(item.Measures, measureItemInfo{
					commonInfo,
					measureInfo.measureOrAttributeId})
			}
		}
	}

	itemBuf, err := json.Marshal(item)
	if err != nil {
		return nil
	}

	return []*RequestItem{{string(buf), string(itemBuf)}}
}
