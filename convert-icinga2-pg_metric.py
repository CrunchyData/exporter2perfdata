import sys
import re
from collections import OrderedDict

with open(sys.argv[1], 'r') as f:
    lines = f.readlines()
    host=OrderedDict()
    add_attributes = 0
    p= re.compile('^apply[]\w\s]+"(.+)"')
    skiplines=False
    for l in lines:
        line = l.strip()
        if len(line) == 0 or line.startswith('//'):
            continue
        elif line.startswith('/*') and line.endswith('*/') :
            continue
        elif line.startswith('/*'):
            skiplines=True
            continue
        elif line.endswith('*/'):
            skiplines=False
            continue
        elif skiplines:
            continue
        if '}' in line:
            add_attributes -= 1
        if  add_attributes > 0 and '=' in line and not 'host.vars' in line:
            linesplit=line.split('=')
            host[name][linesplit[0].strip()]=linesplit[1].strip()
        if p.search(line):
            name=p.search(line).group(1)
            add_attributes += 1
            host[name]=OrderedDict()
        elif "{" in line:
            add_attributes += 1
    print host
    for k, v in host.iteritems():
        cmd=''
        for k1, v1 in v.iteritems():
            print 1, k1, v1
            if k1 == 'vars.pg_action':
                cmd+=' --action=%s' % v1
            elif k1 == 'vars.pg_compare':
                cmd+=' --compare_type=' + v1
            elif k1 == 'vars.pg_critical':
                cmd+=' --critical=' + v1
            elif k1 == 'vars.pg_warning':
                cmd+=' --warning=' + v1
            elif k1 == 'vars.pg_include':
                cmd+=' --include=%s' % v1
            elif k1 == 'vars.pg_exclude':
                cmd+=' --exclude=%s' % v1
            elif k1 == 'vars.pg_expression':
                cmd+=' --expression=%s' % v1
            elif k1 == 'vars.pg_text':
                cmd+=' --text_values=' + v1
        print >> sys.stderr, './pg_metric --url=file://%s' % sys.argv[2], cmd
