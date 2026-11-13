package geo

import (
	"math"
	"testing"
)

// 参照值与前端 coords.test.js 一致，保证双语言实现相同
func TestWGS84ToGCJ02(t *testing.T) {
	lng, lat := WGS84ToGCJ02(116.404, 39.915)
	if math.Abs(lng-116.41024449916938) > 1e-9 || math.Abs(lat-39.91640428150164) > 1e-9 {
		t.Fatalf("北京参照点不符: %v %v", lng, lat)
	}
}

func TestGCJ02ToWGS84(t *testing.T) {
	lng, lat := GCJ02ToWGS84(116.404, 39.915)
	if math.Abs(lng-116.39775550083061) > 1e-9 || math.Abs(lat-39.91359571849836) > 1e-9 {
		t.Fatalf("北京参照点不符: %v %v", lng, lat)
	}
}

func TestOutOfChina(t *testing.T) {
	if OutOfChina(116.404, 39.915) {
		t.Fatal("北京应在境内")
	}
	if !OutOfChina(139.6917, 35.6895) {
		t.Fatal("东京应在境外")
	}
	lng, lat := WGS84ToGCJ02(139.6917, 35.6895)
	if lng != 139.6917 || lat != 35.6895 {
		t.Fatal("境外坐标应原样返回")
	}
}

func TestRoundTrip(t *testing.T) {
	lng, lat := 120.15507, 30.274085
	gLng, gLat := WGS84ToGCJ02(lng, lat)
	bLng, bLat := GCJ02ToWGS84(gLng, gLat)
	if math.Abs(bLng-lng) > 2e-5 || math.Abs(bLat-lat) > 2e-5 {
		t.Fatalf("往返误差过大: %v %v", bLng, bLat)
	}
}
