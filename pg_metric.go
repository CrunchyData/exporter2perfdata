package main

import (
  "bufio"
  "sort"
  "flag"
  "fmt"
  "log"
  "net/http"
  "os"
  "strings"
  "strconv"
  "go/ast"
  "go/parser"
  "go/token"
)

const PGMetricVersion = "2.0"

type nvMap map[string]string
type kvMap map[string]nvMap
var literalMap map[string]bool
var metricMap kvMap

const (
  OKY = 0
  WRN = 1
  ERR = 2
  UNK = 3
)

const (
  NOTHING = 0
  c_GT = 1
  c_LT = 2
  c_NEQ = 3
)

func populate_map(exp ast.Expr) {
  switch exp := exp.(type) {
  case *ast.Ident:
    literalMap[exp.Name]=false
  case *ast.ParenExpr:
    populate_map(exp.X)
  case *ast.BinaryExpr:
    populate_map(exp.X)
    populate_map(exp.Y)
  case *ast.BasicLit:
    return;
 }
}

func Eval(k string, exp ast.Expr) float64 {
  switch exp := exp.(type) {
  case *ast.BinaryExpr:
    return EvalBinaryExpr(k, exp)
  case *ast.ParenExpr:
    return Eval(k, exp.X)
  case *ast.BasicLit:
    switch exp.Kind {
    case token.FLOAT:
      i, _ := strconv.ParseFloat(exp.Value,64)
      return i
    case token.INT:
      i, _ := strconv.ParseFloat(exp.Value,64)
      return i
    }
  case *ast.Ident:
    i, _ := strconv.ParseFloat( metricMap[exp.Name][k], 64 )
    return i
  }
  return 0
}

func EvalBinaryExpr(k string, exp *ast.BinaryExpr) float64 {
  left := Eval(k, exp.X)
  right := Eval(k, exp.Y)
  switch exp.Op {
  case token.ADD:
    return left + right
  case token.SUB:
    return left - right
  case token.MUL:
    return left * right
  case token.QUO:
    return left / right
  default:
    return 0.0
  }
}

func LoadMetric(scanner *bufio.Scanner, actionMap map[string]bool, includeMap nvMap, excludeMap nvMap) {
  var include_matched, exclude_matched bool
  for scanner.Scan() {
    include_matched, exclude_matched = false, false
    t := scanner.Text()
    if t[0] == '#' {
      continue
    }
    var at []string
    var naction string
    kv := strings.Split(t," ")
    nmat := strings.Split(kv[0],"{")
    nm := nmat[0]
    _, ok := actionMap[nm]
    if ok {
      actionMap[nm]=true
      _, ok := metricMap[nm]
      if !ok {
        metricMap[nm] = make(nvMap)
      }
      if len(nmat) >1 {
        at = strings.Split(strings.TrimSuffix(nmat[1],"}"),",")
        var sep = ""
        naction=""
        for i := range at {
          nstr := strings.Replace(strings.Replace(at[i],"\"","",-1),"=","_",-1)
          naction = naction + sep + nstr
          sep = "_"
          _, ok = includeMap[nstr]
          if ok {
            include_matched = true
          }
          _, ok = excludeMap[nstr]
          if ok {
            exclude_matched = true
          }
        }
      }
      if len(includeMap) > 0 && include_matched == false {
        continue
      }
      if len(excludeMap) > 0 && exclude_matched == true {
        continue
      }
      if kv[1] == "NaN" {
        metricMap[nm][naction]="-1"
      } else {
        metricMap[nm][naction]=kv[1]
      }
    } else if len(metricMap) > 0 {
      allDone := true
      for _, v := range actionMap {
          if !v {
            allDone = false
            break
          }
      }
      if allDone {
        break
      }
    }
  }
}

func sortedMapKeys(m map[string]string) ([]string) {
        keys := make([]string, len(m))
        i := 0
        for k := range m {
            keys[i] = k
            i++
        }
        sort.Strings(keys)
        return keys
}

func toIntStr(fkv float64) string {
  ikv := int64(fkv)
  return fmt.Sprintf("%d",ikv)
}

