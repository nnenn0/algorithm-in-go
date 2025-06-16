package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"hash"
	"math"
)

// BloomFilter はBloom Filterのデータ構造
type BloomFilter struct {
	bitArray  []bool      // ビット配列
	size      int         // ビット配列のサイズ
	hashFuncs []hash.Hash // ハッシュ関数のリスト
	numHashes int         // ハッシュ関数の数
	numItems  int         // 追加されたアイテム数
}

// NewBloomFilter は新しいBloom Filterを作成
// expectedItems: 予想されるアイテム数
// falsePositiveRate: 偽陽性率 (0.0 < rate < 1.0)
func NewBloomFilter(expectedItems int, falsePositiveRate float64) *BloomFilter {
	// 最適なビット配列サイズを計算
	size := int(math.Ceil(float64(expectedItems) * math.Log(falsePositiveRate) / math.Log(1.0/math.Pow(2.0, math.Log(2.0)))))

	// 最適なハッシュ関数の数を計算
	numHashes := int(math.Ceil(float64(size) / float64(expectedItems) * math.Log(2.0)))

	// 最小値を保証
	if size < 1 {
		size = 1
	}
	if numHashes < 1 {
		numHashes = 1
	}

	return &BloomFilter{
		bitArray:  make([]bool, size),
		size:      size,
		hashFuncs: createHashFunctions(numHashes),
		numHashes: numHashes,
		numItems:  0,
	}
}

// createHashFunctions は指定された数のハッシュ関数を作成
func createHashFunctions(numHashes int) []hash.Hash {
	funcs := make([]hash.Hash, numHashes)

	// 異なるハッシュ関数を使用（実際にはより多くの種類が必要な場合がある）
	for i := 0; i < numHashes; i++ {
		switch i % 3 {
		case 0:
			funcs[i] = md5.New()
		case 1:
			funcs[i] = sha1.New()
		case 2:
			funcs[i] = sha256.New()
		}
	}

	return funcs
}

// getHashes はデータに対してすべてのハッシュ値を計算
func (bf *BloomFilter) getHashes(data []byte) []int {
	hashes := make([]int, bf.numHashes)

	for i, hashFunc := range bf.hashFuncs {
		hashFunc.Reset()
		hashFunc.Write(data)

		// ハッシュ値の最初の4バイトを使用してインデックスを計算
		hashBytes := hashFunc.Sum(nil)
		hashValue := 0
		for j := 0; j < 4 && j < len(hashBytes); j++ {
			hashValue = (hashValue << 8) | int(hashBytes[j])
		}

		// 負の値を正に変換し、配列サイズで割った余りを取る
		if hashValue < 0 {
			hashValue = -hashValue
		}
		hashes[i] = hashValue % bf.size
	}

	return hashes
}

// Add はBloom Filterにアイテムを追加
func (bf *BloomFilter) Add(item string) {
	hashes := bf.getHashes([]byte(item))

	for _, hash := range hashes {
		bf.bitArray[hash] = true
	}

	bf.numItems++
}

// Test はアイテムがBloom Filterに存在する可能性があるかテスト
// true: 存在する可能性がある（偽陽性の可能性あり）
// false: 確実に存在しない
func (bf *BloomFilter) Test(item string) bool {
	hashes := bf.getHashes([]byte(item))

	for _, hash := range hashes {
		if !bf.bitArray[hash] {
			return false // 確実に存在しない
		}
	}

	return true // 存在する可能性がある
}

// EstimateFalsePositiveRate は現在の偽陽性率を推定
func (bf *BloomFilter) EstimateFalsePositiveRate() float64 {
	if bf.numItems == 0 {
		return 0.0
	}

	// 偽陽性率の理論値: (1 - e^(-kn/m))^k
	// k: ハッシュ関数の数, n: アイテム数, m: ビット配列サイズ
	k := float64(bf.numHashes)
	n := float64(bf.numItems)
	m := float64(bf.size)

	return math.Pow(1.0-math.Exp(-k*n/m), k)
}

