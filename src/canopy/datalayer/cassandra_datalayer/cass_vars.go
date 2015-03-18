/*
 * Copyright 2015 Canopy Services, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package cassandra_datalayer

import (
    "canopy/canolog"
    "canopy/config"
    "canopy/datalayer"
    "canopy/datalayer/cassandra_datalayer/migrations"
    "fmt"
    "github.com/gocql/gocql"
)

struct lodTier {
}

// bucketSizeEnum describes the "size" (i.e. time duration) of a bucket of
// samples.
type bucketSizeEnum int
const (
    BUCKET_SIZE_INVALID bucketSizeEnum = iota,
    BUCKET_SIZE_HOUR,   // Bucket contains 1 hour's worth of samples
    BUCKET_SIZE_DAY,    // Bucket contains 1 day's worth of samples
    BUCKET_SIZE_WEEK,   // Bucket contains 1 week's worth of samples
    BUCKET_SIZE_MONTH,  // You get the idea..
    BUCKET_SIZE_YEAR,
    LAST_BUCKET_SIZE,

    // Certain routines will return buckets of mixed sizes
    BUCKET_SIZE_MIXED,
)

type cloudVarTierEnum int
const (
    TIER_STANDARD cloudVarTierEnum = iota
    TIER_ENHANCED // Extra data storage
    TIER_ULTRA  // Even more data storage
)

// Get the bucket name based on time and bucket size.
func getBucketName(t time.Time, bucketSize bucketSizeEnum) string {
    switch bucketSize {
        case BUCKET_SIZE_HOUR:
            t = t.Truncate(time.Hour)
            return fmt.Sprintf("%2d%2d%2d%2d",
                    t.Year() % 100,
                    t.Month(),
                    t.Day(),
                    t.Hour())
        case BUCKET_SIZE_DAY:
            t = t.Truncate(time.Hour)
            return fmt.Sprintf("%2d%2d%2d", t.Year() % 100, t.Month(), t.Day())
        case BUCKET_SIZE_WEEK:
            // Rewind to Sunday
            dayOfWeek := t.Weekday()
            t = t.Add(-dayOfWeek*24*time.HOUR)
            t = t.Truncate(time.Hour)
            return fmt.Sprintf("%2d%2d%2dw", t.Year() % 100, t.Month(), t.Day())
        case BUCKET_SIZE_MONTH:
            t := time.Date(t.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
            return fmt.Sprintf("%2d%2d", t.Year() % 100, t.Month())
        case BUCKET_SIZE_YEAR:
            t := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
            return fmt.Sprintf("%2d%2d", t.Year() % 100, t.Month())
        default:
            panic("Problemo")
    }
}

// Get the time corresponding to the start of the "bucket" that the time falls
// into.
func roundTimeToBucketStart(t time.Time, bucketSize bucketSizeEnum) time.Time {
    switch bucketSize {
        case BUCKET_SIZE_HOUR:
            t = t.Truncate(time.Hour)
        case BUCKET_SIZE_DAY:
            t = t.Truncate(time.Hour)
        case BUCKET_SIZE_WEEK:
            // Rewind to Sunday
            dayOfWeek := t.Weekday()
            t = t.Add(-dayOfWeek*24*time.HOUR)
            t = t.Truncate(time.Hour)
        case BUCKET_SIZE_MONTH:
            t := time.Date(t.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
        case BUCKET_SIZE_YEAR:
            t := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
        default:
            panic("Problemo")
    }
    return t
}

// Increment a rounded time by 1 bucket size
// into.
func incTimeByBucketSize(t time.Time, bucketSize bucketSizeEnum) time.Time {
    switch bucketSize {
        case BUCKET_SIZE_HOUR:
            t = t.Add(time.Hour)
        case BUCKET_SIZE_DAY:
            t = t.AddDate(0, 0, 1)
        case BUCKET_SIZE_WEEK:
            t = t.AddDate(0, 0, 7)
        case BUCKET_SIZE_MONTH:
            t := time.Date(t.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
            t = t.AddDate(0, 1, 0)
        case BUCKET_SIZE_YEAR:
            t := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
            t = t.AddDate(1, 0, 0)
        default:
            panic("Problemo")
    }
    return t
}

// Returns True if t0 and t1 are in different buckets (for given bucketSize)
func crossedBoundary(t0, t1 time.Time, bucketSize bucketSizeEnum) {
    t0 = roundTimeToBucketStart(t0, bucketSize)
    t1 = roundTimeToBucketStart(t1, bucketSize)
    return !t0.Equal(t1)
}

// Get list of bucket names that span the <start> and <end> time.  If <end> is
// the zero value, then time.UTCNow() is used.  <bucketSize> determines the size of the
// buckets to return.
func getBucketNamesForTimeRange(start, end time.Time, bucketSize bucketSizeEnum) []string {
    if time.IsZero(end) {
        end = time.UTCNow()
    }

    out := []string{}

    // truncate <start> based on bucketSize

    switch bucketSize {
    case BUCKET_SIZE_HOUR:
        rangeStart := start.Truncate(time.Hour)
        rangeEnd := end.Truncate(time.Hour).Add(time.Hour)
    case BUCKET_SIZE_DAY:
        rangeStart := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
        rangeEnd := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
    case BUCKET_SIZE_WEEK:
        rangeEnd := start.AddDate(0, 0, 7)
    case BUCKET_SIZE_MONTH:
        rangeEnd := start.AddDate(0, 1, 0)
    }

    for t := start; t.Before(rangeEnd){
        switch 
        out = out.append(
    }
    switch bucketSizeEnum
}


func (device *CassDevice) insertSampleLOD(varDef sddl.VarDef, t time.Time, value interface{}) error {

func (device *CassDevice) InsertSample(varDef sddl.VarDef, t time.Time, value interface{}) error {
    // Get the most recent update time for this variable.
    // TBD: Do we need strong consistency for this (QUORUM?)

    // For each bucketSize, insert or discard sample based on our
    // stratification algorithm.
    for bucketSize=BUCKET_SIZE_HOUR; bucketSize < LAST_BUCKET_SIZE; bucketSize++ {
        err = insertSampleLOD(varDef, lastUpdateTime, bucketSize, t, value)
        if err != nil {
            // TODO: Transactionize/rollback?
            return err
        }
    }
}

func fetchAndAppendBucketSamples(apendee []cloudvar.CloudVarSample, conn CassConnection, startTime, endTime time.Time, bucketName string) error {
    tableName := "varsample_float" // TODO: Make generic

    query := device.conn.session.Query(`
            SELECT time, value
            FROM ` + tableName + `
            WHERE device_id = ?
                AND propname = ?
                AND timprefix = ?
                AND time >= ?
                AND time <= ?
    `, device.ID(), propname, bucketName, startTime, endTime).Consistency(gocql.One)

    iter := query.Iter()
    apendee := []cloudvar.CloudVarSample{}

    switch datatype {
    case sddl.DATATYPE_VOID:
        var value interface{}
        for iter.Scan(&timestamp) {
            apendee = append(apendee, cloudvar.CloudVarSample{timestamp, value})
        }
    case sddl.DATATYPE_STRING:
        var value string
        for iter.Scan(&timestamp, &value) {
            apendee = append(apendee, cloudvar.CloudVarSample{timestamp, value})
        }
    case sddl.DATATYPE_BOOL:
        var value bool
        for iter.Scan(&timestamp, &value) {
            apendee = append(apendee, cloudvar.CloudVarSample{timestamp, value})
        }
    case sddl.DATATYPE_INT8:
        var value int8
        for iter.Scan(&timestamp, &value) {
            apendee = append(apendee, cloudvar.CloudVarSample{timestamp, value})
        }
    case sddl.DATATYPE_UINT8:
        var value uint8
        for iter.Scan(&timestamp, &value) {
            apendee = append(apendee, cloudvar.CloudVarSample{timestamp, value})
        }
    case sddl.DATATYPE_INT16:
        var value int16
        for iter.Scan(&timestamp, &value) {
            apendee = append(apendee, cloudvar.CloudVarSample{timestamp, value})
        }
    case sddl.DATATYPE_UINT16:
        var value uint16
        for iter.Scan(&timestamp, &value) {
            apendee = append(apendee, cloudvar.CloudVarSample{timestamp, value})
        }
    case sddl.DATATYPE_INT32:
        var value int32
        for iter.Scan(&timestamp, &value) {
            apendee = append(apendee, cloudvar.CloudVarSample{timestamp, value})
        }
    case sddl.DATATYPE_UINT32:
        var value uint32
        for iter.Scan(&timestamp, &value) {
            apendee = append(apendee, cloudvar.CloudVarSample{timestamp, value})
        }
    case sddl.DATATYPE_FLOAT32:
        var value float32
        for iter.Scan(&timestamp, &value) {
            apendee = append(apendee, cloudvar.CloudVarSample{timestamp, value})
        }
    case sddl.DATATYPE_FLOAT64:
        var value float64
        for iter.Scan(&timestamp, &value) {
            apendee = append(apendee, cloudvar.CloudVarSample{timestamp, value})
        }
    case sddl.DATATYPE_DATETIME:
        var value time.Time
        for iter.Scan(&timestamp, &value) {
            apendee = append(apendee, cloudvar.CloudVarSample{timestamp, value})
        }
    case sddl.DATATYPE_INVALID:
        return []cloudvar.CloudVarSample{}, fmt.Errorf("Cannot get property values for DATATYPE_INVALID");
    default:
        return []cloudvar.CloudVarSample{}, fmt.Errorf("Cannot get property values for datatype %d", datatype);
    }

    if err := iter.Close(); err != nil {
        return []cloudvar.CloudVarSample{}, err
    }

    return apendee, nil
}

func (device *CassDevice) historicDataLOD(
    varDef sddl.VarDef, 
    startTime, 
    endTime time.Time,
    bucketSize bucketSizeEnum) ([]cloudvar.CloudVarSample, error) {

    samples := []cloudvar.CloudVarSample{}

    // Get list of all buckets containing samples we are interested in.
    // TODO: This could happen in parallel w/ map-reduce-like algo
    bucketNames = getBucketNamesForTimeRange(start, end time.Time, bucketSize)
    for _, bucketName := range bucketNames {
        err := fetchAndAppendBucketSamples(samples, device.conn, startTime, endTime, bucketName)
        if err != nil {
            return []cloudvar.CloudVarSample{}, err
        }
    }
    return sample, nil
}

// For a given bucket size, we store several buckets of that size depending on
// the Cloud Variable's upgrade tier.  This returns the time duration spanned
// by the buckets of a particular size, given the upgrade tier.
func cloudVarLODDuration(loudVarTierEnum, lod bucketSizeEnum) time.Duration {
    type LODTier struct {
        lod bucketSizeEnum
        tier cloudVarTierEnum
    }
    return map[LODTier]time.Duration {
        LODTier{BUCKET_SIZE_HOUR, TIER_STANDARD}: time.HOUR,
        LODTier{BUCKET_SIZE_HOUR, TIER_ENHANCED}: 24*time.HOUR,
        LODTier{BUCKET_SIZE_HOUR, TIER_ULTRA}: 7*24*time.HOUR,

        LODTier{BUCKET_SIZE_DAY, TIER_STANDARD}: 7*time.HOUR,
        LODTier{BUCKET_SIZE_DAY, TIER_ENHANCED}: 7*24*time.HOUR,
        LODTier{BUCKET_SIZE_DAY, TIER_ULTRA}: 31*24*time.HOUR, // TBD

        LODTier{BUCKET_SIZE_WEEK, TIER_STANDARD}: 7*24*time.HOUR,
        LODTier{BUCKET_SIZE_WEEK, TIER_ENHANCED}: 31*24*time.HOUR, // TBD
        LODTier{BUCKET_SIZE_WEEK, TIER_ULTRA}: 365*24*time.HOUR,

        LODTier{BUCKET_SIZE_MONTH, TIER_STANDARD}: 31*24*time.HOUR, // TBD
        LODTier{BUCKET_SIZE_MONTH, TIER_ENHANCED}: 365*24*time.HOUR,
        LODTier{BUCKET_SIZE_MONTH, TIER_ULTRA}: 4*365*24*time.HOUR,

        LODTier{BUCKET_SIZE_YEAR, TIER_STANDARD}: 365*31*24*time.HOUR,
        LODTier{BUCKET_SIZE_YEAR, TIER_ENHANCED}: 4*365*24*time.HOUR,
        LODTier{BUCKET_SIZE_YEAR, TIER_ULTRA}: 4*365*24*time.HOUR,
    }
}

func (device *CassDevice) HistoricData(
    varDef sddl.VarDef, 
    startTime, 
    endTime time.Time) ([]cloudvar.CloudVarSample, error) {

    // Figure out which resolution to use.
    // Pick the lowest resolution that covers the entire requested period.
    duration := endTime.Sub(startTime)
    var bucketSize bucketSizeEnum
    for bucketSize=BUCKET_SIZE_HOUR; bucketSize<LAST_BUCKET_SIZE; bucketSize++ {
        lodDuration := cloudVarLODDuration(TIER_STANDARD, bucketSize)
        if duration < lodDuration {
            break;
        }
    }
    if bucketSize == LAST_BUCKET_SIZE {
        bucketSize = BUCKET_SIZE_YEAR
    }

    // Fetch the data from that LOD
    return historicDataLOD(varDef, startTime, endTime time.Time, bucketSize)
}
