# exporter2perfdata

exporter2perfdata collects metric from the Prometheus exporters using RESTAPI and converts it to perfdata format for Icinga 2 / Nagios. This module helps Icinga 2 / Nagios to consume metric from Prometheus exporters.

## How to compile

Incremental make
```
make
```

or

Clean and make
```
make clean all
```


## How to test

```
make test
```

Install
```
sudo cp exporter2perfdata /usr/lib64/nagios/plugins/
sudo chmod 755 /usr/lib64/nagios/plugins/exporter2perfdata
```

## Usage

```
./exporter2perfdata -h

exporter2perfdata: 2.0

Usage of exporter2perfdata:
  -action string
    	exporter Metric name [REQUIRED]
  -compare-type int
    	Compare Type 0=None,1=GT,2=LT,3=NEQ
  -critical string
    	Critical threshold
  -exclude string
    	<domain><value> to include
  -expression string
    	expression for calculated values
  -force-ok
    	Force UNKNOWN to return OK status
  -include string
    	<domain><value> to include
  -text-values int
    	Treat values as TEXT
  -url string
    	exporter url http(s)://<ip | domain><:port> [REQUIRED]
  -version
    	Print version
  -warning string
    	Warning threshold
```

### Simple capture the metric
```
/exporter2perfdata --url=<exporter url > --action="<metric name>"
```

### Capture metric and validate results

Metric Value > 3600 will cause a CRITICAL Alert
Metric Value > 1800 will cause a WARNING Alert

```
./exporter2perfdata --url=<exporter url> --action="<metric name>" --compare-type="1" --warning="1800" --critical="3600"
```

Metric Value < 600 will cause a CRITICAL Alert
Metric Value < 800 will cause a WARNING Alert

```
./exporter2perfdata --url=<exporter url> --action="<metric name>" --compare-type="2" --warning="800" --critical="600"
```

## Configuration

Configuration file `exporter2perfdata.conf` needs to be copied to `plugins-contrib.d/` folder.

### Icigna2 config

```
cp exporter2perfdata.conf to /usr/share/icinga2/include/plugins-contrib.d/
```

# Icinga2 metric and alert configuration

```
apply Service "ccp_pg_ready" {
  import "generic-service"

  check_command    = "exporter2perfdata"
  vars.pg_url      = host.vars.pg_url_node

  vars.pg_action   = "ccp_pg_ready"
  /* Checks to see if the given system is NOT (compare = 3) a primary (return value 2) by default.
   * Set on a per-host basis to be able to check whether a system is no longer a replica (return value 1)
   */
  vars.pg_compare  = "3"
  vars.pg_critical  = host.vars.ccp_pg_ready_critical
  if (host.vars.ccp_pg_ready_critical == "") {
          vars.pg_critical = "2"
  }

  assign where host.vars.node_common == "true"
}
```
