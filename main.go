package main

import (
	// "bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"

	// "net"

	"net/http"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Block struct {
	Index     int
	Timestamp string
	FileHash  string
	PubKey    string
	Hash      string
	PrevHash  string
}

var Blockchain []Block
var bcServer chan []Block

func calculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + block.FileHash + block.PubKey + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func generateBlock(oldBlock Block, FileHash string, PubKey string) (Block, error) {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.FileHash = FileHash
	newBlock.PubKey = PubKey
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
}

func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

func replaceChain(newBlocks []Block) {
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
}

func run() error {
	mux := makeMuxRouter()
	httpAddr := os.Getenv("PORT")
	log.Println("Listening on ", os.Getenv("PORT"))
	s := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	return muxRouter
}
func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(Blockchain, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

type Message struct {
	FileHash string
	PubKey   string
}

func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	var m Message

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()

	newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], m.FileHash, m.PubKey)
	if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, m)
		return
	}
	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
		newBlockchain := append(Blockchain, newBlock)
		replaceChain(newBlockchain)
		spew.Dump(Blockchain)
	}

	respondWithJSON(w, r, http.StatusCreated, newBlock)

}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// bcServer = make(chan []Block)

	// create genesis block
	go func() {
		t := time.Now()
		genesisBlock := Block{0, t.String(), "", "", "", ""}
		spew.Dump(genesisBlock)
		Blockchain = append(Blockchain, genesisBlock)
	}()
	log.Fatal(run())

	// start TCP and serve TCP server
	// 	server, err := net.Listen("tcp", ":"+os.Getenv("ADDR"))
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	defer server.Close()

	// 	for {
	// 		conn, err := server.Accept()
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		go handleConn(conn)
	// 	}
	// }

	// func handleConn(conn net.Conn) {

	// 	defer conn.Close()
	// 	scanner := bufio.NewReader(conn)

	// 	io.WriteString(conn, "Enter a new FileHash:")

	// 	// take in FileHash from stdin and add it to blockchain after conducting necessary validation
	// 	go func() {
	// 		// for scanner.() {
	// 		FileHash, _ := scanner.ReadString('\n')
	// 		// if err != nil {
	// 		// 	log.Printf("%v not a number: %v", scanner.Text(), err)
	// 		// 	continue
	// 		// }
	// 		io.WriteString(conn, "\a Enter your Public Key:")
	// 		PubKey, _ := scanner.ReadString('\n')

	// 		// newBlockkey, err := generateBlock(Blockchain[len(Blockchain)-1], PubKey)
	// 		// if err != nil {
	// 		// 	log.Println(err)
	// 		// }
	// 		// if isBlockValid(newBlockkey, Blockchain[len(Blockchain)-1]) {
	// 		// 	newBlockchain := append(Blockchain, newBlockkey)
	// 		// 	replaceChain(newBlockchain)
	// 		// }
	// 		newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], FileHash[0:len(FileHash)-1], PubKey[0:len(PubKey)-1])
	// 		if err != nil {
	// 			log.Println(err)
	// 			// continue
	// 		}
	// 		if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
	// 			newBlockchain := append(Blockchain, newBlock)
	// 			replaceChain(newBlockchain)
	// 		}

	// 		bcServer <- Blockchain
	// 		// io.WriteString(conn, "\a new FileHash")
	// 		// }
	// 	}()

	// 	// simulate receiving broadcast
	// 	go func() {
	// 		for {
	// 			time.Sleep(30 * time.Second)
	// 			output, err := json.Marshal(Blockchain)
	// 			if err != nil {
	// 				log.Fatal(err)
	// 			}
	// 			io.WriteString(conn, string(output))
	// 		}
	// 	}()

	// 	for _ = range bcServer {
	// 		spew.Dump(Blockchain)
	// 	}

}
