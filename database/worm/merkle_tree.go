package worm

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// MerkleTree represents a deterministic cryptographic binary Merkle tree.
type MerkleTree struct {
	Leaves [][]byte
	Layers [][][]byte
	Root   []byte
}

// NewMerkleTree constructs a Merkle tree from a slice of leaf data hashes.
func NewMerkleTree(dataHashes [][]byte) (*MerkleTree, error) {
	if len(dataHashes) == 0 {
		return nil, errors.New("cannot create Merkle tree from empty slice")
	}

	leaves := make([][]byte, len(dataHashes))
	for i, h := range dataHashes {
		leafHash := sha256.Sum256(h)
		leaves[i] = leafHash[:]
	}

	layers := [][][]byte{leaves}
	currentLayer := leaves

	for len(currentLayer) > 1 {
		var nextLayer [][]byte
		for i := 0; i < len(currentLayer); i += 2 {
			if i+1 < len(currentLayer) {
				combined := append(currentLayer[i], currentLayer[i+1]...)
				h := sha256.Sum256(combined)
				nextLayer = append(nextLayer, h[:])
			} else {
				// Odd element: duplicate to maintain binary structure
				combined := append(currentLayer[i], currentLayer[i]...)
				h := sha256.Sum256(combined)
				nextLayer = append(nextLayer, h[:])
			}
		}
		layers = append(layers, nextLayer)
		currentLayer = nextLayer
	}

	return &MerkleTree{
		Leaves: leaves,
		Layers: layers,
		Root:   currentLayer[0],
	}, nil
}

// RootHex returns the hexadecimal string of the Merkle root.
func (mt *MerkleTree) RootHex() string {
	if len(mt.Root) == 0 {
		return ""
	}
	return hex.EncodeToString(mt.Root)
}

// MerkleProof represents an inclusion proof for a specific leaf.
type MerkleProof struct {
	LeafHash string   `json:"leaf_hash"`
	Path     []string `json:"path"`
	Indices  []int    `json:"indices"` // 0 for left, 1 for right
}

// GenerateProof produces an inclusion proof for leaf at index.
func (mt *MerkleTree) GenerateProof(index int) (*MerkleProof, error) {
	if index < 0 || index >= len(mt.Leaves) {
		return nil, errors.New("leaf index out of bounds")
	}

	proof := &MerkleProof{
		LeafHash: hex.EncodeToString(mt.Leaves[index]),
		Path:     make([]string, 0),
		Indices:  make([]int, 0),
	}

	currentIndex := index
	for layerIdx := 0; layerIdx < len(mt.Layers)-1; layerIdx++ {
		layer := mt.Layers[layerIdx]
		var siblingIndex int
		if currentIndex%2 == 0 {
			// Current is left, sibling is right
			if currentIndex+1 < len(layer) {
				siblingIndex = currentIndex + 1
			} else {
				siblingIndex = currentIndex // duplicated odd node
			}
			proof.Indices = append(proof.Indices, 1)
		} else {
			// Current is right, sibling is left
			siblingIndex = currentIndex - 1
			proof.Indices = append(proof.Indices, 0)
		}

		proof.Path = append(proof.Path, hex.EncodeToString(layer[siblingIndex]))
		currentIndex = currentIndex / 2
	}

	return proof, nil
}

// VerifyProof cryptographically verifies a Merkle inclusion proof against a root.
func VerifyProof(proof *MerkleProof, rootHex string) bool {
	currentBytes, err := hex.DecodeString(proof.LeafHash)
	if err != nil {
		return false
	}

	for i, siblingHex := range proof.Path {
		siblingBytes, err := hex.DecodeString(siblingHex)
		if err != nil {
			return false
		}

		var combined []byte
		if proof.Indices[i] == 1 {
			// Sibling is right
			combined = append(currentBytes, siblingBytes...)
		} else {
			// Sibling is left
			combined = append(siblingBytes, currentBytes...)
		}

		h := sha256.Sum256(combined)
		currentBytes = h[:]
	}

	return hex.EncodeToString(currentBytes) == rootHex
}
