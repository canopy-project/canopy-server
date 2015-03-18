/*
 * Copyright 2014 Gregory Prisament
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

func (device *CassDevice) HistoricData(
    varDef sddl.VarDef, 
    startTime, 
    endTime time.Time) ([]cloudvar.CloudVarSample, error) {

}
