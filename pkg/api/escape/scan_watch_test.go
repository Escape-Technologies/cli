package escape

import (
	"reflect"
	"testing"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
)

func TestFilterEvents(t *testing.T) {
	t.Parallel()
	cases := []struct {
		events      []string
		lastEventID string
		expected    []string
	}{
		{
			events:      []string{"1", "2", "3"},
			lastEventID: "",
			expected:    []string{"1", "2", "3"},
		},
		{
			events:      []string{"1", "2", "3"},
			lastEventID: "not in list",
			expected:    []string{"1", "2", "3"},
		},
		{
			events:      []string{"1", "2", "3"},
			lastEventID: "1",
			expected:    []string{"2", "3"},
		},
		{
			events:      []string{"1", "2", "3"},
			lastEventID: "2",
			expected:    []string{"3"},
		},
		{
			events:      []string{"1", "2", "3"},
			lastEventID: "3",
			expected:    []string{},
		},
		{
			events:      []string{},
			lastEventID: "3",
			expected:    []string{},
		},
	}
	for _, c := range cases {
		t.Run(c.lastEventID, func(t *testing.T) {
			events := make([]v2.ListEvents200ResponseDataInner, len(c.events))
			for i, event := range c.events {
				events[i] = v2.ListEvents200ResponseDataInner{
					Id: event,
				}
			}
			actual := filterEvents(events, c.lastEventID)
			res := []string{}
			for _, event := range actual {
				res = append(res, event.Id)
			}
			if !reflect.DeepEqual(res, c.expected) {
				t.Errorf("expected %v, got %v", c.expected, res)
			}
		})
	}
}
