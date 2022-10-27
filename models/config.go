package models

type Config struct {
	Listen        string `json:"listen"`
	MysqlName     string `json:"mysql_name"`
	MysqlPassword string `json:"mysql_password"`
	MysqlPort     string `json:"mysql_port"`
	DBName        string `json:"db_name"`
}
