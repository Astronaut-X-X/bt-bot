package model

type Permissions struct {
	UUID                  string `gorm:"column:uuid;type:varchar(255);primaryKey"`
	Type                  string `gorm:"column:type;type:varchar(255)"`
	AsyncDownloadQuantity int    `gorm:"column:async_download_quantity;type:int"`
	DailyDownloadQuantity int    `gorm:"column:daily_download_quantity;type:int"`
	DailyDownloadRemain   int    `gorm:"column:daily_download_remain;type:int"`
	DailyDownloadDate     int64  `gorm:"column:daily_download_date;type:int64"`
	FileDownloadSize      int64  `gorm:"column:file_download_size;type:int64"`
}

const (
	PermissionsTypeBasic   = "basic"
	PermissionsTypePremium = "premium"
)

var BasicPermissions = Permissions{
	UUID:                  "",
	Type:                  PermissionsTypeBasic,
	AsyncDownloadQuantity: 1,
	DailyDownloadQuantity: 5,
	DailyDownloadRemain:   5,
	DailyDownloadDate:     1770881636,
	FileDownloadSize:      1024 * 1024 * 1024 * 2,
}

var PremiumPermissions = Permissions{
	UUID:                  "",
	Type:                  PermissionsTypePremium,
	AsyncDownloadQuantity: 3,
	DailyDownloadQuantity: 100,
	DailyDownloadRemain:   100,
	DailyDownloadDate:     1770881636,
	FileDownloadSize:      1024 * 1024 * 1024 * 10,
}
