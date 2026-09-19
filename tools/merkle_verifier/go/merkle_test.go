package merkle

import (
	"encoding/hex"
	"testing"
)

func TestKeccak256_StandardVectors(t *testing.T) {
	emptyHash := Keccak256([]byte(""))
	expected := "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
	if hex.EncodeToString(emptyHash[:]) != expected {
		t.Fatalf("expected %s, got %s", expected, hex.EncodeToString(emptyHash[:]))
	}

	helloHash := Keccak256([]byte("hello"))
	expectedHello := "1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8"
	if hex.EncodeToString(helloHash[:]) != expectedHello {
		t.Fatalf("expected %s, got %s", expectedHello, hex.EncodeToString(helloHash[:]))
	}
}

func TestComputeLeaf_CrossPlatformParity(t *testing.T) {
	leaf := ComputeLeaf("investor_42", 100000000)
	expected := "0x7b0a80c907f59398dd28e80e4715e6046bd8a42af289a3d2608a418a9fb82634"
	if Hex(leaf) != expected {
		t.Fatalf("Cross platform parity mismatch: expected %s, got %s", expected, Hex(leaf))
	}
}

func TestHashPair_SortedInvariance(t *testing.T) {
	nodeA := Keccak256([]byte("alpha"))
	nodeB := Keccak256([]byte("beta"))

	pairAB := HashPair(nodeA, nodeB)
	pairBA := HashPair(nodeB, nodeA)

	if pairAB != pairBA {
		t.Fatalf("HashPair should be sorted canonical: %s != %s", Hex(pairAB), Hex(pairBA))
	}
}

func TestSolvencyTree_SingleAccount(t *testing.T) {
	accounts := []AccountLiability{
		{UserID: "usr_sole", BalanceE8: 500000000},
	}

	tree, err := BuildSolvencyTree(accounts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	proof, err := tree.GetProof(0)
	if err != nil {
		t.Fatalf("unexpected error getting proof: %v", err)
	}

	if len(proof) != 0 {
		t.Fatalf("single node tree should have empty proof, got %d elements", len(proof))
	}

	leaf := ComputeLeaf("usr_sole", 500000000)
	if !VerifyProof(leaf, proof, tree.Root) {
		t.Fatalf("single leaf must verify against root")
	}
}

func TestSolvencyTree_MultiAccounts(t *testing.T) {
	testSizes := []int{2, 3, 4, 7, 8, 16}

	for _, size := range testSizes {
		accounts := make([]AccountLiability, size)
		for i := 0; i < size; i++ {
			accounts[i] = AccountLiability{
				UserID:    string(rune('A' + i)),
				BalanceE8: uint64((i + 1) * 1000),
			}
		}

		tree, err := BuildSolvencyTree(accounts)
		if err != nil {
			t.Fatalf("size %d: unexpected error: %v", size, err)
		}

		for i := 0; i < size; i++ {
			proof, err := tree.GetProof(i)
			if err != nil {
				t.Fatalf("size %d, index %d: get proof error: %v", size, i, err)
			}

			leaf := accounts[i].LeafHash
			if !VerifyProof(leaf, proof, tree.Root) {
				t.Fatalf("size %d, index %d: verification failed", size, i)
			}

			// Tampered leaf must fail
			tamperedLeaf := ComputeLeaf(accounts[i].UserID, accounts[i].BalanceE8+1)
			if VerifyProof(tamperedLeaf, proof, tree.Root) {
				t.Fatalf("size %d, index %d: tampered leaf should fail verification", size, i)
			}
		}
	}
}

func TestValidateSolvency(t *testing.T) {
	// 1. Solvent: 105% backed
	ratio, err := ValidateSolvency(100_00000000, 105_00000000)
	if err != nil {
		t.Fatalf("expected solvent, got error: %v", err)
	}
	if ratio < 1.0 {
		t.Fatalf("expected ratio >= 1.0, got %f", ratio)
	}

	// 2. Insolvent: 95% backed
	_, err = ValidateSolvency(100_00000000, 95_00000000)
	if err == nil {
		t.Fatalf("expected insolvency error, got nil")
	}
}
