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
    "canopy/cloudvar"
    "canopy/sddl"
    "fmt"
    "github.com/gocql/gocql"
    "time"
)

// Canopy stores cloud variable timeseries data at multiple resolutions.  This
// approach keeps data storage under control and allows fast data lookup over
// any time period.  High res data is stored for short durations, whereas low
// res data is stored for longer durations.  The different resolutions are
// achieved by discarding fewer or more samples.
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
// A "stratifictation technique" is used to generate low resolution data.  Our
// approach breaks time up into calendar-aligned chunks (conceptually similar
// to the way we generate buckets).  Each chunk may contain at most a single
// data sample.  Further data samples that fall in the same chunk are
// discarded.  This approach gives us a (more-or-less) evenly-spaced
// downsampling of the input signal, while also making it easy to determine
// whether to keep or discard particular samples.
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
    LOD_0 lodEnum = iota
    LOD_1
    LOD_2
    LOD_3
    LOD_4
    LOD_5
    LOD_END
)

// bucketSizeEnum describes the "size" (i.e. time duration) of a bucket of
// samples.
type bucketSizeEnum int
const (
    BUCKET_SIZE_INVALID bucketSizeEnum = iota
    BUCKET_SIZE_15MIN  // Bucket contains 15 minutes worth of samples
    BUCKET_SIZE_HOUR   // Bucket contains 1 hour's worth of samples
    BUCKET_SIZE_DAY    // Bucket contains 1 day's worth of samples
    BUCKET_SIZE_WEEK   // Bucket contains 1 week's worth of samples
    BUCKET_SIZE_MONTH  // You get the idea..
    BUCKET_SIZE_YEAR
    LAST_BUCKET_SIZE

    // Certain routines will return buckets of mixed sizes
    BUCKET_SIZE_MIXED
)


// lodBucketSize maps LOD # to bucket size
var lodBucketSize = map[lodEnum]bucketSizeEnum{
    LOD_0: BUCKET_SIZE_15MIN,
    LOD_1: BUCKET_SIZE_HOUR,
    LOD_2: BUCKET_SIZE_DAY,
    LOD_3: BUCKET_SIZE_WEEK,
    LOD_4: BUCKET_SIZE_MONTH,
    LOD_5: BUCKET_SIZE_YEAR,
}

// stratificationSize describes the "size" (i.e. time duration) of a stratification
// chunk (which may contain at most 1 sample).
type stratificationSizeEnum int
const (
    STRATIFICATION_SIZE_INVALID stratificationSizeEnum = iota
    STRATIFICATION_1_SEC              // Store at most 1 sample / 1 sec.
    STRATIFICATION_5_SEC              // Store at most 1 sample / 5 sec.
    STRATIFICATION_2_MIN              // Store at most 1 sample / 2 min.
    STRATIFICATION_15_MIN             // Store at most 1 sample / 15 min.
    STRATIFICATION_1_HOUR             // Store at most 1 sample / hour.
    STRATIFICATION_12_HOUR            // Store at most 1 sample / hour.
    STRATIFICATION_END
)

// lodStratificationSize maps LOD # to stratificationSize
var lodStratificationSize = map[lodEnum]stratificationSizeEnum{
    LOD_0: STRATIFICATION_1_SEC,
    LOD_1: STRATIFICATION_5_SEC,
    LOD_2: STRATIFICATION_2_MIN,
    LOD_3: STRATIFICATION_15_MIN,
    LOD_4: STRATIFICATION_1_HOUR,
    LOD_5: STRATIFICATION_12_HOUR,
}

// stratificationPeriod maps stratificationSizeEnum value to time duration
var stratificationPeriod = map[stratificationSizeEnum]time.Duration {
    STRATIFICATION_1_SEC: time.Second,
    STRATIFICATION_5_SEC: 5*time.Second,
    STRATIFICATION_2_MIN: 2*time.Minute,
    STRATIFICATION_15_MIN: 15*time.Minute,
    STRATIFICATION_1_HOUR: 1*time.Hour,
    STRATIFICATION_12_HOUR: 12*time.Hour,
}

// storateTierEnum describes the cloud variable's "storage tier".
type storageTierEnum int
const (
    TIER_STANDARD storageTierEnum = iota
    TIER_ENHANCED // Extra data storage
    TIER_ULTRA  // Even more data storage
)

