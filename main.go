package main

import (
    "bufio"
    "fmt"
    "golang.org/x/net/html"
    "log"
    "net/http"
    "os"
    "strings"
)

type link struct {
    url        string
    statusCode int
    attempt int
}

func main() {
    for _, url := range os.Args[1:] {

        links, err := parseLinks(url)
        if err != nil {
            fmt.Println("Не удалось получить информацию, ошибка:", err)
            os.Exit(1)
        }

        sortLinks(&links, url)
        write(links)

        c := make(chan link)
        count := len(links)
        for _, url := range links {
            l := link{url, 0, 1}
            go getStatusCode(l, c)
        }

        answers := 1
        for s := range c {
            fmt.Println("Link:", answers, "Код", s.statusCode, s.url)
            answers++
            if answers > count {
                close(c)
            }
        }

        fmt.Println("Найдено ссылкок:", len(links))
    }
}

func parseLinks(url string) ([]string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("getting %s: %s", url, resp.Status)
    }

    doc, err := html.Parse(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("parsing %s as HTML: %v", url, err)
    }

    return visit(nil, doc), nil
}

func visit(links []string, n *html.Node) []string {
    if n.Type == html.ElementNode && n.Data == "a" {
        for _, a := range n.Attr {
            if a.Key == "href" {
                links = append(links, a.Val)
            }
        }
    }

    for c := n.FirstChild; c != nil; c = c.NextSibling {
        links = visit(links, c)
    }

    return links
}

func sortLinks(links *[]string, url string) {
    var tmp []string
    for _, link := range *links {

        // todo подумать над регуляркой
        if strings.Contains(link, "#") {
            link = strings.Split(link, "#")[0]
        }
        if strings.Contains(link, "?") {
            link = strings.Split(link, "?")[0]
        }

        if len(link) > 1 && string(link[0]) != "/" {
            continue
        }

        if !strings.Contains(link, "https://") || !strings.Contains(link, "http://") {
            link = url + link
        }

        tmp = append(tmp, link)
    }

    // delete duplicates
    keys := make(map[string]bool)
    var list []string
    for _, entry := range tmp {
        if _, value := keys[entry]; !value {
            keys[entry] = true
            list = append(list, entry)
        }
    }

    *links = list
}

func write(links []string) {
    file, err := os.OpenFile("result.txt", os.O_CREATE|os.O_WRONLY, 0644)

    if err != nil {
        log.Fatalf("failed creating file: %s", err)
    }

    datawriter := bufio.NewWriter(file)

    for _, data := range links {
        _, _ = datawriter.WriteString(data + "\n")
    }

    datawriter.Flush()
    file.Close()
}

func getStatusCode(link link, c chan link) {
    resp, err := http.Get(link.url)
    if err != nil {
        fmt.Println(err)
        link.statusCode = 000
        c <- link
        return
    }
    defer resp.Body.Close()

    link.statusCode = resp.StatusCode
    c <- link
}
