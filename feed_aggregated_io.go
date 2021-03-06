package getstream

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"
)

// AggregatedFeedActivity is a getstream Activity
// Use it to post activities to AggregatedFeeds
// It is also the response from AggregatedFeed Fetch and List Requests
type AggregatedFeedActivity struct {
	ID        string
	Actor     FeedID
	Verb      string
	Object    FeedID
	Target    FeedID
	Origin    FeedID
	TimeStamp *time.Time

	ForeignID string
	Data      *json.RawMessage
	MetaData  map[string]string

	To []Feed
}

// MarshalJSON is the custom marshal function for AggregatedFeedActivities
// It will be used by json.Marshal()
func (a AggregatedFeedActivity) MarshalJSON() ([]byte, error) {

	payload := make(map[string]interface{})

	for key, value := range a.MetaData {
		payload[key] = value
	}

	payload["actor"] = a.Actor.Value()
	payload["verb"] = a.Verb
	payload["object"] = a.Object.Value()
	payload["origin"] = a.Origin.Value()

	if a.ID != "" {
		payload["id"] = a.ID
	}
	if a.Target != "" {
		payload["target"] = a.Target.Value()
	}

	if a.Data != nil {
		payload["data"] = a.Data
	}

	if a.ForeignID != "" {
		r, err := regexp.Compile("^[a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}$")
		if err != nil {
			return nil, err
		}
		if !r.MatchString(a.ForeignID) {
			return nil, errors.New("invalid ForeignID")
		}
		payload["foreign_id"] = a.ForeignID
	}

	if a.TimeStamp == nil {
		payload["time"] = time.Now().Format("2006-01-02T15:04:05.999999")
	} else {
		payload["time"] = a.TimeStamp.Format("2006-01-02T15:04:05.999999")
	}

	var tos []string
	for _, feed := range a.To {
		to := feed.FeedID().Value()
		if feed.Token() != "" {
			to += " " + feed.Token()
		}
		tos = append(tos, to)
	}

	if len(tos) > 0 {
		payload["to"] = tos
	}

	return json.Marshal(payload)

}

// UnmarshalJSON is the custom unmarshal function for AggregatedFeedActivities
// It will be used by json.Unmarshal()
func (a *AggregatedFeedActivity) UnmarshalJSON(b []byte) (err error) {

	rawPayload := make(map[string]*json.RawMessage)
	metadata := make(map[string]string)

	err = json.Unmarshal(b, &rawPayload)
	if err != nil {
		return err
	}

	for key, value := range rawPayload {
		lowerKey := strings.ToLower(key)

		if value == nil {
			continue
		}

		if lowerKey == "id" {
			var strValue string
			json.Unmarshal(*value, &strValue)
			a.ID = strValue
		} else if lowerKey == "actor" {
			var strValue string
			json.Unmarshal(*value, &strValue)
			a.Actor = FeedID(strValue)
		} else if lowerKey == "verb" {
			var strValue string
			json.Unmarshal(*value, &strValue)
			a.Verb = strValue
		} else if lowerKey == "foreign_id" {
			var strValue string
			json.Unmarshal(*value, &strValue)
			a.ForeignID = strValue
		} else if lowerKey == "object" {
			var strValue string
			json.Unmarshal(*value, &strValue)
			a.Object = FeedID(strValue)
		} else if lowerKey == "origin" {
			var strValue string
			json.Unmarshal(*value, &strValue)
			a.Origin = FeedID(strValue)
		} else if lowerKey == "target" {
			var strValue string
			json.Unmarshal(*value, &strValue)
			a.Target = FeedID(strValue)
		} else if lowerKey == "time" {
			var strValue string
			err := json.Unmarshal(*value, &strValue)
			if err != nil {
				continue
			}
			timeStamp, err := time.Parse("2006-01-02T15:04:05.999999", strValue)
			if err != nil {
				continue
			}
			a.TimeStamp = &timeStamp
		} else if lowerKey == "data" {
			a.Data = value
		} else if lowerKey == "to" {

			var to1D []string
			var to2D [][]string

			err := json.Unmarshal(*value, &to1D)
			if err != nil {
				err = nil
				err = json.Unmarshal(*value, &to2D)
				if err != nil {
					continue
				}

				for _, to := range to2D {
					if len(to) == 2 {
						feedStr := to[0] + " " + to[1]
						to1D = append(to1D, feedStr)
					} else if len(to) == 1 {
						to1D = append(to1D, to[0])
					}
				}
			}

			for _, to := range to1D {

				feed := GeneralFeed{}

				match, err := regexp.MatchString(`^\w+:\w+ .*?$`, to)
				if err != nil {
					continue
				}

				if match {
					firstSplit := strings.Split(to, ":")
					secondSplit := strings.Split(firstSplit[1], " ")

					feed.FeedSlug = firstSplit[0]
					feed.UserID = secondSplit[0]
					feed.token = secondSplit[1]
					a.To = append(a.To, &feed)
					continue
				}

				match = false
				err = nil

				match, err = regexp.MatchString(`^\w+:\w+$`, to)
				if err != nil {
					continue
				}

				if match {
					firstSplit := strings.Split(to, ":")

					feed.FeedSlug = firstSplit[0]
					feed.UserID = firstSplit[1]
					a.To = append(a.To, &feed)
					continue
				}
			}
		} else {
			var strValue string
			json.Unmarshal(*value, &strValue)
			metadata[key] = strValue
		}
	}

	a.MetaData = metadata
	return nil

}

