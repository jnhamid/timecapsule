package main

import (
    "fmt"
    "log"
    "net/http"
    "html/template"
    "strings"
    "strconv"
	"context"
	"encoding/json"
  	"net"
  	"time"
  	"encoding/gob"
  	"flag"
)

type Block struct {
	Height int `json:"Height"`
	Parenthash string `json:"Parenthash"`
	Roothash string `json:"Roothash"`
	Data []string `json:Data`
	Value string `json:Value`
}

type Blockchain struct {
	Blocks []Block
}

type FeMessage struct {
	Value string
}

func getDoc() string {
	return `<!DOCTYPE html>
	<html>
	<head>
	  <meta charset="UTF-8" />
	  <link rel="stylesheet" type="text/css" href="css/style.css" />
	</head>
	<body>
	<div class="contianer">
		<div class="center">
		<h1 class="center">Jaden&apos;s Time Capsule </h1>
		<form method="POST" action="/">
			<div class="center">
				<label>Data</label>
				<input name="data" type="text" value="" />
				<button type="sumbit">Enter</button>
			</div>
		</form>
	  <div/>
	  <div class="flex center">
	  	{{range .Blocks}}

	  	<div class="flex-item">
	  		<h5>Height:  {{.Height}} </h5>
	  		<h5>value:  {{.Value}} </h5>
	  		<p> Parenthash:  {{.Parenthash}} </p>
	  		<p> Roothash:  {{.Roothash}} </p>
	  	</div>
	  	{{end}}

	  </div>
	</div>
	</body>
	</html>`
}

func listen(port int, Blocks* []Block) {
	fmt.Println("listening")
	portStr := ":" + strconv.Itoa(port)
	l, err := net.Listen("tcp", portStr)
	if err != nil {
		fmt.Println("error:", err)
	}
	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("error:", err)
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(c net.Conn) {
			dec := gob.NewDecoder(c)
			var block Block
			var val []byte
			dec.Decode(&val)
			err := json.Unmarshal(val, &block)
			if err != nil {
				log.Fatal(err)
			}
			(*Blocks) = append(*Blocks, block)
			defer c.Close()
			return
		}(conn)

	}
	return
}

func dial(port int, value FeMessage){

	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	reqStr := "localhost:" + strconv.Itoa(port)

	conn, err := d.DialContext(ctx, "tcp", reqStr)
	if err != nil {
		log.Printf("Failed to dial: %v redailing in 2 Seconds", err)
	}

	defer conn.Close()
	enc := gob.NewEncoder(conn)
 	enc.Encode(value)
}

func main() {
	var n int;
	flag.IntVar(&n, "n", 5, "total number of process")
	flag.Parse()


	b := make([]Block, 0)

	blockchain := Blockchain{b}

	go listen(8100, &blockchain.Blocks)

	fmt.Printf("%+v\n", blockchain)


	doc := getDoc()
    http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
        w.Header().Add("Content Type", "text/html")
        // The template name "template" does not matter here
        templates := template.New("template")
        // "doc" is the constant that holds the HTML content
        templates.New("doc").Parse(doc)
        templates.Lookup("doc").Execute(w, blockchain)
        value := ""
        req.ParseForm()
        fmt.Printf("%+v\n", req.Form)
        if(!(req.Form == nil) || !(len(req.Form) == 0)){
        	for _, v := range req.Form {
       			value += strings.Join(v, "")
    		}
    		fmt.Println("value: ", value)
        }

        if(value != ""){
        	msg := FeMessage{value}
        	for i := 0; i < n; i++{
        		fmt.Println("dialing")
        		dial(3000 + (i * 20), msg)
        	}
        }

    })


    http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
    fmt.Printf("Starting server at port 8080\n")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
