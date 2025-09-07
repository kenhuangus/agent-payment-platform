package hashchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Block represents a block in the hash chain
type Block struct {
	Index        int       `json:"index"`
	Timestamp    time.Time `json:"timestamp"`
	PreviousHash string    `json:"previousHash"`
	Data         string    `json:"data"`
	Hash         string    `json:"hash"`
}

// HashChain manages the blockchain-like hash chain for transactions
type HashChain struct {
	blocks []*Block
}

// NewHashChain creates a new hash chain with a genesis block
func NewHashChain() *HashChain {
	genesisBlock := &Block{
		Index:        0,
		Timestamp:    time.Now(),
		PreviousHash: "0",
		Data:         "Genesis Block - Agent Payments Platform",
		Hash:         "",
	}
	genesisBlock.Hash = calculateHash(genesisBlock)

	return &HashChain{
		blocks: []*Block{genesisBlock},
	}
}

// AddBlock adds a new block to the hash chain
func (hc *HashChain) AddBlock(data string) *Block {
	previousBlock := hc.blocks[len(hc.blocks)-1]

	newBlock := &Block{
		Index:        previousBlock.Index + 1,
		Timestamp:    time.Now(),
		PreviousHash: previousBlock.Hash,
		Data:         data,
		Hash:         "",
	}

	newBlock.Hash = calculateHash(newBlock)
	hc.blocks = append(hc.blocks, newBlock)

	return newBlock
}

// VerifyChain verifies the integrity of the entire hash chain
func (hc *HashChain) VerifyChain() bool {
	for i := 1; i < len(hc.blocks); i++ {
		currentBlock := hc.blocks[i]
		previousBlock := hc.blocks[i-1]

		// Verify hash integrity
		if currentBlock.Hash != calculateHash(currentBlock) {
			return false
		}

		// Verify chain linkage
		if currentBlock.PreviousHash != previousBlock.Hash {
			return false
		}
	}
	return true
}

// GetLatestBlock returns the most recent block
func (hc *HashChain) GetLatestBlock() *Block {
	return hc.blocks[len(hc.blocks)-1]
}

// GetChainLength returns the number of blocks in the chain
func (hc *HashChain) GetChainLength() int {
	return len(hc.blocks)
}

// GetBlockByIndex returns a block by its index
func (hc *HashChain) GetBlockByIndex(index int) (*Block, bool) {
	if index < 0 || index >= len(hc.blocks) {
		return nil, false
	}
	return hc.blocks[index], true
}

// GetBlocks returns all blocks in the chain
func (hc *HashChain) GetBlocks() []*Block {
	return hc.blocks
}

// calculateHash calculates the SHA-256 hash of a block
func calculateHash(block *Block) string {
	record := fmt.Sprintf("%d%s%s%d", block.Index, block.Timestamp.String(), block.PreviousHash, block.Data)
	h := sha256.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}

// TransactionHashData represents the data to be hashed for a transaction
type TransactionHashData struct {
	TransactionID string
	AgentID       string
	Description   string
	Amount        float64
	Currency      string
	Timestamp     time.Time
	Postings      []PostingHashData
}

// PostingHashData represents posting data for hashing
type PostingHashData struct {
	AccountID string
	Amount    float64
	Currency  string
}

// GenerateTransactionHash generates a hash for transaction data
func GenerateTransactionHash(data TransactionHashData) string {
	// Sort postings by account ID for consistent hashing
	sort.Slice(data.Postings, func(i, j int) bool {
		return data.Postings[i].AccountID < data.Postings[j].AccountID
	})

	var postingStrings []string
	for _, posting := range data.Postings {
		postingStrings = append(postingStrings,
			fmt.Sprintf("%s:%.2f:%s", posting.AccountID, posting.Amount, posting.Currency))
	}

	record := fmt.Sprintf("%s|%s|%s|%.2f|%s|%s|%s",
		data.TransactionID,
		data.AgentID,
		data.Description,
		data.Amount,
		data.Currency,
		data.Timestamp.Format(time.RFC3339),
		strings.Join(postingStrings, "|"))

	h := sha256.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}

// PaymentHashData represents the data to be hashed for a payment
type PaymentHashData struct {
	PaymentID    string
	AgentID      string
	AmountUSD    float64
	Counterparty string
	Rail         string
	Description  string
	Timestamp    time.Time
	Status       string
}

// GeneratePaymentHash generates a hash for payment data
func GeneratePaymentHash(data PaymentHashData) string {
	record := fmt.Sprintf("%s|%s|%.2f|%s|%s|%s|%s|%s",
		data.PaymentID,
		data.AgentID,
		data.AmountUSD,
		data.Counterparty,
		data.Rail,
		data.Description,
		data.Timestamp.Format(time.RFC3339),
		data.Status)

	h := sha256.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyHash verifies if the provided data matches the expected hash
func VerifyHash(data, expectedHash string) bool {
	h := sha256.New()
	h.Write([]byte(data))
	computedHash := hex.EncodeToString(h.Sum(nil))
	return computedHash == expectedHash
}

// MerkleTree represents a Merkle tree for efficient verification
type MerkleTree struct {
	Root   string     `json:"root"`
	Leaves []string   `json:"leaves"`
	Levels [][]string `json:"levels"`
}

// BuildMerkleTree builds a Merkle tree from a list of hashes
func BuildMerkleTree(hashes []string) *MerkleTree {
	if len(hashes) == 0 {
		return &MerkleTree{Root: "", Leaves: []string{}, Levels: [][]string{}}
	}

	tree := &MerkleTree{
		Leaves: hashes,
		Levels: [][]string{hashes},
	}

	// Build tree levels
	currentLevel := hashes
	for len(currentLevel) > 1 {
		var nextLevel []string

		for i := 0; i < len(currentLevel); i += 2 {
			if i+1 < len(currentLevel) {
				combined := currentLevel[i] + currentLevel[i+1]
				h := sha256.New()
				h.Write([]byte(combined))
				nextLevel = append(nextLevel, hex.EncodeToString(h.Sum(nil)))
			} else {
				// Odd number of elements, duplicate the last one
				combined := currentLevel[i] + currentLevel[i]
				h := sha256.New()
				h.Write([]byte(combined))
				nextLevel = append(nextLevel, hex.EncodeToString(h.Sum(nil)))
			}
		}

		tree.Levels = append(tree.Levels, nextLevel)
		currentLevel = nextLevel
	}

	if len(currentLevel) > 0 {
		tree.Root = currentLevel[0]
	}

	return tree
}

// VerifyMerkleProof verifies if a leaf is part of the Merkle tree
func (mt *MerkleTree) VerifyMerkleProof(leafHash string, proof []string) bool {
	currentHash := leafHash

	for _, proofHash := range proof {
		// Combine current hash with proof hash (order matters for verification)
		combined := currentHash + proofHash
		h := sha256.New()
		h.Write([]byte(combined))
		currentHash = hex.EncodeToString(h.Sum(nil))
	}

	return currentHash == mt.Root
}
