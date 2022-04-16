#  Service Monitor - Ping

Monitor the availability of one or more servers by pinging them on a set interval and storing the results.
Attempts to provide both location and local time of the server based on the IP.

## Usage

### Supply ips on the command line
You can supply one or more addresses via the command line:
```bash
./service-monitor-ping 123.123.123.123
```

### Supply IPs via a txt file
Read input IPs from a file, where each ip is on it's own line.
```bash
INPUT_FILE=./ips.txt ./service-monitor-ping
```

## Configuration
|Name|Default|Description|
|---|---|---|
|PING_INTERVAL|30|Number of seconds between each ping to a host|
|INPUT_FILE|""|Input file of IPs to be monitored|
|GEOMIND_DATABASE|""|GeoMind database to use for location lookups|