// lodStratificationPeriod maps LOD # to stratification time duration
func lodStratificationPeriod(lod lodEnum) time.Duration{
    return stratificationPeriod[lodStratificationSize[lod]]
}

// bucketStruct represents a "bucket" of samples, corresponding to a particular
// LOD level and calendar-aligned start time.
type bucketStruct struct {
    lod lodEnum
    startTime time.Time
}

// Get the LOD of a bucket
func (bucket bucketStruct)LOD() lodEnum {
    return bucket.lod
}

// Get the bucket size enum value of a bucket
func (bucket bucketStruct)BucketSize() bucketSizeEnum {
    return lodBucketSize[bucket.lod]
}

// Get the start time (inclusive) of a bucket
func (bucket bucketStruct)StartTime() time.Time {
    return bucket.startTime
}

// Get the end time (exclusive) of a bucket
func (bucket bucketStruct)EndTime() time.Time {
    return incTimeByBucketSize(bucket.startTime, bucket.BucketSize())
}

// Get the preceding bucket
/*func (bucket bucketStruct)Prev() bucketStruct {
    t := decTimeByBucketSize(bucket.startTime, bucket.BucketSize())
    return getBucket(t, bucket.BucketSize())
}*/

// Get the following bucket
func (bucket bucketStruct)Next() bucketStruct {
    return getBucket(bucket.EndTime(), bucket.LOD())
}

// Get the name of the bucket.  The bucket's name is also sometimes referred to
// as the "timeprefix".  For example:
//  "201503" for the 1-month bucket (Mar2015-Apr2015).
//  "20150314" for the 1-day bucket (Mar 14, 2015 - Mar 15, 2015)
func (bucket bucketStruct)Name() string {
    t := bucket.StartTime()

    // Assumes StartTime has already been rounded.
    switch bucket.BucketSize() {
    case BUCKET_SIZE_15MIN:
        return fmt.Sprintf("%02d%02d%02d%02d%02dq",
                t.Year() % 100,
                t.Month(),
                t.Day(),
                t.Hour(),
                t.Minute())
    case BUCKET_SIZE_HOUR:
        return fmt.Sprintf("%02d%02d%02d%02d",
                t.Year() % 100,
                t.Month(),
                t.Day(),
                t.Hour())
    case BUCKET_SIZE_DAY:
        return fmt.Sprintf("%02d%02d%02d", t.Year() % 100, t.Month(), t.Day())
    case BUCKET_SIZE_WEEK:
        return fmt.Sprintf("%02d%02d%02dw", t.Year() % 100, t.Month(), t.Day())
    case BUCKET_SIZE_MONTH:
        return fmt.Sprintf("%02d%02d", t.Year() % 100, t.Month())
    case BUCKET_SIZE_YEAR:
        return fmt.Sprintf("%02d%02d", t.Year() % 100, t.Month())
    default:
        panic("Problemo")
    }
}

// Get bucket object that contains time <t> for LOD <lod>.
func getBucket(t time.Time, lod lodEnum) bucketStruct {
    bucketSize := lodBucketSize[lod]
    startTime := roundTimeToBucketStart(t, bucketSize)
    return bucketStruct{
        lod: lod,
        startTime: startTime,
    }
}

// Get the time corresponding to the start of the "bucket" that the time falls
// into.
func roundTimeToBucketStart(t time.Time, bucketSize bucketSizeEnum) time.Time {
    switch bucketSize {
    case BUCKET_SIZE_15MIN:
        t = t.Truncate(15*time.Minute) // TODO: Does this work?
    case BUCKET_SIZE_HOUR:
        t = t.Truncate(time.Hour)
    case BUCKET_SIZE_DAY:
        t = t.Truncate(time.Hour)
    case BUCKET_SIZE_WEEK:
        // Rewind to Sunday
        dayOfWeek := t.Weekday()
        t = t.Add(-time.Duration(int(dayOfWeek))*24*time.Hour)
        t = t.Truncate(time.Hour)
    case BUCKET_SIZE_MONTH:
        t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
    case BUCKET_SIZE_YEAR:
        t = time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
    default:
        panic("Problemo")
    }
    return t
}