// Stats はBloom Filterの統計情報を返す
func (bf *BloomFilter) Stats() map[string]interface{} {
	setBits := 0
	for _, bit := range bf.bitArray {
		if bit {
			setBits++
		}
	}

	return map[string]interface{}{
		"size":           bf.size,
		"num_hashes":     bf.numHashes,
		"num_items":      bf.numItems,
		"set_bits":       setBits,
		"load_factor":    float64(setBits) / float64(bf.size),
		"false_positive": bf.EstimateFalsePositiveRate(),
	}
}

// PrintStats は統計情報を表示
func (bf *BloomFilter) PrintStats() {
	stats := bf.Stats()
	fmt.Println("=== Bloom Filter Statistics ===")
	fmt.Printf("Size: %d bits\n", stats["size"])
	fmt.Printf("Hash functions: %d\n", stats["num_hashes"])
	fmt.Printf("Items added: %d\n", stats["num_items"])
	fmt.Printf("Set bits: %d\n", stats["set_bits"])
	fmt.Printf("Load factor: %.3f\n", stats["load_factor"])
	fmt.Printf("Estimated false positive rate: %.6f (%.4f%%)\n",
		stats["false_positive"], stats["false_positive"].(float64)*100)
}

// 使用例とテスト
func main() {
	fmt.Println("=== Bloom Filter Demo ===")

	// 1000アイテム、1%の偽陽性率でBloom Filterを作成
	bf := NewBloomFilter(1000, 0.01)

	// テストデータを追加
	items := []string{
		"apple", "banana", "cherry", "date", "elderberry",
		"fig", "grape", "honeydew", "kiwi", "lemon",
		"mango", "nectarine", "orange", "papaya", "quince",
	}

	fmt.Printf("Adding %d items to Bloom Filter...\n", len(items))
	for _, item := range items {
		bf.Add(item)
	}

	// 統計情報を表示
	fmt.Println()
	bf.PrintStats()

	// 存在テスト
	fmt.Println("\n=== Existence Tests ===")

	// 確実に存在するアイテムのテスト
	fmt.Println("Testing items that were added:")
	for _, item := range items[:5] {
		exists := bf.Test(item)
		fmt.Printf("'%s': %v\n", item, exists)
	}

	// 存在しないアイテムのテスト
	fmt.Println("\nTesting items that were NOT added:")
	nonExistentItems := []string{"watermelon", "strawberry", "blueberry", "raspberry", "blackberry"}
	falsePositives := 0

	for _, item := range nonExistentItems {
		exists := bf.Test(item)
		fmt.Printf("'%s': %v", item, exists)
		if exists {
			fmt.Print(" (FALSE POSITIVE)")
			falsePositives++
		}
		fmt.Println()
	}

	fmt.Printf("\nFalse positives: %d/%d (%.1f%%)\n",
		falsePositives, len(nonExistentItems),
		float64(falsePositives)/float64(len(nonExistentItems))*100)

	// 大量データでのテスト
	fmt.Println("\n=== Large Scale Test ===")
	largeBF := NewBloomFilter(10000, 0.001)

	// 10000個のアイテムを追加
	for i := 0; i < 10000; i++ {
		largeBF.Add(fmt.Sprintf("item_%d", i))
	}

	// 存在しないアイテムをテスト
	falsePositiveCount := 0
	testCount := 1000

	for i := 10000; i < 10000+testCount; i++ {
		if largeBF.Test(fmt.Sprintf("item_%d", i)) {
			falsePositiveCount++
		}
	}

	actualFPRate := float64(falsePositiveCount) / float64(testCount)

	fmt.Printf("Large scale test results:\n")
	fmt.Printf("Added items: 10,000\n")
	fmt.Printf("Test items (non-existent): %d\n", testCount)
	fmt.Printf("False positives: %d\n", falsePositiveCount)
	fmt.Printf("Actual false positive rate: %.4f%% (target: 0.1%%)\n", actualFPRate*100)

	largeBF.PrintStats()
}
