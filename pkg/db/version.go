package db

import (
	"strconv"
	"sync/atomic"
	"time"
)

var tempPk = atomic.Int64{}
var recordVersion = atomic.Int64{}

const recordVersionMask uint64 = 0x03ffffff

// InstanceId is an id used to identify the current instance of the application. It is important when multiple instances
// of the same application that uses the ORM are accessing the same database.
//
// The value is assigned automatically as the microsecond that the application starts at, in an 8-year interval.
// If your application is running behind a load balancer that can assign instance ids that are unique and sequential,
// you might consider setting InstanceId to that value.
var InstanceId = time.Now().UnixMicro() & int64(^(recordVersionMask << 38))

// RecordVersionFunc is a function that will return a new record version value given a previous value.
// Only set this one if the default behavior does not work for you. See RecordVersion.
var RecordVersionFunc func(prev int64) int64

// TemporaryPrimaryKey returns an atomically unique negative value to be used as a temporary primary key for new records.
// This aids in accessing individual records from a group by id.
func TemporaryPrimaryKey() string {
	i := tempPk.Add(-1)
	return strconv.FormatInt(i, 10)
}

// RecordVersion produces an atomically unique record version value that is different from prev.
// The value returned can be used to determine when a record has changed for optimistic locking.
// Many database implementations provide a mechanism to do row-level locking, and in those cases a basic incrementer
// would work. This only needs to be unique within a record and previous versions of that sinclge record.
//
// However, some NoSQL databases (DynamoDB for one) do not have a mechanism to lock rows ahead of changes, but rather expect
// a condition to be given to the database to check whether a value (like a version number) remains constant throughout
// the transaction, and will report an error after the transaction attempt has been made. In these situations, it is important
// that all instances of the application produce unique numbers for version changes.
//
// A suitable default method is used that will guarantee uniqueness provided that:
//  1. Instances are restarted at least every 8 years,
//  2. Multiple instances are not cold started at the same microsecond, and
//  3. Individual instances do not create more than 67 million records before being restarted.
//
// Even if these parameters cannot be guaranteed, it is still extremely unlikely that a collision will occur.
// You can replace the default method by setting the RecordVersionFunction value.
func RecordVersion(prev int64) (v int64) {
	if RecordVersionFunc != nil {
		return RecordVersionFunc(prev)
	}

	for { // repeat to make sure we come up with a number different from prev
		v = recordVersion.Add(1)
		v = v & 0x03ffffff
		v = v | (InstanceId << 26)
		if v != prev {
			break
		}
	}
	return
}
