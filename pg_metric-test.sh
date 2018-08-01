#!/bin/bash
TESTTYPE=$1
TESTNUM=$2
REPORT_FILE=$3
TEMP_FILE=$(basename $REPORT_FILE .txt).tmp
EX=0

if [ $TESTTYPE = "NODE" ]; then
  case  $TESTNUM in
    1)
      python convert-icinga2-pg_metric.py node-services.conf ob1papropgs03_primary_node.txt >cmd.sh
      ;;
    *)
      exit 2
      ;;
  esac
fi

if [ $TESTTYPE = "PG" ]; then
  case  $TESTNUM in
    1)
      python convert-icinga2-pg_metric.py postgresql-services.conf ob1papropgs03_primary_pg.txt >cmd.sh
      ;;
    *)
      exit 2
      ;;
  esac
fi

bash cmd.sh >$TEMP_FILE
diff -Nau $REPORT_FILE $TEMP_FILE &>/dev/null
EX=$?
if [ $EX -eq 0 ]; then
  echo "Test Passed"
else
  echo "Test Failed"
fi
exit $EX