// Increment a rounded time by 1 bucket size.
func incTimeByBucketSize(t time.Time, bucketSize bucketSizeEnum) time.Time {
    switch bucketSize {
        case BUCKET_SIZE_15MIN:
            t = t.Add(15*time.Minute)
        case BUCKET_SIZE_HOUR:
            t = t.Add(time.Hour)
        case BUCKET_SIZE_DAY:
            t = t.AddDate(0, 0, 1)
        case BUCKET_SIZE_WEEK:
            t = t.AddDate(0, 0, 7)
        case BUCKET_SIZE_MONTH:
            t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
            t = t.AddDate(0, 1, 0)
        case BUCKET_SIZE_YEAR:
            t = time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
            t = t.AddDate(1, 0, 0)
        default:
            panic("Problemo")
    }
    return t
}

// Returns True if t0 and t1 are in different buckets (for given bucketSize)
func crossesBucketBoundary(t0, t1 time.Time, bucketSize bucketSizeEnum) bool {
    t0 = roundTimeToBucketStart(t0, bucketSize)
    t1 = roundTimeToBucketStart(t1, bucketSize)
    return !t0.Equal(t1)
}

// Get list of buckets that span the <start> and <end> time.
// If <end> is zero value, then time.UTCNow() is used.
func getBucketsForTimeRange(start, end time.Time, lod lodEnum) []bucketStruct {
    if end.IsZero() {
        end = time.Now().UTC()
    }

    out := []bucketStruct{}
    startBucket := getBucket(start, lod)
    for bucket := startBucket; bucket.StartTime().Before(end); bucket = bucket.Next() {
        out = append(out, bucket)
    }

    return out
}

// Get the closest time before or equal to <t> that is an integer multiple of
// <period>.
func stratificationBoundary(t time.Time, period time.Duration) time.Time {
    return t.Round(period)
    /*secsSinceEpoch := t.Unix()
    // TODO: Is it safe to use Unix for this?  Will leap seconds screw us up?
    periodSecs := int64(period)/int64(time.Second)
    // Round down to nearest stratification
    boundarySecsSinceEpoch := secsSinceEpoch - (secsSinceEpoch % periodSecs)
    return time.Unix(boundarySecsSinceEpoch)*/
}

// Determine if <t0> and <t1> fall within the same stratification chunk.
func crossesStratificationBoundary(t0, t1 time.Time, 
        stratification stratificationSizeEnum) bool {

    period := stratificationPeriod[stratification]
    sb0 := stratificationBoundary(t0, period)
    sb1 := stratificationBoundary(t1, period)
    canolog.Info("Stratification boundary", sb0, "(matches/mismatches ", sb1, ")")
    return !sb0.Equal(sb1)
}

// Track a bucket in the database for garbage collection purposes
func (device *CassDevice)addBucket(varName string, bucket *bucketStruct) error {
    err := device.conn.session.Query(`
            UPDATE var_buckets
            SET endtime = ?
            WHERE device_id = ?
                AND var_name = ?
                AND lod = ?
                AND timeprefix = ?
    `, bucket.EndTime(), device.ID(), varName, bucket.LOD(), bucket.Name()).Consistency(gocql.One).Exec()

    if err != nil {
        return err
    }
    return nil
}

func varTableNameByDatatype(datatype sddl.DatatypeEnum) (string, error) {
    switch datatype {
    case sddl.DATATYPE_VOID:
        return "varsample_void", nil
    case sddl.DATATYPE_STRING:
        return "varsample_string", nil
    case sddl.DATATYPE_BOOL:
        return "varsample_boolean", nil
    case sddl.DATATYPE_INT8:
        return "varsample_int", nil
    case sddl.DATATYPE_UINT8:
        return "varsample_int", nil
    case sddl.DATATYPE_INT16:
        return "varsample_int", nil
    case sddl.DATATYPE_UINT16:
        return "varsample_int", nil
    case sddl.DATATYPE_INT32:
        return "varsample_int", nil
    case sddl.DATATYPE_UINT32:
        return "varsample_int", nil
    case sddl.DATATYPE_FLOAT32:
        return "varsample_float", nil
    case sddl.DATATYPE_FLOAT64:
        return "varsample_double", nil
    case sddl.DATATYPE_DATETIME:
        return "varsample_timestamp", nil
    case sddl.DATATYPE_INVALID:
        return "", fmt.Errorf("DATATYPE_INVALID not allowed in varTableNameByDatatype");
    default: 
        return "", fmt.Errorf("Unexpected datatype in varTableNameByDatatype: %d", datatype);
    }
}

