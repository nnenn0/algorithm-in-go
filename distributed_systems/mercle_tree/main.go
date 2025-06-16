package main

import (
	"crypto/sha256"
	"fmt"
)

// Node はMerkle Treeのノードを表す
type Node struct {
	Hash  []byte
	Left  *Node
	Right *Node
	Data  []byte // リーフノードのみ使用
}

// MerkleTree はMerkle Tree構造を表す
type MerkleTree struct {
	Root *Node
}

// hash はデータのSHA256ハッシュを計算
func hash(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

// NewLeafNode は新しいリーフノードを作成
func NewLeafNode(data []byte) *Node {
	return &Node{
		Hash: hash(data),
		Data: data,
	}
}

// NewInternalNode は2つの子ノードから内部ノードを作成
func NewInternalNode(left, right *Node) *Node {
	// 左の子と右の子のハッシュを結合してハッシュ化
	combinedHash := append(left.Hash, right.Hash...)
	return &Node{
		Hash:  hash(combinedHash),
		Left:  left,
		Right: right,
	}
}

// NewMerkleTree はデータリストからMerkle Treeを構築
func NewMerkleTree(data [][]byte) *MerkleTree {
	if len(data) == 0 {
		return &MerkleTree{}
	}

	// リーフノードを作成
	var nodes []*Node
	for _, d := range data {
		nodes = append(nodes, NewLeafNode(d))
	}

	// ツリーを下から上へ構築
	for len(nodes) > 1 {
		var nextLevel []*Node

		// ペアごとに処理
		for i := 0; i < len(nodes); i += 2 {
			left := nodes[i]
			var right *Node

			if i+1 < len(nodes) {
				right = nodes[i+1]
			} else {
				// 奇数個の場合、最後のノードを複製
				right = nodes[i]
			}

			parent := NewInternalNode(left, right)
			nextLevel = append(nextLevel, parent)
		}

		nodes = nextLevel
	}

	return &MerkleTree{Root: nodes[0]}
}

// GetRootHash はルートハッシュを取得
func (mt *MerkleTree) GetRootHash() []byte {
	if mt.Root == nil {
		return nil
	}
	return mt.Root.Hash
}

// GetRootHashString はルートハッシュを16進文字列で取得
func (mt *MerkleTree) GetRootHashString() string {
	hash := mt.GetRootHash()
	if hash == nil {
		return ""
	}
	return fmt.Sprintf("%x", hash)
}

// GetProof は指定されたデータのMerkle Proofを取得
func (mt *MerkleTree) GetProof(data []byte) [][]byte {
	if mt.Root == nil {
		return nil
	}

	targetHash := hash(data)
	var proof [][]byte

	// ルートから目標のリーフまでのパスを辿る
	if mt.getProofHelper(mt.Root, targetHash, &proof) {
		return proof
	}

	return nil
}

// getProofHelper はGetProofのヘルパー関数
func (mt *MerkleTree) getProofHelper(node *Node, targetHash []byte, proof *[][]byte) bool {
	if node == nil {
		return false
	}

	// リーフノードの場合
	if node.Left == nil && node.Right == nil {
		return string(node.Hash) == string(targetHash)
	}

	// 左の子ツリーで検索
	if mt.getProofHelper(node.Left, targetHash, proof) {
		// 右の子のハッシュを証明に追加
		*proof = append(*proof, node.Right.Hash)
		return true
	}

	// 右の子ツリーで検索
	if mt.getProofHelper(node.Right, targetHash, proof) {
		// 左の子のハッシュを証明に追加
		*proof = append(*proof, node.Left.Hash)
		return true
	}

	return false
}

// VerifyProof はMerkle Proofを検証
func VerifyProof(data []byte, proof [][]byte, rootHash []byte) bool {
	currentHash := hash(data)

	// プルーフの各ハッシュと結合してルートまで計算
	for _, proofHash := range proof {
		// 結合順序を決定（通常は辞書順）
		if string(currentHash) <= string(proofHash) {
			combined := append(currentHash, proofHash...)
			currentHash = hash(combined)
		} else {
			combined := append(proofHash, currentHash...)
			currentHash = hash(combined)
		}
	}

	return string(currentHash) == string(rootHash)
}

// PrintTree はツリー構造を表示（デバッグ用）
func (mt *MerkleTree) PrintTree() {
	if mt.Root == nil {
		fmt.Println("Empty tree")
		return
	}
	mt.printNode(mt.Root, "", true)
}

func (mt *MerkleTree) printNode(node *Node, prefix string, isLast bool) {
	if node == nil {
		return
	}

	// ノードの情報を表示
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	hashStr := fmt.Sprintf("%x", node.Hash)[:8] // 最初の8文字のみ表示
	if node.Data != nil {
		fmt.Printf("%s%s[LEAF] %s (data: %s)\n", prefix, connector, hashStr, string(node.Data))
	} else {
		fmt.Printf("%s%s[NODE] %s\n", prefix, connector, hashStr)
	}

	// 子ノードを表示
	if node.Left != nil || node.Right != nil {
		newPrefix := prefix
		if isLast {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}

		if node.Right != nil {
			mt.printNode(node.Right, newPrefix, node.Left == nil)
		}
		if node.Left != nil {
			mt.printNode(node.Left, newPrefix, true)
		}
	}
}

// 使用例
func main() {
	// テストデータ
	data := [][]byte{
		[]byte("apple"),
		[]byte("banana"),
		[]byte("cherry"),
		[]byte("date"),
		[]byte("elderberry"),
	}

	fmt.Println("=== Merkle Tree Demo ===")
	fmt.Println("データ:", []string{"apple", "banana", "cherry", "date", "elderberry"})

	// Merkle Treeを構築
	tree := NewMerkleTree(data)

	fmt.Println("\n=== Tree Structure ===")
	tree.PrintTree()

	// ルートハッシュを表示
	fmt.Printf("\n=== Root Hash ===\n%s\n", tree.GetRootHashString())

	// Merkle Proofのテスト
	fmt.Println("\n=== Merkle Proof Test ===")
	testData := []byte("banana")

	proof := tree.GetProof(testData)
	if proof != nil {
		fmt.Printf("'%s'のMerkle Proof:\n", string(testData))
		for i, p := range proof {
			fmt.Printf("  %d: %x\n", i, p)
		}

		// 証明を検証
		isValid := VerifyProof(testData, proof, tree.GetRootHash())
		fmt.Printf("\n検証結果: %v\n", isValid)
	} else {
		fmt.Printf("'%s'のプルーフが見つかりません\n", string(testData))
	}

	// 存在しないデータのテスト
	fmt.Println("\n=== Invalid Data Test ===")
	invalidData := []byte("grape")
	invalidProof := tree.GetProof(invalidData)
	if invalidProof == nil {
		fmt.Printf("'%s'は存在しません（正常）\n", string(invalidData))
	}

	// データの変更を検出するテスト
	fmt.Println("\n=== Tamper Detection Test ===")
	tamperedData := []byte("BANANA") // 大文字に改変
	isValid := VerifyProof(tamperedData, proof, tree.GetRootHash())
	fmt.Printf("改変されたデータ'%s'の検証: %v（改変が検出された）\n", string(tamperedData), isValid)
}
