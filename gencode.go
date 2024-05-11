//go:build ignore

package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "text/template"
    "path"

    "github.com/stoewer/go-strcase"

    "gopkg.in/yaml.v3"
)

func main() {
    funcmap := template.FuncMap{
        "title": strcase.UpperCamelCase,
    }
    data, err := ioutil.ReadFile(os.Args[2])
    if err != nil {
        panic(err)
    }
    var schemas []map[string]any
    if err = yaml.Unmarshal(data, &schemas); err != nil {
        panic(err)
    }
    for _, a := range schemas {
        t := template.New(path.Base(os.Args[1])).Funcs(funcmap)
        t = template.Must(t.ParseFiles(os.Args[1]))
        var filepath string
        if tmp, ok := a["mkdir"].(bool); ok && tmp {
            os.Mkdir(path.Join(os.Args[3], a["name"].(string)), os.ModePerm)
            filepath = path.Join(os.Args[3], a["name"].(string), fmt.Sprintf("%s_gen.go", a["name"]))
        } else {
            filepath = path.Join(os.Args[3], fmt.Sprintf("%s_gen.go", a["name"]))
        }
        f, err := os.Create(filepath)
        if err != nil {
            panic(err)
        }
        if err = t.Execute(f, a); err != nil {
            panic(err)
        }
        if err = f.Close(); err != nil {
            panic(err)
        }
    }
}
