package archive

import (
	"encoding/json"
	"fmt"
	"github.com/imsat-spb/go-apkdk-core"
	"time"
)

type sdsEventInfo struct {
	ObjectId uint32 `json:"objectId"`
	StateId  uint16 `json:"stateId"`
}

type failureEventInfo struct {
	ObjectId    uint32 `json:"objectId"`
	Fault       uint32 `json:"faultId"`
	IsStarted   bool   `json:"isStarted"`
	FailureTime int64  `json:"failureTime"`
}

type accidentEventInfo struct {
	ObjectId       uint32 `json:"objectId"`
	AlgorithmId    int32  `json:"algorithmId"`
	AccidentTypeId byte   `json:"accidentType"`
	StartTime      int64  `json:"startTime"`
	EndTime        int64  `json:"endTime,omitempty"`
}

type nwaEventInfo struct {
	ObjectId    uint32 `json:"objectId"`
	AlgorithmId uint32 `json:"algorithmId"`
	StateId     int32  `json:"stateId"`
	IsStarted   bool   `json:"isStarted"`
	EventTime   int64  `json:"time"`
}

type fpEventInfo struct {
	ObjectId    uint32 `json:"objectId"`
	AlgorithmId uint32 `json:"algorithmId"`
	StepIndex   int32  `json:"stepIndex"`
	EventTime   int64  `json:"time"`
}

type nwaStateEventInfo struct {
	ObjectId  uint32 `json:"objectId"`
	StateId   int32  `json:"stateId"`
	EventTime int64  `json:"time"`
}

type eventChangeItemInfo struct {
	Time      int64               `json:"time"`
	RawData   string              `json:"rawData"`
	Stations  []int               `json:"stations"`
	DeviceId  int32               `json:"deviceId"`
	Format    byte                `json:"format"`
	Sds       []sdsEventInfo      `json:"sds,omitempty"`
	Failures  []failureEventInfo  `json:"failures,omitempty"`
	Accidents []accidentEventInfo `json:"accidents,omitempty"`
	Nwa       []nwaEventInfo      `json:"anr,omitempty"`
	Fp        []fpEventInfo       `json:"ap,omitempty"`
	NwaState  []nwaStateEventInfo `json:"sanr, omitempty"`
}

type objectChangeEventUpdateEventInfo struct {
	processingTime time.Time
	packageInfo    *core.DataPackage
	events         *core.PackageEvents
	stations       []int
}

func (update *objectChangeEventUpdateEventInfo) getArchiveServerRequest() []*RequestItem {
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

	item := &eventChangeItemInfo{
		Time:     processingTime,
		Stations: update.stations,
		DeviceId: update.packageInfo.DeviceId,
		Format:   update.packageInfo.Format,
		RawData:  update.packageInfo.GetBase64String()}

	sdsEvents := len(update.events.ObjectStates)

	if sdsEvents > 0 {
		item.Sds = make([]sdsEventInfo, sdsEvents)
		i := 0
		for k, v := range update.events.ObjectStates {
			item.Sds[i] = sdsEventInfo{k, v}
			i++
		}
	}

	failureEvents := len(update.events.ObjectFailuresChangeState)
	if failureEvents > 0 {
		item.Failures = make([]failureEventInfo, failureEvents)
		i := 0
		for k, v := range update.events.ObjectFailuresChangeState {

			item.Failures[i] = failureEventInfo{ObjectId: k.ObjectId,
				Fault:       k.FailureId,
				IsStarted:   v.IsStarted,
				FailureTime: core.GetUnixMillisecondsFromTime(v.EventTime)}
			i++
		}
	}

	accidentEvents := len(update.events.ObjectAccidentsChangeState)
	if accidentEvents > 0 {
		item.Accidents = make([]accidentEventInfo, accidentEvents)
		i := 0
		for k, v := range update.events.ObjectAccidentsChangeState {

			item.Accidents[i] = accidentEventInfo{ObjectId: k.ObjectId,
				AlgorithmId:    k.AccidentId,
				AccidentTypeId: v.AccidentType,
				StartTime:      core.GetUnixMillisecondsFromTime(v.StartTime),
				EndTime:        core.GetUnixMillisecondsFromTime(v.EndTime)}
			i++
		}
	}

	nwaEvents := len(update.events.ObjectNwaChangeState)
	if nwaEvents > 0 {
		item.Nwa = make([]nwaEventInfo, nwaEvents)
		i := 0
		for _, v := range update.events.ObjectNwaChangeState {

			item.Nwa[i] = nwaEventInfo{ObjectId: v.ObjectId,
				AlgorithmId: v.AlgorithmId,
				StateId:     v.StateId,
				IsStarted:   v.IsStarted,
				EventTime:   core.GetUnixMillisecondsFromTime(v.EventTime)}
			i++
		}
	}

	fpEvents := len(update.events.ObjectFpChangeState)
	if fpEvents > 0 {
		item.Fp = make([]fpEventInfo, fpEvents)
		i := 0
		for _, v := range update.events.ObjectFpChangeState {

			item.Fp[i] = fpEventInfo{ObjectId: v.ObjectId,
				AlgorithmId: v.AlgorithmId,
				StepIndex:   v.StepIndex,
				EventTime:   core.GetUnixMillisecondsFromTime(v.EventTime)}
			i++
		}
	}

	nwaStateEvents := len(update.events.ObjectNwaStateLeaveEnter)
	if nwaStateEvents > 0 {
		item.NwaState = make([]nwaStateEventInfo, nwaStateEvents)
		i := 0
		for _, v := range update.events.ObjectNwaStateLeaveEnter {

			item.NwaState[i] = nwaStateEventInfo{ObjectId: v.ObjectId,
				StateId:   v.NwaStateId,
				EventTime: core.GetUnixMillisecondsFromTime(v.EventTime)}
			i++
		}
	}

	itemBuf, err := json.Marshal(item)
	if err != nil {
		return nil
	}

	return []*RequestItem{{string(buf), string(itemBuf)}}
}
