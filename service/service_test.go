package service

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/alibabacloud-go/tea/utils"
)

type validatorTest struct {
	Num  *int       `json:"num" require:"true"`
	Str  *string    `json:"str" pattern:"^[a-d]*$" maxLength:"4"`
	Test *errLength `json:"test"`
	List []*string  `json:"list" pattern:"^[a-d]*$" maxLength:"4"`
}

type errLength struct {
	Num *int `json:"num" maxLength:"a"`
}

func Test_GetEndpoint(t *testing.T) {
	endpoint := GetEndpoint(tea.String("common.aliyuncs.com"), tea.Bool(true), tea.String("internal"))
	utils.AssertEqual(t, "common-internal.aliyuncs.com", tea.StringValue(endpoint))

	endpoint = GetEndpoint(tea.String("common.aliyuncs.com"), tea.Bool(true), tea.String("accelerate"))
	utils.AssertEqual(t, "oss-accelerate.aliyuncs.com", tea.StringValue(endpoint))

	endpoint = GetEndpoint(tea.String("common.aliyuncs.com"), tea.Bool(true), tea.String(""))
	utils.AssertEqual(t, "common.aliyuncs.com", tea.StringValue(endpoint))
}

type TestCommon struct {
	Body io.Reader `json:"Body"`
	Test string    `json:"Test"`
}

func Test_Convert(t *testing.T) {
	in := &TestCommon{
		Body: strings.NewReader("common"),
		Test: "ok",
	}
	out := new(TestCommon)
	Convert(in, &out)
	utils.AssertEqual(t, "ok", out.Test)
}

func Test_GetTimestamp(t *testing.T) {
	stamp := GetTimestamp()
	utils.AssertNotNil(t, stamp)
}

func Test_GetSignatureV1(t *testing.T) {
	signed := map[string]*string{
		"test": tea.String("ok"),
	}

	sign := GetSignatureV1(signed, tea.String(""), tea.String("accessKeySecret"))
	utils.AssertEqual(t, "jHx/oHoHNrbVfhncHEvPdHXZwHU=", tea.StringValue(sign))
}

func Test_GetSignature(t *testing.T) {
	req := tea.NewRequest()
	req.Query["test"] = tea.String("ok")

	sign := GetSignature(req, tea.String("accessKeySecret"))
	utils.AssertEqual(t, "jHx/oHoHNrbVfhncHEvPdHXZwHU=", tea.StringValue(sign))
}

func Test_HasError(t *testing.T) {
	iserror := HasError(nil)
	utils.AssertEqual(t, true, tea.BoolValue(iserror))

	body := map[string]interface{}{
		"Code": "200",
	}
	iserror = HasError(body)
	utils.AssertEqual(t, true, tea.BoolValue(iserror))

	body = map[string]interface{}{
		"Code": 0,
	}
	iserror = HasError(body)
	utils.AssertEqual(t, false, tea.BoolValue(iserror))

	body = map[string]interface{}{
		"Code": 0.00,
	}
	iserror = HasError(body)
	utils.AssertEqual(t, false, tea.BoolValue(iserror))

	body = make(map[string]interface{})
	iserror = HasError(body)
	utils.AssertEqual(t, false, tea.BoolValue(iserror))
}

func Test_Query(t *testing.T) {
	filter := map[string]interface{}{
		"client": "test",
		"tag": map[string]string{
			"key": "value",
		},
		"strs": []string{"str1", "str2"},
		"num":  123456789,
	}

	result := Query(filter)
	utils.AssertEqual(t, "test", tea.StringValue(result["client"]))
	utils.AssertEqual(t, "value", tea.StringValue(result["tag.key"]))
	utils.AssertEqual(t, "str1", tea.StringValue(result["strs.1"]))
	utils.AssertEqual(t, "str2", tea.StringValue(result["strs.2"]))
	utils.AssertEqual(t, "123456789", tea.StringValue(result["num"]))
}

func Test_QueryWithErrorMatchesQueryForNonNilParameters(t *testing.T) {
	filter := map[string]interface{}{
		"client":  "test",
		"enabled": false,
		"number":  int64(9223372036854775807),
		"tag": map[string]interface{}{
			"key": "value",
		},
		"strs": []interface{}{"str1", "", "str3"},
		"listeners": []interface{}{
			map[string]interface{}{
				"Enabled": true,
				"Port":    443,
			},
		},
	}

	result, err := QueryWithError(filter)
	if err != nil {
		t.Fatalf("QueryWithError returned an unexpected error: %v", err)
	}
	legacyResult := Query(filter)
	if !reflect.DeepEqual(legacyResult, result) {
		t.Fatalf("QueryWithError result differs from Query: got %#v, want %#v", result, legacyResult)
	}
	utils.AssertEqual(t, "test", tea.StringValue(result["client"]))
	utils.AssertEqual(t, "false", tea.StringValue(result["enabled"]))
	utils.AssertEqual(t, "9223372036854775807", tea.StringValue(result["number"]))
	utils.AssertEqual(t, "value", tea.StringValue(result["tag.key"]))
	utils.AssertEqual(t, "str1", tea.StringValue(result["strs.1"]))
	utils.AssertEqual(t, "", tea.StringValue(result["strs.2"]))
	utils.AssertEqual(t, "str3", tea.StringValue(result["strs.3"]))
	utils.AssertEqual(t, "true", tea.StringValue(result["listeners.1.Enabled"]))
	utils.AssertEqual(t, "443", tea.StringValue(result["listeners.1.Port"]))
}

