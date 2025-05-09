package archive

import (
	"fmt"
	"github.com/deckarep/golang-set"
	"github.com/imsat-spb/go-apkdk-core"
	"sync"
	"time"
)

type runtimeSensorMappingInfo struct {
	measures       []*archiveMeasureOrAttributeInfo
	currentValue   float32
	lastUpdateTime time.Time
}

func (rt *runtimeSensorMappingInfo) tryUpdateValue(newValue float32, now time.Time) bool {
	if rt.IsValueAssigned() {
		if newValue == rt.currentValue || core.IsNaN(newValue) && core.IsNaN(rt.currentValue) {
			if rt.lastUpdateTime.Add(time.Minute).After(now) {
				return false
			}
		}
	}

	rt.currentValue = newValue
	rt.lastUpdateTime = now

	return true
}

func (rt *runtimeSensorMappingInfo) IsValueAssigned() bool {
	return !rt.lastUpdateTime.IsZero()
}

type runtimeConfiguration struct {
	lock              sync.Mutex
	mappings          map[int32]map[uint16]*runtimeSensorMappingInfo
	objectsToStations map[int]int
	hostToObjects     map[int]mapset.Set
}

func (runtimeConfig *runtimeConfiguration) GetUpdateRequestItemsFromPackage(dataPackage *core.DataPackage) ([]*RequestItem, error) {

	updateResult, err := runtimeConfig.updateFromPackage(dataPackage)

	if err != nil {
		return nil, err
	}

	if updateResult == nil {
		return nil, nil
	}
	return updateResult.getArchiveServerRequest(), nil
}

func NewRuntimeConfiguration(info *ConfigurationInfo) *runtimeConfiguration {
	result := &runtimeConfiguration{
		mappings:          make(map[int32]map[uint16]*runtimeSensorMappingInfo),
		objectsToStations: make(map[int]int),
		hostToObjects:     make(map[int]mapset.Set)}

	for deviceId, deviceMapping := range info.mappings {
		runTimeDeviceMap := make(map[uint16]*runtimeSensorMappingInfo)
		result.mappings[int32(deviceId)] = runTimeDeviceMap

		for sensorId, sensorMapping := range deviceMapping {

			runTimeDeviceMap[uint16(sensorId)] = &runtimeSensorMappingInfo{measures: sensorMapping}
		}
	}

	for _, obj := range info.Objects {
		result.objectsToStations[obj.objectId] = obj.stationId
		if obj.hostId != 0 {
			if aSet, ok := result.hostToObjects[obj.hostId]; ok {
				aSet.Add(obj.objectId)
			} else {
				aSet := mapset.NewSet()
				aSet.Add(obj.objectId)
				result.hostToObjects[obj.hostId] = aSet
			}
		}
	}

	return result
}

func (runtimeConfig *runtimeConfiguration) updateFromFullStatePackage(packageInfo *core.DataPackage) (updateEventInfo, error) {
	if !(packageInfo.Format == core.PackageFormatFullObjectStates || packageInfo.Format == core.PackageFormatFullFailureStates ||
		packageInfo.Format == core.PackageFormatFullAccidentStates) {
		return nil, nil
	}

	stations := runtimeConfig.getStationsForSpecialDevice(packageInfo.DeviceId)
	if len(stations) == 0 {
		return nil, fmt.Errorf("no stations found for special device {%d}", packageInfo.DeviceId)
	}
	now := packageInfo.GetPackageTime()
	result := &objectFullStateUpdateEventInfo{
		processingTime: now,
		packageInfo:    packageInfo,
		stations:       stations}

	return result, nil
}

