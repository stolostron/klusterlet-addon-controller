package addon

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func validateValues(values, expectedValues string) error {
	if values == expectedValues {
		return nil
	}

	v := map[string]interface{}{}
	err := json.Unmarshal([]byte(values), &v)
	if err != nil {
		return fmt.Errorf("failed to unmarshal values %v", err)
	}
	ev := map[string]interface{}{}
	err = json.Unmarshal([]byte(expectedValues), &ev)
	if err != nil {
		return fmt.Errorf("failed to unmarshal expectedValues %v", err)
	}

	if !reflect.DeepEqual(v, ev) {
		return fmt.Errorf("the values and expected values are different")
	}
	return nil
}

func Test_updateAnnotationValues(t *testing.T) {
	cases := []struct {
		name             string
		gv               globalValues
		annotationValues string
		expectedValues   string
		expectedErr      bool
	}{
		{
			name: "annotation no global",
			gv: globalValues{Global: global{
				ImageOverrides: map[string]string{"multicloud_manager": "myquay.io/multicloud_manager:2.5"},
				NodeSelector:   map[string]string{"infraNode": "true"},
				ProxyConfig:    nil,
			}},
			annotationValues: `{"logLevel":1}`,
			expectedErr:      false,
			expectedValues:   `{"logLevel":1,"global":{"imageOverrides":{"multicloud_manager":"myquay.io/multicloud_manager:2.5"},"nodeSelector":{"infraNode":"true"}}}`,
		},
		{
			name: "annotation no value",
			gv: globalValues{Global: global{
				ImageOverrides: map[string]string{"multicloud_manager": "myquay.io/multicloud_manager:2.5"},
				NodeSelector:   map[string]string{"infraNode": "true"},
				ProxyConfig:    nil,
			}},
			annotationValues: "",
			expectedErr:      false,
			expectedValues:   `{"global":{"imageOverrides":{"multicloud_manager":"myquay.io/multicloud_manager:2.5"},"nodeSelector":{"infraNode":"true"}}}`,
		},
		{
			name: "annotation global image override",
			gv: globalValues{Global: global{
				ImageOverrides: map[string]string{"multicloud_manager": "myquay.io/multicloud_manager:2.5"},
				NodeSelector:   map[string]string{"infraNode": "false"},
				ProxyConfig:    map[string]string{"HTTP_PROXY": "1.1.1.1", "HTTPS_PROXY": "2.2.2.2", "NO_PROXY": "3.3.3.3"},
			}},
			annotationValues: `{"logLevel":1,"global":{"imageOverrides":{"multicloud_manager":"myquay.io/multicloud_manager:2.4"},"nodeSelector":{"infraNode":"true"}}}`,
			expectedErr:      false,
			expectedValues:   `{"logLevel":1,"global":{"imageOverrides":{"multicloud_manager":"myquay.io/multicloud_manager:2.5"},"nodeSelector":{"infraNode":"false"},"proxyConfig":{"HTTPS_PROXY":"2.2.2.2","HTTP_PROXY":"1.1.1.1","NO_PROXY":"3.3.3.3"}}}`,
		},
		{
			name: "annotation global no image override",
			gv: globalValues{Global: global{
				NodeSelector: map[string]string{"infraNode": "false"},
				ProxyConfig:  map[string]string{"HTTP_PROXY": "1.1.1.1", "HTTPS_PROXY": "2.2.2.2", "NO_PROXY": "3.3.3.3"},
			}},
			annotationValues: `{"logLevel":1,"global":{"pullPolicy":"Always","imageOverrides":{"multicloud_manager":"myquay.io/multicloud_manager:2.4"},"nodeSelector":{"infraNode":"true"}}}`,
			expectedErr:      false,
			expectedValues:   `{"logLevel":1,"global":{"pullPolicy":"Always","imageOverrides":{"multicloud_manager":"myquay.io/multicloud_manager:2.4"},"nodeSelector":{"infraNode":"false"},"proxyConfig":{"HTTPS_PROXY":"2.2.2.2","HTTP_PROXY":"1.1.1.1","NO_PROXY":"3.3.3.3"}}}`,
		},
		{
			name: "annotation no change",
			gv: globalValues{Global: global{
				ImageOverrides: map[string]string{"multicloud_manager": "myquay.io/multicloud_manager:2.5"},
				NodeSelector:   map[string]string{"infraNode": "false"},
				ProxyConfig:    map[string]string{"HTTP_PROXY": "1.1.1.1", "HTTPS_PROXY": "2.2.2.2", "NO_PROXY": "3.3.3.3"},
			}},
			annotationValues: `{"logLevel":1,"global":{"imageOverrides":{"multicloud_manager":"myquay.io/multicloud_manager:2.5"},"nodeSelector":{"infraNode":"false"},"proxyConfig":{"HTTPS_PROXY":"2.2.2.2","HTTP_PROXY":"1.1.1.1","NO_PROXY":"3.3.3.3"}}}`,
			expectedErr:      false,
			expectedValues:   "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			values, err := updateAnnotationValues(c.gv, c.annotationValues)
			if !c.expectedErr && err != nil {
				t.Errorf("expected no error but got %v", err)
			}
			if c.expectedErr && err == nil {
				t.Errorf("expected error but got nil")
			}
			if err := validateValues(values, c.expectedValues); err != nil {
				t.Errorf("expected values %v, but got %v. error:%v", c.expectedValues, values, err)
			}
		})
	}
}