// Insert a sample into the database for a particular LOD level, discarding the
// sample if the stratification chunk already contains a sample.
func (device *CassDevice) insertOrDiscardSampleLOD(varDef sddl.VarDef, 
        lastUpdateTime time.Time,
        lod lodEnum, 
        t time.Time, 
        value interface{}) error {

    // Discard sample if it doesn't cross a stratification boundary
    stratificationSize := lodStratificationSize[lod]
    if !crossesStratificationBoundary(lastUpdateTime, t, stratificationSize) {
        // discard sample
        canolog.Info("LOD", lod, "discarded")
        return nil
    }

    // Get table name
    tableName, err := varTableNameByDatatype(varDef.Datatype())
    if err != nil {
        return err
    }

    // insert sample
    bucket := getBucket(t, lod)
    propname := varDef.Name()
    err = device.conn.session.Query(`
            INSERT INTO ` + tableName + ` 
                (device_id, propname, timeprefix, time, value)
            VALUES (?, ?, ?, ?, ?)
    `, device.ID(), propname, bucket.Name(), t, value).Exec()
    if err != nil {
        return err
    }
    canolog.Info("LOD", lod, "sample inserted into bucket", bucket.Name())

    // Track new bucket (if any) for garbage collection purposes.
    // And garbage collect.
    if crossesBucketBoundary(lastUpdateTime, t, bucket.BucketSize()) {
        err := device.addBucket(propname, &bucket)
        canolog.Info("New bucket", bucket, "created")
        if err != nil {
            canolog.Error("Error adding sample bucket: ", err)
            // don't return!  We still need to do garbage collection!
        }
        device.garbageCollectLOD(t, varDef, lod, false)
    }

    return nil
}

// Get the last time a cloud variable was updated.  Returns (time.Time{} (zero
// value), nil) if the Cloud Variable has never been set.
func (device *CassDevice)varLastUpdateTime(varName string) (time.Time, error) {
    var t time.Time
    // We use Quorum consistency here for strong conistency guarantee.
    query := device.conn.session.Query(`
            SELECT last_update
            FROM var_lastupdatetime
            WHERE device_id = ?
                AND var_name = ?
    `, device.ID(), varName).Consistency(gocql.Quorum)
    err := query.Scan(&t)
    if err != nil {
        switch err {
        case gocql.ErrNotFound:
            return time.Time{}, nil
        default:
            return time.Time{}, err
        }
    }
    return t, nil
}

// Update cloud variable's last update time.  Needed for stratified
// downsampling.
func (device *CassDevice)varSetLastUpdateTime(varName string, t time.Time) error {
    // TODO: How to use QUORUM for writes?
    err := device.conn.session.Query(`
            UPDATE var_lastupdatetime
            SET last_update = ?
            WHERE device_id = ?
                AND var_name = ?
    `, t, device.ID(), varName).Consistency(gocql.Quorum).Exec()
    if err != nil {
        return err
    }
    return nil
}

// Insert a cloud variable data sample.
func (device *CassDevice) InsertSample(varDef sddl.VarDef, t time.Time, value interface{}) error {
    // Convert to UTC before inserting
    t = t.UTC()
    canolog.Info("Inserting sample", varDef.Name(), t)

    // check last update time
    lastUpdateTime, err := device.varLastUpdateTime(varDef.Name())
    if err != nil {
        return err
    }
    canolog.Info("Last update time was", lastUpdateTime)

    if t.Before(lastUpdateTime) {
        canolog.Error("Insertion time before last update time: ", t, lastUpdateTime)
        return fmt.Errorf("Insertion time %s before last update time %s", t, lastUpdateTime)
    }

    // update last update time
    err = device.varSetLastUpdateTime(varDef.Name(), t)
    if err != nil {
        return err
    }

    // For each LOD, insert or discard sample based on our
    // stratification algorithm.
    for lod := LOD_0; lod < LOD_END; lod++ {
        err = device.insertOrDiscardSampleLOD(varDef, lastUpdateTime, lod, t, value)
        if err != nil {
            // TODO: Transactionize/rollback?
            return err
        }
    }

    // TODO: Do we need to update in-memory device object?
    return nil
}

