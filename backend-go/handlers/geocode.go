package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"travel-footprints/geo"
	"travel-footprints/models"
)

// DeriveRegionCodes 从 6 位国标 adcode 截取省码（前 2 位补 0）与市码（前 4 位补 0）
func DeriveRegionCodes(adcode string) (string, string) {
	adcode = strings.TrimSpace(adcode)
	if len(adcode) != 6 {
		return "", ""
	}
	return adcode[:2] + "0000", adcode[:4] + "00"
}

type GeocodeClient struct {
	key    string
	client *http.Client
}

func NewGeocodeClient(amapKey string) *GeocodeClient {
	return &GeocodeClient{
		key:    strings.TrimSpace(amapKey),
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// ReverseGeocode 输入 WGS-84 坐标（库内标准），转 GCJ-02 后调高德 regeo。
// 任何失败都返回 nil，由调用方降级为空字段——补码失败不阻断记录足迹。
func (g *GeocodeClient) ReverseGeocode(lng, lat string) *models.RegionInfo {
	if g == nil || g.key == "" {
		return nil
	}
	var lngF, latF float64
	if _, err := fmt.Sscanf(strings.TrimSpace(lng)+" "+strings.TrimSpace(lat), "%f %f", &lngF, &latF); err != nil {
		return nil
	}
	if geo.OutOfChina(lngF, latF) {
		return nil // 境外坐标无 adcode，短路省一次无效调用（为境外扩展预留）
	}
	gLng, gLat := geo.WGS84ToGCJ02(lngF, latF)

	endpoint := "https://restapi.amap.com/v3/geocode/regeo?" + url.Values{
		"key":      {g.key},
		"location": {fmt.Sprintf("%.6f,%.6f", gLng, gLat)},
	}.Encode()

	res, err := g.client.Get(endpoint)
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	body := make([]byte, 0, 4096)
	buf := make([]byte, 1024)
	for {
		n, readErr := res.Body.Read(buf)
		body = append(body, buf[:n]...)
		if readErr != nil || n == 0 {
			break
		}
	}
	info, err := ParseRegeoPayload(body)
	if err != nil {
		return nil
	}
	return info
}

// ParseRegeoPayload 解析 regeo 响应；status 非 "1" 或结构异常视为失败
func ParseRegeoPayload(body []byte) (*models.RegionInfo, error) {
	var payload struct {
		Status    string `json:"status"`
		Info      string `json:"info"`
		Regeocode struct {
			AddressComponent struct {
				Province interface{} `json:"province"`
				City     interface{} `json:"city"`
				District interface{} `json:"district"`
				Adcode   string      `json:"adcode"`
			} `json:"addressComponent"`
		} `json:"regeocode"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if payload.Status != "1" {
		return nil, fmt.Errorf("regeo 失败: %s", payload.Info)
	}

	provinceName, _ := payload.Regeocode.AddressComponent.Province.(string)
	cityName, _ := payload.Regeocode.AddressComponent.City.(string)
	if cityName == "" {
		cityName = provinceName // 直辖市/省直辖县市的 city 可能为空数组，用省名兜底
	}
	districtName, _ := payload.Regeocode.AddressComponent.District.(string)
	if districtName == "" {
		districtName = cityName // 直辖市的 district 也可能为空数组，用市名兜底
	}
	provinceCode, cityCode := DeriveRegionCodes(payload.Regeocode.AddressComponent.Adcode)
	if provinceCode == "" {
		return nil, fmt.Errorf("regeo 无 adcode")
	}
	return &models.RegionInfo{
		ProvinceCode: provinceCode,
		CityCode:     cityCode,
		ProvinceName: provinceName,
		CityName:     cityName,
		DistrictCode: payload.Regeocode.AddressComponent.Adcode,
		DistrictName: districtName,
	}, nil
}
