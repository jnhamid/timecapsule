package main

import (
	"log"
	"net"
	"fmt"
	"strconv"
	"context"
	"time"
	"sync"
	"flag"
	"crypto/rand"
  	"crypto/dsa"
  	"os"
  	"encoding/gob"
  	"math/big"
  	"crypto/sha512"
  	"encoding/hex"
  	"encoding/json"
)

//Pem, pkcs, p12

type Message struct {
	Value string //marshalled string of Block 
	HashedBlock string //Value hashed
	S  *big.Int // Sig 
	R *big.Int // Sig 
	PublicKey string // Marsalled publicKey 
}

type Block struct {
	Height int `json:"Height"`
	Parenthash string `json:"Parenthash"`
	Roothash string `json:"Roothash"`
	Data []string `json:Data`
	Value string `json:Value`
}

type FeMessage struct {
	Value string
}

func getBlockFromMessage(message Message) Block {
	blockStr := message.Value

	var block Block
	err := json.Unmarshal([]byte(blockStr), &block)
	if err != nil {
		fmt.Println("error:", err)
	}

	return block
}

func getPublicKeyFromMessage(message Message) dsa.PublicKey {
	var pubKey dsa.PublicKey
	err := json.Unmarshal([]byte(message.PublicKey), &pubKey)
	if err != nil {
		fmt.Println("error:", err)
	}
	return pubKey
}

func generateKey(privatekey* dsa.PrivateKey) {
	params := new(dsa.Parameters)
   // see http://golang.org/pkg/crypto/dsa/#ParameterSizes
   if err := dsa.GenerateParameters(params, rand.Reader, dsa.L1024N160); err != nil {
      fmt.Println(err)
      os.Exit(1)
   }
   privatekey.PublicKey.Parameters = *params
   dsa.GenerateKey(privatekey, rand.Reader) // this generates a public & private key pair
}

func listen(port int, numOfcalls int, wg* sync.WaitGroup, values* chan Message) {

	fmt.Println("listening")
	numOfConns := 0
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
			val := Message{}
			dec.Decode(&val)
			(*values) <- val
			numOfConns++ 
			defer c.Close()

			if(numOfConns == numOfcalls){
				defer wg.Done()	
				numOfConns = 0
				return
			}
		}(conn)

	}
	return
}

func listenFrontEnd(port int, value* string) {
	fmt.Println("listening for frontend")
	portStr := ":" + strconv.Itoa(port)
	l, err := net.Listen("tcp", portStr)
	if err != nil {
		fmt.Println("error:", err)
	}
	defer l.Close()
	for {
		fmt.Print("connection")
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
			val := FeMessage{}
			dec.Decode(&val)
			fmt.Println(val);
			*value = val.Value 
			defer c.Close()
			return 
		}(conn)

	}
	return
}


func dial(port int, value Message){

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


func dialFe(port int, value []byte){

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

func inExtracted(extracted []string, x string) bool {
	for _, b := range extracted {
		if(x == b){
			return true
		}
	}
	return false
}

func findMin(values []string) string{
	min := values[0]
	for _, value := range values {
		if value < min{
			min = value
		}
	}
	return min
}

func getWinner(blocks []Block, hash string) Block {
	for _, b := range blocks{
		if(hash == b.Roothash){
			return b
		}
	}
	return Block{}

}

func main() {
	// ClI args
	var port, index, n int

	flag.IntVar(&port, "port", 1234, "port for process")
	flag.IntVar(&index, "index", 0, "index of the process")
	flag.IntVar(&n, "n", 5, "total number of process")
	flag.Parse()

	privatekey := new(dsa.PrivateKey)
	generateKey(privatekey)

	strIndex := strconv.Itoa(index)

	data := make([]string, 0)
	values := make(chan Message, n)
	extracted := make([]string, 0)
	blocks := make([]Block, 0)
	blockchain := make([]Block, 0)

	h := sha512.New()
	h.Write([]byte(strIndex))
	hashed := hex.EncodeToString(h.Sum(nil))

	data = append(data, strIndex)

	block := Block{0, "", hashed, data, strIndex}

	b, err := json.Marshal(block)
	if err != nil {
		fmt.Println("error:", err)
	}

	jsonPub, err := json.Marshal(privatekey.PublicKey)
	if err != nil {
		fmt.Println("error:", err)
	}

	h.Write(b)
	hashedBlock := hex.EncodeToString(h.Sum(nil))
	r,s,_ := dsa.Sign(rand.Reader, privatekey, []byte(hashedBlock))
	

	msg := Message{string(b), hashedBlock, s, r, string(jsonPub)}
	
	values <- msg

	var feMsg string;

	// setup wait group
	var wg sync.WaitGroup
	go listen(port, n-1, &wg, &values)
	
	go listenFrontEnd(3000 + (index *20), &feMsg)

	wg.Add(1)

	time.Sleep(4 * time.Second)
	//listen on port

	for {
		//dial all other
		for j := 0; j <= n-1; j++ {
			if(j == index){
				continue
			}
			dial(2000  + (j * 20), msg)
		}

		wg.Wait()
		wg.Add(1)

		for i := 0; i < n; i++ {
			x :=  <- values
			pubKey := getPublicKeyFromMessage(x)

			if(dsa.Verify(&pubKey, []byte(x.HashedBlock), x.R, x.S)){ //verify
				block := getBlockFromMessage(x)
				blocks = append(blocks, block)

				fmt.Println("Before winner: ", block)

				if(!inExtracted(extracted, block.Roothash)){
					extracted = append(extracted, block.Roothash)
				}
			}
		}

		block = getWinner(blocks, findMin(extracted))

		//new block 

		blockchain = append(blockchain, block)
		fmt.Println("block: ", block)
		//dial 

		bFe, err := json.Marshal(block)
		if err != nil {
			fmt.Println("error:", err)
		}
		if(index == 0){
			dialFe(8100, bFe)
		}

		for(feMsg == ""){
			time.Sleep(1 * time.Second)
		}
 
		block.Height += 1
		block.Parenthash = block.Roothash
		block.Value = feMsg

		temp := feMsg
		for _ , bv := range extracted{
			temp += bv
		}


		h.Write([]byte(temp))
		rootHash := hex.EncodeToString(h.Sum(nil))

		block.Roothash = rootHash
		block.Data = extracted
		extracted = nil

		b, err := json.Marshal(block)
		if err != nil {
			fmt.Println("error:", err)
		}

		h.Write(b)
		hashedBlock := hex.EncodeToString(h.Sum(nil))
		r,s,_ := dsa.Sign(rand.Reader, privatekey, []byte(hashedBlock))

		msg = Message{string(b), hashedBlock, s, r, string(jsonPub)}

		values <- msg

		feMsg =""
		

	}

	//ParseDSAPrivateKey returns a DSA private key from its ASN.1 DER encoding, as
	// specified by the OpenSSL DSA man page.
}