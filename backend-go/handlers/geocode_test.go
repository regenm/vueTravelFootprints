package handlers

import "testing"

func TestDeriveRegionCodes(t *testing.T) {
	cases := []struct {
		adcode, wantProvince, wantCity string
	}{
		{"110105", "110000", "110100"}, // 北京朝阳区
		{"410102", "410000", "410100"}, // 郑州中原区
		{"419001", "410000", "419000"}, // 济源市（省直辖县级市）
		{"810002", "810000", "810000"}, // 香港湾仔区
		{"330106", "330000", "330100"}, // 杭州西湖区
	}
	for _, c := range cases {
		p, ci := DeriveRegionCodes(c.adcode)
		if p != c.wantProvince || ci != c.wantCity {
			t.Fatalf("adcode %s: got %s/%s want %s/%s", c.adcode, p, ci, c.wantProvince, c.wantCity)
		}
	}
	if p, _ := DeriveRegionCodes(""); p != "" {
		t.Fatal("空 adcode 应返回空码")
	}
}

func TestParseRegeoPayload(t *testing.T) {
	body := []byte(`{"status":"1","regeocode":{"formattedAddress":"浙江省杭州市西湖区","addressComponent":{"province":"浙江省","city":"杭州市","district":"西湖区","adcode":"330106"}}}`)
	info, err := ParseRegeoPayload(body)
	if err != nil {
		t.Fatal(err)
	}
	if info.ProvinceCode != "330000" || info.CityCode != "330100" {
		t.Fatalf("省市码不符: %+v", info)
	}
	if info.ProvinceName != "浙江省" || info.CityName != "杭州市" {
		t.Fatalf("省市名不符: %+v", info)
	}
	if info.DistrictCode != "330106" || info.DistrictName != "西湖区" {
		t.Fatalf("区县码/名不符: %+v", info)
	}
}

func TestParseRegeoPayloadCityArray(t *testing.T) {
	// 直辖市以外的省直辖情况 city 字段可能返回空数组
	body := []byte(`{"status":"1","regeocode":{"addressComponent":{"province":"河南省","city":[],"district":"济源市","adcode":"419001"}}}`)
	info, err := ParseRegeoPayload(body)
	if err != nil {
		t.Fatal(err)
	}
	if info.CityName != "河南省" {
		t.Fatalf("city 为空数组时应用省名兜底: %+v", info)
	}
	if info.DistrictCode != "419001" || info.DistrictName != "济源市" {
		t.Fatalf("省直辖县市的区县码应为其自身: %+v", info)
	}
}

func TestParseRegeoPayloadFailure(t *testing.T) {
	if _, err := ParseRegeoPayload([]byte(`{"status":"0","info":"INVALID_USER_KEY"}`)); err == nil {
		t.Fatal("status 非 1 应返回错误")
	}
	if _, err := ParseRegeoPayload([]byte(`not json`)); err == nil {
		t.Fatal("坏 JSON 应返回错误")
	}
}
