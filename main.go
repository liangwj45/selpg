package main

import (
  "bufio"
  "fmt"
  "io"
  "os"
  "os/exec"
  "errors"

  flag "github.com/spf13/pflag"
)

var (
  startPage = flag.IntP("start-page", "s", -1, "start page")
  endPage = flag.IntP("end-page", "e", -1, "end page")
  pageLen = flag.IntP("page-len", "l", 72, "number of lines in a page")
  pageBreak = flag.BoolP("use-page-break", "f", false, "pages devided by page break")
  printDest = flag.StringP("print-dest", "d", "", "printing destination")
  pageType string
)

func main() {
  handleError(parseArgs())
  handleError(readAndPrint())
}

func parseArgs() error {
  flag.Parse()  // 使用pflag提供的方法解析命令行参数

  if *pageBreak {  // 判断文件分页格式
    pageType = "f"
  } else {
    pageType = "l"
  }

  if *startPage == -1 || *endPage == -1 {  // 判断参数是否完整（起始页和结束页）
    printUsage()
    return errors.New("arguments are not enough")
  }

  if *startPage <= 0 || *endPage <= 0 {  // 判断页码是否合理
    return errors.New("page number can not be negative")
  }

  if *startPage > *endPage {  // 判断页码是否合理
    return errors.New("start page cannot be greater than end page")
  }

  if pageType == "l" && *pageLen <= 0 {  // 判断页长是否合理
    return errors.New("line number can not be negative")
  }

  if pageType == "f" && *pageLen != 72 {  // 判断是否同时设置两种分页格式
    return errors.New("-f and -lNumber cannot be set at the same time")
  }

  return nil
}

func readAndPrint() error {
  var reader *bufio.Reader  // 定义一个输入流
  var writer *bufio.Writer  // 定义一个输出流

  if flag.NArg() == 0 {  // 根据[FILE]参数给输入流赋值
    reader = bufio.NewReader(os.Stdin)
  } else {
    input, err := os.Open(flag.Arg(0))
    if (err != nil) {
      return err
    }
    defer input.Close()
    reader = bufio.NewReader(input)
  }
  
  if len(*printDest) == 0 {  // 根据-d参数给输出流赋值
    writer = bufio.NewWriter(os.Stdout)
  } else {
    cmd := exec.Command("lp", "-d"+*printDest)
    output, err := cmd.StdinPipe()
    if err != nil {
      return err
    }
    defer output.Close()
    cmd.Stdout = os.Stdout
    if err := cmd.Start(); err != nil {
      return err
    }
    writer = bufio.NewWriter(output)
  }
  defer writer.Flush()

  var pageSpliter byte  // 获取分页格式标识符
  if pageType == "f" {
    pageSpliter = '\f'
  } else {
    pageSpliter = '\n'
  }

  pages, lines := 1, 0
  for {
    sub, err := reader.ReadBytes(pageSpliter)  // 根据标识符读取一段内容
    if err == io.EOF {
      break
    } else if err != nil {
      return err
    }
    if pageType == "f" {  // 进行行/页相应处理
      pages++
    } else {
      lines++
      if lines > *pageLen {
        lines = 1
        pages++
      }
    }
    if pages >= *startPage && pages <= *endPage {
      if _, err := writer.Write(sub); err != nil {
        return err
      }
    } else {
      break
    }
  }

  return nil
}

func printUsage() {  // 打印使用方法
  fmt.Fprintln(os.Stderr, "Usage: selpg -sNumber -eNumber [-lNumber/-f] [-dDestination] [filename]")
  flag.PrintDefaults()
  os.Exit(2)
}

func handleError(err error) {  // 错误处理
  if err != nil {
    if _, err2 := fmt.Fprintf(os.Stderr, "%s\n", err.Error()); err2 != nil {
      panic(err2)
    }
    os.Exit(1)
  }
}