// #Middleware: {
//   Constructor: {
//     Name: string
//     Args: [..._]
//   }
// }

Port: int16 | *3000
Backend: string & =~ "^https?://[a-z0-9_\\.-]+(:[0-9]+)?$"

Memcached: [
  string & =~ "^[a-z0-9_\\.-]+:[0-9]+$"
] | *[]

DumpDir?: string
