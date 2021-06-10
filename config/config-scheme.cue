LogLevel: "debug" | *"info" | "warn" | "error" | "critical"
Port:     int16 | *3000

DB: {
	Driver:   string | *"mysql" | "postgres" | "sqlite3"
	Host:     string
	Port:     uint16 & *3306
	Username: string
	Password: string
	DBName:   string
}
