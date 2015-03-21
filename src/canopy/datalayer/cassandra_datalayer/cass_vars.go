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

// Canopy stores cloud variable timeseries data at multiple resolutions.  This
// approach keeps data storage under control and allows fast data lookup over
// any time period.  High res data is stored for short durations,
// whereas low res data is stored for longer durations.  The different
// resolutions are achieved by discarding fewer or more samples.
//
// The LOD (Level of Detail) is an integer. 0=highest resolution with shortest
// duration, 5=lowest resolution with longest duration.
//
// Internally, samples are collected in "buckets".  The LOD determines the
// "bucket size" which is the time duration that a bucket represents:
//
//  LOD         BUCKET SIZE
//  ------------------------------
//  LOD_0       15 minute
//  LOD_1       1 hour
//  LOD_2       1 day
//  LOD_3       1 week
//  LOD_4       1 month
//  LOD_5       1 year
//
// For each LOD, buckets are created aligned to the calendar.  Here are some
// example buckets:
//
//  EXAMPLE BUCKETS:
//  LOD    Bucket Start         Bucket End (exclusive)   Bucket Duration
//  ------------------------------------------------------------------------
//  LOD_0  2015-04-03 10:15     2015-04-03 10:30         15 min
//  LOD_0  2015-04-03 10:30     2015-04-03 10:45         15 min
//  LOD_1  2015-04-03 10:00     2015-04-03 11:00         1 hour
//  LOD_2  2015-04-03           2015-04-04               1 day
//  LOD_4  2015-04              2015-05                  1 month
//  LOD_5  2015                 2016                     1 year
//
// A sample may appear in multiple buckets.  For example, a data sample that
// occurred at "2015-04-03 10:22:43" may end up in several of the above
// buckets.
//
// A garbage collection mechanism deletes old buckets once they have expired.
// The time until a bucket expired is determined by both the LOD and the cloud
// variable's "storage tier".
//
//  STORAGE TIER        LOD     EXPIRATION           MAX ACTIVE BUCKETS FOR LOD
//  ------------------------------------------------------------------------
//  STANDARD            LOD_0   15 min after Bucket End         2
//  STANDARD            LOD_1   1 hour after Bucket End         2
//  STANDARD            LOD_2   1 day after Bucket End          2
//  STANDARD            LOD_3   1 week after Bucket End         2
//  STANDARD            LOD_4   1 month after Bucket End        2
//  STANDARD            LOD_5   1 year after Bucket End         2
//  
//  DELUXE              LOD_0   1 hour after Bucket End         5
//  DELUXE              LOD_1   1 day after Bucket End          25
//  DELUXE              LOD_2   1 week after Bucket End         8
//  DELUXE              LOD_3   1 month after Bucket End        5
//  DELUXE              LOD_4   1 year after Bucket End         13
//  DELUXE              LOD_5   4 years after Bucket End        5
//
//  ULTRA               LOD_0   1 day after Bucket End          97
//  ULTRA               LOD_1   1 week after Bucket End         169
//  ULTRA               LOD_2   1 month after Bucket End        31
//  ULTRA               LOD_3   1 year after Bucket End         53
//  ULTRA               LOD_4   4 years after Bucket End        49
//  ULTRA               LOD_5   16 years after Bucket End       17
//
//
// A "stratifictation approach" is used to generate low resolution data.  Our
// approach breaks time up into calendar-aligned chunks (conceptually similar
// to the way we generate buckets).  Each chunk may contain at most a single
// data sample.  Further data samples that fall in the same chunk are
// discarded.  This approach gives us a (more-or-less) evenly-spaced
// downsampling of the input signal, while also making it easy to determine
// whether to keep or discard a sample.
//
//  LOD         STRATIFICATION PERIOD
//  ------------------------------
//  LOD_0       1 sec
//  LOD_1       5 sec
//  LOD_2       2 min
//  LOD_3       15 min
//  LOD_4       1 hour
//  LOD_5       12 hour

type lodEnum int
const (
    LOD_0 lodEnum = iota,
    LOD_1,
    LOD_2,
    LOD_3,
    LOD_4,
    LOD_5,
    LOD_END,
)

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

func stratificationBoundary(t time.Time, maxSamplingPeriod time.Duration) time.Time {
    // TODO
}


func (device *CassDevice) insertOrDiscardSampleLOD(varDef sddl.VarDef, last_insert_t, t time.Time, bucketSize bucketSizeEnum, value interface{}) error {
    // Get the stratification boundary occuring most recently before <t>.
    // Only insert sample if <last_insert_time> is before the stratification
    // boundary.
    samplingPeriod := LODSamplingPeriod(bucketSize)
    strat_boundary_t := stratificationBoundary(t, samplingPeriod)
    if last_insert_t < strat_boundary_t {
        bucketName := getBucketName(t, bucketSize)

        // insert sample
        err := device.conn.session.Query(`
                INSERT INTO varsample_float (device_id, propname, timeprefix, time, value)
                VALUES (?, ?, ?, ?)
        `, device.ID(), propname, t, value).Exec()
        if err != nil {
            return err;
        }
        return nil;
    } else {
        // discard sample
        return nil
    }
}

func (device *CassDevice) InsertSample(varDef sddl.VarDef, t time.Time, value interface{}) error {
    // Get the most recent update time for this variable.
    // TBD: Do we need strong consistency for this (QUORUM?)

    // For each bucketSize, insert or discard sample based on our
    // stratification algorithm.
    for bucketSize=BUCKET_SIZE_HOUR; bucketSize < LAST_BUCKET_SIZE; bucketSize++ {
        err = insertOrDiscardSampleLOD(varDef, lastUpdateTime, bucketSize, t, value)
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

func LODSamplingPeriod(lod bucketSizeEnum) time.Duration {
    return map[bucketSizeEnum]time.Duration {
        BUCKET_SIZE_HOUR: 5*time.SECOND,
        BUCKET_SIZE_DAY: 2*time.MINUTE,
        BUCKET_SIZE_WEEK: 15*time.MINUTE,
        BUCKET_SIZE_MONTH: 1*time.HOUR,
        BUCKET_SIZE_YEAR: 12*time.HOUR
    }[lod]
}

// For a given bucket size, we store several buckets of that size depending on
// the Cloud Variable's upgrade tier.  This returns the time duration spanned
// by the buckets of a particular size, given the upgrade tier.
func cloudVarLODDuration(cloudVarTierEnum, lod bucketSizeEnum) time.Duration {
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
    }[cloudVarTierEnum, lod]
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

func crossedBucketThreshold(t0, t1 time.Time, bucketSize bucketSizeEnum) bool {
    bucket0 := getBucketName(t0, bucketSize)
    bucket1 := getBucketName(t1, bucketSize)
    return bucket0 != bucket1
}

// Remove old buckets for a single cloud variable
func garbageCollect(varDef sddl.VarDef) {
    // For each bucketSize, did we cross a threshold since last sample?
    // If so, cleanup older buckets.

    var bucketSize bucketSizeEnum
    for bucketSize=BUCKET_SIZE_HOUR; bucketSize<LAST_BUCKET_SIZE; bucketSize++ {
        lastSampleBucket := getBucketName(prev_t, bucketSize)
        curSampleBucket := getBucketName(cur_t bucketSize)
        if lastSampleBucket != curSampleBucket {
        }

    func getBucketName(t time.Time, bucketSize bucketSizeEnum) string {
        }
    }
}
