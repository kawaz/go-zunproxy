#Middleware: {
  Constructor: {
    Name: string
    Args: [..._]
  }
}


Port: int | *3000
Addr: string | *(":" Port)
Backend: string & =~ "^https?://[a-z0-9\\.\\-](:[0-9]+)$"
Memcached: [
  string & =~ "^[a-z0-9¥._-]+:[0-9]+$"
]