package main

import (
	"crypto/sha1"
	"fmt"
	"sort"
	"strconv"
)

// ConsistentHash はコンシステントハッシュリングを表す構造体
type ConsistentHash struct {
	replicas int            // 各ノードの仮想ノード数
	keys     []int          // ソートされたハッシュ値のリスト
	hashMap  map[int]string // ハッシュ値からノード名へのマップ
}

// New は新しいConsistentHashインスタンスを作成
func New(replicas int) *ConsistentHash {
	return &ConsistentHash{
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
}

// hash は文字列をハッシュ値に変換
func (ch *ConsistentHash) hash(key string) int {
	h := sha1.New()
	h.Write([]byte(key))
	hashBytes := h.Sum(nil)

	// 最初の4バイトを使ってintに変換
	hash := int(hashBytes[0])<<24 + int(hashBytes[1])<<16 + int(hashBytes[2])<<8 + int(hashBytes[3])
	if hash < 0 {
		hash = -hash
	}
	return hash
}

// Add はハッシュリングにノードを追加
func (ch *ConsistentHash) Add(nodes ...string) {
	for _, node := range nodes {
		// 各ノードに対して複数の仮想ノードを作成
		for i := 0; i < ch.replicas; i++ {
			// 仮想ノード名を作成（ノード名 + レプリカ番号）
			virtualNode := node + "#" + strconv.Itoa(i)
			hash := ch.hash(virtualNode)
			ch.keys = append(ch.keys, hash)
			ch.hashMap[hash] = node
		}
	}
	// ハッシュ値でソート
	sort.Ints(ch.keys)
}

// Remove はハッシュリングからノードを削除
func (ch *ConsistentHash) Remove(node string) {
	for i := 0; i < ch.replicas; i++ {
		virtualNode := node + "#" + strconv.Itoa(i)
		hash := ch.hash(virtualNode)

		// ハッシュマップから削除
		delete(ch.hashMap, hash)

		// keysスライスから削除
		idx := ch.search(hash)
		if idx < len(ch.keys) && ch.keys[idx] == hash {
			ch.keys = append(ch.keys[:idx], ch.keys[idx+1:]...)
		}
	}
}

// search はソートされたkeysスライス内でハッシュ値の挿入位置を検索
func (ch *ConsistentHash) search(hash int) int {
	return sort.Search(len(ch.keys), func(i int) bool {
		return ch.keys[i] >= hash
	})
}

// Get は指定されたキーに対応するノードを取得
func (ch *ConsistentHash) Get(key string) string {
	if len(ch.keys) == 0 {
		return ""
	}

	hash := ch.hash(key)

	// ハッシュ値以上の最初のノードを検索
	idx := ch.search(hash)

	// リングの最後を超えた場合は最初のノードを返す
	if idx == len(ch.keys) {
		idx = 0
	}

	return ch.hashMap[ch.keys[idx]]
}

// GetNodes は現在登録されている全ノードのリストを取得
func (ch *ConsistentHash) GetNodes() []string {
	nodeSet := make(map[string]bool)
	for _, node := range ch.hashMap {
		nodeSet[node] = true
	}

	nodes := make([]string, 0, len(nodeSet))
	for node := range nodeSet {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)
	return nodes
}

// 使用例
func main() {
	// 仮想ノード数3でコンシステントハッシュを作成
	ch := New(3)

	// ノードを追加
	ch.Add("server1", "server2", "server3")

	fmt.Println("初期ノード:", ch.GetNodes())

	// キーの分散をテスト
	keys := []string{"user1", "user2", "user3", "user4", "user5", "data1", "data2", "data3"}

	fmt.Println("\n各キーの分散:")
	for _, key := range keys {
		node := ch.Get(key)
		fmt.Printf("Key: %s -> Node: %s\n", key, node)
	}

	// ノードを追加
	fmt.Println("\nserver4を追加:")
	ch.Add("server4")
	fmt.Println("ノード:", ch.GetNodes())

	fmt.Println("\nserver4追加後の分散:")
	for _, key := range keys {
		node := ch.Get(key)
		fmt.Printf("Key: %s -> Node: %s\n", key, node)
	}

	// ノードを削除
	fmt.Println("\nserver2を削除:")
	ch.Remove("server2")
	fmt.Println("ノード:", ch.GetNodes())

	fmt.Println("\nserver2削除後の分散:")
	for _, key := range keys {
		node := ch.Get(key)
		fmt.Printf("Key: %s -> Node: %s\n", key, node)
	}
}
