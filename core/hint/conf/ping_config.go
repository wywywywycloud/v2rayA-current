package conf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/xtls/xray-core/common/errors"
)

func (e *extendedJSON) UnmarshalJSON(data []byte) error {
	var raw struct {
		Multi *struct {
			Observers []json.RawMessage `json:"observers"`
		} `json:"multiObservatory"`
		Burst json.RawMessage `json:"burstObservatory"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Multi != nil {
		e.MultiObservatory = &multiObsJSON{}
		for i, entry := range raw.Multi.Observers {
			var observer multiObsEntryJSON
			if err := json.Unmarshal(entry, &observer); err != nil {
				var identity struct {
					Tag string `json:"tag"`
				}
				_ = json.Unmarshal(entry, &identity)
				return errors.New(fmt.Sprintf("multiObservatory.observers[%d] tag %q", i, identity.Tag)).Base(err)
			}
			e.MultiObservatory.Observers = append(e.MultiObservatory.Observers, observer)
		}
	}
	if len(raw.Burst) > 0 {
		if err := json.Unmarshal(raw.Burst, &e.BurstObservatory); err != nil {
			return errors.New("burstObservatory").Base(err)
		}
	}
	return nil
}

// Accept the native Xray spelling as well as the v5 spelling used by v2rayA.
func (p *pingConfigJSON) UnmarshalJSON(data []byte) error {
	type plain pingConfigJSON
	value := struct {
		*plain
		Sampling      *int32 `json:"sampling"`
		SamplingCount *int32 `json:"samplingCount"`
	}{plain: (*plain)(p)}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	if value.Sampling != nil {
		p.SamplingCount = int(*value.Sampling)
	}
	if value.SamplingCount != nil {
		if value.Sampling != nil && *value.Sampling != *value.SamplingCount {
			return errors.New("pingConfig sampling and samplingCount disagree")
		}
		p.SamplingCount = int(*value.SamplingCount)
	}
	if p.SamplingCount < 0 {
		return errors.New("pingConfig samplingCount must not be negative")
	}
	if p.Timeout < 0 || p.Interval < 0 {
		return errors.New("pingConfig timeout and interval must not be negative")
	}
	if p.HTTPMethod != "" {
		method := strings.TrimSpace(p.HTTPMethod)
		if method == "" {
			return errors.New("pingConfig httpMethod must not be whitespace")
		}
		if _, err := http.NewRequest(method, "http://localhost/", nil); err != nil {
			return errors.New("invalid pingConfig httpMethod").Base(err)
		}
	}
	return nil
}