func (runtimeConfig *runtimeConfiguration) updateFromEventsPackage(packageInfo *core.DataPackage) (updateEventInfo, error) {
	if !(packageInfo.Format == core.PackageFormatEvents ||
		packageInfo.Format == core.PackageFormatChangeObjectStates ||
		packageInfo.Format == core.PackageFormatChangeFailureStates) {
		return nil, nil
	}

	events, err := packageInfo.ParseEventsPackage()

	if err != nil {
		return nil, err
	}

	stations := runtimeConfig.getStationsForEvents(events)

	if len(stations) == 0 {
		return nil, fmt.Errorf("no stations found for special device {%d}", packageInfo.DeviceId)
	}
	now := packageInfo.GetPackageTime()
	result := &objectChangeEventUpdateEventInfo{
		processingTime: now,
		packageInfo:    packageInfo,
		events:         events,
		stations:       stations}

	return result, nil
}

func (runtimeConfig *runtimeConfiguration) getStationsForSpecialDevice(specialDeviceId int32) []int {

	hostId, err := core.GetHostForSpecialDevice(specialDeviceId)
	if err != nil {
		return []int{}
	}

	if objects, ok := runtimeConfig.hostToObjects[hostId]; ok {
		return runtimeConfig.getStationsForObjects(objects)
	}

	return []int{}
}

func getIntSlice(aSet mapset.Set) []int {
	result := make([]int, aSet.Cardinality())
	for i, s := range aSet.ToSlice() {
		result[i] = s.(int)
	}

	return result
}

func (runtimeConfig *runtimeConfiguration) getStationsForObjects(objects mapset.Set) []int {
	stations := mapset.NewSet()

	for o := range objects.Iter() {
		if stationId, ok := runtimeConfig.objectsToStations[o.(int)]; ok {
			stations.Add(stationId)
		}
	}

	return getIntSlice(stations)
}

func (runtimeConfig *runtimeConfiguration) getStationsForEvents(events *core.PackageEvents) []int {

	objects := events.GetObjects()

	return runtimeConfig.getStationsForObjects(objects)
}

func (runtimeConfig *runtimeConfiguration) updateFromRawDataPackage(packageInfo *core.DataPackage) (updateEventInfo, error) {
	if packageInfo.Format != core.PackageFormatData {
		return nil, nil
	}

	if packageInfo.IsCompressed() {
		return nil, fmt.Errorf("decompress package first")
	}

	// TODO: bits per sensor get for device

	deviceMapping := runtimeConfig.mappings[packageInfo.DeviceId]
	if deviceMapping == nil {
		return nil, nil
	}

	handler, err := core.GetDataConverterFunction(packageInfo.BitsPerSensor)
	if err != nil {
		return nil, err
	}

	now := packageInfo.GetPackageTime()

	stations := mapset.NewSet()

	result := &measuresUpdateEventInfo{
		processingTime: now,
		packageInfo:    packageInfo,
		changedValues:  make(map[uint16]*updatedMeasures)}

	for sensorId, item := range deviceMapping {
		newValue := handler(packageInfo.Data, sensorId)

		// Изменение в данных или прошло больше минуты со времени последнего изменения
		if !item.tryUpdateValue(newValue, now) {
			continue
		}

		// Добавляем станции для измерений
		for _, am := range item.measures {
			stations.Add(am.stationId)
		}
		result.changedValues[sensorId] = &updatedMeasures{newValue, item.measures}
	}

	result.stations = getIntSlice(stations)

	return result, nil
}

func (runtimeConfig *runtimeConfiguration) updateFromPackage(packageInfo *core.DataPackage) (updateEventInfo, error) {
	runtimeConfig.lock.Lock()
	defer runtimeConfig.lock.Unlock()

	switch packageInfo.Format {
	case core.PackageFormatData:
		return runtimeConfig.updateFromRawDataPackage(packageInfo)
	case core.PackageFormatEvents,
		core.PackageFormatChangeObjectStates,
		core.PackageFormatChangeFailureStates:
		return runtimeConfig.updateFromEventsPackage(packageInfo)
	case core.PackageFormatFullObjectStates,
		core.PackageFormatFullFailureStates,
		core.PackageFormatFullAccidentStates:
		return runtimeConfig.updateFromFullStatePackage(packageInfo)
	default:
		return nil, nil
	}
}
