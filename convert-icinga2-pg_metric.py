import sys
import re
from collections import OrderedDict

with open(sys.argv[1], 'r') as f:
    lines = f.readlines()
    host=OrderedDict()
    add_attributes = False
    p= re.compile('^apply[]\w\s]+"(.+)"')
    for line in lines:
        if '}' in line:
            add_attributes = False
        if  add_attributes and '=' in line:
            host[name][line.split('=')[0]]=line.split('=')[1]
        if p.search(line):
            name=p.search(line).group(1)
            add_attributes = True
            host[name]=OrderedDict()
    for k, v in host.iteritems():
        cmd=""
        for k1, v1 in v.iteritems():
            k1strip = k1.strip()
            v1strip = v1.strip()
            if v1strip.startswith("host."):
                continue
            if k1strip == "vars.pg_action":
                cmd+=" --action=%s" % v1strip
            elif k1strip == "vars.pg_compare":
                cmd+=" --compare_type=" + v1strip
            elif k1strip == "vars.pg_critical":
                cmd+=" --critical=" + v1strip
            elif k1strip == "vars.pg_warning":
                cmd+=" --warning=" + v1strip
            elif k1strip == "vars.pg_include":
                cmd+=" --include=%s" % v1strip
            elif k1strip == "vars.pg_exclude":
                cmd+=" --exclude=%s" % v1strip
            elif k1strip == "vars.pg_expression":
                cmd+=" --expression=%s" % v1strip
            elif k1strip == "vars.pg_text":
                cmd+=" --text_values=" + v1strip
        print "./pg_metric --url=file://%s" % sys.argv[2], cmd