// Append the samples in bucket <bucketName> that fall between <startTime> and
// <endTime> to <apendee>
func (device *CassDevice) fetchAndAppendBucketSamples(varDef sddl.VarDef, 
        apendee []cloudvar.CloudVarSample, 
        startTime, 
        endTime time.Time, 
        bucketName string) ([]cloudvar.CloudVarSample, error) {

    // Get table name
    tableName, err := varTableNameByDatatype(varDef.Datatype())
    if err != nil {
        return []cloudvar.CloudVarSample{}, err
    }

    query := device.conn.session.Query(`
            SELECT time, value
            FROM ` + tableName + `
            WHERE device_id = ?
                AND propname = ?
                AND timeprefix = ?
                AND time >= ?
                AND time <= ?
    `, device.ID(), varDef.Name(), bucketName, startTime, endTime).Consistency(gocql.One)

    iter := query.Iter()

    var timestamp time.Time
    switch varDef.Datatype() {
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
        return []cloudvar.CloudVarSample{}, fmt.Errorf("Cannot get property values for datatype %d", varDef.Datatype());
    }

    err = iter.Close(); 
    if err != nil {
        return []cloudvar.CloudVarSample{}, err
    }

    return apendee, nil
}

// Fetch the historic timeseries data for a particular LOD.
func (device *CassDevice) historicDataLOD(
    varDef sddl.VarDef, 
    start, 
    end time.Time,
    lod lodEnum) ([]cloudvar.CloudVarSample, error) {

    var err error
    samples := []cloudvar.CloudVarSample{}

    // Get list of all buckets containing samples we are interested in.
    // TODO: This could happen in parallel w/ map-reduce-like algo
    buckets := getBucketsForTimeRange(start, end, lod)
    canolog.Info("Using buckets: ", buckets)
    for _, bucket := range buckets {
        samples, err = device.fetchAndAppendBucketSamples(varDef, samples, start, end, bucket.Name())

        if err != nil {
            canolog.Info("Error: ", err)
            return samples, err
        }
        canolog.Info("Fetched ", len(samples), "samples from bucket", bucket.Name())
    }
    return samples, nil
}

// For a given bucket size, we store several buckets of that size depending on
// the Cloud Variable's upgrade tier.  This returns the time duration spanned
// by the buckets of a particular size, given the upgrade tier.
func cloudVarLODDuration(tier storageTierEnum, lod lodEnum) time.Duration {
    type LODTier struct {
        lod lodEnum
        tier storageTierEnum
    }
    return map[LODTier]time.Duration {
        LODTier{LOD_0, TIER_STANDARD}: 15*time.Minute,
        LODTier{LOD_0, TIER_ENHANCED}: time.Hour,
        LODTier{LOD_0, TIER_ULTRA}: 24*time.Hour,

        LODTier{LOD_1, TIER_STANDARD}: time.Hour,
        LODTier{LOD_1, TIER_ENHANCED}: 24*time.Hour,
        LODTier{LOD_1, TIER_ULTRA}: 7*24*time.Hour,

        LODTier{LOD_2, TIER_STANDARD}: 7*time.Hour,
        LODTier{LOD_2, TIER_ENHANCED}: 7*24*time.Hour,
        LODTier{LOD_2, TIER_ULTRA}: 31*24*time.Hour, // TBD

        LODTier{LOD_3, TIER_STANDARD}: 7*24*time.Hour,
        LODTier{LOD_3, TIER_ENHANCED}: 31*24*time.Hour, // TBD
        LODTier{LOD_3, TIER_ULTRA}: 365*24*time.Hour,

        LODTier{LOD_4, TIER_STANDARD}: 31*24*time.Hour, // TBD
        LODTier{LOD_4, TIER_ENHANCED}: 365*24*time.Hour,
        LODTier{LOD_4, TIER_ULTRA}: 4*365*24*time.Hour,

        LODTier{LOD_5, TIER_STANDARD}: 365*31*24*time.Hour,
        LODTier{LOD_5, TIER_ENHANCED}: 4*365*24*time.Hour,
        LODTier{LOD_5, TIER_ULTRA}: 4*365*24*time.Hour,
    }[LODTier{lod, tier}]
}