type postAggregatedFeedOutputActivities struct {
	Activities []*AggregatedFeedActivity `json:"activities"`
}

// GetAggregatedFeedInput is used to Get a list of Activities from a AggregatedFeed
type GetAggregatedFeedInput struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`

	IDGTE string `json:"id_gte,omitempty"`
	IDGT  string `json:"id_gt,omitempty"`
	IDLTE string `json:"id_lte,omitempty"`
	IDLT  string `json:"id_lt,omitempty"`

	Ranking string `json:"ranking,omitempty"`
}

// GetAggregatedFeedOutput is the response from a AggregatedFeed Activities Get Request
type GetAggregatedFeedOutput struct {
	Duration string
	Next     string
	Results  []*struct {
		Activities    []*AggregatedFeedActivity
		ActivityCount int
		ActorCount    int
		CreatedAt     string
		Group         string
		ID            string
		UpdatedAt     string
		Verb          string
	}
}

type getAggregatedFeedOutput struct {
	Duration string                           `json:"duration"`
	Next     string                           `json:"next"`
	Results  []*getAggregatedFeedOutputResult `json:"results"`
}

func (a getAggregatedFeedOutput) output() *GetAggregatedFeedOutput {

	output := GetAggregatedFeedOutput{
		Duration: a.Duration,
		Next:     a.Next,
	}

	var results []*struct {
		Activities    []*AggregatedFeedActivity
		ActivityCount int
		ActorCount    int
		CreatedAt     string
		Group         string
		ID            string
		UpdatedAt     string
		Verb          string
	}

	for _, result := range a.Results {

		outputResult := struct {
			Activities    []*AggregatedFeedActivity
			ActivityCount int
			ActorCount    int
			CreatedAt     string
			Group         string
			ID            string
			UpdatedAt     string
			Verb          string
		}{
			ActivityCount: result.ActivityCount,
			ActorCount:    result.ActorCount,
			CreatedAt:     result.CreatedAt,
			Group:         result.Group,
			ID:            result.ID,
			UpdatedAt:     result.UpdatedAt,
			Verb:          result.Verb,
		}

		for _, activity := range result.Activities {
			outputResult.Activities = append(outputResult.Activities, activity)
		}

		results = append(results, &outputResult)
	}

	output.Results = results

	return &output
}

type getAggregatedFeedOutputResult struct {
	Activities    []*AggregatedFeedActivity `json:"activities"`
	ActivityCount int                       `json:"activity_count"`
	ActorCount    int                       `json:"actor_count"`
	CreatedAt     string                    `json:"created_at"`
	Group         string                    `json:"group"`
	ID            string                    `json:"id"`
	UpdatedAt     string                    `json:"updated_at"`
	Verb          string                    `json:"verb"`
}

type getAggregatedFeedFollowersInput struct {
	Limit int `json:"limit"`
	Skip  int `json:"offset"`
}

type getAggregatedFeedFollowersOutput struct {
	Duration string                                    `json:"duration"`
	Results  []*getAggregatedFeedFollowersOutputResult `json:"results"`
}

type getAggregatedFeedFollowersOutputResult struct {
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	FeedID    string `json:"feed_id"`
	TargetID  string `json:"target_id"`
}

type postAggregatedFeedFollowingInput struct {
	Target            string `json:"target"`
	ActivityCopyLimit int    `json:"activity_copy_limit"`
}
