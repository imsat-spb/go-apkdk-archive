package archive

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/imsat-spb/go-apkdk-core"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestEventsServerRequest(t *testing.T) {

	const objectId1 = 100
	const objectId2 = 200
	const stationId1 = 30000
	const stationId2 = 33000
	const hostId = 800

	const objectState1 byte = 10
	const objectState2 byte = 17

	deviceId := core.GetSpecialDeviceForHost(hostId)

	aTime := time.Now()
	testPackage := &core.DataPackage{
		Time:     core.GetUnixMicrosecondsFromTime(aTime),
		DeviceId: deviceId,
		Format:   core.PackageFormatEvents,
		Data: []byte{
			core.PackageEventTypeObjectState, objectId1, 0, 0, 0, objectState1, 0,
			core.PackageEventTypeObjectState, objectId2, 0, 0, 0, objectState2, 0},
		BitsPerSensor: 8,
		DataSize:      14,
		SensorCount:   14}

	eventsInPackage, err := testPackage.ParseEventsPackage()
	assert.Nil(t, err)

	updateResult := &objectChangeEventUpdateEventInfo{packageInfo: testPackage, processingTime: aTime,
		stations: []int{stationId1, stationId2}, events: eventsInPackage}

	res := updateResult.getArchiveServerRequest()

	assert.Len(t, res, 1)

	item := res[0]

	var eventInfo eventChangeItemInfo
	err = json.Unmarshal([]byte(item.item), &eventInfo)

	assert.Nil(t, err)

	assert.Equal(t, testPackage.DeviceId, eventInfo.DeviceId)
	assert.ElementsMatch(t, []int{stationId2, stationId1}, eventInfo.Stations)
	assert.Equal(t, testPackage.GetBase64String(), eventInfo.RawData)
	assert.Equal(t, core.PackageFormatEvents, eventInfo.Format)

	expectedEventTime := core.GetUnixMillisecondsFromTime(aTime)
	assert.Equal(t, expectedEventTime, eventInfo.Time)

	assert.ElementsMatch(t, []sdsEventInfo{
		{ObjectId: objectId2, StateId: uint16(objectState2)},
		{ObjectId: objectId1, StateId: uint16(objectState1)},
	}, eventInfo.Sds)

	var request map[string]*createRequest

	err = json.Unmarshal([]byte(item.request), &request)

	assert.Nil(t, err)

	create := request["create"]
	assert.Equal(t, "_doc", create.DocType)
	assert.Equal(t, "events", create.Index)
	assert.Equal(t, fmt.Sprintf("%d_1_%d", deviceId, eventInfo.Time), create.Id)
}

func getSliceFromInt32(value int32) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, value)
	return buf.Bytes()
}

func getTimeAndSlice(timeValue time.Time) (time.Time, []byte) {
	timeUnix := core.GetUnixMicrosecondsFromTime(timeValue) // В пакете время в микросекундах
	dataTime := make([]byte, 8)
	binary.LittleEndian.PutUint64(dataTime, uint64(timeUnix))

	timeInMcs := core.GetTimeFromUnixMicroseconds(binary.LittleEndian.Uint64(dataTime))
	return timeInMcs, dataTime
}

func TestAccidentEventsServerRequest(t *testing.T) {

	const objectId1 = 100
	const stationId1 = 30000
	const hostId = 800

	algId := getSliceFromInt32(-1)

	testData := append([]byte{core.PackageEventTypeAccidentInfo, 1}, algId...)

	testData = append(testData, []byte{objectId1, 0, 0, 0}...)

	aTime := time.Now()

	_, tBuf := getTimeAndSlice(aTime)
	testData = append(testData, tBuf...)

	_, tBuf1 := getTimeAndSlice(aTime)
	testData = append(testData, tBuf1...)

	deviceId := core.GetSpecialDeviceForHost(hostId)
	testPackage := &core.DataPackage{
		Time:          core.GetUnixMicrosecondsFromTime(aTime),
		DeviceId:      deviceId,
		Format:        core.PackageFormatEvents,
		Data:          testData,
		BitsPerSensor: 8,
		DataSize:      26,
		SensorCount:   26}

	eventsInPackage, err := testPackage.ParseEventsPackage()
	assert.Nil(t, err)

	updateResult := &objectChangeEventUpdateEventInfo{packageInfo: testPackage, processingTime: aTime,
		stations: []int{stationId1}, events: eventsInPackage}

	res := updateResult.getArchiveServerRequest()

	assert.Len(t, res, 1)

	item := res[0]

	var eventInfo eventChangeItemInfo
	err = json.Unmarshal([]byte(item.item), &eventInfo)

	assert.Nil(t, err)

	assert.Equal(t, testPackage.DeviceId, eventInfo.DeviceId)
	assert.ElementsMatch(t, []int{stationId1}, eventInfo.Stations)
	assert.Equal(t, testPackage.GetBase64String(), eventInfo.RawData)
	assert.Equal(t, core.PackageFormatEvents, eventInfo.Format)

	expectedEventTime := core.GetUnixMillisecondsFromTime(aTime)
	assert.Equal(t, expectedEventTime, eventInfo.Time)

	assert.ElementsMatch(t, []accidentEventInfo{
		{ObjectId: objectId1,
			StartTime:      expectedEventTime,
			EndTime:        expectedEventTime,
			AlgorithmId:    -1,
			AccidentTypeId: 1},
	}, eventInfo.Accidents)

	var request map[string]*createRequest

	err = json.Unmarshal([]byte(item.request), &request)

	assert.Nil(t, err)

	create := request["create"]
	assert.Equal(t, "_doc", create.DocType)
	assert.Equal(t, "events", create.Index)
	assert.Equal(t, fmt.Sprintf("%d_1_%d", deviceId, eventInfo.Time), create.Id)
}
