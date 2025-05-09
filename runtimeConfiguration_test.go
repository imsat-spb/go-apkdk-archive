package archive

import (
	"github.com/deckarep/golang-set"
	"github.com/imsat-spb/go-apkdk-core"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestNewRuntimeConfiguration(t *testing.T) {
	const deviceId = 5
	const sensorId1 = 1
	const objectId1 = 100
	const objectId2 = 200
	const stationId = 30000
	const hostId = 800
	const measureId = 100

	testMeasureInfo := &archiveMeasureOrAttributeInfo{
		objectId:             objectId1,
		stationId:            stationId,
		isAttribute:          false,
		measureOrAttributeId: measureId,
	}

	configInfo := &ConfigurationInfo{
		Objects: map[int]*ObjectInfo{
			objectId1: {objectId: objectId1, stationId: stationId, hostId: hostId},
			objectId2: {objectId: objectId2, stationId: stationId, hostId: hostId},
		},
		mappings: map[int]map[int][]*archiveMeasureOrAttributeInfo{
			deviceId: {sensorId1: []*archiveMeasureOrAttributeInfo{testMeasureInfo}},
		},
	}

	info := NewRuntimeConfiguration(configInfo)

	assert.Equal(t, map[int]int{
		objectId1: stationId,
		objectId2: stationId,
	}, info.objectsToStations)

	assert.Equal(t, map[int]mapset.Set{
		hostId: mapset.NewSet(objectId2, objectId1),
	}, info.hostToObjects)

	assert.Len(t, info.mappings, 1)

	devMappings := info.mappings[deviceId]
	assert.Len(t, devMappings, 1)
	sensorMappings := devMappings[uint16(sensorId1)]
	assert.Len(t, sensorMappings.measures, 1)

	assert.False(t, sensorMappings.measures[0].isAttribute)
	assert.Equal(t, objectId1, sensorMappings.measures[0].objectId)
	assert.Equal(t, measureId, sensorMappings.measures[0].measureOrAttributeId)
	assert.False(t, sensorMappings.IsValueAssigned())
}

func TestUpdateObjectFullState(t *testing.T) {
	const objectId1 = 100
	const objectId2 = 200
	const stationId1 = 30000
	const stationId2 = 33000
	const hostId = 800

	configInfo := &ConfigurationInfo{
		Objects: map[int]*ObjectInfo{
			objectId1: {objectId: objectId1, stationId: stationId1, hostId: hostId},
			objectId2: {objectId: objectId2, stationId: stationId2, hostId: hostId},
		},
	}

	info := NewRuntimeConfiguration(configInfo)

	now := time.Now().Add(-time.Minute * 10)
	testPackage := &core.DataPackage{
		Time:     core.GetUnixMicrosecondsFromTime(now),
		DeviceId: core.GetSpecialDeviceForHost(hostId),
		Format:   core.PackageFormatFullObjectStates,
		Data: []byte{
			core.PackageEventTypeObjectState, objectId1, 0, 0, 0, 1, 0,
			core.PackageEventTypeObjectState, objectId2, 0, 0, 0, 1, 0},
		BitsPerSensor: 8,
		DataSize:      14,
		SensorCount:   14}

	updateItem, err := info.updateFromPackage(testPackage)

	assert.Nil(t, err)

	update := updateItem.(*objectFullStateUpdateEventInfo)

	assert.ElementsMatch(t, []int{stationId2, stationId1}, update.stations)
	assert.Equal(t, testPackage.GetPackageTime(), update.processingTime)
}

func TestUpdateFromEvents(t *testing.T) {
	const objectId1 = 100
	const objectId2 = 200
	const stationId1 = 30000
	const stationId2 = 33000
	const hostId = 800

	const objectState1 byte = 10
	const objectState2 byte = 17

	configInfo := &ConfigurationInfo{
		Objects: map[int]*ObjectInfo{
			objectId1: {objectId: objectId1, stationId: stationId1, hostId: hostId},
			objectId2: {objectId: objectId2, stationId: stationId2, hostId: hostId},
		},
	}

	info := NewRuntimeConfiguration(configInfo)

	now := time.Now().Add(-time.Minute * 10)
	testPackage := &core.DataPackage{
		Time:     core.GetUnixMicrosecondsFromTime(now),
		DeviceId: core.GetSpecialDeviceForHost(hostId),
		Format:   core.PackageFormatEvents,
		Data: []byte{
			core.PackageEventTypeObjectState, objectId1, 0, 0, 0, objectState1, 0,
			core.PackageEventTypeObjectState, objectId2, 0, 0, 0, objectState2, 0},
		BitsPerSensor: 8,
		DataSize:      14,
		SensorCount:   14}

	updateItem, err := info.updateFromPackage(testPackage)

	assert.Nil(t, err)

	update := updateItem.(*objectChangeEventUpdateEventInfo)

	assert.ElementsMatch(t, []int{stationId2, stationId1}, update.stations)
	assert.Equal(t, testPackage.GetPackageTime(), update.processingTime)
	assert.Len(t, update.events.ObjectStates, 2)
	assert.Equal(t, uint16(objectState1), update.events.ObjectStates[objectId1])
	assert.Equal(t, uint16(objectState2), update.events.ObjectStates[objectId2])
}

func TestUpdateMeasures(t *testing.T) {
	const deviceId = 5
	const sensorId1 = 1
	const objectId1 = 100
	const objectId2 = 200
	const stationId = 30000
	const hostId = 800

	testMeasureInfo := &archiveMeasureOrAttributeInfo{
		objectId:             objectId1,
		isAttribute:          false,
		measureOrAttributeId: 100,
	}

	configInfo := &ConfigurationInfo{
		Objects: map[int]*ObjectInfo{
			objectId1: {objectId: objectId1, stationId: stationId, hostId: hostId},
			objectId2: {objectId: objectId2, stationId: stationId, hostId: hostId},
		},
		mappings: map[int]map[int][]*archiveMeasureOrAttributeInfo{
			deviceId: {sensorId1: []*archiveMeasureOrAttributeInfo{testMeasureInfo}},
		},
	}

	info := NewRuntimeConfiguration(configInfo)

	now := time.Now().Add(-time.Minute * 10)
	testPackage := &core.DataPackage{
		Time:          core.GetUnixMicrosecondsFromTime(now),
		DeviceId:      deviceId,
		Format:        core.PackageFormatData,
		Data:          []byte{0, 0, 0, 0, 255, 0, 0, 0, 200, 0, 0, 0},
		BitsPerSensor: 32,
		DataSize:      12,
		SensorCount:   3}

	updateItem, err := info.updateFromPackage(testPackage)

	update := updateItem.(*measuresUpdateEventInfo)

	assert.Nil(t, err)
	assert.Len(t, update.changedValues, 1)
	cv := update.changedValues[sensorId1]
	assert.Equal(t, float32(0.255), cv.value)
	assert.Len(t, cv.measures, 1)
	mi := cv.measures[0]
	assert.True(t, reflect.DeepEqual(testMeasureInfo, mi))

	now = now.Add(time.Second * 50)
	// нет изменений и не прошло 1 минуты
	testPackage.Time = core.GetUnixMicrosecondsFromTime(now)
	updateItem, err = info.updateFromPackage(testPackage)
	update = updateItem.(*measuresUpdateEventInfo)

	assert.Nil(t, err)
	assert.Len(t, update.changedValues, 0)
	assert.Equal(t, testPackage.GetPackageTime(), update.processingTime)

	// прошла минута. должны появиться изменения
	now = now.Add(time.Second * 50)
	testPackage.Time = core.GetUnixMicrosecondsFromTime(now)
	updateItem, err = info.updateFromPackage(testPackage)
	update = updateItem.(*measuresUpdateEventInfo)

	assert.Nil(t, err)
	assert.Len(t, update.changedValues, 1)
	assert.Equal(t, testPackage.GetPackageTime(), update.processingTime)

	cv = update.changedValues[sensorId1]
	assert.Equal(t, float32(0.255), cv.value)
	assert.Len(t, cv.measures, 1)
	mi = cv.measures[0]
	assert.True(t, reflect.DeepEqual(testMeasureInfo, mi))

	testPackage1 := &core.DataPackage{
		DeviceId:      deviceId,
		Format:        core.PackageFormatData,
		Data:          []byte{0, 0, 0, 0, 122, 0, 0, 0, 200, 0, 0, 0},
		BitsPerSensor: 32,
		DataSize:      12,
		SensorCount:   3}

	now = now.Add(time.Second * 5)
	testPackage1.Time = core.GetUnixMicrosecondsFromTime(now)
	updateItem, err = info.updateFromPackage(testPackage1)
	update = updateItem.(*measuresUpdateEventInfo)

	assert.Nil(t, err)
	assert.Len(t, update.changedValues, 1)
	assert.Equal(t, testPackage1.GetPackageTime(), update.processingTime)

	cv = update.changedValues[sensorId1]
	assert.Equal(t, float32(0.122), cv.value)
	assert.Len(t, cv.measures, 1)
	mi = cv.measures[0]
	assert.True(t, reflect.DeepEqual(testMeasureInfo, mi))
}

func TestGetStationsFromEvents(t *testing.T) {

	config := RuntimeConfiguration{objectsToStations: map[int]int{
		1000: 30001,
		2000: 30002,
		3000: 30001,
	}}
	tests := []struct {
		name     string
		data     core.PackageEvents
		expected []int
	}{
		{"test1", core.PackageEvents{ObjectStates: map[uint32]uint16{
			1000: 1,
			2000: 3,
		}}, []int{30001, 30002}},

		{"test2", core.PackageEvents{ObjectStates: map[uint32]uint16{
			1000: 1,
			3000: 3,
			4000: 17,
		}}, []int{30001}},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			result := config.getStationsForEvents(&test.data)

			assert.ElementsMatch(t, result, test.expected)

		})
	}
}

func TestGetStationsForSpecialDevice(t *testing.T) {

	objSetOnHost1 := mapset.NewSet(1000, 2000)
	objSetOnHost2 := mapset.NewSet(3000, 4000)
	config := RuntimeConfiguration{objectsToStations: map[int]int{
		1000: 30001,
		2000: 30002,
		3000: 30003,
		4000: 30004,
	}, hostToObjects: map[int]mapset.Set{
		100: objSetOnHost1,
		200: objSetOnHost2,
	}}
	tests := []struct {
		name     string
		hostId   int
		expected []int
	}{
		{"test1", 100, []int{30001, 30002}},
		{"test2", 200, []int{30003, 30004}},
		{"noHost", 300, []int{}},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			devId := core.GetSpecialDeviceForHost(test.hostId)
			result := config.getStationsForSpecialDevice(devId)

			assert.ElementsMatch(t, result, test.expected)

		})
	}
}
