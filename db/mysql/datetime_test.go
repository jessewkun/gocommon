package mysql

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateTime_Value(t *testing.T) {
	// Test with a non-zero time
	tm, _ := time.Parse("2006-01-02 15:04:05", "2024-01-10 12:30:00")
	dt := DateTime(tm)
	val, err := dt.Value()
	assert.NoError(t, err)
	assert.Equal(t, "2024-01-10 12:30:00", val)

	// Test with a zero time
	var zeroDt DateTime
	val, err = zeroDt.Value()
	assert.NoError(t, err)
	assert.Nil(t, val, "Zero DateTime should produce a nil value for the database")
}

func TestDateTime_Scan(t *testing.T) {
	var dt DateTime

	// Test scanning from nil
	err := dt.Scan(nil)
	assert.NoError(t, err)
	assert.True(t, time.Time(dt).IsZero(), "Scanning nil should result in a zero time")

	// Test scanning from time.Time
	tm, _ := time.Parse("2006-01-02 15:04:05", "2024-01-10 12:30:00")
	err = dt.Scan(tm)
	assert.NoError(t, err)
	assert.Equal(t, DateTime(tm), dt)

	// Test scanning from string
	err = dt.Scan("2024-01-10 12:31:00")
	expectedTm, _ := time.ParseInLocation("2006-01-02 15:04:05", "2024-01-10 12:31:00", time.Local)
	assert.NoError(t, err)
	assert.Equal(t, DateTime(expectedTm), dt)

	// Test scanning from []byte
	err = dt.Scan([]byte("2024-01-10 12:32:00"))
	expectedTmBytes, _ := time.ParseInLocation("2006-01-02 15:04:05", "2024-01-10 12:32:00", time.Local)
	assert.NoError(t, err)
	assert.Equal(t, DateTime(expectedTmBytes), dt)

	// Test scanning invalid type
	err = dt.Scan(12345)
	assert.Error(t, err)
}

func TestDateTime_MarshalJSON(t *testing.T) {
	// Test marshalling a non-zero time
	tm, _ := time.Parse("2006-01-02 15:04:05", "2024-01-10 12:30:00")
	dt := DateTime(tm)
	jsonBytes, err := json.Marshal(dt)
	assert.NoError(t, err)
	assert.Equal(t, `"2024-01-10 12:30:00"`, string(jsonBytes))

	// Test marshalling a zero time
	var zeroDt DateTime
	jsonBytes, err = json.Marshal(zeroDt)
	assert.NoError(t, err)
	assert.Equal(t, "null", string(jsonBytes), "Zero DateTime should marshal to null")
}

func TestDateTime_UnmarshalJSON(t *testing.T) {
	var dt DateTime

	// Test unmarshalling from a valid time string
	err := json.Unmarshal([]byte(`"2024-01-10 12:30:00"`), &dt)
	expectedTm, _ := time.ParseInLocation("2006-01-02 15:04:05", "2024-01-10 12:30:00", time.Local)
	assert.NoError(t, err)
	assert.Equal(t, DateTime(expectedTm), dt)

	// Test unmarshalling from null
	err = json.Unmarshal([]byte("null"), &dt)
	assert.NoError(t, err)
	assert.True(t, time.Time(dt).IsZero(), "Unmarshalling null should result in a zero time")

	// Test unmarshalling from an invalid format
	err = json.Unmarshal([]byte(`"not-a-time"`), &dt)
	assert.Error(t, err)

	// Test unmarshalling from a non-string value
	err = json.Unmarshal([]byte("12345"), &dt)
	assert.Error(t, err)
}

func TestDateTime_JSON_Roundtrip(t *testing.T) {
	type TempStruct struct {
		EventTime DateTime `json:"event_time"`
	}

	// Test with a value
	s1 := TempStruct{EventTime: DateTime(time.Now())}
	jsonData, err := json.Marshal(s1)
	assert.NoError(t, err)

	var s2 TempStruct
	err = json.Unmarshal(jsonData, &s2)
	assert.NoError(t, err)

	// Compare formatted strings because `time.Time` objects can have subtle differences
	assert.Equal(t, s1.EventTime.String(), s2.EventTime.String())

	// Test with null
	s3 := TempStruct{EventTime: DateTime(time.Time{})}
	jsonData, err = json.Marshal(s3)
	assert.NoError(t, err)
	assert.Equal(t, `{"event_time":null}`, string(jsonData))

	var s4 TempStruct
	err = json.Unmarshal(jsonData, &s4)
	assert.NoError(t, err)
	assert.True(t, time.Time(s4.EventTime).IsZero())
}
