#!/bin/bash
TESTTYPE=$1
TESTNUM=$2
CONF_FILE=${TESTTYPE}.conf
DATA_FILE=${TESTTYPE}_${TESTNUM}_REPORT.dat
RESULT_FILE=${TESTTYPE}_${TESTNUM}_REPORT.result
TEMP_FILE=${TESTTYPE}_${TESTNUM}_REPORT.tmp
DBG_FILE=debug.tmp
CMD_FILE=cmd.sh

EX=0

python convert-icinga2-pg_metric.py ${CONF_FILE} ${DATA_FILE} 2>${CMD_FILE} >${DBG_FILE}

bash ${CMD_FILE} >${TEMP_FILE}
diff -Nau ${RESULT_FILE} ${TEMP_FILE} &>>${DBG_FILE}
EX=$?
if [ $EX -eq 0 ]; then
  echo "Test Passed ($(cat ${CMD_FILE} | wc -l))"
else
  echo "Test Failed"
fi
exit $EX
