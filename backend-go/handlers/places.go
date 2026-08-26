package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type PlaceHandler struct {
	key    string
	client *http.Client
}

type PlaceSuggestion struct {
	Value   string  `json:"value"`
	Name    string  `json:"name"`
	Address string  `json:"address"`
	Lng     float64 `json:"lng"`
	Lat     float64 `json:"lat"`
}

func NewPlaceHandler(amapKey string) *PlaceHandler {
	return &PlaceHandler{
		key:    strings.TrimSpace(amapKey),
		client: &http.Client{Timeout: 8 * time.Second},
	}
}

func (h *PlaceHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeOK(w, []PlaceSuggestion{})
		return
	}
	if h.key == "" {
		writeOK(w, []PlaceSuggestion{})
		return
	}

	endpoint := "https://restapi.amap.com/v3/place/text?" + url.Values{
		"key":        {h.key},
		"keywords":   {q},
		"offset":     {"12"},
		"page":       {"1"},
		"extensions": {"base"},
	}.Encode()

	res, err := h.client.Get(endpoint)
	if err != nil {
		writeError(w, http.StatusBadGateway, "地点搜索服务暂时不可用")
		return
	}
	defer res.Body.Close()

	var payload struct {
		Status string `json:"status"`
		Info   string `json:"info"`
		Pois   []struct {
			Name     string      `json:"name"`
			Address  interface{} `json:"address"`
			Location interface{} `json:"location"`
			Pname    string      `json:"pname"`
			Cityname string      `json:"cityname"`
			Adname   string      `json:"adname"`
		} `json:"pois"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadGateway, "地点搜索结果解析失败")
		return
	}
	if payload.Status != "1" {
		writeOK(w, []PlaceSuggestion{})
		return
	}

	list := make([]PlaceSuggestion, 0, len(payload.Pois))
	for _, poi := range payload.Pois {
		lng, lat, ok := parseAmapLocation(poi.Location)
		if !ok {
			continue
		}
		address := joinAmapAddress(poi.Pname, poi.Cityname, poi.Adname, stringifyAmap(poi.Address))
		name := poi.Name
		value := name
		if address != "" {
			value = fmt.Sprintf("%s · %s", name, address)
		}
		list = append(list, PlaceSuggestion{
			Value:   value,
			Name:    name,
			Address: address,
			Lng:     lng,
			Lat:     lat,
		})
	}
	writeOK(w, list)
}

func stringifyAmap(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		return ""
	}
}

func parseAmapLocation(v interface{}) (float64, float64, bool) {
	s := stringifyAmap(v)
	if s == "" {
		return 0, 0, false
	}
	parts := strings.Split(s, ",")
	if len(parts) < 2 {
		return 0, 0, false
	}
	var lng, lat float64
	if _, err := fmt.Sscanf(strings.TrimSpace(parts[0])+" "+strings.TrimSpace(parts[1]), "%f %f", &lng, &lat); err != nil {
		return 0, 0, false
	}
	return lng, lat, true
}

func joinAmapAddress(parts ...string) string {
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return strings.Join(out, "")
}
