package database

// SnapshotTo 在线生成当前数据库的一致性快照文件，不阻塞其他读写
func (db *DB) SnapshotTo(path string) error {
	_, err := db.conn.Exec("VACUUM INTO ?", path)
	return err
}