func Test_QueryWithErrorRejectsNilRepeatedParam(t *testing.T) {
	tests := []struct {
		name       string
		filter     map[string]interface{}
		wantErrMsg string
	}{
		{
			name: "top-level repeated parameter",
			filter: map[string]interface{}{
				"CaCertificateIds": []interface{}{nil, "cert-1"},
			},
			wantErrMsg: `cannot serialize repeated parameter element "CaCertificateIds.1": value is nil`,
		},
		{
			name: "nested repeated parameter",
			filter: map[string]interface{}{
				"Listeners": []interface{}{
					map[string]interface{}{
						"CertificateIds": []interface{}{"cert-1", nil},
					},
				},
			},
			wantErrMsg: `cannot serialize repeated parameter element "Listeners.1.CertificateIds.2": value is nil`,
		},
		{
			name: "typed nil string pointer",
			filter: map[string]interface{}{
				"CertificateIds": []*string{tea.String("cert-1"), nil},
			},
			wantErrMsg: `cannot serialize repeated parameter element "CertificateIds.2": value is nil`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := QueryWithError(tt.filter)
			if err == nil {
				t.Fatal("QueryWithError should reject a nil repeated parameter")
			}
			if result != nil {
				t.Fatal("QueryWithError should not return partial query parameters on error")
			}
			utils.AssertEqual(t, tt.wantErrMsg, err.Error())
		})
	}
}

func Test_QueryWithErrorAllowsUnsetParameters(t *testing.T) {
	filter := map[string]interface{}{
		"unset":    nil,
		"nilSlice": []string(nil),
		"nested": map[string]interface{}{
			"unset": nil,
			"value": "ok",
		},
	}

	result, err := QueryWithError(filter)
	if err != nil {
		t.Fatalf("QueryWithError returned an unexpected error for unset parameters: %v", err)
	}
	if _, ok := result["unset"]; ok {
		t.Fatal("unset parameter should be omitted")
	}
	if _, ok := result["nilSlice"]; ok {
		t.Fatal("nil slice parameter should be omitted")
	}
	if _, ok := result["nested.unset"]; ok {
		t.Fatal("nested unset parameter should be omitted")
	}
	utils.AssertEqual(t, "ok", tea.StringValue(result["nested.value"]))
}

func Test_QueryWithErrorPreservesLegacyMarshalFailureBehavior(t *testing.T) {
	filter := map[string]interface{}{
		"unsupported": make(chan int),
	}

	legacyResult := Query(filter)
	result, err := QueryWithError(filter)
	if err != nil {
		t.Fatalf("QueryWithError changed the legacy marshal failure into an error: %v", err)
	}
	if !reflect.DeepEqual(legacyResult, result) {
		t.Fatalf("QueryWithError result differs from Query after a marshal failure: got %#v, want %#v", result, legacyResult)
	}
}

func Test_flatRepeatedList(t *testing.T) {
	filter := map[string]interface{}{
		"client":  "test",
		"version": "1",
		"null":    nil,
		"slice": []interface{}{
			map[string]interface{}{
				"map": "valid",
			},
			6,
		},
		"map": map[string]interface{}{
			"value": "ok",
		},
	}

	result := make(map[string]*string)
	for key, value := range filter {
		filterValue := reflect.ValueOf(value)
		flatRepeatedList(filterValue, result, key)
	}
	utils.AssertEqual(t, tea.StringValue(result["slice.1.map"]), "valid")
	utils.AssertEqual(t, tea.StringValue(result["slice.2"]), "6")
	utils.AssertEqual(t, tea.StringValue(result["map.value"]), "ok")
	utils.AssertEqual(t, tea.StringValue(result["client"]), "test")
	utils.AssertEqual(t, tea.StringValue(result["slice.1.map"]), "valid")
}

func Test_GetHost(t *testing.T) {
	endpoint := GetHost(tea.String(""), tea.String(""), tea.String("client.aliyuncs.com"))
	utils.AssertEqual(t, "client.aliyuncs.com", tea.StringValue(endpoint))
}

func Test_GetOpenPlatFormEndpoint(t *testing.T) {
	endpoint := GetOpenPlatFormEndpoint(tea.String("openplatform.aliyuncs.com"), tea.String(""))
	utils.AssertEqual(t, `openplatform.aliyuncs.com`, tea.StringValue(endpoint))

	endpoint = GetOpenPlatFormEndpoint(tea.String("openplatform.aliyuncs.com"), tea.String("cn-hangzhou"))
	utils.AssertEqual(t, `openplatform.aliyuncs.com`, tea.StringValue(endpoint))

	endpoint = GetOpenPlatFormEndpoint(tea.String("openplatform.aliyuncs.com"), tea.String("ap-northeast-1"))
	utils.AssertEqual(t, `openplatform.ap-northeast-1.aliyuncs.com`, tea.StringValue(endpoint))
}