func main() {
  var url, action, value string
  var compare_type, text_values int
  var warning, critical, include, exclude, expression string
  var scanner *bufio.Scanner
  var naction string
  var msg string = ""
  var cmsg string = "|"
  var status int = UNK
  var cstatus string = ""
  var fwarning,fcritical, fkv float64
  var isExpression bool = false
  var includeMap, excludeMap nvMap
  var actionMap map[string]bool
  var keys []string
  actionMap = make(map[string]bool)

  flag.StringVar(&url, "url", "", "postgres_exporter url http(s)://<ip | domain><:port>")
  flag.StringVar(&action, "action", "", "postgres_exporter Metric name")
  flag.IntVar(&text_values, "text_values",  0, "Treat values as TEXT")
  flag.IntVar(&compare_type, "compare_type",  0, "Compare Type 0=None,1=GT,2=LT,3=NEQ")
  flag.StringVar(&warning, "warning",  "", "Warning threshold")
  flag.StringVar(&critical, "critical", "", "Critical threshold")
  flag.StringVar(&include, "include", "", "<domain><value> to include")
  flag.StringVar(&exclude, "exclude", "", "<domain><value> to include")
  flag.StringVar(&expression, "expression", "", "expression for calculated values")
  version := flag.Bool("version", false, "Print version")
  flag.Parse()

  if *version {
    fmt.Println(PGMetricVersion)
    os.Exit(0)
  }

  if url == "" {
    fmt.Println("pg_metric:",PGMetricVersion)
    fmt.Println("--url is required")
    os.Exit(UNK)
  }
  if action == "" {
    fmt.Println("pg_metric:",PGMetricVersion)
    fmt.Println("--action is required")
    os.Exit(UNK)
  }
  actionMap[action]=false
  if compare_type == c_NEQ && warning == "" && critical == "" {
    fmt.Println("pg_metric:",PGMetricVersion)
    fmt.Println("--compare_type NEQ requires --warning and not --critical")
    os.Exit(UNK)
  } else if (compare_type > NOTHING && compare_type < c_NEQ && (warning == "" || critical == "")) {
    fmt.Println("pg_metric:",PGMetricVersion)
    fmt.Println("--compare_type requires --warning and --critical")
    os.Exit(UNK)
  }
  if strings.HasPrefix(url, "file") {
    inFile, err := os.Open(strings.TrimPrefix(url, "file://"))
    if err != nil {
      fmt.Println("pg_metric:",PGMetricVersion)
      log.Print(err)
      os.Exit(UNK)
    }
    defer inFile.Close()
    scanner = bufio.NewScanner(inFile)
  }  else {
    response, err := http.Get(url)
    if err != nil {
      fmt.Println("pg_metric:",PGMetricVersion)
      log.Print(err)
      os.Exit(UNK)
    }
    defer response.Body.Close()
    scanner = bufio.NewScanner(response.Body)
  }
  if include != "" {
    includeMap = make(nvMap)
    for _, k := range strings.Split(include, ",") {
      includeMap[k]=""
    }
  }
  if exclude != "" {
    excludeMap = make(nvMap)
    for _,k := range strings.Split(exclude, ",") {
      excludeMap[k]=""
    }
  }
  if compare_type != NOTHING {
    fwarning, _ = strconv.ParseFloat(warning, 32)
    fcritical, _ = strconv.ParseFloat(critical, 32)
  }

  exp, err := parser.ParseExpr(expression)
  if err == nil {
    literalMap = make(map[string]bool)
    populate_map(exp)
  }

  metricMap = make(kvMap)
  if len(literalMap) > 0 {
    isExpression=true
    LoadMetric(scanner, literalMap, includeMap, excludeMap)
  } else {
    LoadMetric(scanner, actionMap, includeMap, excludeMap)
  }
  for _, v := range metricMap {
    if status == UNK {
      status = OKY
    }
    keys = sortedMapKeys(v)
    for _, key := range keys {
      value = v[key]
      if isExpression {
        fkv = Eval(key,exp)
      } else {
        fkv, _ = strconv.ParseFloat(value,64)
      }
      switch compare_type {
        case c_GT:
          if fkv > fcritical {
            status = ERR
          } else if fkv > fwarning {
            if status == OKY {
              status = WRN
            }
          }
        case c_LT:
          if fkv < fcritical {
            status = ERR
          } else if fkv < fwarning {
            if status == OKY {
              status = WRN
            }
          }
        case c_NEQ:
          if critical != "" && value != critical {
            status = ERR
          } else if warning != "" && value != warning {
            if status == OKY {
              status = WRN
            }
          }
      }
      if len(key) > 0 {
        naction = key
      } else {
        naction = action
      }
      if text_values == 1 {
        msg = msg + naction +": "+ value + " "
        cmsg = cmsg + " "+ naction + "=" + value
      } else {
        skv := toIntStr(fkv)
        msg = msg + naction +": "+ skv + " "
        cmsg = cmsg + " "+ naction + "=" +  skv
      }
      if compare_type > NOTHING && compare_type < c_NEQ {
        cmsg = cmsg + ";" + warning + ";" + critical
      }
    }
    break
  }
  switch status {
    case OKY:
      cstatus = "OK:"
    case WRN:
      cstatus = "WARNING:"
    case ERR:
      cstatus = "CRITICAL:"
    case UNK:
      cstatus = "UNKNOWN:"
  }
  fmt.Println(action,cstatus,msg,cmsg)
  os.Exit(status)
}

