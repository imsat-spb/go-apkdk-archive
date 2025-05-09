package archive

import (
	"github.com/imsat-spb/go-apkdk-core"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetArchiveServerRequest(t *testing.T) {

	testMeasureInfo1 := &archiveMeasureOrAttributeInfo{
		isAttribute:          false,
		measureOrAttributeId: 100,
		objectId:             200,
		objectTypeId:         1,
		stationId:            500,
		unitOfMeasure:        "",
	}

	testMeasureInfo2 := &archiveMeasureOrAttributeInfo{
		isAttribute:          false,
		measureOrAttributeId: 101,
		objectId:             200,
		objectTypeId:         1,
		stationId:            500,
		unitOfMeasure:        "",
	}

	testAttributeInfo1 := &archiveMeasureOrAttributeInfo{
		isAttribute:          true,
		measureOrAttributeId: 100,
		objectId:             200,
		objectTypeId:         1,
		stationId:            500,
		unitOfMeasure:        "",
	}

	um1 := &updatedMeasures{value: 100.25, measures: []*archiveMeasureOrAttributeInfo{testMeasureInfo1}}
	um2 := &updatedMeasures{value: core.GetNaN(), measures: []*archiveMeasureOrAttributeInfo{testMeasureInfo2}}
	um3 := &updatedMeasures{value: core.GetNaN(), measures: []*archiveMeasureOrAttributeInfo{testAttributeInfo1}}

	aTime := time.Now()
	packageInfo := &core.DataPackage{Format: 0, DeviceId: 100}
	updateResult := &measuresUpdateEventInfo{aTime, map[uint16]*updatedMeasures{
		1: um1, 2: um2, 3: um3}, packageInfo, []int{500}}

	res := updateResult.getArchiveServerRequest()

	assert.Len(t, res, 1)

	// TODO: check return values
}
