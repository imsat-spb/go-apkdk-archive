package archive

import (
	"github.com/imsat-spb/go-apkdk-configuration"
)

type archiveMeasureOrAttributeInfo struct {
	objectId             int
	measureOrAttributeId int
	unitOfMeasure        string
	stationId            int
	objectTypeId         int
	isAttribute          bool
}

type ParameterOrAttributeMappingKey struct {
	configuration.ParameterMappingKey
	isAttribute bool
}

func NewParameterOrAttributeMappingKey(objectId int, measureId int, isAttribute bool) ParameterOrAttributeMappingKey {
	return ParameterOrAttributeMappingKey{
		configuration.NewParameterMappingKey(objectId, measureId),
		isAttribute}
}

type ConfigurationInfo struct {
	// Устройство на датчик на список измерений или атрибутов
	mappings map[int]map[int][]*archiveMeasureOrAttributeInfo
	// Информация по объекту контроля
	Objects map[int]*ObjectInfo
}

type ObjectInfo struct {
	objectId  int
	stationId int
	hostId    int
}

type deviceMappingItem struct {
	deviceId int
	sensorId int
}

func NewConfigurationInfo(project configuration.ProjectInformation) (*ConfigurationInfo, error) {

	var archiveConfig = &ConfigurationInfo{
		mappings: make(map[int]map[int][]*archiveMeasureOrAttributeInfo),
		Objects:  make(map[int]*ObjectInfo),
	}

	for objectId, objectInfo := range project.GetObjects() {
		archiveConfig.Objects[objectId] = &ObjectInfo{
			objectId:  objectId,
			stationId: objectInfo.StationId,
			hostId:    project.GetObjectHost(objectId),
		}
	}
	measuresToDevices := make(map[ParameterOrAttributeMappingKey]deviceMappingItem)

	// Добавляем маппинг измерений
	measureMappingsMap := project.GetObjectParametersMappingsMap()
	attributesMappingsMap := project.GetObjectAttributeMappingsMap()

	resultMap := make(map[ParameterOrAttributeMappingKey]deviceMappingItem)

	for _, mv := range measureMappingsMap {
		resultMap[NewParameterOrAttributeMappingKey(mv.ObjectId, mv.Id, false)] = deviceMappingItem{
			mv.DeviceId,
			mv.SensorId}
	}

	for _, av := range attributesMappingsMap {
		resultMap[NewParameterOrAttributeMappingKey(av.ObjectId, av.Id, true)] = deviceMappingItem{
			av.DeviceId,
			av.SensorId}
	}

	// Вставка информации об устройствах
	for mapKey, mapping := range resultMap {
		objectId := mapKey.GetObjectId()
		measureId := mapKey.GetMeasureId()

		// Игнорируем не объявленные объекты или измерения
		oInfo := project.GetObjectInfo(objectId)
		if oInfo == nil {
			continue
		}

		var unitOfMeasure string

		if mapKey.isAttribute {
			attributeInfo := project.GetAttributeInfo(measureId)

			if attributeInfo == nil {
				continue
			}

			unitOfMeasure = attributeInfo.GetUnitOfMeasure()
		} else {
			paramInfo := project.GetObjectParameterInfo(measureId)

			if paramInfo == nil {
				continue
			}

			unitOfMeasure = paramInfo.GetUnitOfMeasureDisplayName()
		}

		deviceInfo := project.GetDeviceInfo(mapping.deviceId)
		if deviceInfo == nil {
			// Не нашли устройство
			continue
		}
		if mapping.sensorId >= deviceInfo.SensorCount {
			// Неправильный маппинг датчика
			continue
		}

		if _, ok := measuresToDevices[mapKey]; ok {
			// Измерение уже было добавлено, игнорируем повторное описание
			continue
		}

		mInfo := &archiveMeasureOrAttributeInfo{
			isAttribute:          mapKey.isAttribute,
			objectId:             objectId,
			stationId:            oInfo.StationId,
			objectTypeId:         oInfo.TypeId,
			measureOrAttributeId: measureId,
			unitOfMeasure:        unitOfMeasure}

		devInfo := archiveConfig.mappings[mapping.deviceId]
		if devInfo == nil {
			devInfo = make(map[int][]*archiveMeasureOrAttributeInfo)
			archiveConfig.mappings[mapping.deviceId] = devInfo
		}

		sensorInfo := devInfo[mapping.sensorId]
		if sensorInfo == nil {
			sensorInfo = []*archiveMeasureOrAttributeInfo{mInfo}
			devInfo[mapping.sensorId] = sensorInfo
		} else {
			devInfo[mapping.sensorId] = append(sensorInfo, mInfo)
		}

		measuresToDevices[mapKey] = deviceMappingItem{deviceId: mapping.deviceId, sensorId: mapping.sensorId}
	}

	return archiveConfig, nil
}
