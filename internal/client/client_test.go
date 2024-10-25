package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAgent_PostUpdate(t *testing.T) {
	type fields struct {
		BaseUrl string
		Client  *http.Client
		Stats   Stats
	}
	type args struct {
		metricType  string
		metricName  string
		metricValue string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *http.Response
	}{
		{
			name: "test 1",
			fields: fields{
				BaseUrl: "http://localhost",
				Client:  http.DefaultClient,
				Stats: Stats{
					Gauge: Gauge{
						Alloc: 1.1245,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				BaseUrl: tt.fields.BaseUrl,
				Client:  tt.fields.Client,
				Stats:   tt.fields.Stats,
			}

			fmt.Println(a)
			httptest.NewRequest(http.MethodPost, tt.fields.BaseUrl+tt.args.metricType+tt.args.metricName+tt.args.metricValue, nil)
			//if got := a.PostUpdate(tt.args.metricType, tt.args.metricName, tt.args.metricValue); !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("PostUpdate() = %v, want %v", got, tt.want)
			//}
		})
	}
}