// Fetch historic time series data for a cloud variable. The resolution is
// automatically selected.
func (device *CassDevice) HistoricData(
    varDef sddl.VarDef, 
    curTime,
    startTime, 
    endTime time.Time) ([]cloudvar.CloudVarSample, error) {

    canolog.Info("Fetching historic data for", varDef.Name(), startTime, endTime)

    // Figure out which resolution to use.
    // Pick the highest resolution that covers the entire requested period.
    var lod lodEnum
    for lod = LOD_0; lod < LOD_END; lod++ {
        lodDuration := cloudVarLODDuration(TIER_STANDARD, lod)
        // TODO: Should we use curTime or lastUpdateTime for this?
        if startTime.After(curTime.Add(-lodDuration)) {
            break;
        }
    }
    if lod == LOD_END {
        lod = LOD_5
    }

    canolog.Info("Using LOD", lod)

    // Fetch the data from that LOD
    return device.historicDataLOD(varDef, startTime, endTime, lod)
}

// Determine if bucket has expired (and should be garbage collected).
func bucketExpired(curTime, 
        endTime time.Time, 
        tier storageTierEnum, 
        lod lodEnum) bool {

    // # amount of time after endTime that a bucket should stick around
    ttl := cloudVarLODDuration(tier, lod)
    expireTime := endTime.Add(ttl)
    return expireTime.Before(curTime)
}

// Remove old buckets for a single cloud variable and LOD
// Set <deleteAll> to false for normal garbage collection (only expired buckets
// are removed).  Set <deleteAll> to true to delete all data, expired or not.
func (device *CassDevice)garbageCollectLOD(curTime time.Time, 
        varDef sddl.VarDef,
        lod lodEnum,
        deleteAll bool) error {

    canolog.Info("Running garbage collection for ", varDef.Name(), "LOD", lod)

    // Get list of expired buckets for that LOD
    var bucketName string
    bucketsToRemove := []string{}

    query := device.conn.session.Query(`
            SELECT timeprefix, endtime
            FROM var_buckets
            WHERE device_id = ?
                AND var_name = ?
                AND lod = ?
            ORDER BY timeprefix DESC
    `, device.ID(), varDef.Name(), lod).Consistency(gocql.One)

    iter := query.Iter()

    var endTime time.Time
    // NOTE: As a special case, we never delete the most recent LOD0 bucket,
    // even if it has expired, because we need it for LastUpdateTime.
    skipFirst := (lod == LOD_0)
    for iter.Scan(&bucketName, &endTime) {
        // determine expiration time
        // TODO: Handle tiers
        if deleteAll || bucketExpired(curTime, endTime, TIER_STANDARD, lod) {
            if skipFirst {
                skipFirst = false
            } else {
                bucketsToRemove = append(bucketsToRemove, bucketName)
            }
        }
    }

    err := iter.Close(); 
    if err != nil {
        return fmt.Errorf("Error garbage collecting cloudvar: %s", err.Error())
    }

    // Remove buckets
    for _, bucketName := range bucketsToRemove {
        // Get table name
        tableName, err := varTableNameByDatatype(varDef.Datatype())
        if err != nil {
            return err
        }

        // Remove expired bucket
        canolog.Info("Removing expired bucket", varDef.Name(), bucketName)
        err = device.conn.session.Query(`
                DELETE FROM ` + tableName + `
                WHERE device_id = ?
                    AND propname = ?
                    AND timeprefix = ?
        `, device.ID(), varDef.Name(), bucketName).Consistency(gocql.One).Exec()
        if err != nil {
            canolog.Error("Problem deleting bucket ", device.ID(), varDef.Name(), bucketName)
        } else {
            // Cleanup var_buckets table, but only if we actually deleted the
            // bucket in the previous step
            err := device.conn.session.Query(`
                DELETE FROM var_buckets
                WHERE device_id = ?
                    AND var_name = ?
                    AND lod = ?
                    AND timeprefix = ?
            `, device.ID(), varDef.Name(), lod, bucketName).Consistency(gocql.One).Exec()
            if err != nil {
                canolog.Error("Problem cleaning var_buckets ", device.ID(), varDef.Name(), bucketName, ":", err)
            }
        }
    }
    return nil
}

