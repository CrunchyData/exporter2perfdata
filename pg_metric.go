package main

import (
  "bufio"
  "flag"
  "fmt"
  "log"
  "net/http"
  "os"
  "strings"
  "strconv"
)

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

func main() {
  var url, action string
  var compare_type, text_values int
  var warning, critical, include, exclude string
  var scanner *bufio.Scanner

  flag.StringVar(&url, "url", "", "postgres_exporter url http(s)://<ip | domain><:port>")
  flag.StringVar(&action, "action", "", "postgres_exporter Metric name")
  flag.IntVar(&text_values, "text_values",  0, "Treat values as TEXT")
  flag.IntVar(&compare_type, "compare_type",  0, "Compare Type 0=None,1=GT,2=LT,3=NEQ")
  flag.StringVar(&warning, "warning",  "", "Warning threshold")
  flag.StringVar(&critical, "critical", "", "Critical threshold")
  flag.StringVar(&include, "include", "", "<domain><value> to include")
  flag.StringVar(&exclude, "exclude", "", "<domain><value> to include")
  flag.Parse()

  if url == "" {
    fmt.Println("--url is required")
    os.Exit(UNK)
  }
  if action == "" {
    fmt.Println("--action is required")
    os.Exit(UNK)
  }
  if compare_type == c_NEQ && warning == "" && critical == "" {
    fmt.Println("--compare_type NEQ requires --warning and not --critical")
    os.Exit(UNK)
  } else if (compare_type > NOTHING && compare_type < c_NEQ && (warning == "" || critical == "")) {
    fmt.Println("--compare_type requires --warning and --critical")
    os.Exit(UNK)
  }
  if strings.HasPrefix(url, "file") {
    inFile, err := os.Open(strings.TrimPrefix(url, "file://"))
    if err != nil {
      log.Print(err)
      os.Exit(UNK)
    }
    defer inFile.Close()
    scanner = bufio.NewScanner(inFile)
  }  else {
    response, err := http.Get(url)
    if err != nil {
      log.Print(err)
      os.Exit(UNK)
    }
    defer response.Body.Close()
    scanner = bufio.NewScanner(response.Body)
  }
  var msg string = ""
  var cmsg string = ""
  var status int = UNK
  var cstatus string = ""
  var fwarning,fcritical, fkv float64
  var ikv int64
  var matched bool = false
  var vinclude, vexclude []string
  if include != "" {
    vinclude = strings.Split(include, ",")
  }
  if exclude != "" {
    vexclude = strings.Split(exclude, ",")
  }
  if compare_type != NOTHING {
    fwarning, _ = strconv.ParseFloat(warning, 32)
    fcritical, _ = strconv.ParseFloat(critical, 32)
  }
    cmsg = "|"
  for scanner.Scan() {
    t := scanner.Text()
    var include_matched, exclude_matched bool = false, false
    if t[0] != '#' {
      var at []string
      var naction string
      kv := strings.Split(t," ")
      nmat := strings.Split(kv[0],"{")
      nm := nmat[0]
      if nm == action {
        matched = true
        if status == UNK {
          status = OKY
        }
        if kv[1] == "NaN" {
          kv[1]="-1"
        }
        fkv, _ = strconv.ParseFloat(kv[1], 64)
        ikv = int64(fkv)
        skv := fmt.Sprintf("%d",ikv)
        naction = nm
        if len(nmat) >1 {
          at = strings.Split(strings.TrimSuffix(nmat[1],"}"),",")
          var sep = ""
          naction=""
          for i := range at {
            nstr := strings.Replace(strings.Replace(at[i],"\"","",-1),"=","_",-1)
            naction = naction + sep + nstr
            sep = "_"
            if include != "" {
              for j := range vinclude {
                if nstr == vinclude[j] {
                  include_matched = true
                }
              }
            }
            if exclude != "" {
              for j := range vexclude {
                if nstr == vexclude[j] {
                  exclude_matched = true
                }
              }
            }
          }
        }
        if include != "" && include_matched == false {
          //fmt.Println("CONTINUE on Including.........")
          continue
        }
        if exclude != "" && exclude_matched == true {
          //fmt.Println("CONTINUE on Excluding.........")
          continue
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
            if critical != "" && kv[1] != critical {
              status = ERR
            } else if warning != "" && kv[1] != warning {
              if status == OKY {
                status = WRN
              }
            }
        }
        if text_values == 1 {
          msg = msg + naction +": "+ kv[1] + " "
          cmsg = cmsg + " "+ naction + "=" +  kv[1]
        } else {
          msg = msg + naction +": "+ skv + " "
          cmsg = cmsg + " "+ naction + "=" +  skv
        }
        if compare_type > NOTHING && compare_type < c_NEQ {
          cmsg = cmsg + ";" + warning + ";" + critical
        }
      } else if matched == true {
        break
      }
    }
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

