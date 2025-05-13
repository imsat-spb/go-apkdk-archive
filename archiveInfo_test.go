package archive

import (
	configuration "github.com/imsat-spb/go-apkdk-configuration"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewArchiveEmptyInfo(t *testing.T) {

	projectInfo := &configuration.TestProjectData{}
	info, err := NewConfigurationInfo(projectInfo)

	assert.Nil(t, err)

	assert.Len(t, info.mappings, 0)
}

func TestNewArchiveSingleMeasureInfo(t *testing.T) {

	const typeId = 1
	const typeName = "Type1"
	const objectName = "test1"
	const objectId = 1
	const measureId = 100
	const attributeId = 100
	const measureShortName = "P"
	const measureName = "Parameter"
	const attributeName = "Attribute"
	const stationId = 1000
	const stationName = "TestStation"

	const deviceId = 3
	const sensorId = 0
	const sensorAttrId = 1

	const hostId = 175

	testProject := configuration.TestProjectData{
		Devices: map[int]*configuration.Device{
			deviceId: {Id: deviceId, SensorCount: 100, BitsPerSensor: 32},
		},
		ObjectsToHosts: map[int]int{objectId: hostId},
		Objects: map[int]*configuration.ObjectInfo{
			objectId: {Id: objectId, TypeId: typeId, Name: objectName, StationId: stationId}},
		Parameters: map[int]*configuration.ObjectParameter{
			measureId: {Id: measureId, Name: measureName, ShortName: measureShortName, UnitOfMeasure: "Секунды,с"}},
		Attributes: map[int]*configuration.ObjectAttribute{
			attributeId: {Id: attributeId, Name: attributeName, UnitOfMeasure: "В"}},
		ParameterMappings: map[configuration.ParameterMappingKey]*configuration.ObjectParameterMapping{
			configuration.NewParameterMappingKey(objectId, measureId): {Id: measureId, ObjectId: objectId, DeviceId: deviceId, SensorId: sensorId}},
		AttributeMappings: map[configuration.ParameterMappingKey]*configuration.ObjectAttributeMapping{
			configuration.NewParameterMappingKey(objectId, attributeId): {Id: attributeId, ObjectId: objectId, DeviceId: deviceId, SensorId: sensorAttrId}},
		Hosts: []configuration.NestedHost{{Id: hostId, Devices: []configuration.Device{{Id: deviceId, SensorCount: 100, BitsPerSensor: 32}}}},
	}

	var projectInfo configuration.ProjectInformation = &testProject
	info, err := NewConfigurationInfo(projectInfo)

	assert.Nil(t, err)

	assert.Len(t, info.mappings, 1)
	assert.Len(t, info.mappings[deviceId], 2)
	m := info.mappings[deviceId][sensorId][0]
	a := info.mappings[deviceId][sensorAttrId][0]

	expectedMeasure := archiveMeasureOrAttributeInfo{
		objectTypeId:         typeId,
		measureOrAttributeId: measureId,
		objectId:             objectId,
		stationId:            stationId,
		unitOfMeasure:        "с",
		isAttribute:          false,
	}
	assert.Equal(t, *m, expectedMeasure)

	expectedAttribute := archiveMeasureOrAttributeInfo{
		objectTypeId:         typeId,
		measureOrAttributeId: attributeId,
		objectId:             objectId,
		stationId:            stationId,
		unitOfMeasure:        "В",
		isAttribute:          true,
	}
	assert.Equal(t, *a, expectedAttribute)

	assert.Len(t, info.mappings, 1)
	dm := info.mappings[deviceId]
	assert.Len(t, dm, 2)
	pm := dm[sensorId]
	assert.Len(t, pm, 1)
	assert.Equal(t, *pm[0], expectedMeasure)

	am := dm[sensorAttrId]
	assert.Len(t, am, 1)
	assert.Equal(t, *am[0], expectedAttribute)
}