func (device *CassDevice)ClearVarData(varDef sddl.VarDef) {
    // Delete all buckets
    for lod := LOD_0; lod < LOD_END; lod++ {
        device.garbageCollectLOD(time.Now(), varDef, lod, true)
    }
}

func (device *CassDevice) getLatestData_generic(varname string, datatype sddl.DatatypeEnum) (*cloudvar.CloudVarSample, error) {
    var timestamp time.Time
    var sample *cloudvar.CloudVarSample

    // Get table name
    tableName, err := varTableNameByDatatype(datatype)
    if err != nil {
        return nil, err
    }

    // Get most recent LOD0 bucket
    query := device.conn.session.Query(`
            SELECT timeprefix
            FROM var_buckets
            WHERE device_id = ?
                AND var_name = ?
                AND lod = ?
            ORDER BY timeprefix DESC
            LIMIT 1
    `, device.ID(), varname, LOD_0).Consistency(gocql.One)

    var timeprefix string
    err = query.Scan(&timeprefix)
    if err != nil {
        canolog.Error("Error getting most recent LOD_0 bucket", err)
        return nil, err
    }

    // Get most recent sample in most recent LOD0 bucket
    query = device.conn.session.Query(`
            SELECT time, value
            FROM ` + tableName + `
            WHERE device_id = ?
                AND propname = ?
                AND timeprefix = ?
            ORDER BY time DESC
            LIMIT 1
    `, device.ID(), varname, timeprefix).Consistency(gocql.One)

    switch datatype {
    case sddl.DATATYPE_VOID:
        err = query.Scan(&timestamp)
        sample = &cloudvar.CloudVarSample{timestamp, nil}
    case sddl.DATATYPE_STRING:
        var value string
        err = query.Scan(&timestamp, &value)
        sample = &cloudvar.CloudVarSample{timestamp, value}
    case sddl.DATATYPE_BOOL:
        var value bool
        err = query.Scan(&timestamp, &value)
        sample = &cloudvar.CloudVarSample{timestamp, value}
    case sddl.DATATYPE_INT8:
        var value int8
        err = query.Scan(&timestamp, &value)
        sample = &cloudvar.CloudVarSample{timestamp, value}
    case sddl.DATATYPE_UINT8:
        var value uint8
        err = query.Scan(&timestamp, &value)
        sample = &cloudvar.CloudVarSample{timestamp, value}
    case sddl.DATATYPE_INT16:
        var value int16
        err = query.Scan(&timestamp, &value)
        sample = &cloudvar.CloudVarSample{timestamp, value}
    case sddl.DATATYPE_UINT16:
        var value uint16
        err = query.Scan(&timestamp, &value)
        sample = &cloudvar.CloudVarSample{timestamp, value}
    case sddl.DATATYPE_INT32:
        var value int32
        err = query.Scan(&timestamp, &value)
        sample = &cloudvar.CloudVarSample{timestamp, value}
    case sddl.DATATYPE_UINT32:
        var value uint32
        err = query.Scan(&timestamp, &value)
        sample = &cloudvar.CloudVarSample{timestamp, value}
    case sddl.DATATYPE_FLOAT32:
        var value float32
        err = query.Scan(&timestamp, &value)
        sample = &cloudvar.CloudVarSample{timestamp, value}
    case sddl.DATATYPE_FLOAT64:
        var value float64
        err = query.Scan(&timestamp, &value)
        sample = &cloudvar.CloudVarSample{timestamp, value}
    case sddl.DATATYPE_DATETIME:
        var value time.Time
        err = query.Scan(&timestamp, &value)
        sample = &cloudvar.CloudVarSample{timestamp, value}
    case sddl.DATATYPE_INVALID:
        return nil, fmt.Errorf("Cannot get property values for DATATYPE_INVALID");
    default:
        return nil, fmt.Errorf("Cannot get property values for datatype %d", datatype);
    }

    if err != nil {
        return nil, fmt.Errorf("Error reading latest property value: ", err)
    }

    return sample, nil
}

func (device *CassDevice) LatestData(varDef sddl.VarDef) (*cloudvar.CloudVarSample, error) {
    return device.getLatestData_generic(varDef.Name(), varDef.Datatype())
}
