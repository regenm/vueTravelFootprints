package database

import (
	"testing"
	"travel-footprints/models"
)

func newTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("打开内存库失败: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestTripCRUD(t *testing.T) {
	db := newTestDB(t)

	trip := models.NewTrip("u1", models.CreateTripRequest{Name: "春节云南行", StartDate: "2026-02-10", EndDate: "2026-02-17", Notes: "一家三口"})
	if err := db.CreateTrip(trip); err != nil {
		t.Fatalf("创建失败: %v", err)
	}

	list, err := db.GetTripsByUser("u1")
	if err != nil || len(list) != 1 {
		t.Fatalf("查询失败: %v len=%d", err, len(list))
	}
	if list[0].Name != "春节云南行" || list[0].StartDate != "2026-02-10" {
		t.Fatalf("字段不符: %+v", list[0])
	}

	name := "春节滇西北行"
	notes := "改了大理行程"
	updated, err := db.UpdateTrip(list[0].ID, "u1", models.UpdateTripRequest{Name: &name, Notes: &notes})
	if err != nil || updated.Name != "春节滇西北行" {
		t.Fatalf("更新失败: %v %+v", err, updated)
	}

	// 建一个关联足迹，删除行程后 trip_id 应被清空
	marker := models.NewMarker("u1", models.CreateMarkerRequest{Name: "洱海", Longitude: "100.18", Latitude: "25.60", TripID: trip.ID})
	if err := db.CreateMarker(marker); err != nil {
		t.Fatalf("创建足迹失败: %v", err)
	}
	if err := db.DeleteTrip(trip.ID, "u1"); err != nil {
		t.Fatalf("删除失败: %v", err)
	}
	got, _ := db.GetMarkerByID(marker.ID)
	if got.TripID != "" {
		t.Fatalf("删除行程后足迹 trip_id 应为空: %q", got.TripID)
	}

	// 他人不可删
	other := models.NewTrip("u2", models.CreateTripRequest{Name: "别人的"})
	_ = db.CreateTrip(other)
	if err := db.DeleteTrip(other.ID, "u1"); err == nil {
		t.Fatal("非所有者删除应失败")
	}
}
