package archive

import (
	"encoding/json"
	"github.com/imsat-spb/go-apkdk-core"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestObjectFullStateGetArchiveServerRequest(t *testing.T) {

	const objectId1 = 100
	const objectId2 = 200
	const stationId1 = 30000
	const stationId2 = 33000
	const hostId = 800

	const objectState1 byte = 10
	const objectState2 byte = 17

	aTime := time.Now()
	testPackage := &core.DataPackage{
		Time:     core.GetUnixMicrosecondsFromTime(aTime),
		DeviceId: core.GetSpecialDeviceForHost(hostId),
		Format:   core.PackageFormatFullObjectStates,
		Data: []byte{
			core.PackageEventTypeObjectState, objectId1, 0, 0, 0, objectState1, 0,
			core.PackageEventTypeObjectState, objectId2, 0, 0, 0, objectState2, 0},
		BitsPerSensor: 8,
		DataSize:      14,
		SensorCount:   14}

	updateResult := &objectFullStateUpdateEventInfo{packageInfo: testPackage, processingTime: aTime,
		stations: []int{stationId1, stationId2}}

	res := updateResult.getArchiveServerRequest()

	assert.Len(t, res, 1)

	item := res[0]

	var eventInfo eventItemInfo
	err := json.Unmarshal([]byte(item.item), &eventInfo)

	assert.Nil(t, err)

	assert.Equal(t, testPackage.DeviceId, eventInfo.DeviceId)
	assert.ElementsMatch(t, []int{stationId2, stationId1}, eventInfo.Stations)
	assert.Equal(t, testPackage.GetBase64String(), eventInfo.RawData)
	assert.Equal(t, core.PackageFormatFullObjectStates, eventInfo.Format)
	// TODO: check eventInfo.time

	var request map[string]*createRequest

	err = json.Unmarshal([]byte(item.request), &request)

	assert.Nil(t, err)

	//assert.Equal(t, "_doc", create.DocType)
	//assert.Equal(t, "events", create.Index)

	// TODO: check create.Id

}

func TestObjectFullStateNoStationsGetArchiveServerRequest(t *testing.T) {

	const objectId1 = 100
	const objectId2 = 200
	const hostId = 800

	const objectState1 byte = 10
	const objectState2 byte = 17

	aTime := time.Now()
	testPackage := &core.DataPackage{
		Time:     core.GetUnixMicrosecondsFromTime(aTime),
		DeviceId: core.GetSpecialDeviceForHost(hostId),
		Format:   core.PackageFormatFullObjectStates,
		Data: []byte{
			core.PackageEventTypeObjectState, objectId1, 0, 0, 0, objectState1, 0,
			core.PackageEventTypeObjectState, objectId2, 0, 0, 0, objectState2, 0},
		BitsPerSensor: 8,
		DataSize:      14,
		SensorCount:   14}

	updateResult := &objectFullStateUpdateEventInfo{packageInfo: testPackage, processingTime: aTime,
		stations: []int{}}

	res := updateResult.getArchiveServerRequest()

	assert.Nil(t, res)
}
